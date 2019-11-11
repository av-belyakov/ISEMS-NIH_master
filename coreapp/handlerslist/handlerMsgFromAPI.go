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

				fmt.Printf("func 'handlerMsgFromAPI' - ищем ожидающую в очереди задачу по ее ID: %v\n", msgc.ClientTaskID)

				sourceID, taskID, err := hsm.QTS.SearchTaskForClientIDQueueTaskStorage(msgc.ClientTaskID)
				if err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					fmt.Printf("ERROR :::: Ошибка, по переданному идентификатору '%v' ожидающих или выполняемых задач не обнаружено\n", msgc.ClientTaskID)

					nsErr := notifications.NotificationSettingsToClientAPI{
						MsgType:        "danger",
						MsgDescription: fmt.Sprintf("Ошибка, по переданному идентификатору '%v' ожидающих или выполняемых задач не обнаружено", msgc.ClientTaskID),
					}
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErr, msgc.ClientTaskID, msg.IDClientAPI)

					return
				}

				//проверяем наличие задачи в StoringMemoryTask
				isExist := hsm.SMT.CheckIsExistMemoryTask(taskID)
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

			emt := ErrorMessageType{
				IDClientAPI: msg.IDClientAPI,
				Section:     "download control",
				Instruction: "task processing",
				MsgType:     "danger",
				ChanToAPI:   outCoreChans.OutCoreChanAPI,
			}

			if msgc.MsgInstruction == "to start downloading" {
				var dcts configure.DownloadControlTypeStart

				if err := json.Unmarshal(msgJSON, &dcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				emt.TaskID = dcts.MsgOption.TaskIDApp
				emt.TaskIDClientAPI = dcts.ClientTaskID
				emt.Sources = []int{dcts.MsgOption.ID}

				//ищем источник по указанному идентификатору
				sourceInfo, ok := hsm.ISL.GetSourceSetting(dcts.MsgOption.ID)
				if !ok {
					emt.MsgHuman = fmt.Sprintf("Ошибка, источника с ID %v не существует", dcts.MsgOption.ID)
					_ = saveMessageApp.LogMessage("error", fmt.Sprintf("source ID %v was not found%v", dcts.MsgOption.ID, funcName))

					//сообщение о том что задача была отклонена
					if err := ErrorMessage(emt); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					return
				}

				//проверяем подключение источника
				if !sourceInfo.ConnectionStatus {
					emt.MsgHuman = fmt.Sprintf("Ошибка, источник с ID %v не подключен", dcts.MsgOption.ID)
					_ = saveMessageApp.LogMessage("error", fmt.Sprintf("source ID %v is not connected%v", dcts.MsgOption.ID, funcName))

					//сообщение о том что задача была отклонена
					if err := ErrorMessage(emt); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					return
				}

				//проверяем наличие в очереди задачи с указанным ID
				_, ti, err := hsm.QTS.SearchTaskForIDQueueTaskStorage(dcts.MsgOption.TaskIDApp)
				if err == nil {
					var errMsg string

					if ti.TaskStatus == "execution" {
						//проверяем наличие выполняемой задачи
						if smti, ok := hsm.SMT.GetStoringMemoryTask(dcts.MsgOption.TaskIDApp); ok {
							//проверяем завершена ли задача
							if smti.TaskStatus {
								errMsg = fmt.Sprintf("Task with ID '%v' for source ID %v rejected. You cannot add a task with the same ID many times in a short period of time.", dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID)
								emt.MsgHuman = "Задача отклонена. Нельзя добавлять задачу с одним и тем же идентификатором множество раз в течении небольшого периода времени"
							} else {
								errMsg = fmt.Sprintf("You cannot add a task with ID '%v' to a source with ID %v because it is already running", dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID)
								emt.MsgHuman = fmt.Sprintf("Невозможно добавить задачу с ID '%v', для источника с ID %v, так как она уже выполняется", dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID)
							}
						}
					} else if ti.TaskStatus == "wait" {
						errMsg = fmt.Sprintf("Unable to add task with ID '%v' because it is already pending", dcts.MsgOption.TaskIDApp)
						emt.MsgHuman = fmt.Sprintf("Невозможно добавить задачу с ID '%v' так как она уже ожидает выполнения", dcts.MsgOption.TaskIDApp)
					} else {
						errMsg = fmt.Sprintf("Unable to add task with ID '%v'. The task has been completed, but has not yet been removed from the pending task list", dcts.MsgOption.ID)
						emt.MsgHuman = fmt.Sprintf("Невозможно добавить задачу с ID '%v'. Задача была выполнена, однако из списка задач ожидающих выполнения пока не удалена", dcts.MsgOption.ID)
					}

					if len(errMsg) > 0 {
						_ = saveMessageApp.LogMessage("error", errMsg)

						//сообщение о том что задача была отклонена
						if err := ErrorMessage(emt); err != nil {
							_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
						}

						return
					}
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
					emt.MsgHuman = fmt.Sprintf("Ошибка, запись для источника с ID %v отсутствует в памяти приложения", dcts.MsgOption.ID)
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					//сообщение о том что задача была отклонена
					if err := ErrorMessage(emt); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

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

				//информационное сообщение о том что задача добавлена в очередь
				notifications.SendNotificationToClientAPI(
					outCoreChans.OutCoreChanAPI,
					notifications.NotificationSettingsToClientAPI{
						MsgType:        "success",
						MsgDescription: fmt.Sprintf("Задача по скачиванию файлов с источника ID %v, добавлена в очередь", dcts.MsgOption.ID),
						Sources:        []int{dcts.MsgOption.ID},
					},
					msgc.ClientTaskID,
					msg.IDClientAPI)
			}

			if msgc.MsgInstruction == "to cancel downloading" {
				var dcts configure.DownloadControlTypeStart

				if err := json.Unmarshal(msgJSON, &dcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				emt.TaskID = dcts.MsgOption.TaskIDApp
				emt.TaskIDClientAPI = dcts.ClientTaskID
				emt.Sources = []int{dcts.MsgOption.ID}
				emt.MsgHuman = fmt.Sprintf("Невозможен останов задачи по скачиванию файлов с источника ID %v, не найдена задача с заданным идентификатором", dcts.MsgOption.ID)

				//ищем задачу в очереди задач и в выполняемых задачах
				if _, err := hsm.QTS.GetQueueTaskStorage(dcts.MsgOption.ID, dcts.MsgOption.TaskIDApp); err != nil {
					if err := ErrorMessage(emt); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					return
				}

				if _, ok := hsm.SMT.GetStoringMemoryTask(dcts.MsgOption.TaskIDApp); !ok {

					fmt.Println("func 'handlerMsgFromAPI', StoringMemoryTask task is not found!!!")

					//если задача есть в очереди но еще не выполнялась ставим
					// ей статус 'complete'
					if err := hsm.QTS.ChangeTaskStatusQueueTask(dcts.MsgOption.ID, dcts.MsgOption.TaskIDApp, "complete"); err != nil {

						fmt.Printf("func 'handlerMsgFromAPI', ERROR: %v (hsm.QTS.ChangeTaskStatusQueueTask)\n", err)

						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					//сообщение об успешном снятии задачи из очереди ожидания
					resMsg := configure.DownloadControlTypeInfo{
						MsgOption: configure.DownloadControlMsgTypeInfo{
							ID:        dcts.MsgOption.ID,
							TaskIDApp: dcts.MsgOption.TaskIDApp,
							Status:    "stop",
						},
					}
					resMsg.MsgType = "information"
					resMsg.MsgSection = "download control"
					resMsg.MsgInstruction = "task processing"
					resMsg.ClientTaskID = dcts.ClientTaskID

					msgJSON, err := json.Marshal(resMsg)
					if err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					emt.ChanToAPI <- &configure.MsgBetweenCoreAndAPI{
						MsgGenerator: "Core module",
						MsgRecipient: "API module",
						IDClientAPI:  msg.IDClientAPI,
						MsgJSON:      msgJSON,
					}

					notifications.SendNotificationToClientAPI(
						outCoreChans.OutCoreChanAPI,
						notifications.NotificationSettingsToClientAPI{
							MsgType:        "success",
							MsgDescription: fmt.Sprintf("Задача по скачиванию файлов с источника ID %v, удалена из очереди ожидающих задач", dcts.MsgOption.ID),
							Sources:        []int{dcts.MsgOption.ID},
						},
						dcts.ClientTaskID,
						msg.IDClientAPI)

					return
				}

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:     dcts.MsgOption.TaskIDApp,
					ClientName: msg.ClientName,
					Section:    "download control",
					Command:    "stop",
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
