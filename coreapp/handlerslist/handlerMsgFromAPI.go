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
	hsm HandlersStoringMemory) {

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
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
	_ = saveMessageApp.LogMessage("requests", fmt.Sprintf("client name: '%v' (%v), request: type = %v, section = %v, instruction = %v, client task ID = %v", msg.ClientName, msg.ClientIP, msgc.MsgType, msgc.MsgSection, msgc.MsgInsturction, msgc.ClientTaskID))

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInsturction == "send new source list" {
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
			if msgc.MsgInsturction == "get an updated list of sources" {
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
			if msgc.MsgInsturction == "performing an action" {
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
			if msgc.MsgInsturction == "to start filtering" {
				var fcts configure.FiltrationControlTypeStart
				if err := json.Unmarshal(msgJSON, &fcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				go handlerFiltrationControlTypeStart(outCoreChans.OutCoreChanDB, &fcts, hsm, msg.IDClientAPI, outCoreChans.OutCoreChanAPI)

				return
			}

			//команда на останов фильтрации
			if msgc.MsgInsturction == "to cancel filtering" {
				//ищем выполняемую задачу по ClientTaskID (уникальный ID задачи на стороне клиента)
				taskID, ti, isExist := hsm.SMT.GetStoringMemoryTaskForClientID(msg.IDClientAPI, msgc.ClientTaskID)
				if !isExist {
					nsErr := notifications.NotificationSettingsToClientAPI{
						MsgType:        "danger",
						MsgDescription: fmt.Sprintf("Ошибка, по переданному идентификатору '%v' выполняемых задач не обнаружено", msgc.ClientTaskID),
					}
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErr, msgc.ClientTaskID, msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:     taskID,
					ClientName: msg.ClientName,
					Section:    "filtration control",
					Command:    "stop",
					SourceID:   ti.TaskParameter.FiltrationTask.ID,
				}

				return
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
			_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgInstruction'"+funcName)

			return

		// УПРАВЛЕНИЕ ВЫГРУЗКОЙ ФАЙЛОВ
		case "download control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'download control'")

			if msgc.MsgInsturction == "to start downloading" {
				fmt.Println("START task 'DOWNLOADING'")

				var dcts configure.DownloadControlTypeStart
				if err := json.Unmarshal(msgJSON, &dcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				//ищем источник по указанному идентификатору
				sourceInfo, ok := hsm.ISL.GetSourceSetting(dcts.MsgOption.ID)
				if !ok {
					nsErrJSON.MsgDescription = fmt.Sprintf("Ошибка, источника с ID %v не существует", dcts.MsgOption.ID)
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", fmt.Sprintf("source ID %v was not found%v", dcts.MsgOption.ID, funcName))

					return
				}

				//проверяем подключение источника
				if !sourceInfo.ConnectionStatus {
					nsErrJSON.MsgDescription = fmt.Sprintf("Ошибка, источник с ID %v не подключен", dcts.MsgOption.ID)
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", fmt.Sprintf("source ID %v is not connected%v", dcts.MsgOption.ID, funcName))

					return
				}

				//добавляем задачу в очередь
				hsm.QTS.AddQueueTaskStorage(dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID, configure.CommonTaskInfo{
					IDClientAPI:     msg.IDClientAPI,
					TaskIDClientAPI: dcts.ClientTaskID,
					TaskType:        "download",
				}, &configure.DescriptionParametersReceivedFromUser{
					DownloadList: dcts.MsgOption.FileList,
				})

				//устанавливаем статус источника для данной задачи как подключен
				hsm.QTS.ChangeAvailabilityConnection(dcts.MsgOption.ID, dcts.MsgOption.TaskIDApp)

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

			if msgc.MsgInsturction == "to cancel downloading" {
				fmt.Println("STOP task 'DOWNLOADING'")

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
