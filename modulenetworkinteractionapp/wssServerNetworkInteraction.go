package modulenetworkinteractionapp

/*
* Сервер для взаимодействия с источниками
*
* Версия 0.1, дата релиза 21.02.2019
* */

import (
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

//SettingsHTTPServer параметры необходимые при взаимодействии с HTTP сервером
type SettingsHTTPServer struct {
	Host, Port, Token string
	StorMem           *configure.InformationStoringMemory
}

//SettingsWssServer параметры для взаимодействия с wssServer
type SettingsWssServer struct {
	StorMem *configure.InformationStoringMemory
}

//HandlerRequest обработчик HTTPS запросов
func (settingsHTTPServer *SettingsHTTPServer) HandlerRequest(w http.ResponseWriter, req *http.Request) {
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

	remoteAddr := strings.Split(req.RemoteAddr, ":")[0]

	_, validToken := settingsHTTPServer.StorMem.SearchSourceToken(remoteAddr, stringToken)
	if (len(stringToken) == 0) || !validToken {
		w.Header().Set("Content-Length", strconv.Itoa(utf8.RuneCount(bodyHTTPResponseError)))

		w.WriteHeader(400)
		w.Write(bodyHTTPResponseError)

		_ = saveMessageApp.LogMessage("error", "missing or incorrect identification token (сlient ipaddress "+req.RemoteAddr+")")
	} else {
		http.Redirect(w, req, "https://"+settingsHTTPServer.Host+":"+settingsHTTPServer.Port+"/wss", 301)
	}
}

//ServerWss webSocket запросов
func (sws SettingsWssServer) ServerWss(w http.ResponseWriter, req *http.Request) {
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	remoteIP := strings.Split(req.RemoteAddr, ":")[0]
	_, ipIsExist := sws.StorMem.GetSourceSetting(remoteIP)

	if !ipIsExist || !sws.StorMem.GetAccessIsAllowed(remoteIP) {
		w.WriteHeader(401)
		_ = saveMessageApp.LogMessage("error", "access for the user with ipaddress "+req.RemoteAddr+" is prohibited")
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

		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
	}
	defer func() {

		//		chanEndGoroutin <- struct{}{}

		//закрытие канала связи с источником
		c.Close()

		//изменяем состояние соединения для данного источника
		_ = sws.StorMem.ChangeSourceConnectionStatus(remoteIP)
		_ = saveMessageApp.LogMessage("info", "disconnect for IP address "+remoteIP)

		/*		if _, ok := acc.Addresses[remoteIP]; !ok {
				fmt.Println(ok, "--- --- ---- IPADDRESS ", remoteIP, "NOT FOUND, WEBSOCKET DISCONNECT")
			}*/

		//при разрыве соединения удаляем задачу по скачиванию файлов
		//dfi.DelTaskDownloadFiles(remoteIP)

		fmt.Println("websocket disconnect whis ip", remoteIP)
	}()

	//изменяем состояние соединения для данного источника
	_ = sws.StorMem.ChangeSourceConnectionStatus(remoteIP)

	//добавляем линк соединения по websocket
	sws.StorMem.AddLinkWebsocketConnect(remoteIP, c)

	/*	go func(acc *configure.AccessClientsConfigure) {
		DONE:
			for {
				select {
				case messageText := <-acc.ChanWebsocketTranssmition:
					if _, isExist := acc.Addresses[remoteIP]; isExist {
						if err := acc.Addresses[remoteIP].SendWsMessage(1, messageText); err != nil {
							_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
						}
					}
				case messageBinary := <-acc.ChanWebsocketTranssmitionBinary:
					if _, isExist := acc.Addresses[remoteIP]; isExist {
						if err := acc.Addresses[remoteIP].SendWsMessage(2, messageBinary); err != nil {
							_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
						}
					}
				case <-chanEndGoroutin:

					chanStopSendInfoTranssmition <- struct{}{}

					break DONE
				}
			}

			//		fmt.Println("_!!!_ COUNT GOROUTINE:", runtime.NumGoroutine())
		}(&acc)*/

	if e := recover(); e != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(e))
	}

	//	routes.RouteWebSocketRequest(remoteIP, &acc, &ift, &dfi, &mc, chanStopSendInfoTranssmition)
}

//WssServerNetworkInteraction запуск сервера для обработки запросов с источников
func WssServerNetworkInteraction(appConf *configure.AppConfig, ism *configure.InformationStoringMemory) {
	fmt.Println("START WSS SERVER...")
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	fmt.Println("*** Send request for LIST SOURCES ***")

	//запрос на получение списка источников
	ism.ChannelCollection.ChannelFromMNIService <- configure.ServiceMessageInfoStatusSource{Type: "get_list"}

	for msg := range ism.ChannelCollection.ChannelToMNIService {
		fmt.Println("*** GET SOURCE LIST ***", msg)

		if msg.Type == "send_list" {
			settingsHTTPServer := SettingsHTTPServer{
				Host:    appConf.ServerHTTP.Host,
				Port:    string(appConf.ServerHTTP.Port),
				StorMem: ism,
			}

			settingsWssServer := SettingsWssServer{
				StorMem: ism,
			}

			/* инициализируем HTTPS сервер */
			log.Println("The HTTPS server is running on ip address " + appConf.ServerHTTP.Host + ", port " + string(appConf.ServerHTTP.Port) + "\n")

			http.HandleFunc("/", settingsHTTPServer.HandlerRequest)
			http.HandleFunc("/wss", settingsWssServer.ServerWss)

			port := string(appConf.ServerHTTP.Port)
			if err := http.ListenAndServeTLS(appConf.ServerHTTP.Host+":"+port, appConf.PathCertFile, appConf.PathPrivateKeyFile, nil); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				log.Println(err)
				os.Exit(1)
			}
		}
	}

	/*
		select {
		case msg := <- ism.ChannelCollection.ChannelToMNIService:
			fmt.Println("*** GET SOURCE LIST ***")


			}
	*/
}
