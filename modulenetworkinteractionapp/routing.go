package modulenetworkinteractionapp

/*
* Маршрутизация запросов приходящих через websocket
*
* Версия 0.1, дата релиза 26.02.2019
* */

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/modulenetworkinteractionapp/handlers"
	"ISEMS-NIH_master/modulenetworkinteractionapp/processrequest"
	"ISEMS-NIH_master/savemessageapp"
)

//RouteCoreRequest маршрутизирует запросы от CoreApp и обрабатывает сообщения от wss модулей
func RouteCoreRequest(cwt chan<- configure.MsgWsTransmission, isl *configure.InformationSourcesList, chanOutCore <-chan configure.MsgBetweenCoreAndNI, chanColl map[string]chan [2]string) {
	fmt.Println("START module 'RouteCoreRequest' (network interaction)...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	/*messageHandler := map[string]map[string]int{
		"information": map[string]int{
			"change_status_source": 1,
			"source_telemetry":     2,
			"filtration":            3,
			"download":             4,
			"error_message":        5,
		},
		"instraction": map[string]int{
			"add_source":            1,
			"delete_source":         2,
			"change_setting_source": 3,
			"filtration_start":       4,
			"filtration_stop":        5,
			"download_start":        6,
			"download_stop":         7,
			"download_resume":       8,
		},
	}*/

	//обработка данных получаемых через каналы
	for {
		select {
		/*
			обработка сообщений от МОДУЛЕЙ WSS (информация об источниках)
		*/
		//модуль wssServer
		case msg := <-chanColl["outWssModuleServer"]:
			sourceIP, action := msg[0], msg[1]

			if action == "disconnect" {
				fmt.Println("convert IP to ID and send in channel 'chanOutCore'")
				fmt.Println(action, sourceIP)
			}

		//модуль wssClient
		case msg := <-chanColl["outWssModuleClient"]:
			sourceIP, action := msg[0], msg[1]

			if action == "connect" {
				sourceSettings, _ := isl.GetSourceSetting(sourceIP)
				formatJSON, err := processrequest.SendMsgPingPong("ping", sourceSettings.Settings.MaxCountProcessfiltration)
				if err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}

				//отправляем источнику запрос типа Ping
				cwt <- configure.MsgWsTransmission{
					DestinationHost: sourceIP,
					Data:            formatJSON,
				}
			}

			if action == "disconnect" {
				fmt.Println("convert IP to ID and send in channel 'chanOutCore'")
				fmt.Println(action, sourceIP)
			}

			/*
				обработка сообщений от ЯДРА приложения, тип:
					SourceID
					MsgType //команда/информационное
					TypeRequiredAction string
					Data []byte
			*/

			//обработка сообщения от ядра
		case msg := <-chanOutCore:
			fmt.Println("RESIVED data FROM CoreApp!!!")
			fmt.Printf("%v\n", msg)

			handlers.HandlerMsgFromCore(cwt, isl, msg)
		}
	}
}

//RouteWssConnectionResponse маршрутизирует сообщения от источников
func RouteWssConnectionResponse(cwt chan<- configure.MsgWsTransmission, chanInCore chan<- configure.MsgBetweenCoreAndNI, isl *configure.InformationSourcesList) {
	fmt.Println("START module 'RouteWssConnectionResponse' (network interaction)...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//MessageType содержит тип JSON сообщения
	type MessageType struct {
		Type string `json:"messageType"`
	}

	var messageType MessageType

	sourcesListConnection := isl.GetSourcesListConnection()

	for _, c := range sourcesListConnection {

		sourceIP := c.Link.RemoteAddr().String()

		_, message, err := c.Link.ReadMessage()
		if err != nil {
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			break
		}
		if err = json.Unmarshal(message, &messageType); err != nil {
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
		}

		switch messageType.Type {
		case "ping":
			sourceSettings, _ := isl.GetSourceSetting(sourceIP)
			formatJSON, err := processrequest.SendMsgPingPong("pong", sourceSettings.Settings.MaxCountProcessfiltration)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

			//отправляем источнику запрос типа Ping
			cwt <- configure.MsgWsTransmission{
				DestinationHost: sourceIP,
				Data:            formatJSON,
			}
		case "pong":
			/* Нужно отправить сообщение в RouteCore о том что связь установленна */
		case "source_telemetry":

		case "filtration":

		case "download files":

		}
	}
}
