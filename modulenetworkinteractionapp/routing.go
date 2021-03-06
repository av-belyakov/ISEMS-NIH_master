package modulenetworkinteractionapp

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/modulenetworkinteractionapp/handlers"
	"ISEMS-NIH_master/savemessageapp"
)

//sendNIStopTask сообщение NI module с целью остановить выполнение задачи
// из-за разрыва соединения
func sendNIStopTask(
	chanInCRRF chan<- *configure.MsgChannelReceivingFiles,
	sourceID int,
	qts *configure.QueueTaskStorage) {

	tasks, ok := qts.GetAllTaskQueueTaskStorage(sourceID)
	if !ok {
		return
	}

	msgStop := configure.MsgChannelReceivingFiles{
		SourceID: sourceID,
		Command:  "to stop the task because of a disconnection",
	}

	for tid, taskInfo := range tasks {
		msgStop.TaskID = tid
		qts.ChangeAvailabilityConnectionOnDisconnection(sourceID, tid)

		if taskInfo.TaskStatus == "execution" {
			chanInCRRF <- &msgStop
		}
	}
}

//checkIfThereTaskForSource проверяет есть ли в очереди задачи для данного источника
func checkIfThereTaskForSource(sourceID int, qts *configure.QueueTaskStorage) {
	tasks, ok := qts.GetAllTaskQueueTaskStorage(sourceID)
	if !ok {
		return
	}

	for tid := range tasks {
		qts.ChangeAvailabilityConnectionOnConnection(sourceID, tid)
	}
}

