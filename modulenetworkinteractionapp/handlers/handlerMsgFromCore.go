package handlers

/*
* Обработчик запросов от ядра приложения
*
* Версия 0.2, дата релиза 21.03.2019
* */

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

	funcName := ", function 'HandlerMsgFromCore'"

	//максимальное количество одновременно запущеных процессов фильтрации
	var mcpf int8 = 3

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
				_ = saveMessageApp.LogMessage("error", "NI module - type conversion error"+funcName)

				return
			}

			createSourceList(isl, sl)
		}

		if msg.Command == "load list" {
			ado, ok := msg.AdvancedOptions.(configure.SourceControlMsgTypeFromAPI)
			if !ok {
				_ = saveMessageApp.LogMessage("error", "NI module - type conversion error"+funcName)

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
					HumanDescriptionNotification: "Получен пустой список сенсоров",
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
					HumanDescriptionNotification: "Обновление списка сенсоров выполнено не полностью, параметры сенсоров: " + strSourceID + " содержат некорректные значения",
				}

				chanInCore <- &clientNotify
			} else {
				hdn := "Обновление настроек сенсоров выполнено успешно"
				cm := "success"

				if len(executedSourcesList) > 0 {
					strSourceID := createStringFromSourceList(executedSourcesList)
					hdn = "На сенсорах: " + strSourceID + " выполняются задачи, в настоящее время изменение их настроек невозможно"
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
				_ = saveMessageApp.LogMessage("error", "NI module - "+fmt.Sprint(err))

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
				_ = saveMessageApp.LogMessage("error", "NI module - type conversion error"+funcName)

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
					HumanDescriptionNotification: "Получен пустой список сенсоров",
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

				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "source control",
					TypeActionPerformed:          "perform actions on sources",
					CriticalityMessage:           "warning",
					Sources:                      *listInvalidSource,
					HumanDescriptionNotification: "Невозможно выполнить действия над сенсорами:" + strSourceID + ", приняты некорректные значения",
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

	case "filtration control":
		if msg.Command == "start" {

			fmt.Println("function 'HandlerMsgFromCore', section - 'filtration control', command - 'START'")

			//проверяем наличие подключения для заданного источника
			si, ok := isl.GetSourceSetting(msg.SourceID)
			if !ok || !si.ConnectionStatus {
				humanNotify := fmt.Sprintf("Не возможно отправить запрос на фильтрацию, источник с ID %v не подключен", msg.SourceID)
				if !ok {
					humanNotify = fmt.Sprintf("Источник с ID %v не найден", msg.SourceID)
				}

				if ti, ok := smt.GetStoringMemoryTask(msg.TaskID); !ok {
					//останавливаем передачу списка файлов (найденных в результате поиска по индексам)
					ti.ChanStopTransferListFiles <- struct{}{}
				}

				//отправляем сообщение пользователю
				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "filtration control",
					TypeActionPerformed:          "start",
					CriticalityMessage:           "warning",
					HumanDescriptionNotification: humanNotify,
				}

				chanInCore <- &clientNotify

				//обновляем информацию о задаче фильтрации в памяти приложения
				smt.UpdateTaskFiltrationAllParameters(msg.TaskID, configure.FiltrationTaskParameters{Status: "refused"})

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
				_ = saveMessageApp.LogMessage("error", "NI module - type conversion error"+funcName)

				return
			}

			fmt.Println("function 'HandlerMsgFromCore', section - 'filtration control', send task 'START' to source")

			//передаем задачу источнику
			cwt <- configure.MsgWsTransmission{
				DestinationHost: si.IP,
				Data:            &msgJSON,
			}

			fmt.Println("\tЗадача по фильтрации сет. трафика ушла на источник...")
		}

		if msg.Command == "stop" {

			fmt.Println("function 'HandlerMsgFromCore', section - 'filtration control', command - 'STOP'")

			if si, ok := isl.GetSourceSetting(msg.SourceID); ok {
				//проверяем наличие подключения для заданного источника
				if !si.ConnectionStatus {
					//отправляем сообщение пользователю
					clientNotify.AdvancedOptions = configure.MessageNotification{
						SourceReport:                 "NI module",
						Section:                      "filtration control",
						TypeActionPerformed:          "stop",
						CriticalityMessage:           "warning",
						HumanDescriptionNotification: fmt.Sprintf("Не возможно отправить запрос на фильтрацию, источник с ID %v не подключен", msg.SourceID),
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
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					return
				}

				fmt.Println("function 'HandlerMsgFromCore', section - 'filtration control', send task 'STOP' to source")

				//отправляем источнику сообщение типа 'confirm complete' для того что бы подтвердить останов задачи
				cwt <- configure.MsgWsTransmission{
					DestinationHost: si.IP,
					Data:            &msgJSON,
				}
			}
		}

	case "download control":
		if msg.Command == "start" {

			fmt.Println("function 'HandlerMsgFromCore', section - 'download control', send task 'START' to source")

			chanInCRRF <- &configure.MsgChannelReceivingFiles{
				SourceID: msg.SourceID,
				TaskID:   msg.TaskID,
				Command:  "give my the files",
			}
		}

		if msg.Command == "stop" {

			fmt.Println("function 'HandlerMsgFromCore', section - 'download control', send task 'STOP' to source")

			chanInCRRF <- &configure.MsgChannelReceivingFiles{
				SourceID: msg.SourceID,
				TaskID:   msg.TaskID,
				Command:  "stop receiving files",
			}
		}

	}
}
