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

	taskInfo, ok := hsm.SMT.GetStoringMemoryTask(msg.TaskID)
	if ok {
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
			/*
				   Обработка изменения состояния соединения
				   в том числе РАЗРЫВ соединения

				здесь можно сделать удаление задачи из StoringMemoryTask
				и изменение статуса задачи в StoringMemoryQueueTask
				с 'execute' на 'wait'
			*/

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
			msg.MsgInsturction = "send telemetry"

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

		//отправляем иформацию о ходе фильтрации в БД
		outCoreChans.OutCoreChanDB <- &msgChan

		/* упаковываем в JSON и отправляем информацию о ходе фильтрации клиенту API
		при чем если статус 'execute', то отправляем еще и содержимое поля 'FoundFilesInformation',
		а если статус фильтрации 'stop' или 'complite' то данное поле не заполняем */

		//если задача найдена
		if ok {
			if err := sendInformationFiltrationTask(outCoreChans.OutCoreChanAPI, taskInfo, msg); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}
		}

	case "download control":
		fmt.Println("func 'HandlerMsgFromNI', section DOWNLOAD CONTROL")

		switch msg.Command {
		//при завершении скачивания файла
		case "file download complete":
			dfi, ok := msg.AdvancedOptions.(configure.DetailedFileInformation)
			if !ok {
				_ = saveMessageApp.LogMessage("error", "type conversion error"+funcName)

				return
			}

			fi := map[string]*configure.DownloadFilesInformation{
				dfi.Name: &configure.DownloadFilesInformation{},
			}

			//обновляем список загруженых файлов StoreMemoryQueueTask
			/*
				функцию ChangeIsLoadedFiles() надо ПОТЕСТИТЬ
			*/
			if err := hsm.QTS.ChangeIsLoadedFiles(msg.SourceID, msg.TaskID, fi); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

			//отправляем информацию клиенту API

			//отправляем информацию в БД, НО ТОЛЬКО РАЗ за определенное
			//количество, как с фильтрацией

		//при завершении скачивания ВСЕХ файлов
		case "download complete":
			//отправляем информацию клиенту API

			//отправляем информацию в БД

			//удаляем задачу из StoringMemoryQueueTask

		}

	case "error notification":
		if taskInfo == nil {
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

			if taskInfo == nil {
				_ = saveMessageApp.LogMessage("error", "task with "+msg.TaskID+" not found")

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
			hsm.SMT.CompleteStoringMemoryTask(msg.TaskID)
		}
	}
}
