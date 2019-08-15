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
	SourceList        *configure.InformationSourcesList
}

//SettingsWssServer параметры для взаимодействия с wssServer
type SettingsWssServer struct {
	SourceList *configure.InformationSourcesList
	COut       chan<- [2]string
	CwtReq     chan<- configure.MsgWsTransmission
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
	//если токен валидный изменяем состояние AccessIsAllowed в true
	_, validToken := settingsHTTPServer.SourceList.SearchSourceIPAndToken(remoteAddr, stringToken)

	if (len(stringToken) == 0) || !validToken {
		w.Header().Set("Content-Length", strconv.Itoa(utf8.RuneCount(bodyHTTPResponseError)))

		w.WriteHeader(400)
		w.Write(bodyHTTPResponseError)

		_ = saveMessageApp.LogMessage("error", fmt.Sprintf("missing or incorrect identification token (сlient ipaddress %v)", req.RemoteAddr))

		return
	}

	if id, ok := settingsHTTPServer.SourceList.GetSourceIDOnIP(remoteAddr); ok {
		settingsHTTPServer.SourceList.SetAccessIsAllowed(id)

		http.Redirect(w, req, "https://"+settingsHTTPServer.Host+":"+settingsHTTPServer.Port+"/wss", 301)
	}
}

//ServerWss webSocket запросов
func (sws SettingsWssServer) ServerWss(w http.ResponseWriter, req *http.Request) {
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	remoteIP := strings.Split(req.RemoteAddr, ":")[0]

	clientID, idIsExist := sws.SourceList.GetSourceIDOnIP(remoteIP)
	if !idIsExist {
		w.WriteHeader(401)
		_ = saveMessageApp.LogMessage("error", fmt.Sprintf("access for the user with ipaddress %v is prohibited", remoteIP))

		return
	}

	//проверяем разрешено ли данному ip соединение с сервером wss
	if !sws.SourceList.GetAccessIsAllowed(remoteIP) {
		w.WriteHeader(401)
		_ = saveMessageApp.LogMessage("error", fmt.Sprintf("access for the user with ipaddress %v is prohibited", remoteIP))

		return
	}

	if req.Header.Get("Connection") != "Upgrade" {
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
		if c != nil {
			c.Close()
		}

		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
	}
	defer connClose(sws.COut, c, sws.SourceList, clientID, remoteIP)

	//изменяем состояние соединения для данного источника
	_ = sws.SourceList.ChangeSourceConnectionStatus(clientID, true)

	//добавляем линк соединения по websocket
	sws.SourceList.AddLinkWebsocketConnect(remoteIP, c)

	//отправляем модулю routing сообщение об изменении статуса источника
	sws.COut <- [2]string{remoteIP, "connect"}

	//маршрутизация запросов получаемых с подключенного источника
	for {
		if c == nil {
			break
		}

		_, message, err := c.ReadMessage()
		if err != nil {
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

			break
		}

		sws.CwtReq <- configure.MsgWsTransmission{
			DestinationHost: remoteIP,
			Data:            &message,
		}
	}
}

//WssServerNetworkInteraction соединение в режиме 'Сервер'
func WssServerNetworkInteraction(
	cOut chan<- [2]string,
	appConf *configure.AppConfig,
	isl *configure.InformationSourcesList,
	cwtReq chan<- configure.MsgWsTransmission) {

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	port := strconv.Itoa(appConf.ServerHTTPS.Port)

	settingsHTTPServer := SettingsHTTPServer{
		Host:       appConf.ServerHTTPS.Host,
		Port:       port,
		SourceList: isl,
	}

	settingsWssServer := SettingsWssServer{
		CwtReq:     cwtReq,
		SourceList: isl,
		COut:       cOut,
	}

	/* инициализируем HTTPS сервер */
	fmt.Printf("\tThe HTTPS server Network Integration is running, %v:%v\n", appConf.ServerHTTPS.Host, port)

	http.HandleFunc("/", settingsHTTPServer.HandlerRequest)
	http.HandleFunc("/wss", settingsWssServer.ServerWss)

	if err := http.ListenAndServeTLS(appConf.ServerHTTPS.Host+":"+port, appConf.ServerHTTPS.PathCertFile, appConf.ServerHTTPS.PathPrivateKeyFile, nil); err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

		log.Println(err)
		os.Exit(1)
	}
}
