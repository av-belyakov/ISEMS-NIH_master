package moduleapiapp

/*
* Модуль прикладного программного интерфейса (API)
*
* Версия 0.2, дата релиза 11.03.2019
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

type settingsServerAPI struct {
	IP, Port, Token string
}

type channels struct {
	ChanIn, ChanOut chan configure.MsgBetweenCoreAndAPI
}

var chn channels
var storingMemoryAPI configure.StoringMemoryAPI

//HandlerRequest обработчик HTTPS запроса к "/"
func (settingsServerAPI *settingsServerAPI) HandlerRequest(w http.ResponseWriter, req *http.Request) {
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

	if (len(stringToken) == 0) || (stringToken != settingsServerAPI.Token) {
		w.Header().Set("Content-Length", strconv.Itoa(utf8.RuneCount(bodyHTTPResponseError)))

		w.WriteHeader(400)
		w.Write(bodyHTTPResponseError)

		_ = saveMessageApp.LogMessage("error", "Server API - missing or incorrect identification token (сlient ipaddress "+req.RemoteAddr+")")
	} else {
		remoteAddr := strings.Split(req.RemoteAddr, ":")[0]
		//добавляем нового пользователя которому разрешен доступ
		_ = storingMemoryAPI.AddNewClient(remoteAddr)

		http.Redirect(w, req, "https://"+settingsServerAPI.IP+":"+settingsServerAPI.Port+"/wss", 301)
	}
}

func serverWss(w http.ResponseWriter, req *http.Request) {
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
	}
	defer func() {
		c.Close()

		//удаляем информацию о клиенте
		storingMemoryAPI.DelClientAPI(clientID)
		_ = saveMessageApp.LogMessage("info", "Server API - disconnect for IP address "+remoteIP)

		fmt.Println("Client API whis ip", remoteIP, "is disconnect")
	}()

	storingMemoryAPI.SaveWssClientConnection(clientID, c)

	//обработка ответов получаемых от ядра приложения
	go func() {
		for msg := range chn.ChanOut {
			fmt.Println("resived message from Core to API", msg)

			if msg.MsgGenerator == "Core module" && msg.MsgRecipient == "API module" {
				clientSettings, ok := storingMemoryAPI.GetClientSettings(msg.IDClientAPI)
				if !ok {
					_ = saveMessageApp.LogMessage("error", "Server API - client with id "+msg.IDClientAPI+" not found, he sending data is not possible")
				}

				if err := clientSettings.SendWsMessage(1, msg.MsgJSON); err != nil {
					_ = saveMessageApp.LogMessage("error", "Server API - "+fmt.Sprint(err))
				}
			}
		}
	}()

	//передача сообщений в ядро
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				_ = saveMessageApp.LogMessage("error", "Server API - "+fmt.Sprint(err))
				break
			}

			chn.ChanIn <- configure.MsgBetweenCoreAndAPI{
				MsgGenerator: "API module",
				MsgRecipient: "Core module",
				IDClientAPI:  clientID,
				MsgJSON:      message,
			}
		}
	}()

	if e := recover(); e != nil {
		_ = saveMessageApp.LogMessage("error", "Server API - "+fmt.Sprint(e))
	}
}

func init() {
	chn = channels{
		ChanOut: make(chan configure.MsgBetweenCoreAndAPI, 10),
		ChanIn:  make(chan configure.MsgBetweenCoreAndAPI, 10),
	}

	defer func() {
		close(chn.ChanIn)
		close(chn.ChanOut)
	}()

	storingMemoryAPI.NewRepository()
}

//MainAppAPI обработчик запросов поступающих через API
func MainAppAPI(appConfig *configure.AppConfig, ism *configure.InformationStoringMemory) (chanOut, chanIn chan configure.MsgBetweenCoreAndAPI) {
	fmt.Println("START module 'MainAppAPI'...")

	settingsServerAPI := settingsServerAPI{
		IP:    appConfig.ServerAPI.Host,
		Port:  strconv.Itoa(appConfig.ServerAPI.Port),
		Token: appConfig.AuthenticationTokenClientAPI,
	}

	//создаем сервер wss для подключения клиентов
	http.HandleFunc("/", settingsServerAPI.HandlerRequest)
	http.HandleFunc("/wss", serverWss)

	err := http.ListenAndServeTLS(settingsServerAPI.IP+":"+settingsServerAPI.Port, appConfig.ServerAPI.PathCertFile, appConfig.ServerAPI.PathPrivateKeyFile, nil)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	/* ПОКА ПРОСТО ТЕСТОВОЕ СООБЩЕНИЕ С НОВЫМ СПИСКОМ ИСТОЧНИКОВ */
	// --- ТЕСТОВЫЙ ОТВЕТ ---
	/*	chanIn <- configure.MsgBetweenCoreAndAPI{
				MsgGenerator: "API module",
				MsgRecipient: "Core module",
				IDClientAPI:  "du68whfh733hjf9393",
		MsgJSON:
	*/

	/*
	   ПОДГОТОВИТЬ ТЕСТОВЫЙ JSON КОТОРЫЙ ДОБАВЛЯЕТСЯ В MsgJSON


	   ОТПРАВЛЯТЬ СООБЩЕНИЕ С ПРЕКЛКПЛЕННОМ К НЕМУ JSON данными полученными
	   от клиента API

	   MsgType:      "information",
	   		MsgSection:   "source control",
	   		IDClientAPI:  "du68whfh733hjf9393",
	   		AdvancedOptions: configure.MsgInfoChangeStatusSource{
	   			SourceListIsExist: true,
	   			SourceList: []configure.MainOperatingParametersSource{
	   				{9, "127.0.0.1", "fmdif3o444fdf344k0fiif", false, configure.SourceDetailedInformation{}},
	   				{10, "192.168.0.10", "fmdif3o444fdf344k0fiif", false, configure.SourceDetailedInformation{}},
	   				{11, "192.168.0.11", "ttrr9gr9r9e9f9fadx94", false, configure.SourceDetailedInformation{}},
	   				{12, "192.168.0.12", "2n3n3iixcxcc3444xfg0222", false, configure.SourceDetailedInformation{}},
	   				{13, "192.168.0.13", "osdsoc9c933cc9cn939f9f33", true, configure.SourceDetailedInformation{}},
	   				{14, "192.168.0.14", "hgdfffffff9333ffodffodofff0", true, configure.SourceDetailedInformation{}},
	   			},
	   		},
	   	}*/
	//------------------------

	/*
	   if message := <-*ism.ChannelCollection.ChanMessageToAPI {
	   	*ism.ChannelCollection.ChanMessageFromAPI<- configure.MessageAPI{
	   		MsgID: "2",
	   		MsgType: "response",
	   		MsgDate: 838283,
	   	}
	   }
	   	fmt.Println("MESSAGE TO API:", <-*ism.ChannelCollection.ChanMessageToAPI)
	*/

	return chn.ChanOut, chn.ChanIn
}