//RouteCoreRequest маршрутизирует запросы от CoreApp и обрабатывает сообщения от wss модулей
// cwt - канал для передачи данных источникам
// chanInCore - канал для взаимодействия с Ядром приложения (ИСХОДЯЩИЙ)
// chanInCRRF - канал для взаимодействия с контроллером запрошенных принимаемых файлов
// isl - информация по источникам
// smt - информация по выполняемым задачам
// qts - информация об ожидающих и выполняющихся задачах
// saveMessageApp - объект для записи логов
// chanColl - коллекция каналов для взаимодействия с WssServer и WssClient
// chanOutCore - канал для взаимодействия с Ядром приложения (ВХОДЯЩИЙ)
func RouteCoreRequest(
	cwt chan<- configure.MsgWsTransmission,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	chanInCRRF chan<- *configure.MsgChannelReceivingFiles,
	isl *configure.InformationSourcesList,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanColl map[string]chan [2]string,
	chanOutCore <-chan *configure.MsgBetweenCoreAndNI) {

	fn := "RouteCoreRequest"

	for {
		select {
		/*
			обработка сообщений от МОДУЛЕЙ WSS (информация об источниках)
		*/
		//модуль wssServer
		case msg := <-chanColl["outWssModuleServer"]:
			sourceIP, action := msg[0], msg[1]

			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				TypeMessage: "info",
				Description: fmt.Sprintf("SERVER: source with IP %v has success %v", sourceIP, action),
				FuncName:    fn,
			})

			sourceID, ok := isl.GetSourceIDOnIP(sourceIP)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprintf("it is impossible to find the ID of the source ip address of %v", sourceIP),
					FuncName:    fn,
				})

				break
			}

			chanInCore <- &configure.MsgBetweenCoreAndNI{
				Section: "source control",
				Command: "change connection status source",
				AdvancedOptions: configure.SettingsChangeConnectionStatusSource{
					ID:     sourceID,
					Status: action,
				},
			}

			//остановить скачивание файлов если соединение с источником было разорвано
			if action == "disconnect" {
				sendNIStopTask(chanInCRRF, sourceID, qts)
			}

			if action == "connect" {
				err := handlers.SendPing(sourceIP, sourceID, isl, cwt)
				if err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    fn,
					})

					continue
				}

				//проверяем есть ли в очереди задачи для данного источника
				checkIfThereTaskForSource(sourceID, qts)

				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("SERVER: send msg type PING source %v", sourceID),
					FuncName:    fn,
				})
			}

		//модуль wssClient
		case msg := <-chanColl["outWssModuleClient"]:
			sourceIP, action := msg[0], msg[1]

			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				TypeMessage: "info",
				Description: fmt.Sprintf("CLIENT: source with IP %v has success %v", sourceIP, action),
				FuncName:    fn,
			})

			sourceID, ok := isl.GetSourceIDOnIP(sourceIP)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprintf("it is impossible to find the ID of the source ip address of %v", sourceIP),
					FuncName:    fn,
				})

				break
			}

			chanInCore <- &configure.MsgBetweenCoreAndNI{
				Section: "source control",
				Command: "change connection status source",
				AdvancedOptions: configure.SettingsChangeConnectionStatusSource{
					ID:     sourceID,
					Status: action,
				},
			}

			//остановить скачивание файлов если соединение с источником было разорвано
			if action == "disconnect" {
				sendNIStopTask(chanInCRRF, sourceID, qts)
			}

			if action == "connect" {
				err := handlers.SendPing(sourceIP, sourceID, isl, cwt)
				if err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    fn,
					})

					continue
				}

				//проверяем есть ли в очереди задачи для данного источника
				checkIfThereTaskForSource(sourceID, qts)

				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("CLIENT: send msg type PING source %v", sourceID),
					FuncName:    fn,
				})
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
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	chanInCRRF chan<- *configure.MsgChannelReceivingFiles,
	chanInHMRFF chan<- *handlers.ChanHandlerMessagesReceivedFilesFiltering,
	isl *configure.InformationSourcesList,
	smt *configure.StoringMemoryTask,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	cwtReq <-chan configure.MsgWsTransmission) {

	//MessageType содержит тип JSON сообщения
	type MessageType struct {
		Type string `json:"messageType"`
	}

	var messageType MessageType
	fn := "RouteWssConnectionResponse"

	for msg := range cwtReq {
		sourceIP := msg.DestinationHost
		message := msg.Data

		sourceID, ok := isl.GetSourceIDOnIP(sourceIP)
		if !ok {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprintf("not found the ID of the source ip address %v", sourceIP),
				FuncName:    fn,
			})
		}

		if msg.MsgType == 1 {
			/* обработка текстовых данных */

			if err := json.Unmarshal(*message, &messageType); err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    fn,
				})
			}

			switch messageType.Type {
			case "pong":
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					Section:         "source control",
					Command:         "received version app",
					SourceID:        sourceID,
					AdvancedOptions: message,
				}

				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("resived message type 'PONG' from IP %v", sourceIP),
					FuncName:    fn,
				})

			case "telemetry":
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					Section:         "source control",
					Command:         "telemetry",
					SourceID:        sourceID,
					AdvancedOptions: message,
				}

			case "filtration":
				chanInHMRFF <- &handlers.ChanHandlerMessagesReceivedFilesFiltering{
					SourceID: sourceID,
					SourceIP: sourceIP,
					Message:  message,
				}

			case "download files":
				var mtd configure.MsgTypeDownload
				if err := json.Unmarshal(*message, &mtd); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    fn,
					})

					return
				}

				chanInCRRF <- &configure.MsgChannelReceivingFiles{
					SourceID: sourceID,
					SourceIP: sourceIP,
					TaskID:   mtd.Info.TaskID,
					Command:  "taken from the source",
					MsgType:  msg.MsgType,
					Message:  message,
				}

				mtd = configure.MsgTypeDownload{}

			case "notification":
				var notify configure.MsgTypeNotification
				err := json.Unmarshal(*message, &notify)
				if err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    fn,
					})

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
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    fn,
					})

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
		} else if msg.MsgType == 2 {
			/* обработка бинарных данных */

			//определяем принадлежность пакета
			checkBytes := (*message)[:1]

			if string(checkBytes) == "1" {
				/*
					raw файл (сет. трафик)
				*/

				chanInCRRF <- &configure.MsgChannelReceivingFiles{
					SourceID: sourceID,
					SourceIP: sourceIP,
					TaskID:   string((*message)[2:34]),
					Command:  "taken from the source",
					MsgType:  msg.MsgType,
					Message:  message,
				}

			} else if string(checkBytes) == "2" {
				/*
					tar.gz архив (JSON файл с индексами)
				*/

			} else {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprintf("unknown format of data received from source with ID %v (ip %v)", sourceID, sourceIP),
					FuncName:    fn,
				})
			}

		} else {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprintf("unknown data type received from source with ID %v (ip %v)", sourceID, sourceIP),
				FuncName:    fn,
			})
		}
	}
}
