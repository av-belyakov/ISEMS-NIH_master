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
func RouteCoreRequest(
	cwt chan<- configure.MsgWsTransmission,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	isl *configure.InformationSourcesList,
	smt *configure.StoringMemoryTask,
	chanColl map[string]chan [2]string,
	chanOutCore <-chan *configure.MsgBetweenCoreAndNI) {

	fmt.Println("START module 'RouteCoreRequest' (network interaction)...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//обработка данных получаемых через каналы
	for {
		select {
		/*
			обработка сообщений от МОДУЛЕЙ WSS (информация об источниках)
		*/
		//модуль wssServer
		case msg := <-chanColl["outWssModuleServer"]:
			sourceIP, action := msg[0], msg[1]

			fmt.Println("--- SERVER: SOURCE WITH IP", sourceIP, " has success ", action)
			fmt.Println(action, sourceIP)

			sourceID, ok := isl.GetSourceIDOnIP(sourceIP)
			if !ok {
				break
			}

			sendMsg := configure.MsgBetweenCoreAndNI{
				Section: "source control",
				Command: "change connection status source",
				AdvancedOptions: configure.SettingsChangeConnectionStatusSource{
					ID:     sourceID,
					Status: action,
				},
			}

			chanInCore <- &sendMsg

			if action == "connect" {
				if id, ok := isl.GetSourceIDOnIP(sourceIP); ok {
					ss, _ := isl.GetSourceSetting(id)
					formatJSON, err := processrequest.SendMsgPing(ss)
					if err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					fmt.Printf("send msg type PING source %v (action SERVER)\n", id)

					//отправляем источнику запрос типа Ping
					cwt <- configure.MsgWsTransmission{
						DestinationHost: sourceIP,
						Data:            &formatJSON,
					}
				}
			}

		//модуль wssClient
		case msg := <-chanColl["outWssModuleClient"]:
			sourceIP, action := msg[0], msg[1]

			fmt.Println("--- CLIENT: SOURCE WITH IP", sourceIP, " has success ", action)
			fmt.Println(action, sourceIP)

			sourceID, ok := isl.GetSourceIDOnIP(sourceIP)
			if !ok {
				break
			}

			sendMsg := configure.MsgBetweenCoreAndNI{
				Section: "source control",
				Command: "change connection status source",
				AdvancedOptions: configure.SettingsChangeConnectionStatusSource{
					ID:     sourceID,
					Status: action,
				},
			}

			chanInCore <- &sendMsg

			if action == "connect" {
				if id, ok := isl.GetSourceIDOnIP(sourceIP); ok {
					ss, _ := isl.GetSourceSetting(id)
					formatJSON, err := processrequest.SendMsgPing(ss)
					if err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					fmt.Println("send msg type PING (action CLIENT)")

					//отправляем источнику запрос типа Ping
					cwt <- configure.MsgWsTransmission{
						DestinationHost: sourceIP,
						Data:            &formatJSON,
					}
				}
			}

			//обработка сообщения от ядра
		case msg := <-chanOutCore:
			go handlers.HandlerMsgFromCore(cwt, isl, msg, smt, chanInCore)
		}
	}
}

//RouteWssConnectionResponse маршрутизирует сообщения от источников
func RouteWssConnectionResponse(
	cwtRes chan<- configure.MsgWsTransmission,
	isl *configure.InformationSourcesList,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	cwtReq <-chan configure.MsgWsTransmission) {

	fmt.Println("START module 'RouteWssConnectionResponse' (network interaction)...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//MessageType содержит тип JSON сообщения
	type MessageType struct {
		Type string `json:"messageType"`
		//		Info []byte `json:"info"`
	}

	var messageType MessageType

	//sourcesListConnection := isl.GetSourcesListConnection()

	for msg := range cwtReq {
		sourceIP := msg.DestinationHost
		message := msg.Data

		fmt.Println("RESIVED source ip", sourceIP)
		//fmt.Printf("%v\n", *message)

		if err := json.Unmarshal(*message, &messageType); err != nil {
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
		}

		switch messageType.Type {
		case "pong":

			fmt.Println("RESIVED message type 'PONG' from IP", sourceIP)

		case "telemetry":
			fmt.Println("RESIVED message type 'TELEMETRY' from IP", sourceIP)
			fmt.Printf("%v\n", messageType)

			sourceID, _ := isl.GetSourceIDOnIP(sourceIP)

			chanInCore <- &configure.MsgBetweenCoreAndNI{
				Section:         "source control",
				Command:         "telemetry",
				SourceID:        sourceID,
				AdvancedOptions: message,
			}

		case "filtration":

		case "download files":

		}
	}

	/*for _, c := range sourcesListConnection {

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
		if id, ok := isl.GetSourceIDOnIP(sourceIP); ok {
			sourceSettings, _ := isl.GetSourceSetting(id)
			formatJSON, err := processrequest.SendMsgPingPong("pong", sourceSettings.Settings.MaxCountProcessFiltration)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

			//отправляем источнику запрос типа Ping
			cwtRes <- configure.MsgWsTransmission{
				DestinationHost: sourceIP,
				Data:            formatJSON,
			}
		}

	case "pong":
		/* Нужно отправить сообщение в RouteCore о том что связь установленна */
	/*case "source_telemetry":

		case "filtration":

		case "download files":

		}
	}*/
}
