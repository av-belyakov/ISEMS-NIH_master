package handlers

/*
* Обработчик запросов от ядра приложения
*
* Версия 0.2, дата релиза 21.03.2019
* */

import (
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
	chanInCore chan<- *configure.MsgBetweenCoreAndNI) {

	fmt.Println("START func HandlerMsgFromCore... (NI module)")
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
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

			//fmt.Println("====== CREATE LIST RESIVED FROM DB =======")

			sl, ok := msg.AdvancedOptions.([]configure.InformationAboutSource)
			if !ok {
				_ = saveMessageApp.LogMessage("error", "NI module - type conversion error"+funcName)

				return
			}

			createSourceList(isl, sl)

			//fmt.Printf("curent list %v \n=======================\n", isl.GetSourceList())
		}

		if msg.Command == "load list" {

			//fmt.Println("====== CREATE LIST RESIVED FROM CLIENT API =======", msg.ClientName, "====")

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

			executedSourcesList, listInvalidSource := updateSourceList(isl, ado.SourceList, msg.ClientName, mcpf)
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
			fmt.Println("====== PERFOM ACTIONS ON SOURCES RESIVED FROM CLIENT API =======", msg.ClientName, "====")

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

			//fmt.Printf("BEFORE update MEMORY: %v\n", isl.GetSourceList())

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

			listActionType, listInvalidSource, err := performActionSelectedSources(isl, &ado.MsgOptions.SourceList, msg.ClientName, mcpf)
			if err != nil {
				strSourceID := createStringFromSourceList(*listInvalidSource)

				clientNotify.AdvancedOptions = configure.MessageNotification{
					SourceReport:                 "NI module",
					Section:                      "source control",
					TypeActionPerformed:          "perform actions on sources",
					CriticalityMessage:           "warning",
					Sources:                      *listInvalidSource,
					HumanDescriptionNotification: "невозможно выполнить действия над сенсорами:" + strSourceID + ", приняты некорректные значения",
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

			//fmt.Printf("\nLIST SOURCES IN MEMORY\n%v\n", isl.GetSourceList())

			//fmt.Println("List Action Type", listActionType)

			// получаем ID источников по которым нужно обновить информацию
			// в БД, к ним относятся источники для которых выполненно действие
			// la - add, lu - update, ld - delete
			la, lu, ld := getSourceListsForWriteToBD(&ado.MsgOptions.SourceList, listActionType, msg.ClientName, mcpf)

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

			fmt.Printf("|-|-|-|-|-| RESIVED MESSAGE 'filtration control', 'START'\n%v\n", msg)

			//проверяем наличие подключения для заданного источника
			if si, ok := isl.GetSourceSetting(msg.SourceID); ok {
				if !si.ConnectionStatus {
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
						HumanDescriptionNotification: fmt.Sprintf("Не возможно отправить запрос на фильтрацию, источник с ID %v не подключен", msg.SourceID),
					}

					chanInCore <- &clientNotify

					//обнавляем информацию о задаче фильтрации в памяти приложения
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

				//передаем задачу источнику
				cwt <- configure.MsgWsTransmission{
					DestinationHost: si.IP,
					Data:            &msgJSON,
				}
			}
		}

		if msg.Command == "stop" {
			/*
				//снять отслеживание выполнения задачи
					chanInCore <- &configure.MsgBetweenCoreAndNI{
						TaskID:  msg.TaskID,
						Section: "monitoring task performance",
						Command: "complete task",
					}
			*/
		}

	case "download control":
		if msg.Command == "start" {

		}

		if msg.Command == "stop" {
			/*
				//снять отслеживание выполнения задачи
					chanInCore <- &configure.MsgBetweenCoreAndNI{
						TaskID:  msg.TaskID,
						Section: "monitoring task performance",
						Command: "complete task",
					}
			*/
		}

	}
}
