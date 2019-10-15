package handlerslist

import (
	"encoding/json"
	"fmt"
	"time"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromAPI обработчик сообщений приходящих от модуля API
func HandlerMsgFromAPI(
	outCoreChans HandlerOutChans,
	msg *configure.MsgBetweenCoreAndAPI,
	hsm HandlersStoringMemory,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) {

	funcName := ", function 'HeaderMsgFromAPI'"

	msgc := configure.MsgCommon{}

	nsErrJSON := notifications.NotificationSettingsToClientAPI{
		MsgType:        "danger",
		MsgDescription: "Ошибка, получен некорректный формат JSON сообщения",
	}

	msgJSON, ok := msg.MsgJSON.([]byte)
	if !ok {
		notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
		_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

		return
	}

	if err := json.Unmarshal(msgJSON, &msgc); err != nil {
		notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
		_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

		return
	}

	//логируем запросы клиентов
	_ = saveMessageApp.LogMessage("requests", fmt.Sprintf("client name: '%v' (%v), request: type = %v, section = %v, instruction = %v, client task ID = %v", msg.ClientName, msg.ClientIP, msgc.MsgType, msgc.MsgSection, msgc.MsgInstruction, msgc.ClientTaskID))

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInstruction == "send new source list" {
				var scmo configure.SourceControlMsgOptions

				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				taskID := common.GetUniqIDFormatMD5(msg.IDClientAPI)

				//добавляем новую задачу
				hsm.SMT.AddStoringMemoryTask(taskID, configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					ClientName:      msg.ClientName,
					Section:         "source control",
					Command:         "load list",
					AdvancedOptions: scmo.MsgOptions,
				}

				return
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
			_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgSection'"+funcName)

			return
		}

		notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
		_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgSection'"+funcName)

		return
	}

	if msgc.MsgType == "command" {
		switch msgc.MsgSection {

		// УПРАВЛЕНИЕ ИСТОЧНИКАМИ
		case "source control":
			//получить актуальный список источников
			if msgc.MsgInstruction == "get an updated list of sources" {
				taskID := common.GetUniqIDFormatMD5(msg.IDClientAPI)

				//добавляем новую задачу
				hsm.SMT.AddStoringMemoryTask(taskID, configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
					MsgGenerator: "API module",
					MsgRecipient: "DB module",
					MsgSection:   "source control",
					Instruction:  "find_all",
					TaskID:       taskID,
				}

				return
			}

			//выполнить какие либо действия над источниками
			if msgc.MsgInstruction == "performing an action" {
				var scmo configure.SourceControlMsgOptions
				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				taskID := common.GetUniqIDFormatMD5(msg.IDClientAPI)

				//добавляем новую задачу
				hsm.SMT.AddStoringMemoryTask(taskID, configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					ClientName:      msg.ClientName,
					Section:         "source control",
					Command:         "perform actions on sources",
					AdvancedOptions: scmo,
				}

				fmt.Printf("---- func 'handlerMsgFromAPI' -----\n%v\n", scmo)

				return
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
			_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgInstruction'"+funcName)

			return

		// УПРАВЛЕНИЕ ФИЛЬТРАЦИЕЙ
		case "filtration control":
			//обработка команды на запуск фильтрации
			if msgc.MsgInstruction == "to start filtering" {
				var fcts configure.FiltrationControlTypeStart
				if err := json.Unmarshal(msgJSON, &fcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				go handlerFiltrationControlTypeStart(&fcts, hsm, msg.IDClientAPI, saveMessageApp, outCoreChans.OutCoreChanAPI)

				return
			}

			//команда на останов фильтрации
			if msgc.MsgInstruction == "to cancel filtering" {
				//ищем ожидающую в очереди задачу по ее ID
				sourceID, taskID, err := hsm.QTS.SearchTaskForClientIDQueueTaskStorage(msgc.ClientTaskID)
				if err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					nsErr := notifications.NotificationSettingsToClientAPI{
						MsgType:        "danger",
						MsgDescription: fmt.Sprintf("Ошибка, по переданному идентификатору '%v' ожидающих или выполняемых задач не обнаружено", msgc.ClientTaskID),
					}
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErr, msgc.ClientTaskID, msg.IDClientAPI)

					return
				}

				//проверяем наличие задачи в StoringMemoryTask
				_, isExist := hsm.SMT.GetStoringMemoryTask(taskID)
				if !isExist {

					return
				}

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:     taskID,
					ClientName: msg.ClientName,
					Section:    "filtration control",
					Command:    "stop",
					SourceID:   sourceID,
				}

				return
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
			_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgInstruction'"+funcName)

			return

		// УПРАВЛЕНИЕ ВЫГРУЗКОЙ ФАЙЛОВ
		case "download control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'download control'")

			//отправляем сообщение о том что задача была отклонена
			resMsgRefused := configure.DownloadControlTypeInfo{
				MsgOption: configure.DownloadControlMsgTypeInfo{
					Status: "refused",
				},
			}
			resMsgRefused.MsgType = "information"
			resMsgRefused.MsgSection = "download control"
			resMsgRefused.MsgInstruction = "task processing"

			if msgc.MsgInstruction == "to start downloading" {
				fmt.Println("START task 'DOWNLOADING'")

				var dcts configure.DownloadControlTypeStart

				if err := json.Unmarshal(msgJSON, &dcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				resMsgRefused.ClientTaskID = dcts.ClientTaskID
				resMsgRefused.MsgOption.ID = dcts.MsgOption.ID
				resMsgRefused.MsgOption.TaskIDApp = dcts.MsgOption.TaskIDApp
				msgJSONRefused, err := json.Marshal(resMsgRefused)
				if err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					return
				}

				msgToAPI := configure.MsgBetweenCoreAndAPI{
					MsgGenerator: "Core module",
					MsgRecipient: "API module",
					IDClientAPI:  msg.IDClientAPI,
					MsgJSON:      msgJSONRefused,
				}

				//ищем источник по указанному идентификатору
				sourceInfo, ok := hsm.ISL.GetSourceSetting(dcts.MsgOption.ID)
				if !ok {
					nsErrJSON.MsgDescription = fmt.Sprintf("Ошибка, источника %v не существует", dcts.MsgOption.ID)
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", fmt.Sprintf("source ID %v was not found%v", dcts.MsgOption.ID, funcName))

					//сообщение о том что задача была отклонена
					outCoreChans.OutCoreChanAPI <- &msgToAPI

					return
				}

				//проверяем подключение источника
				if !sourceInfo.ConnectionStatus {
					nsErrJSON.MsgDescription = fmt.Sprintf("Ошибка, источник %v не подключен", dcts.MsgOption.ID)
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", fmt.Sprintf("source ID %v is not connected%v", dcts.MsgOption.ID, funcName))

					//сообщение о том что задача была отклонена
					outCoreChans.OutCoreChanAPI <- &msgToAPI

					return
				}

				//проверяем наличие в очереди задачи с указанным ID
				_, ti, err := hsm.QTS.SearchTaskForIDQueueTaskStorage(dcts.MsgOption.TaskIDApp)
				if err == nil {
					var errMsg string

					if ti.TaskStatus == "wait" {
						errMsg = fmt.Sprintf("Unable to add task with ID '%v' because it is already pending", dcts.MsgOption.TaskIDApp)
						nsErrJSON.MsgDescription = fmt.Sprintf("Невозможно добавить задачу с ID '%v' так как она уже ожидает выполнения", dcts.MsgOption.TaskIDApp)
					} else if ti.TaskStatus == "execution" {
						errMsg = fmt.Sprintf("You cannot add a task with ID '%v' to a source with ID %v because it is already running", dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID)
						nsErrJSON.MsgDescription = fmt.Sprintf("Невозможно добавить задачу с ID '%v', для источника с ID %v, так как она уже выполняется", dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID)
					} else {
						errMsg = fmt.Sprintf("Unable to add task with ID '%v'. The task has been completed, but has not yet been removed from the pending task list", dcts.MsgOption.ID)
						nsErrJSON.MsgDescription = fmt.Sprintf("Невозможно добавить задачу с ID '%v'. Задача была выполнена, однако из списка задач ожидающих выполнения пока не удалена", dcts.MsgOption.ID)
					}

					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", errMsg)

					//сообщение о том что задача была отклонена
					outCoreChans.OutCoreChanAPI <- &msgToAPI

					return
				}

				//добавляем задачу в очередь
				hsm.QTS.AddQueueTaskStorage(dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID, configure.CommonTaskInfo{
					IDClientAPI:     msg.IDClientAPI,
					TaskIDClientAPI: dcts.ClientTaskID,
					TaskType:        "download control",
				}, &configure.DescriptionParametersReceivedFromUser{
					DownloadList: dcts.MsgOption.FileList,
				})

				//устанавливаем проверочный статус источника для данной задачи как подключен
				if err := hsm.QTS.ChangeAvailabilityConnectionOnConnection(dcts.MsgOption.ID, dcts.MsgOption.TaskIDApp); err != nil {
					nsErrJSON.MsgDescription = fmt.Sprintf("Ошибка, запись для источника %v отсутствует в памяти приложения", dcts.MsgOption.ID)
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					//сообщение о том что задача была отклонена
					outCoreChans.OutCoreChanAPI <- &msgToAPI

					return
				}

				//запрос в БД о наличии задачи с указанным ID и файлов для скачивания
				outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
					MsgGenerator:    "Core module",
					MsgRecipient:    "DB module",
					MsgSection:      "download control",
					Instruction:     "finding information about a task",
					IDClientAPI:     msg.IDClientAPI,
					TaskID:          dcts.MsgOption.TaskIDApp,
					TaskIDClientAPI: dcts.ClientTaskID,
				}
			}

			if msgc.MsgInstruction == "to cancel downloading" {
				fmt.Println("STOP task 'DOWNLOADING'")

				var dcts configure.DownloadControlTypeStart

				if err := json.Unmarshal(msgJSON, &dcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				/*
					   Проверить, выполняется ли задача с указанным ID

					   	//сообщение о том что задача была отклонена
						outCoreChans.OutCoreChanAPI <- &msgToAPI
				*/

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:     dcts.MsgOption.TaskIDApp,
					ClientName: msg.ClientName,
					Section:    "download control",
					Command:    "stop receiving files",
					SourceID:   dcts.MsgOption.ID,
				}
			}

			return

		// УПРАВЛЕНИЕ ПОИСКОМ ИНФОРМАЦИИ В БД ПРИЛОЖЕНИЯ
		case "information search control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'information search control'")

			return

		default:
			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
			_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgInstruction'"+funcName)

			return
		}
	}
}
