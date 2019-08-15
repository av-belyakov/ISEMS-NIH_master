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
	IP, Port       string
	Tokens         []configure.SettingsAuthenticationTokenClientsAPI
	SaveMessageApp *savemessageapp.PathDirLocationLogFiles
}

type channels struct {
	ChanIn, ChanOut chan *configure.MsgBetweenCoreAndAPI
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
	//fmt.Println("RESIVED http request '/api'")

	defer func() {
		if err := recover(); err != nil {
			_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - %v", fmt.Sprint(err)))
		}
	}()

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

		_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - missing or incorrect identification token (сlient ipaddress %v)", req.RemoteAddr))
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

	_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - missing or incorrect identification token (сlient ipaddress %v)", req.RemoteAddr))
}

func (settingsServerAPI *settingsServerAPI) serverWss(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - %v", fmt.Sprint(err)))
		}
	}()

	remoteIP := strings.Split(req.RemoteAddr, ":")[0]

	//проверяем прошел ли клиент аутентификацию
	clientID, _, ok := storingMemoryAPI.SearchClientForIP(remoteIP)
	if !ok {
		w.WriteHeader(401)
		_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - access for the user with ipaddress %v is prohibited", req.RemoteAddr))
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

		_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - %v", fmt.Sprint(err)))

		log.Println("Client API whis ip", remoteIP, "is disconnect")
	}

	//получаем настройки клиента
	clientSettings, ok := storingMemoryAPI.GetClientSettings(clientID)
	if !ok {
		_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - client setup with ID %v not found", clientID))

		return
	}

	storingMemoryAPI.SaveWssClientConnection(clientID, c)

	//при подключении клиента отправляем запрос на получение списка источников
	sendMsgGetSourceList(clientID)

	//обработка ответов получаемых от ядра приложения
	go func() {
		for msg := range chn.ChanOut {
			if msg.MsgGenerator == "Core module" && msg.MsgRecipient == "API module" {
				msgjson, ok := msg.MsgJSON.([]byte)
				if !ok {
					_ = settingsServerAPI.SaveMessageApp.LogMessage("error", "Server API - failed to send json message, error while casting type")

					continue
				}

				clientSettings, ok := storingMemoryAPI.GetClientSettings(msg.IDClientAPI)
				//если клиент с таким ID не найден, отправляем широковещательное сообщение
				if !ok {
					cl := storingMemoryAPI.GetClientList()
					for _, cs := range cl {
						if err := cs.SendWsMessage(1, msgjson); err != nil {
							_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - %v", fmt.Sprint(err)))
						}
					}

					continue
				}

				if err := clientSettings.SendWsMessage(1, msgjson); err != nil {
					_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - %v", fmt.Sprint(err)))
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
				_ = settingsServerAPI.SaveMessageApp.LogMessage("error", fmt.Sprintf("Server API - %v", fmt.Sprint(err)))

				log.Println("Client API whis ip", remoteIP, "is disconnect")

				break
			}

			chn.ChanIn <- &configure.MsgBetweenCoreAndAPI{
				MsgGenerator: "API module",
				MsgRecipient: "Core module",
				IDClientAPI:  clientID,
				ClientName:   clientSettings.ClientName,
				ClientIP:     remoteIP,
				MsgJSON:      message,
			}
		}
	}()
}

func init() {
	chn = channels{
		ChanOut: make(chan *configure.MsgBetweenCoreAndAPI),
		ChanIn:  make(chan *configure.MsgBetweenCoreAndAPI),
	}

	storingMemoryAPI = configure.NewRepositorySMAPI()
}

//MainAPIApp обработчик запросов поступающих через API
func MainAPIApp(appConfig *configure.AppConfig, saveMessageApp *savemessageapp.PathDirLocationLogFiles) (chanOut, chanIn chan *configure.MsgBetweenCoreAndAPI) {
	settingsServerAPI := settingsServerAPI{
		IP:             appConfig.ServerAPI.Host,
		Port:           strconv.Itoa(appConfig.ServerAPI.Port),
		Tokens:         appConfig.AuthenticationTokenClientsAPI,
		SaveMessageApp: saveMessageApp,
	}

	go func() {
		//создаем сервер wss для подключения клиентов
		http.HandleFunc("/api", settingsServerAPI.HandlerRequest)
		http.HandleFunc("/api_wss", settingsServerAPI.serverWss)

		err := http.ListenAndServeTLS(settingsServerAPI.IP+":"+settingsServerAPI.Port, appConfig.ServerAPI.PathCertFile, appConfig.ServerAPI.PathPrivateKeyFile, nil)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}()

	fmt.Printf("\tAPI server successfully started at %v:%v\n", settingsServerAPI.IP, settingsServerAPI.Port)

	return chn.ChanOut, chn.ChanIn
}
