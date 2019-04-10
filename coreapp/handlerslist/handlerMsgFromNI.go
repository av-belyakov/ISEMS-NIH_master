package handlerslist

/*
* Обработчик запросов поступающих от модуля сетевого взаимодействия
*
* Версия 0.2, дата релиза 26.03.2019
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

	//	fmt.Printf("--- START function 'HandlerMsgFromNI'... (Core module) %v\n", msg.Command)

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'HandlerMsgFromNI'"

	taskInfo, ok := smt.GetStoringMemoryTask(msg.TaskID)
	if ok {
		smt.TimerUpdateStoringMemoryTask(msg.TaskID)
	}

	//	fmt.Printf("%v\n", msg)

	switch msg.Section {
	case "source control":
		fmt.Printf("func 'HandlerMsgFromNI', section SOURCE CONTROL '%v'\n", msg.Command)

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
			fmt.Println("MSG Core module 'change connection status source'")
			fmt.Printf("%v\n", msg.AdvancedOptions)

			sendChanStatusSourceForAPI(chanToAPI, msg)

		case "telemetry":
			//клиенту API
			fmt.Println("TELEMETRY func 'handlerMsgFromNI'")

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
		fmt.Println("func 'HandlerMsgFromNI', section FILTRATION CONTROL")

	case "download control":
		fmt.Println("func 'HandlerMsgFromNI', section DOWNLOAD CONTROL")

	case "error notification":
		fmt.Println("func 'HandlerMsgFromNI', section ERROR NOTIFICATION")

	case "message notification":
		fmt.Println("func 'HandlerMsgFromNI', section MESSAGE NOTIFICATION")

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

			fmt.Printf("TASK INFO ClientID: %v", taskInfo)

			notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)
		}

	case "monitoring task performance":
		fmt.Println("func 'HandlerMsgFromNI', section MESSAGE MONITORING TASKPERFORMANCE")

		if msg.Command == "complete task" {
			smt.StoringMemoryTaskComplete(msg.TaskID)
		}
	}
}

//getSourceListToLoadDB подготаваливаем список источников полученный от модуля
//NI для загрузки его в БД
/*func getSourceListToLoadDB(l interface{}) (*[]configure.MainOperatingParametersSource, error) {
	ls, ok := l.(*map[int]configure.SourceSetting)
	if !ok {
		return nil, errors.New("type conversion error, function 'getSourceListToLoadDB'")
	}

	list := make([]configure.MainOperatingParametersSource, 0, len(*ls))

	for id, s := range *ls {
		list = append(list, configure.MainOperatingParametersSource{
			ID:       id,
			IP:       s.IP,
			Token:    s.Token,
			AsServer: s.AsServer,
			Options: configure.SourceDetailedInformation{
				StorageFolders:            s.Settings.StorageFolders,
				EnableTelemetry:           s.Settings.EnableTelemetry,
				MaxCountProcessFiltration: s.Settings.MaxCountProcessFiltration,
			},
		})
	}

	return &list, nil
}

//getSourceListToAPI подготавливаем списко источников отправляемых пользователю
func getSourceListToAPI(l interface{}) (*[]configure.DetailedListSources, error) {
	ls, ok := l.(*map[int]configure.SourceSetting)
	if !ok {
		return nil, errors.New("type conversion error, function 'getSourceListToAPI'")
	}

	list := make([]configure.DetailedListSources, 0, len(*ls))
	for id, s := range *ls {
		list = append(list, configure.DetailedListSources{
			ID:         id,
			ActionType: "none",
			Argument: configure.SourceArguments{
				IP:    s.IP,
				Token: "", //думаю что токен не стоит возвращать, особенно если он пойдет не тому клиенту
				Settings: configure.SourceSettings{
					AsServer:                  s.AsServer,
					EnableTelemetry:           s.Settings.EnableTelemetry,
					MaxCountProcessFiltration: s.Settings.MaxCountProcessFiltration,
					StorageFolders:            s.Settings.StorageFolders,
				},
			},
		})
	}

	return &list, nil
}*/
