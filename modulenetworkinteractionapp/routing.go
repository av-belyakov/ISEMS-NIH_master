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
	"ISEMS-NIH_master/modulenetworkinteractionapp/processrequest"
	"ISEMS-NIH_master/savemessageapp"
)

//RouteCoreRequest маршрутизирует запросы от CoreApp и обрабатывает сообщения от wss модулей
func RouteCoreRequest(cwt chan<- configure.MsgWsTransmission, ism *configure.InformationStoringMemory, chanOutCore <-chan configure.MsgBetweenCoreAndNI, chanColl map[string]chan [2]string) {
	fmt.Println("START 'Routing' module network interaction...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	/*messageHandler := map[string]map[string]int{
		"information": map[string]int{
			"change_status_source": 1,
			"source_telemetry":     2,
			"filtering":            3,
			"download":             4,
			"error_message":        5,
		},
		"instraction": map[string]int{
			"add_source":            1,
			"delete_source":         2,
			"change_setting_source": 3,
			"filtering_start":       4,
			"filtering_stop":        5,
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
				sourceSettings, _ := ism.GetSourceSetting(sourceIP)
				formatJSON, err := processrequest.SendMsgPingPong("ping", sourceSettings.Settings.MaxCountProcessFiltering)
				if err != nil {
					_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
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
		case msg := <-chanOutCore:
			fmt.Println("RESIVED data FROM CoreApp!!!")
			fmt.Println(msg)

			/*

				при delete_source нужно закрыть соединение через c.Close()

					   !!! ОБРАБОТКА ЗАПРОСОВ от CoreApp ТИПА

					"add_source":            1,
					"delete_source":         2,
					"change_setting_source": 3,
					"filtering_start":       4,
					"filtering_stop":        5,
					"download_start":        6,
					"download_stop":         7,
					"download_resume":
					   !!!
			*/

		}
	}
}

//RouteWssConnectionResponse маршрутизирует сообщения от источников
func RouteWssConnectionResponse(cwt chan<- configure.MsgWsTransmission, chanInCore chan<- configure.MsgBetweenCoreAndNI, ism *configure.InformationStoringMemory) {
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//MessageType содержит тип JSON сообщения
	type MessageType struct {
		Type string `json:"messageType"`
	}

	var messageType MessageType

	for _, c := range ism.SourcesListConnection {

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
			sourceSettings, _ := ism.GetSourceSetting(sourceIP)
			formatJSON, err := processrequest.SendMsgPingPong("pong", sourceSettings.Settings.MaxCountProcessFiltering)
			if err != nil {
				_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
			}

			//отправляем источнику запрос типа Ping
			cwt <- configure.MsgWsTransmission{
				DestinationHost: sourceIP,
				Data:            formatJSON,
			}
		case "pong":
			/* Нужно отправить сообщение в RouteCore о том что связь установленна */
		case "source_telemetry":

		case "filtering":

		case "download files":

		}
	}
}
