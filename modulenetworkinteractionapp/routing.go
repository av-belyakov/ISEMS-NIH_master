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

			if action == "disconnect" {
				fmt.Println("--- SOURCE WITH IP", sourceIP, " has success ", action)
				fmt.Println(action, sourceIP)
			}

			if action == "connect" {
				fmt.Println("--- SOURCE WITH IP", sourceIP, " has success ", action)
				fmt.Println(action, sourceIP)
			}

		//модуль wssClient
		case msg := <-chanColl["outWssModuleClient"]:
			sourceIP, action := msg[0], msg[1]

			if action == "connect" {
				if id, ok := isl.GetSourceIDOnIP(sourceIP); ok {
					sourceSettings, _ := isl.GetSourceSetting(id)
					formatJSON, err := processrequest.SendMsgPingPong("ping", sourceSettings.Settings.MaxCountProcessFiltration)
					if err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					//отправляем источнику запрос типа Ping
					cwt <- configure.MsgWsTransmission{
						DestinationHost: sourceIP,
						Data:            &formatJSON,
					}
				}
			}

			if action == "disconnect" {
				fmt.Println("convert IP to ID and send in channel 'chanOutCore'")
				fmt.Println(action, sourceIP)
			}

			//обработка сообщения от ядра
		case msg := <-chanOutCore:
			go handlers.HandlerMsgFromCore(cwt, isl, msg, chanInCore)
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
	}

	var messageType MessageType

	//sourcesListConnection := isl.GetSourcesListConnection()

	for msg := range cwtReq {
		sourceIP := msg.DestinationHost
		message := msg.Data

		if err := json.Unmarshal(*message, &messageType); err != nil {
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
					Data:            &formatJSON,
				}
			}

		case "pong":
			/* Нужно отправить сообщение в RouteCore о том что связь установленна */
		case "source_telemetry":

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
