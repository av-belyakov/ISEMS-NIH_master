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
	SourceList                *configure.InformationSourcesList
	MsgChangeSourceConnection chan<- [2]string
	CwtReq                    chan<- configure.MsgWsTransmission
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

	fmt.Printf("_____________ RESIVED HTTP REQUEST FROM IP:%v _______________ %v\n", req.RemoteAddr, req.Header)

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

		_ = saveMessageApp.LogMessage("error", "missing or incorrect identification token (сlient ipaddress "+req.RemoteAddr+")")
	} else {
		if id, ok := settingsHTTPServer.SourceList.GetSourceIDOnIP(remoteAddr); ok {
			settingsHTTPServer.SourceList.SetAccessIsAllowed(id)

			http.Redirect(w, req, "https://"+settingsHTTPServer.Host+":"+settingsHTTPServer.Port+"/wss", 301)
		}
	}
}

//ServerWss webSocket запросов
func (sws SettingsWssServer) ServerWss(w http.ResponseWriter, req *http.Request) {
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	remoteIP := strings.Split(req.RemoteAddr, ":")[0]

	id, idIsExist := sws.SourceList.GetSourceIDOnIP(remoteIP)
	if !idIsExist {
		w.WriteHeader(401)
		_ = saveMessageApp.LogMessage("error", "access for the user with ipaddress "+remoteIP+" is prohibited")

		return
	}

	sett, _ := sws.SourceList.GetSourceSetting(id)

	fmt.Println("Request WSS", req)

	fmt.Println("Access is allowed:", sett.AccessIsAllowed)

	//проверяем разрешено ли данному ip соединение с сервером wss
	if !sws.SourceList.GetAccessIsAllowed(remoteIP) {
		w.WriteHeader(401)
		_ = saveMessageApp.LogMessage("error", "access for the user with ipaddress "+remoteIP+" is prohibited")

		return
	}

	if req.Header.Get("Connection") != "Upgrade" {
		return
	}

	connectClose := func(c *websocket.Conn) {
		//закрытие канала связи с источником
		if c != nil {
			c.Close()
		}

		if id, ok := sws.SourceList.GetSourceIDOnIP(remoteIP); ok {

			fmt.Println("1111111111")

			//изменяем состояние соединения для данного источника
			_ = sws.SourceList.ChangeSourceConnectionStatus(id)
		}

		//удаляем линк соединения
		sws.SourceList.DelLinkWebsocketConnection(remoteIP)

		//при разрыве соединения отправляем модулю routing сообщение об изменении статуса источников
		sws.MsgChangeSourceConnection <- [2]string{remoteIP, "disconnect"}

		_ = saveMessageApp.LogMessage("info", "websocket disconnect whis ip "+remoteIP)
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

	defer connectClose(c)

	if id, ok := sws.SourceList.GetSourceIDOnIP(remoteIP); ok {

		fmt.Println("22222222222")

		//изменяем состояние соединения для данного источника
		_ = sws.SourceList.ChangeSourceConnectionStatus(id)
	}

	//добавляем линк соединения по websocket
	sws.SourceList.AddLinkWebsocketConnect(remoteIP, c)

	//отправляем модулю routing сообщение об изменении статуса источника
	sws.MsgChangeSourceConnection <- [2]string{remoteIP, "connect"}

	//маршрутизация запросов получаемых с подключенного источника
	for {
		if c == nil {
			break
		}

		_, message, err := c.ReadMessage()
		if err != nil {
			connectClose(c)
		}

		sws.CwtReq <- configure.MsgWsTransmission{
			DestinationHost: remoteIP,
			Data:            &message,
		}
	}
}

//WssServerNetworkInteraction запуск сервера для обработки запросов с источников
func WssServerNetworkInteraction(cOut chan<- [2]string, appConf *configure.AppConfig, isl *configure.InformationSourcesList, cwtReq chan<- configure.MsgWsTransmission) {
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	port := strconv.Itoa(appConf.ServerHTTPS.Port)

	settingsHTTPServer := SettingsHTTPServer{
		Host:       appConf.ServerHTTPS.Host,
		Port:       port,
		SourceList: isl,
	}

	settingsWssServer := SettingsWssServer{
		SourceList:                isl,
		MsgChangeSourceConnection: cOut,
		CwtReq:                    cwtReq,
	}

	/* инициализируем HTTPS сервер */
	log.Println("\tThe HTTPS server Network Integration is running on ip address " + appConf.ServerHTTPS.Host + ", port " + port + "\n")

	http.HandleFunc("/", settingsHTTPServer.HandlerRequest)
	http.HandleFunc("/wss", settingsWssServer.ServerWss)

	if err := http.ListenAndServeTLS(appConf.ServerHTTPS.Host+":"+port, appConf.ServerHTTPS.PathCertFile, appConf.ServerHTTPS.PathPrivateKeyFile, nil); err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

		log.Println(err)
		os.Exit(1)
	}
}
