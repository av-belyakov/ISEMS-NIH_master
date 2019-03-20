package moduleapiapp

/*
* Модуль прикладного программного интерфейса (API)
*
* Версия 0.21, дата релиза 12.03.2019
* */

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

type settingsServerAPI struct {
	IP, Port string
	Tokens   []configure.SettingsAuthenticationTokenClientsAPI
}

type channels struct {
	ChanIn, ChanOut chan configure.MsgBetweenCoreAndAPI
}

var chn channels
var storingMemoryAPI *configure.StoringMemoryAPI

//запрос на получение списка источников
func sendMsgGetSourceList(clientID string) error {
	errMsg := "unable to send message no customer information available"

	clientSettings, ok := storingMemoryAPI.GetClientSettings(clientID)
	if !ok {
		return errors.New(errMsg)
	}

	msgjson, err := json.Marshal(configure.MsgCommon{
		MsgType:        "command",
		MsgSection:     "source control",
		MsgInsturction: "get new source list",
	})
	if err != nil {
		return err
	}

	clientSettings.SendWsMessage(1, msgjson)

	return nil
}

//HandlerRequest обработчик HTTPS запроса к "/"
func (settingsServerAPI *settingsServerAPI) HandlerRequest(w http.ResponseWriter, req *http.Request) {
	fmt.Println("RESIVED http request '/api'")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	bodyHTTPResponseError := []byte(`<!DOCTYPE html>
		<html lang="en"
		<head><meta charset="utf-8"><title>Server Nginx</title></head>
		<body><h1>Access denied. For additional information, please contact the webmaster.</h1></body>
		</html>`)

	stringToken := ""
	for headerName := range req.Header {
		if headerName == "Token" {
			stringToken = req.Header[headerName][0]
			continue
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Language", "en")

	if req.Method != "GET" {
		http.Error(w, "Method not allowed", 405)

		return
	}

	if len(stringToken) == 0 {
		w.Header().Set("Content-Length", strconv.Itoa(utf8.RuneCount(bodyHTTPResponseError)))

		w.WriteHeader(400)
		w.Write(bodyHTTPResponseError)

		_ = saveMessageApp.LogMessage("error", "Server API - missing or incorrect identification token (сlient ipaddress "+req.RemoteAddr+")")
	}

	for _, sc := range settingsServerAPI.Tokens {
		if stringToken == sc.Token {
			remoteAddr := strings.Split(req.RemoteAddr, ":")[0]
			//добавляем нового пользователя которому разрешен доступ
			_ = storingMemoryAPI.AddNewClient(remoteAddr, sc.Name)

			http.Redirect(w, req, "https://"+settingsServerAPI.IP+":"+settingsServerAPI.Port+"/api_wss", 301)

			return
		}
	}

	w.Header().Set("Content-Length", strconv.Itoa(utf8.RuneCount(bodyHTTPResponseError)))

	w.WriteHeader(400)
	w.Write(bodyHTTPResponseError)

	_ = saveMessageApp.LogMessage("error", "Server API - missing or incorrect identification token (сlient ipaddress "+req.RemoteAddr+")")
}

func serverWss(w http.ResponseWriter, req *http.Request) {
	fmt.Println("START server wss")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	remoteIP := strings.Split(req.RemoteAddr, ":")[0]

	//проверяем прошел ли клиент аутентификацию
	clientID, _, ok := storingMemoryAPI.SearchClientForIP(remoteIP)
	if !ok {
		w.WriteHeader(401)
		_ = saveMessageApp.LogMessage("error", "Server API - access for the user with ipaddress "+req.RemoteAddr+" is prohibited")
		return
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
		//ReadBufferSize:    1024,
		//WriteBufferSize:   100000000,
		HandshakeTimeout: (time.Duration(1) * time.Second),
	}

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		c.Close()

		//удаляем информацию о клиенте
		storingMemoryAPI.DelClientAPI(clientID)

		_ = saveMessageApp.LogMessage("error", "Server API - "+fmt.Sprint(err))

		log.Println("Client API whis ip", remoteIP, "is disconnect")
	}

	//получаем настройки клиента
	clientSettings, ok := storingMemoryAPI.GetClientSettings(clientID)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "Server API - client setup with ID "+clientID+" not found")

		return
	}

	/*defer func() {
		c.Close()

		//удаляем информацию о клиенте
		storingMemoryAPI.DelClientAPI(clientID)
		_ = saveMessageApp.LogMessage("info", "Server API - disconnect for IP address "+remoteIP)

		fmt.Println("Client API whis ip", remoteIP, "is disconnect")
	}()*/

	storingMemoryAPI.SaveWssClientConnection(clientID, c)

	//при подключении клиента отправляем запрос на получение списка источников
	sendMsgGetSourceList(clientID)

	//обработка ответов получаемых от ядра приложения
	go func() {
		for msg := range chn.ChanOut {
			//			fmt.Println("resived message from Core to API", msg)
			//			fmt.Println("Storage ClientID:", clientID, ", resived ClientID:", msg.IDClientAPI)

			if msg.MsgGenerator == "Core module" && msg.MsgRecipient == "API module" {
				clientSettings, ok := storingMemoryAPI.GetClientSettings(msg.IDClientAPI)

				if !ok {
					_ = saveMessageApp.LogMessage("error", "Server API - client with id "+msg.IDClientAPI+" not found, he sending data is not possible")

					continue
				}

				msgjson, ok := msg.MsgJSON.([]byte)
				if !ok {
					_ = saveMessageApp.LogMessage("error", "Server API - failed to send json message, error while casting type")

					continue
				}

				if err := clientSettings.SendWsMessage(1, msgjson); err != nil {
					_ = saveMessageApp.LogMessage("error", "Server API - "+fmt.Sprint(err))
				}
			}
		}
	}()

	//маршрутизация сообщений приходящих от клиентов API
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				c.Close()

				//удаляем информацию о клиенте
				storingMemoryAPI.DelClientAPI(clientID)
				_ = saveMessageApp.LogMessage("error", "Server API - "+fmt.Sprint(err))

				log.Println("Client API whis ip", remoteIP, "is disconnect")
				break
			}

			fmt.Println("resived message from API client")

			chn.ChanIn <- configure.MsgBetweenCoreAndAPI{
				MsgGenerator: "API module",
				MsgRecipient: "Core module",
				IDClientAPI:  clientID,
				ClientName:   clientSettings.ClientName,
				MsgJSON:      message,
			}
		}
	}()

	if err := recover(); err != nil {
		_ = saveMessageApp.LogMessage("error", "Server API - "+fmt.Sprint(err))
	}
}

