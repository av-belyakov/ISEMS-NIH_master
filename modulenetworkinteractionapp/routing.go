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

		sourceID, _ := isl.GetSourceIDOnIP(sourceIP)

		if err := json.Unmarshal(*message, &messageType); err != nil {
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
		}

		switch messageType.Type {
		case "pong":

			fmt.Println("RESIVED message type 'PONG' from IP", sourceIP)

		case "telemetry":
			fmt.Println("RESIVED message type 'TELEMETRY' from IP", sourceIP)
			fmt.Printf("%v\n", messageType)

			chanInCore <- &configure.MsgBetweenCoreAndNI{
				Section:         "source control",
				Command:         "telemetry",
				SourceID:        sourceID,
				AdvancedOptions: message,
			}

		case "filtration":
			/*

				!!! ВНИМАНИЕ !!!
				При получении сообщения о завершении фильтрации, 'stop' или 'complite'
				и сборке всего списка файлов полученного в результате фильтрации,
				то есть обработать параметр 'NumberMessagesFrom' [2]int
				ОТПРАВИТЬ сообщение "confirm complite" что бы удалить задачу на
				стороне ISEMS-NIH_slave

					//конвертирование принятой информации из формата JSON

									//запись информации о ходе выполнения задачи в память
									smt.UpdateTaskFiltrationAllParameters(taskID, configure.FiltrationTaskParameters{
										ID:                       190,
										Status:                   "execute",
										NumberFilesToBeFiltered:  231,
										SizeFilesToBeFiltered:    4738959669055,
										CountDirectoryFiltartion: 3,
										NumberFilesProcessed:     12,
										NumberFilesFound:         3,
										SizeFilesFound:           0,
										PathStorageSource:        "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
										FoundFilesInformation: map[string]*configure.FoundFilesInformation{
											"1438535410_2015_08_02____20_10_10_644263.tdp": &configure.FoundFilesInformation{
												Size: 1448375,
												Hex:  "fj9j939j9t88232",
											},
										},
									})

									//формируем из ответа от ISEMS_NIH_slave
								dif := configure.DetailedInformationFiltering{
									TaskStatus                      string
									TimeIntervalTaskExecution       TimeInterval
									WasIndexUsed                    bool
									NumberFilesMeetFilterParameters int
									NumberProcessedFiles            int
									NumberFilesFoundResultFiltering int
									NumberDirectoryFiltartion       int
									SizeFilesMeetFilterParameters   int64
									SizeFilesFoundResultFiltering   int64
									PathDirectoryForFilteredFiles   string
									ListFilesFoundResultFiltering   []*InformationFilesFoundResultFiltering
								}

									//отправка инофрмации на запись в БД
									chanInCore<-&configure.MsgBetweenCoreAndDB{
										Section:         "filtration",
									Command:         "update",
									SourceID:        sourceID,
									TaskID: "получаем из ответа ISEMS_NIH_slave",
									AdvancedOptions: dif,
									}
			*/

		case "download files":

		case "notification":
			fmt.Println("--- RESIVED MESSAGE TYPE 'NOTIFICATION' ---")
			fmt.Println("------------------------------------")

			var notify configure.MsgTypeNotification
			err := json.Unmarshal(*message, &notify)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			clientNotify := &configure.MsgBetweenCoreAndNI{
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

			chanInCore <- clientNotify

		case "error":
			fmt.Println("--- RESIVED MESSAGE TYPE 'ERROR' ---")
			fmt.Println("------------------------------------")

			var errMsg configure.MsgTypeError
			err := json.Unmarshal(*message, errMsg)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			errNotify := &configure.MsgBetweenCoreAndNI{
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

			chanInCore <- errNotify
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
