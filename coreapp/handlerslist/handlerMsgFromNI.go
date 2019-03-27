package handlerslist

/*
* Обработчик запросов поступающих от модуля сетевого взаимодействия
*
* Версия 0.2, дата релиза 26.03.2019
* */

import (
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

	fmt.Println("--- START function 'HandlerMsgFromNI'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'HandlerMsgFromNI'"

	taskInfo, ok := smt.GetStoringMemoryTask(msg.TaskID)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "task with "+msg.TaskID+" not found")

		return
	}

	fmt.Printf("%v\n", msg)

	switch msg.Section {
	case "source control":
		fmt.Println("func 'HandlerMsgFromNI', section SOURCE CONTROL", msg.Command)

		//в БД
		if msg.Command == "keep list sources in database" {
			fmt.Println(":INSERT (Core module)")

			chanToDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator:    "NI module",
				MsgRecipient:    "DB module",
				MsgSection:      "source control",
				Instruction:     "insert",
				TaskID:          msg.TaskID,
				AdvancedOptions: msg.AdvancedOptions,
			}
		}

		//в БД
		if msg.Command == "delete sources in database" {
			fmt.Println(":DELETE (Core module)")

			chanToDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator:    "NI module",
				MsgRecipient:    "DB module",
				MsgSection:      "source control",
				Instruction:     "delete",
				TaskID:          msg.TaskID,
				AdvancedOptions: msg.AdvancedOptions,
			}
		}

		//в БД
		if msg.Command == "update sources in database" {
			fmt.Println(":UPDATE (Core module)")

			chanToDB <- &configure.MsgBetweenCoreAndDB{
				MsgGenerator:    "NI module",
				MsgRecipient:    "DB module",
				MsgSection:      "source control",
				Instruction:     "update",
				TaskID:          msg.TaskID,
				AdvancedOptions: msg.AdvancedOptions,
			}
		}

		//клиенту API
		if msg.Command == "confirm the action" {
			go getConfirmActionSourceListForAPI(chanToAPI, msg, smt)
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

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType:        ao.CriticalityMessage,
				MsgDescription: ao.HumanDescriptionNotification,
				Sources:        ao.Sources,
			}

			fmt.Printf("TASK INFO ClientID: %v", taskInfo)

			notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)
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
