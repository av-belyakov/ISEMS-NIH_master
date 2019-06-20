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
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	msg *configure.MsgBetweenCoreAndNI,
	smt *configure.StoringMemoryTask,
	chanToDB chan<- *configure.MsgBetweenCoreAndDB) {

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'HandlerMsgFromNI'"

	taskInfo, ok := smt.GetStoringMemoryTask(msg.TaskID)
	if ok {
		smt.TimerUpdateStoringMemoryTask(msg.TaskID)
	}

	switch msg.Section {
	case "source control":
		switch msg.Command {
		case "keep list sources in database":
			//в БД
			fmt.Println(":INSERT (Core module)")

			chanToDB <- &configure.MsgBetweenCoreAndDB{
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

			chanToDB <- &configure.MsgBetweenCoreAndDB{
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

			chanToDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator:    "NI module",
				MsgRecipient:    "DB module",
				MsgSection:      "source control",
				Instruction:     "update",
				TaskID:          msg.TaskID,
				AdvancedOptions: msg.AdvancedOptions,
			}

		case "confirm the action":
			//клиенту API
			getConfirmActionSourceListForAPI(chanToAPI, msg, smt)

		case "change connection status source":
			//клиенту API
			sendChanStatusSourceForAPI(chanToAPI, msg)

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

			chanToAPI <- &configure.MsgBetweenCoreAndAPI{
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
		chanToDB <- &msgChan

		/* упаковываем в JSON и отправляем информацию о ходе фильтрации клиенту API
		при чем если статус 'execute', то отправляем еще и содержимое поля 'FoundFilesInformation',
		а если статус фильтрации 'stop' или 'complite' то данное поле не заполняем */
		sendInformationFiltrationTask(chanToAPI, taskInfo, msg)

	case "download control":
		fmt.Println("func 'HandlerMsgFromNI', section DOWNLOAD CONTROL")

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

		notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)

		//останавливаем выполнение задачи
		smt.CompleteStoringMemoryTask(msg.TaskID)

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

			notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)
		}

	case "monitoring task performance":
		if msg.Command == "complete task" {
			smt.CompleteStoringMemoryTask(msg.TaskID)
		}
	}
}
