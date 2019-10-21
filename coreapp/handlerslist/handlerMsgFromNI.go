package handlerslist

/*
* Обработчик запросов поступающих от модуля сетевого взаимодействия
*
* Версия 0.3, дата релиза 10.06.2019
* */

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromNI обработчик запросов поступающих от модуля сетевого взаимодействия
func HandlerMsgFromNI(
	outCoreChans HandlerOutChans,
	msg *configure.MsgBetweenCoreAndNI,
	hsm HandlersStoringMemory,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) {

	funcName := ", function 'HandlerMsgFromNI'"

	taskInfo, taskInfoIsExist := hsm.SMT.GetStoringMemoryTask(msg.TaskID)
	if taskInfoIsExist {
		hsm.SMT.TimerUpdateStoringMemoryTask(msg.TaskID)
	}

	switch msg.Section {
	case "source control":
		switch msg.Command {
		case "keep list sources in database":
			//в БД
			fmt.Println(":INSERT (Core module)")

			outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator:    "NI module",
				MsgRecipient:    "DB module",
				MsgSection:      "source control",
				Instruction:     "insert",
				TaskID:          msg.TaskID,
				AdvancedOptions: msg.AdvancedOptions,
			}

		case "delete sources in database":
			//в БД
			fmt.Println(":DELETE (Core module)")

			outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator:    "NI module",
				MsgRecipient:    "DB module",
				MsgSection:      "source control",
				Instruction:     "delete",
				TaskID:          msg.TaskID,
				AdvancedOptions: msg.AdvancedOptions,
			}

		case "update sources in database":
			//в БД
			fmt.Println(":UPDATE (Core module)")

			outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator:    "NI module",
				MsgRecipient:    "DB module",
				MsgSection:      "source control",
				Instruction:     "update",
				TaskID:          msg.TaskID,
				AdvancedOptions: msg.AdvancedOptions,
			}

		case "confirm the action":
			//клиенту API
			if err := getConfirmActionSourceListForAPI(outCoreChans.OutCoreChanAPI, msg, hsm.SMT); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

		case "change connection status source":
			//клиенту API
			if err := sendChanStatusSourceForAPI(outCoreChans.OutCoreChanAPI, msg); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

		case "telemetry":
			//клиенту API
			jsonIn, ok := msg.AdvancedOptions.(*[]byte)
			if !ok {
				_ = saveMessageApp.LogMessage("error", "type conversion error"+funcName)

				return
			}

			var st configure.SourceTelemetry
			err := json.Unmarshal(*jsonIn, &st)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			msg := configure.Telemetry{
				MsgOptions: configure.TelemetryOptions{
					SourceID:    msg.SourceID,
					Information: st.Info,
				},
			}

			msg.MsgType = "information"
			msg.MsgSection = "source control"
			msg.MsgInstruction = "send telemetry"

			jsonOut, err := json.Marshal(msg)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			outCoreChans.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
				MsgGenerator: "Core module",
				MsgRecipient: "API module",
				MsgJSON:      jsonOut,
			}
		}

	case "filtration control":
		msgChan := configure.MsgBetweenCoreAndDB{
			MsgGenerator:    "NI module",
			MsgRecipient:    "DB module",
			MsgSection:      "filtration control",
			Instruction:     "update",
			TaskID:          msg.TaskID,
			AdvancedOptions: msg.AdvancedOptions,
		}

		if ao, ok := msg.AdvancedOptions.(configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus); ok && taskInfoIsExist {
			/* упаковываем в JSON и отправляем информацию о ходе фильтрации клиенту API
			при чем если статус 'execute', то отправляем еще и содержимое поля 'FoundFilesInformation',
			а если статус фильтрации 'stop' или 'complete' то данное поле не заполняем */
			if err := sendInformationFiltrationTask(outCoreChans.OutCoreChanAPI, taskInfo, &ao, msg.SourceID, msg.TaskID); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

			if (ao.TaskStatus == "complete") || (ao.TaskStatus == "stop") {
				//для удаления задачи и из storingMemoryTask и storingMemoryQueueTask
				hsm.SMT.CompleteStoringMemoryTask(msg.TaskID)

				if err := hsm.QTS.ChangeTaskStatusQueueTask(taskInfo.TaskParameter.FiltrationTask.ID, msg.TaskID, "complete"); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}
			}

			//отправляем иформацию о ходе фильтрации в БД
			outCoreChans.OutCoreChanDB <- &msgChan
		}

	case "download control":
		fmt.Println("func 'HandlerMsgFromNI', section DOWNLOAD CONTROL")

		if !taskInfoIsExist {
			_ = saveMessageApp.LogMessage("error", fmt.Sprintf("there is no task with the specified ID %v", msg.TaskID))

			return
		}

		sourceID := taskInfo.TaskParameter.DownloadTask.ID

		msgToAPI := configure.MsgBetweenCoreAndAPI{
			MsgGenerator: "Core module",
			MsgRecipient: "API module",
			IDClientAPI:  taskInfo.ClientID,
		}

		ns := notifications.NotificationSettingsToClientAPI{
			Sources: []int{sourceID},
		}

		//отправляем сообщение о том что задача была отклонена
		resMsgInfo := configure.DownloadControlTypeInfo{
			MsgOption: configure.DownloadControlMsgTypeInfo{
				ID:                                  sourceID,
				Status:                              "execute",
				TaskIDApp:                           msg.TaskID,
				NumberFilesTotal:                    taskInfo.TaskParameter.DownloadTask.NumberFilesTotal,
				NumberFilesDownloaded:               taskInfo.TaskParameter.DownloadTask.NumberFilesDownloaded,
				NumberFilesDownloadedError:          taskInfo.TaskParameter.DownloadTask.NumberFilesDownloadedError,
				PathDirectoryStorageDownloadedFiles: taskInfo.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles,
				DetailedFileInformation: configure.MoreFileInformation{
					Name:                taskInfo.TaskParameter.DownloadTask.FileInformation.Name,
					Hex:                 taskInfo.TaskParameter.DownloadTask.FileInformation.Hex,
					FullSizeByte:        taskInfo.TaskParameter.DownloadTask.FileInformation.FullSizeByte,
					AcceptedSizeByte:    taskInfo.TaskParameter.DownloadTask.FileInformation.AcceptedSizeByte,
					AcceptedSizePercent: taskInfo.TaskParameter.DownloadTask.FileInformation.AcceptedSizePercent,
				},
			},
		}
		resMsgInfo.MsgType = "information"
		resMsgInfo.MsgSection = "download control"
		resMsgInfo.MsgInstruction = "task processing"

		hdtsct := handlerDownloadTaskStatusCompleteType{
			SourceID:       sourceID,
			TaskID:         msg.TaskID,
			TI:             taskInfo,
			QTS:            hsm.QTS,
			NS:             ns,
			ResMsgInfo:     resMsgInfo,
			OutCoreChanAPI: outCoreChans.OutCoreChanAPI,
			OutCoreChanDB:  outCoreChans.OutCoreChanDB,
		}

		switch msg.Command {
		//завершение записи части файла кратной 1%
		case "file download process":
			//отправляем информацию клиенту API
			msgJSONInfo, err := json.Marshal(resMsgInfo)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			msgToAPI.MsgJSON = msgJSONInfo
			outCoreChans.OutCoreChanAPI <- &msgToAPI

		//при завершении скачивания файла
		case "file download complete":
			/*
				записываем информацию в БД
				Модуль БД сам определяет когда стоит добавить запись в БД
				а когда (основываясь на таймере) добавление записи в БД не происходит
			*/
			outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator: "NI module",
				MsgRecipient: "DB module",
				MsgSection:   "download control",
				Instruction:  "update",
				TaskID:       msg.TaskID,
			}

			//отправляем информацию клиенту API
			msgJSONInfo, err := json.Marshal(resMsgInfo)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			msgToAPI.MsgJSON = msgJSONInfo
			outCoreChans.OutCoreChanAPI <- &msgToAPI

		//при завершении задачи по скачиванию файлов
		case "task completed":
			hdtsct.NS.MsgType = "success"
			hdtsct.NS.MsgDescription = fmt.Sprintf("Задача по скачиванию файлов с источника '%v' выполнена успешно", sourceID)

			hdtsct.ResMsgInfo.MsgOption.Status = "complete"
			hdtsct.ResMsgInfo.MsgOption.DetailedFileInformation = configure.MoreFileInformation{}

			if err := handlerDownloadTaskStatusComplete(hdtsct); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

		//останов задачи пользователем
		case "file transfer stopped":
			hdtsct.NS.MsgType = "success"
			hdtsct.NS.MsgDescription = fmt.Sprintf("Задача по скачиванию файлов с источника '%v' была успешно остановлена", sourceID)

			hdtsct.ResMsgInfo.MsgOption.Status = "complete"
			hdtsct.ResMsgInfo.MsgOption.DetailedFileInformation = configure.MoreFileInformation{}

			if err := handlerDownloadTaskStatusComplete(hdtsct); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

		//останов задачи в связи с разрывом соединения с источником
		case "task stoped disconnect":
			//записываем информацию в БД
			hdtsct.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator: "NI module",
				MsgRecipient: "DB module",
				MsgSection:   "download control",
				Instruction:  "update",
				TaskID:       hdtsct.TaskID,
			}

			//отправляем информационное сообщение клиенту API
			ns.MsgType = "warning"
			ns.MsgDescription = fmt.Sprintf("Задача по скачиванию файлов с источника %v была аварийно завершена из-за потери сетевого соединения", msg.SourceID)
			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)

			hdtsct.ResMsgInfo.MsgOption.Status = "stop"
			hdtsct.ResMsgInfo.MsgOption.DetailedFileInformation = configure.MoreFileInformation{}

			//изменяем статус задачи на 'wait' (storingMemoryQueueTask)
			if err := hsm.QTS.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "wait"); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

			//отправляем информацию по задаче клиенту API
			msgJSONInfo, err := json.Marshal(hdtsct.ResMsgInfo)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}
			msgToAPI.MsgJSON = msgJSONInfo
			outCoreChans.OutCoreChanAPI <- &msgToAPI

		//задача остановлена из-за внутренней ошибки приложения
		case "task stoped error":
			hdtsct.NS.MsgType = "danger"
			hdtsct.NS.MsgDescription = fmt.Sprintf("Задача по скачиванию файлов с источника '%v' была остановлена из-за внутренней ошибки приложения", sourceID)

			hdtsct.ResMsgInfo.MsgOption.Status = "stop"
			hdtsct.ResMsgInfo.MsgOption.DetailedFileInformation = configure.MoreFileInformation{}

			if err := handlerDownloadTaskStatusComplete(hdtsct); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

		}

	case "error notification":
		if !taskInfoIsExist {
			_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with %v not found", msg.TaskID))

			return
		}

		ao, ok := msg.AdvancedOptions.(configure.ErrorNotification)
		if !ok {
			_ = saveMessageApp.LogMessage("error", "type conversion error"+funcName)

			return
		}

		//записываем информацию об ошибках в лог приложения
		_ = saveMessageApp.LogMessage("error", ao.HumanDescriptionError)

		//стандартное информационное сообщение пользователю
		ns := notifications.NotificationSettingsToClientAPI{
			MsgType:        "danger",
			MsgDescription: "Непредвиденная ошибка, выполнение задачи остановлено. Подробнее о возникшей проблеме в логах администратора приложения.",
			Sources:        ao.Sources,
		}

		notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)

		//останавливаем выполнение задачи
		hsm.SMT.CompleteStoringMemoryTask(msg.TaskID)

	case "message notification":
		if msg.Command == "send client API" {
			ao, ok := msg.AdvancedOptions.(configure.MessageNotification)
			if !ok {
				_ = saveMessageApp.LogMessage("error", "type conversion error"+funcName)

				return
			}

			if !taskInfoIsExist {
				_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with %v not found", msg.TaskID))

				return
			}

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType:        ao.CriticalityMessage,
				MsgDescription: ao.HumanDescriptionNotification,
				Sources:        ao.Sources,
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)
		}

	case "monitoring task performance":
		if msg.Command == "complete task" {

			fmt.Printf("------ SECTOR: 'monitoring task performance', Command: 'complete task', %v\n", msg)

			if !taskInfoIsExist {
				return
			}

			hsm.SMT.CompleteStoringMemoryTask(msg.TaskID)

			/*if !taskInfoIsExist {
				_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with %v not found", msg.TaskID))

				return
			}*/

			if err := hsm.QTS.ChangeTaskStatusQueueTask(taskInfo.TaskParameter.FiltrationTask.ID, msg.TaskID, "complete"); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}
		}
	}
}
