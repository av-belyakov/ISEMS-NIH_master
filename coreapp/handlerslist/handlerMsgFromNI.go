package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//HandlerMsgFromNI обработчик запросов поступающих от модуля сетевого взаимодействия
func HandlerMsgFromNI(
	outCoreChans HandlerOutChans,
	msg *configure.MsgBetweenCoreAndNI,
	hsm HandlersStoringMemory) error {

	funcName := "HandlerMsgFromNI"

	var err error
	taskInfo, taskInfoIsExist := hsm.SMT.GetStoringMemoryTask(msg.TaskID)
	if taskInfoIsExist {
		hsm.SMT.TimerUpdateStoringMemoryTask(msg.TaskID)
	}

	switch msg.Section {
	case "source control":
		switch msg.Command {
		case "keep list sources in database":
			//в БД
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
			if err := getConfirmActionSourceListForAPI(outCoreChans.OutCoreChanAPI, msg, taskInfo.ClientID, taskInfo.ClientTaskID); err != nil {
				return err
			}

		case "change connection status source":
			//клиенту API
			if err := sendChanStatusSourceForAPI(outCoreChans.OutCoreChanAPI, msg); err != nil {
				return err
			}

		case "telemetry":
			//клиенту API
			jsonIn, ok := msg.AdvancedOptions.(*[]byte)
			if !ok {
				return fmt.Errorf("type conversion error%v", funcName)
			}

			var st configure.SourceTelemetry
			err := json.Unmarshal(*jsonIn, &st)
			if err != nil {
				return err
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
				return err
			}

			outCoreChans.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
				MsgGenerator: "Core module",
				MsgRecipient: "API module",
				MsgJSON:      jsonOut,
			}

		case "received version app":
			//клиенту API
			jsonIn, ok := msg.AdvancedOptions.(*[]byte)
			if !ok {
				return fmt.Errorf("type conversion error%v", funcName)
			}

			var mtp configure.MsgTypePong
			err := json.Unmarshal(*jsonIn, &mtp)
			if err != nil {
				return err
			}

			msg := configure.SourceVersionApp{
				MsgOptions: configure.SourceVersionAppOptions{
					SourceID:       msg.SourceID,
					AppVersion:     mtp.Info.AppVersion,
					AppReleaseDate: mtp.Info.AppReleaseDate,
				},
			}

			msg.MsgType = "information"
			msg.MsgSection = "source control"
			msg.MsgInstruction = "send version app"

			jsonOut, err := json.Marshal(msg)
			if err != nil {
				return err
			}

			outCoreChans.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
				MsgGenerator: "Core module",
				MsgRecipient: "API module",
				MsgJSON:      jsonOut,
			}

		}

	case "filtration control":
		//отправляем иформацию о ходе фильтрации в БД
		outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
			MsgGenerator:    "NI module",
			MsgRecipient:    "DB module",
			MsgSection:      "filtration control",
			Instruction:     "update",
			TaskID:          msg.TaskID,
			AdvancedOptions: msg.AdvancedOptions,
		}

		//клиенту API
		ao, ok := msg.AdvancedOptions.(configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus)
		if ok && taskInfoIsExist {
			//упаковываем в JSON и отправляем информацию о ходе фильтрации клиенту API
			// при чем если статус 'execute', то отправляем еще и содержимое поля 'FoundFilesInformation',
			// а если статус фильтрации 'stop' или 'complete' то данное поле не заполняем
			err = sendInformationFiltrationTask(outCoreChans.OutCoreChanAPI, taskInfo, &ao, msg.SourceID, msg.TaskID)

			if (ao.TaskStatus == "complete") || (ao.TaskStatus == "stop") {
				//для удаления задачи и из storingMemoryTask и storingMemoryQueueTask
				hsm.SMT.CompleteStoringMemoryTask(msg.TaskID)

				err = hsm.QTS.ChangeTaskStatusQueueTask(taskInfo.TaskParameter.FiltrationTask.ID, msg.TaskID, "complete")
			}
		}

	case "download control":
		if !taskInfoIsExist {
			return fmt.Errorf("there is no task with the specified ID %v", msg.TaskID)
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
		resMsgInfo.ClientTaskID = taskInfo.ClientTaskID

		hdtsct := handlerDownloadTaskStatusCompleteType{
			SourceID:       sourceID,
			TaskID:         msg.TaskID,
			ClientTaskID:   taskInfo.ClientTaskID,
			QTS:            hsm.QTS,
			SMT:            hsm.SMT,
			NS:             ns,
			ResMsgInfo:     resMsgInfo,
			OutCoreChanAPI: outCoreChans.OutCoreChanAPI,
			OutCoreChanDB:  outCoreChans.OutCoreChanDB,
		}

		switch msg.Command {
		//завершение записи части файла кратной 1%
		case "file download process":
			if fi, ok := msg.AdvancedOptions.(configure.MoreFileInformation); ok {
				if fi.Hex == resMsgInfo.MsgOption.DetailedFileInformation.Hex {
					resMsgInfo.MsgOption.DetailedFileInformation.AcceptedSizeByte = fi.AcceptedSizeByte
					resMsgInfo.MsgOption.DetailedFileInformation.AcceptedSizePercent = fi.AcceptedSizePercent
				}
			}

			//отправляем информацию клиенту API
			msgJSONInfo, err := json.Marshal(resMsgInfo)
			if err != nil {
				return err
			}

			msgToAPI.MsgJSON = msgJSONInfo
			outCoreChans.OutCoreChanAPI <- &msgToAPI

		//при завершении скачивания файла
		case "file download complete":
			//записываем информацию в БД
			// Модуль БД сам определяет когда стоит добавить запись в БД
			// а когда (основываясь на таймере) добавление записи в БД не происходит
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
				return err
			}

			msgToAPI.MsgJSON = msgJSONInfo
			outCoreChans.OutCoreChanAPI <- &msgToAPI

		//при завершении задачи по скачиванию файлов
		case "task completed":
			hdtsct.NS.MsgType = "success"
			hdtsct.NS.MsgDescription = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   sourceID,
				TaskType:   "скачивание файлов",
				TaskAction: "задача успешно выполнена",
			})

			hdtsct.ResMsgInfo.MsgOption.Status = "complete"
			hdtsct.ResMsgInfo.MsgOption.DetailedFileInformation = configure.MoreFileInformation{}

			if err := handlerDownloadTaskStatusComplete(hdtsct); err != nil {
				return err
			}

		//останов задачи пользователем
		case "file transfer stopped":
			hdtsct.NS.MsgType = "success"
			hdtsct.NS.MsgDescription = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   sourceID,
				TaskType:   "скачивание файлов",
				TaskAction: "задача успешно остановлена",
			})

			hdtsct.ResMsgInfo.MsgOption.Status = "complete"
			hdtsct.ResMsgInfo.MsgOption.DetailedFileInformation = configure.MoreFileInformation{}

			if err := handlerDownloadTaskStatusComplete(hdtsct); err != nil {
				return err
			}

		//останов задачи в связи с разрывом соединения с источником
		case "task stoped disconnect":
			//обновление статуса задачи
			if err := setStatusCompleteDownloadTask(hdtsct.TaskID, hdtsct.SMT); err != nil {
				return err
			}

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
			ns.MsgDescription = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   msg.SourceID,
				TaskType:   "скачивание файлов",
				TaskAction: "задача аварийно завершена",
				Message:    "задача была аварийно завершена из-за потери сетевого соединения с источником",
			})

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)

			hdtsct.ResMsgInfo.MsgOption.Status = "stop"
			hdtsct.ResMsgInfo.MsgOption.DetailedFileInformation = configure.MoreFileInformation{}

			//изменяем статус задачи на 'pause'
			// теперь задача будет ожидать соединения с источником в течении сутое
			// если соединения не произойдет то будет удалена или продолжит выполнятся
			// если соединение будет установлено
			_ = hsm.QTS.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "pause")

			//отмечаем задачу как завершенную для ее последующего удаления
			hsm.SMT.CompleteStoringMemoryTask(msg.TaskID)

			//отправляем информацию по задаче клиенту API
			msgJSONInfo, err := json.Marshal(hdtsct.ResMsgInfo)
			if err != nil {
				return err
			}

			msgToAPI.MsgJSON = msgJSONInfo
			outCoreChans.OutCoreChanAPI <- &msgToAPI

		//задача остановлена из-за внутренней ошибки приложения
		case "task stoped error":
			hdtsct.NS.MsgType = "danger"
			hdtsct.NS.MsgDescription = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID: sourceID,
				TaskType: "скачивание файлов",
				Message:  "задача была остановлена из-за внутренней ошибки приложения",
			})

			hdtsct.ResMsgInfo.MsgOption.Status = "stop"
			hdtsct.ResMsgInfo.MsgOption.DetailedFileInformation = configure.MoreFileInformation{}

			err = handlerDownloadTaskStatusComplete(hdtsct)

		}

	case "error notification":
		if !taskInfoIsExist {
			return fmt.Errorf("task with %v not found", msg.TaskID)
		}

		ao, ok := msg.AdvancedOptions.(configure.ErrorNotification)
		if !ok {
			return fmt.Errorf("type conversion error%v", funcName)
		}

		//стандартное информационное сообщение пользователю
		ns := notifications.NotificationSettingsToClientAPI{
			MsgType: "danger",
			MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
				Message: "непредвиденная ошибка, подробнее о возникшей проблеме в логах администратора приложения",
			}),
			Sources: ao.Sources,
		}

		notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)

		//останавливаем выполнение задачи
		hsm.SMT.CompleteStoringMemoryTask(msg.TaskID)

		err = fmt.Errorf(ao.HumanDescriptionError)

	case "message notification":
		if msg.Command == "send client API" {
			ao, ok := msg.AdvancedOptions.(configure.MessageNotification)
			if !ok {
				return fmt.Errorf("type conversion error%v", funcName)
			}

			taskActionPattern := map[string]string{
				"start":     "инициализация выполнения задачи",
				"stop":      "останов задачи",
				"load list": "загрузка списка",
			}
			taskTypePattern := map[string]string{
				"source control":     "управление источниками",
				"filtration control": "фильтрация",
				"download files":     "скачивание файлов",
			}
			var sourceID int
			if len(ao.Sources) != 0 {
				sourceID = ao.Sources[0]
			}
			ns := notifications.NotificationSettingsToClientAPI{
				MsgType: ao.CriticalityMessage,
				MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   sourceID,
					TaskType:   taskTypePattern[ao.Section],
					TaskAction: taskActionPattern[ao.TypeActionPerformed],
					Message:    ao.HumanDescriptionNotification,
				}),
				Sources: ao.Sources,
			}

			if !taskInfoIsExist {
				_, qti, err := hsm.QTS.SearchTaskForIDQueueTaskStorage(msg.TaskID)
				if err != nil {
					return err
				}

				notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, qti.TaskIDClientAPI, qti.IDClientAPI)

				return fmt.Errorf("task with %v not found", msg.TaskID)
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)
		}

	case "monitoring task performance":
		if msg.Command == "complete task" {
			hsm.SMT.CompleteStoringMemoryTask(msg.TaskID)

			if !taskInfoIsExist {
				return fmt.Errorf("Section: 'monitoring task performance', task with %v not found", msg.TaskID)
			}

			err = hsm.QTS.ChangeTaskStatusQueueTask(taskInfo.TaskParameter.FiltrationTask.ID, msg.TaskID, "complete")
		}
	}

	return err
}
