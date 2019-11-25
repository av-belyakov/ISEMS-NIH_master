package modulenetworkinteractionapp

/*
* 		Клиент для взаимодействия с источниками
* осуществляет соединение с источниками если те выступают в роли сервера
* */

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"

	"github.com/gorilla/websocket"
)

type clientSetting struct {
	ID             int
	IP, Port       string
	InfoSourceList *configure.InformationSourcesList
	COut           chan<- [2]string
	CwtReq         chan<- configure.MsgWsTransmission
}

func (cs clientSetting) redirectPolicyFunc(req *http.Request, rl []*http.Request) error {

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	funcName := "redirectPolicyFunc"

	go func() {
		header := http.Header{}
		header.Add("Content-Type", "text/plain;charset=utf-8")
		header.Add("Accept-Language", "en")
		header.Add("User-Agent", "Mozilla/5.0 (ISEMS-NIH_slave)")

		d := websocket.Dialer{
			HandshakeTimeout:  (time.Duration(1) * time.Second),
			EnableCompression: false,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		c, res, err := d.Dial("wss://"+cs.IP+":"+cs.Port+"/wss", header)
		if err != nil {
			_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    funcName,
			})

			return
		}
		defer connClose(cs.COut, c, cs.InfoSourceList, cs.ID, cs.IP)

		if res.StatusCode == 101 {
			//изменяем статус подключения клиента
			_ = cs.InfoSourceList.ChangeSourceConnectionStatus(cs.ID, true)

			//добавляем линк соединения
			cs.InfoSourceList.AddLinkWebsocketConnect(cs.IP, c)

			//отправляем через канал сообщение о том что соединение установлено
			cs.COut <- [2]string{cs.IP, "connect"}

			//обработчик запросов приходящих через websocket
			for {
				msgType, message, err := c.ReadMessage()
				if err != nil {
					_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					break
				}

				cs.CwtReq <- configure.MsgWsTransmission{
					DestinationHost: cs.IP,
					MsgType:         msgType,
					Data:            &message,
				}
			}
		}
	}()

	//отправляем ошибку что бы предотвратить еще один редирект который вызывается после обработки этой функции
	return errors.New("stop redirect")
}

//WssClientNetworkInteraction соединение в режиме 'Клиент'
func WssClientNetworkInteraction(
	cOut chan<- [2]string,
	appc *configure.AppConfig,
	isl *configure.InformationSourcesList,
	cwt chan<- configure.MsgWsTransmission) {

	/* инициализируем HTTPS клиента */
	fmt.Println("\tThe HTTPS client Network Integration is running")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	funcName := "WssClientNetworkInteraction"

	//читаем сертификат CA для того что бы наш клиент доверял сертификату переданному сервером
	rootCA, err := ioutil.ReadFile(appc.PathRootCA)
	if err != nil {
		_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})
	}

	//создаем новый пул доверенных центров серификации и добавляем в него корневой сертификат
	cp := x509.NewCertPool()
	if ok := cp.AppendCertsFromPEM(rootCA); !ok {
		_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "root certificate was not added to the pool",
			FuncName:    funcName,
		})
	}

	conf := &tls.Config{
		ServerName: "isems_nih_slave",
		RootCAs:    cp,
	}
	conf.BuildNameToCertificate()

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
			TLSClientConfig:    conf,
		},
	}

	//цикличные попытки установления соединения в режиме клиент
	ticker := time.NewTicker(time.Duration(appc.TimeReconnectClient) * time.Second)
	for range ticker.C {
		sl := isl.GetSourceList()

		if len(*sl) == 0 {
			continue
		}

		for id, s := range *sl {
			if s.AsServer && !s.ConnectionStatus {
				port := strconv.Itoa(s.Settings.IfAsServerThenPort)

				cs := clientSetting{
					ID:             id,
					IP:             s.IP,
					Port:           port,
					InfoSourceList: isl,
					COut:           cOut,
					CwtReq:         cwt,
				}
				client.CheckRedirect = cs.redirectPolicyFunc

				req, err := http.NewRequest("GET", "https://"+cs.IP+":"+cs.Port+"/", nil)
				if err != nil {
					_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					continue
				}

				req.Header.Add("Content-Type", "text/plain;charset=utf-8")
				req.Header.Add("Accept-Language", "en")
				req.Header.Add("User-Agent", "Mozilla/5.0 (ISEMS-NIH_slave)")
				req.Header.Add("Token", s.Token)

				_, err = client.Do(req)
				if err != nil {
					strErr := fmt.Sprint(err)
					if !strings.Contains(strErr, "stop redirect") {
						_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: strErr,
							FuncName:    funcName,
						})
					}

					continue
				}
			}
		}
	}
}
