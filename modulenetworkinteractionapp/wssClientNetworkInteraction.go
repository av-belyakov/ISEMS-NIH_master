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
	saveMessageApp *savemessageapp.PathDirLocationLogFiles
	TLSConf        *tls.Config
	COut           chan<- [2]string
	CwtReq         chan<- configure.MsgWsTransmission
}

type handlerRequestParameters struct {
	id, port  int
	ip, token string
}

func (cs clientSetting) redirectPolicyFunc(req *http.Request, rl []*http.Request) error {
	funcName := "redirectPolicyFunc"

	go func() {
		header := http.Header{}
		header.Add("Content-Type", "text/plain;charset=utf-8")
		header.Add("Accept-Language", "en")
		header.Add("User-Agent", "Mozilla/5.0 (ISEMS-NIH_slave)")

		d := websocket.Dialer{
			HandshakeTimeout:  (time.Duration(3) * time.Second),
			EnableCompression: false,
			TLSClientConfig:   cs.TLSConf, /*&tls.Config{
				InsecureSkipVerify: true,
			},*/
		}

		c, res, err := d.Dial("wss://"+cs.IP+":"+cs.Port+"/wss", header)
		if err != nil {
			cs.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprintf("Error Dial: '%v' (ip %v)", err, cs.IP),
				FuncName:    funcName,
			})

			return
		}
		defer connClose(cs.COut, c, cs.InfoSourceList, cs.ID, cs.IP, cs.saveMessageApp)

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
					cs.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprintf("Error ReadMessage: '%v', status code: '%v' (ip %v)", err, res.StatusCode, cs.IP),
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
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	cwt chan<- configure.MsgWsTransmission) {

	/* инициализируем HTTPS клиента */
	fmt.Println("\tThe HTTPS client Network Integration is running")

	funcName := "WssClientNetworkInteraction"

	//читаем сертификат CA для того что бы наш клиент доверял сертификату переданному сервером
	rootCA, err := ioutil.ReadFile(appc.PathRootCA)
	if err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})
	}

	//создаем новый пул доверенных центров серификации и добавляем в него корневой сертификат
	cp := x509.NewCertPool()
	if ok := cp.AppendCertsFromPEM(rootCA); !ok {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "root certificate was not added to the pool",
			FuncName:    funcName,
		})
	}

	conf := &tls.Config{
		ServerName: "isems_nih_slave",
		RootCAs:    cp,
	}
	conf.BuildNameToCertificate()

	isCycleProcessing := "non-blocking"
	cycleProcessing := func(l *map[int]configure.SourceSetting) {

		fmt.Println("func 'cycleProcessing'......")

		var countGoroutine int
		chanDone := make(chan struct{})

		handlerRequest := func(chanDone chan<- struct{}, hrp *handlerRequestParameters) {
			port := strconv.Itoa(hrp.port)

			cs := clientSetting{
				ID:             hrp.id,
				IP:             hrp.ip,
				Port:           port,
				InfoSourceList: isl,
				TLSConf:        conf,
				saveMessageApp: saveMessageApp,
				COut:           cOut,
				CwtReq:         cwt,
			}

			client := &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:       0,
					IdleConnTimeout:    10 * time.Second,
					DisableCompression: true,
					TLSClientConfig:    conf,
				},
				CheckRedirect: cs.redirectPolicyFunc,
			}

			req, err := http.NewRequest("GET", "https://"+hrp.ip+":"+port+"/", nil)
			if err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				chanDone <- struct{}{}

				return
			}

			req.Header.Add("Content-Type", "text/plain;charset=utf-8")
			req.Header.Add("Accept-Language", "en")
			req.Header.Add("User-Agent", "Mozilla/5.0 (ISEMS-NIH_slave)")
			req.Header.Add("Token", hrp.token)

			_, err = client.Do(req)
			if err != nil {
				strErr := fmt.Sprint(err)
				if !strings.Contains(strErr, "stop redirect") {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						TypeMessage: "info",
						Description: strErr,
						FuncName:    funcName,
					})
				}
			}

			chanDone <- struct{}{}
		}

		var processHandlerRequest bool
		for id, s := range *l {
			if id == 1000 {
				fmt.Printf("func 'wssClientNetworkInterface', --- ID '%d', STATUS: '%v'\n", id, s.ConnectionStatus)
			}

			if s.AsServer && !s.ConnectionStatus {

				if id == 1000 {
					fmt.Printf("func 'wssClientNetworkInterface', --- CONNECTION to ID '%d', IP '%s'\n", id, s.IP)
				}

				go handlerRequest(chanDone, &handlerRequestParameters{
					id:    id,
					ip:    s.IP,
					port:  s.Settings.IfAsServerThenPort,
					token: s.Token,
				})

				processHandlerRequest = true
				countGoroutine++
			}
		}

		if !processHandlerRequest {
			isCycleProcessing = "non-blocking"

			fmt.Printf("func 'handlerRequest' processHandlerRequest: '%v'\n", processHandlerRequest)

			return
		}

		for {
			<-chanDone

			fmt.Printf("func 'handlerRequest', Before: minus groutine, count go:'%v'\n", countGoroutine)

			countGoroutine--

			fmt.Printf("func 'handlerRequest', After: minus groutine, count go:'%v'\n", countGoroutine)

			if countGoroutine == 0 {

				fmt.Println("func 'handlerRequest', countGoroutine = 00000000")

				break
			}
		}

		fmt.Println("func 'handlerRequest', END")

		isCycleProcessing = "non-blocking"
	}

	//цикличные попытки установления соединения в режиме клиент
	ticker := time.NewTicker(time.Duration(appc.TimeReconnectClient) * time.Second)
	for range ticker.C {
		sl := isl.GetSourceList()
		if len(*sl) == 0 {
			continue
		}

		if isCycleProcessing == "blocking" {

			fmt.Println("isCycleProcessing == 'blocking'")

			continue
		}

		isCycleProcessing = "blocking"

		go cycleProcessing(sl)
	}
}
