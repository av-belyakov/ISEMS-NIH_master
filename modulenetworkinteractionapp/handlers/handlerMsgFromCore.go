package handlers

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromCore обработчик сообщений от ядра приложения
func HandlerMsgFromCore(
	cwt chan<- configure.MsgWsTransmission,
	isl *configure.InformationSourcesList,
	msg *configure.MsgBetweenCoreAndNI,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	chanInCRRF chan<- *configure.MsgChannelReceivingFiles) {

	//максимальное количество одновременно запущеных процессов фильтрации
	var mcpf int8 = 3

	funcName := ", function 'HandlerMsgFromCore'"
	clientNotify := configure.MsgBetweenCoreAndNI{
		TaskID:  msg.TaskID,
		Section: "message notification",
		Command: "send client API",
	}

	switch msg.Section {
	case "source control":
		if msg.Command == "create list" {
			sl, ok := msg.AdvancedOptions.([]configure.InformationAboutSource)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error",
					FuncName:    funcName,
				})

				return
			}

			createSourceList(isl, sl)
		}

		if msg.Command == "load list" {
			ado, ok := msg.AdvancedOptions.(configure.SourceControlMsgTypeFromAPI)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error",
					FuncName:    funcName,
				})

				//снять отслеживание выполнения задачи
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "monitoring task performance",
					Command: "complete task",
				}

				return
			}

			//проверяем прислал ли пользователь данные по источникам
			if len(ado.SourceList) == 0 {
				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "source control",
					TypeActionPerformed:          "load list",
					CriticalityMessage:           "warning",
					HumanDescriptionNotification: "получен пустой список источников",
				}

				chanInCore <- &clientNotify

				//снять отслеживание выполнения задачи
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "monitoring task performance",
					Command: "complete task",
				}

				return
			}

			executedSourcesList, listInvalidSource := updateSourceList(isl, qts, ado.SourceList, msg.ClientName, mcpf)
			if len(listInvalidSource) != 0 {
				strSourceID := createStringFromSourceList(listInvalidSource)

				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "source control",
					TypeActionPerformed:          "load list",
					CriticalityMessage:           "warning",
					Sources:                      listInvalidSource,
					HumanDescriptionNotification: fmt.Sprintf("обновление списка источников выполнено не полностью, параметры источников: %v, содержат некорректные значения", strSourceID),
				}

				chanInCore <- &clientNotify
			} else {
				hdn := "обновление настроек источников выполнено успешно"
				cm := "success"

				if len(executedSourcesList) > 0 {
					strSourceID := createStringFromSourceList(executedSourcesList)
					hdn = fmt.Sprintf("на источниках: %v выполняются задачи, в настоящее время изменение их настроек невозможно", strSourceID)

					cm = "info"
				}

				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "source control",
					TypeActionPerformed:          "load list",
					CriticalityMessage:           cm,
					Sources:                      executedSourcesList,
					HumanDescriptionNotification: hdn,
				}

				chanInCore <- &clientNotify
			}

			lc, ld := isl.GetListsConnectedAndDisconnectedSources()
			lcd := []map[int]string{lc, ld}

			ts := make([]int, 0, (len(lc) + len(ld)))
			for _, item := range lcd {
				for id := range item {
					ts = append(ts, id)
				}
			}

			sltsdb, err := getSourceListToStoreDB(ts, &ado.SourceList, msg.ClientName, mcpf)
			if err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				//снять отслеживание выполнения задачи
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "monitoring task performance",
					Command: "complete task",
				}

				return
			}

			msgToCore := configure.MsgBetweenCoreAndNI{
				TaskID:          msg.TaskID,
				Section:         "source control",
				Command:         "keep list sources in database",
				AdvancedOptions: sltsdb,
			}

			//новый список источников для сохранения в БД
			chanInCore <- &msgToCore

			//снять отслеживание выполнения задачи
			chanInCore <- &configure.MsgBetweenCoreAndNI{
				TaskID:  msg.TaskID,
				Section: "monitoring task performance",
				Command: "complete task",
			}
		}

		if msg.Command == "perform actions on sources" {
			ado, ok := msg.AdvancedOptions.(configure.SourceControlMsgOptions)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error",
					FuncName:    funcName,
				})

				//снять отслеживание выполнения задачи
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "monitoring task performance",
					Command: "complete task",
				}

				return
			}

			//проверяем прислал ли пользователь данные по источникам
			if len(ado.MsgOptions.SourceList) == 0 {
				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "source control",
					TypeActionPerformed:          "load list",
					CriticalityMessage:           "warning",
					HumanDescriptionNotification: "получен пустой список источников",
				}

				chanInCore <- &clientNotify

				//снять отслеживание выполнения задачи
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "monitoring task performance",
					Command: "complete task",
				}

				return
			}

			listActionType, listInvalidSource, err := performActionSelectedSources(isl, qts, &ado.MsgOptions.SourceList, msg.ClientName, mcpf)
			if err != nil {
				strSourceID := createStringFromSourceList(*listInvalidSource)
				strSource := "источником"

				if len(*listInvalidSource) > 1 {
					strSource = "источниками"
				}

				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "source control",
					TypeActionPerformed:          "perform actions on sources",
					CriticalityMessage:           "warning",
					Sources:                      *listInvalidSource,
					HumanDescriptionNotification: fmt.Sprintf("невозможно выполнить действия над %v: %v, приняты некорректные значения", strSource, strSourceID),
				}

				chanInCore <- &clientNotify

				//снять отслеживание выполнения задачи
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "monitoring task performance",
					Command: "complete task",
				}
				return
			}

			// получаем ID источников по которым нужно обновить информацию
			// в БД, к ним относятся источники для которых выполненно действие
			// la - add, lu - update, ld - delete
			la, lu, ld := getSourceListsForWriteToDB(&ado.MsgOptions.SourceList, listActionType, msg.ClientName, mcpf)

			//актуализируем информацию в БД
			if len(*la) > 0 {
				//добавить
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:          msg.TaskID,
					Section:         "source control",
					Command:         "keep list sources in database",
					AdvancedOptions: la,
				}
			}

			if len(*ld) > 0 {
				//удалить
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:          msg.TaskID,
					Section:         "source control",
					Command:         "delete sources in database",
					AdvancedOptions: ld,
				}
			}

			if len(*lu) > 0 {
				//обновить
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:          msg.TaskID,
					Section:         "source control",
					Command:         "update sources in database",
					AdvancedOptions: lu,
				}
			}

			//отправляем сообщение пользователю
			chanInCore <- &configure.MsgBetweenCoreAndNI{
				TaskID:          msg.TaskID,
				Section:         "source control",
				Command:         "confirm the action",
				AdvancedOptions: listActionType,
			}

			//снимаем отслеживание выполнения задачи
			chanInCore <- &configure.MsgBetweenCoreAndNI{
				TaskID:  msg.TaskID,
				Section: "monitoring task performance",
				Command: "complete task",
			}
		}

		if msg.Command == "get telemetry" {
			csl, ok := msg.AdvancedOptions.(map[int]*configure.DetailedSourceInformation)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error",
					FuncName:    funcName,
				})

				return
			}

			reqTelemetry := configure.MsgTypeTelemetryControl{
				MsgType: "telemetry",
				Info: configure.SettingsTelemetryControlRequest{
					TaskID:  msg.TaskID,
					Command: "give me telemetry",
				},
			}

			msgJSON, err := json.Marshal(reqTelemetry)
			if err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				return
			}

			for sourceID := range csl {
				//проверяем наличие подключения для заданного источника
				if si, ok := isl.GetSourceSetting(sourceID); ok {
					//передаем задачу источнику
					cwt <- configure.MsgWsTransmission{
						DestinationHost: si.IP,
						Data:            &msgJSON,
					}
				}
			}
		}

	case "filtration control":
		//проверяем наличие подключения для заданного источника
		si, ok := isl.GetSourceSetting(msg.SourceID)
		if !ok {
			//отправляем сообщение пользователю
			clientNotify.AdvancedOptions = configure.MessageNotification{
				SourceReport:                 "NI module",
				Section:                      "filtration control",
				TypeActionPerformed:          "start",
				CriticalityMessage:           "warning",
				HumanDescriptionNotification: "источник не найден",
			}

			chanInCore <- &clientNotify

			return
		}

		if msg.Command == "start" {
			if !si.ConnectionStatus {
				if ti, ok := smt.GetStoringMemoryTask(msg.TaskID); ok {
					//останавливаем передачу списка файлов (найденных в результате поиска по индексам)
					ti.ChanStopTransferListFiles <- struct{}{}
				}

				//отправляем сообщение пользователю
				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "filtration control",
					TypeActionPerformed:          "start",
					CriticalityMessage:           "warning",
					HumanDescriptionNotification: "не возможно отправить запрос для выполнения задачи, источник не подключен",
				}

				chanInCore <- &clientNotify

				//обновляем информацию о задаче фильтрации в памяти приложения
				smt.UpdateTaskFiltrationAllParameters(msg.TaskID, &configure.FiltrationTaskParameters{Status: "refused"})

				//отправляем сообщение в БД информирующее о необходимости записи новых параметров
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:   msg.TaskID,
					Section:  "filtration control",
					Command:  "update",
					SourceID: msg.SourceID,
				}

				//снимаем отслеживание выполнения задачи
				chanInCore <- &configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "monitoring task performance",
					Command: "complete task",
				}

				return
			}

			msgJSON, ok := msg.AdvancedOptions.([]byte)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error",
					FuncName:    funcName,
				})

				return
			}

			//передаем задачу источнику
			cwt <- configure.MsgWsTransmission{
				DestinationHost: si.IP,
				Data:            &msgJSON,
			}
		}

		if msg.Command == "stop" {
			//проверяем наличие подключения для заданного источника
			if !si.ConnectionStatus {
				//отправляем сообщение пользователю
				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "filtration control",
					TypeActionPerformed:          "stop",
					CriticalityMessage:           "warning",
					HumanDescriptionNotification: "не возможно отправить запрос на останов задачи, источник не подключен",
				}

				chanInCore <- &clientNotify

				return
			}

			reqTypeStop := configure.MsgTypeFiltrationControl{
				MsgType: "filtration",
				Info: configure.SettingsFiltrationControl{
					TaskID:  msg.TaskID,
					Command: "stop",
				},
			}

			msgJSON, err := json.Marshal(reqTypeStop)
			if err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				return
			}

			//отправляем источнику сообщение типа 'confirm complete' для того что бы подтвердить останов задачи
			cwt <- configure.MsgWsTransmission{
				DestinationHost: si.IP,
				Data:            &msgJSON,
			}
		}

	case "download control":
		if msg.Command == "start" {
			chanInCRRF <- &configure.MsgChannelReceivingFiles{
				SourceID: msg.SourceID,
				TaskID:   msg.TaskID,
				Command:  "give my the files",
			}
		}

		if msg.Command == "stop" {
			chanInCRRF <- &configure.MsgChannelReceivingFiles{
				SourceID: msg.SourceID,
				TaskID:   msg.TaskID,
				Command:  "stop receiving files",
			}
		}
	}
}
