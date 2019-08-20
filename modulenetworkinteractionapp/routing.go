package modulenetworkinteractionapp

/*
* Маршрутизация запросов приходящих через websocket
*
* Версия 0.2, дата релиза 30.05.2019
* */

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/modulenetworkinteractionapp/handlers"
	"ISEMS-NIH_master/modulenetworkinteractionapp/processrequest"
	"ISEMS-NIH_master/modulenetworkinteractionapp/processresponse"
	"ISEMS-NIH_master/savemessageapp"
)

//RouteCoreRequest маршрутизирует запросы от CoreApp и обрабатывает сообщения от wss модулей
func RouteCoreRequest(
	cwt chan<- configure.MsgWsTransmission,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	chanInCRRF chan<- *handlers.MsgChannelReceivingFiles,
	isl *configure.InformationSourcesList,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanColl map[string]chan [2]string,
	chanOutCore <-chan *configure.MsgBetweenCoreAndNI) {

	//обработка данных получаемых через каналы
	for {
		select {
		/*
			обработка сообщений от МОДУЛЕЙ WSS (информация об источниках)
		*/
		//модуль wssServer
		case msg := <-chanColl["outWssModuleServer"]:
			sourceIP, action := msg[0], msg[1]

			_ = saveMessageApp.LogMessage("info", fmt.Sprintf("SERVER: source with IP %v has success %v", sourceIP, action))

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

					_ = saveMessageApp.LogMessage("info", fmt.Sprintf("SERVER: send msg type PING source %v", id))

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

			_ = saveMessageApp.LogMessage("info", fmt.Sprintf("CLIENT: source with IP %v has success %v", sourceIP, action))

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

					_ = saveMessageApp.LogMessage("info", fmt.Sprintf("CLIENT: send msg type PING source %v", id))

					//отправляем источнику запрос типа Ping
					cwt <- configure.MsgWsTransmission{
						DestinationHost: sourceIP,
						Data:            &formatJSON,
					}
				}
			}

			//обработка сообщения от ядра
		case msg := <-chanOutCore:
			go handlers.HandlerMsgFromCore(cwt, isl, msg, smt, qts, saveMessageApp, chanInCore, chanInCRRF)
		}
	}
}

//RouteWssConnectionResponse маршрутизирует сообщения от источников
func RouteWssConnectionResponse(
	cwtRes chan<- configure.MsgWsTransmission,
	isl *configure.InformationSourcesList,
	smt *configure.StoringMemoryTask,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	chanInCRRF chan<- *handlers.MsgChannelReceivingFiles,
	cwtReq <-chan configure.MsgWsTransmission) {

	//MessageType содержит тип JSON сообщения
	type MessageType struct {
		Type string `json:"messageType"`
	}

	var messageType MessageType

	for msg := range cwtReq {
		sourceIP := msg.DestinationHost
		message := msg.Data

		sourceID, _ := isl.GetSourceIDOnIP(sourceIP)

		if err := json.Unmarshal(*message, &messageType); err != nil {
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
		}

		switch messageType.Type {
		case "pong":
			_ = saveMessageApp.LogMessage("info", fmt.Sprintf("resived message type 'PONG' from IP %v", sourceIP))

		case "telemetry":
			chanInCore <- &configure.MsgBetweenCoreAndNI{
				Section:         "source control",
				Command:         "telemetry",
				SourceID:        sourceID,
				AdvancedOptions: message,
			}

		case "filtration":
			pprmtf := processresponse.ParametersProcessingReceivedMsgTypeFiltering{
				CwtRes:         cwtRes,
				ChanInCore:     chanInCore,
				CwtReq:         cwtReq,
				Isl:            isl,
				Smt:            smt,
				Message:        message,
				SourceID:       sourceID,
				SourceIP:       sourceIP,
				SaveMessageApp: saveMessageApp,
			}

			go processresponse.ProcessingReceivedMsgTypeFiltering(pprmtf)

		case "download files":
			chanInCRRF <- &handlers.MsgChannelReceivingFiles{
				SourceID: sourceID,
				SourceIP: sourceIP,
				Command:  "taken from the source",
				Message:  message,
			}

		case "notification":
			var notify configure.MsgTypeNotification
			err := json.Unmarshal(*message, &notify)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			clientNotify := configure.MsgBetweenCoreAndNI{
				TaskID:   notify.Info.TaskID,
				Section:  "message notification",
				Command:  "send client API",
				SourceID: sourceID,
				AdvancedOptions: configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      notify.Info.Section,
					TypeActionPerformed:          notify.Info.TypeActionPerformed,
					CriticalityMessage:           notify.Info.CriticalityMessage,
					HumanDescriptionNotification: notify.Info.Description,
					Sources:                      []int{sourceID},
				},
			}

			chanInCore <- &clientNotify

		case "error":
			var errMsg configure.MsgTypeError
			err := json.Unmarshal(*message, &errMsg)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			errNotify := configure.MsgBetweenCoreAndNI{
				TaskID:   errMsg.Info.TaskID,
				Section:  "error notification",
				SourceID: sourceID,
				AdvancedOptions: configure.ErrorNotification{
					SourceReport:          "NI module",
					ErrorName:             errMsg.Info.ErrorName,
					HumanDescriptionError: errMsg.Info.ErrorDescription,
					Sources:               []int{sourceID},
				},
			}

			chanInCore <- &errNotify
		}
	}
}