func init() {
	chn = channels{
		ChanOut: make(chan configure.MsgBetweenCoreAndAPI, 10),
		ChanIn:  make(chan configure.MsgBetweenCoreAndAPI, 10),
	}

	storingMemoryAPI = configure.NewRepositorySMAPI()
}

//MainAPIApp обработчик запросов поступающих через API
func MainAPIApp(appConfig *configure.AppConfig) (chanOut, chanIn chan configure.MsgBetweenCoreAndAPI) {
	fmt.Println("START module 'MainAppAPI'...")

	settingsServerAPI := settingsServerAPI{
		IP:     appConfig.ServerAPI.Host,
		Port:   strconv.Itoa(appConfig.ServerAPI.Port),
		Tokens: appConfig.AuthenticationTokenClientsAPI,
	}

	go func() {
		//создаем сервер wss для подключения клиентов
		http.HandleFunc("/api", settingsServerAPI.HandlerRequest)
		http.HandleFunc("/api_wss", serverWss)

		err := http.ListenAndServeTLS(settingsServerAPI.IP+":"+settingsServerAPI.Port, appConfig.ServerAPI.PathCertFile, appConfig.ServerAPI.PathPrivateKeyFile, nil)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}()
	log.Println("\tAPI server successfully started at", settingsServerAPI.IP+":"+settingsServerAPI.Port)

	return chn.ChanOut, chn.ChanIn
}
