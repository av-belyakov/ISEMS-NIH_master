package handlerslist

/*
* Обработчик запросов поступающих от модуля сетевого взаимодействия
*
* Версия 0.1, дата релиза 18.03.2019
* */

import (
	"encoding/json"
	"errors"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//HandlerMsgFromNI обработчик запросов поступающих от модуля сетевого взаимодействия
func HandlerMsgFromNI(
	chanToAPI chan<- configure.MsgBetweenCoreAndAPI,
	msg *configure.MsgBetweenCoreAndNI,
	smt *configure.StoringMemoryTask,
	chanToDB chan<- configure.MsgBetweenCoreAndDB) error {

	fmt.Println("--- START function 'HandlerMsgFromNI'...")

	funcName := ", function 'HandlerMsgFromNI'"

	taskInfo, ok := smt.GetStoringMemoryTask(msg.TaskID)
	if !ok {
		return errors.New("task with " + msg.TaskID + " not found")
	}

	switch msg.Section {
	case "source control":
		fmt.Println("func 'HandlerMsgFromNI', section SOURCE CONTROL")

		if msg.Command == "keep list sources in database" {
			//отправить список источников ТОЛЬКО в БД

			sourceList, err := getSourceListToLoadDB(msg.AdvancedOptions)
			if err != nil {
				return err
			}

			fmt.Println("*-**********", sourceList, "---********")

			chanToDB <- configure.MsgBetweenCoreAndDB{
				MsgGenerator: "NI module",
				MsgRecipient: "DB module",
				MsgSection:   "source control",
				Instruction:  "insert",
				TaskID:       msg.TaskID,
				AdvancedOptions: configure.MsgInfoChangeStatusSource{
					SourceListIsExist: true,
					SourceList:        sourceList,
				},
			}
		} else if msg.Command == "send list sources to client api" {
			//отправить список источников ТОЛЬКО клиенту API

			sourceList, err := getSourceListToAPI(msg.AdvancedOptions)
			if err != nil {
				return err
			}

			msg := configure.SourceControlAllListSources{
				SourceList: *sourceList,
			}

			msg.MsgType = "information"
			msg.MsgSection = "source control"
			msg.MsgInsturction = "send current source list"
			msg.ClientTaskID = taskInfo.ClientTaskID

			msgjson, err := json.Marshal(&msg)
			if err != nil {
				return err
			}

			chanToAPI <- configure.MsgBetweenCoreAndAPI{
				MsgGenerator: "Core module",
				MsgRecipient: "API module",
				IDClientAPI:  taskInfo.ClientID,
				MsgJSON:      msgjson,
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
				return errors.New("cannot cast type" + funcName)
			}

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType:        ao.CriticalityMessage,
				MsgDescription: ao.HumanDescriptionNotification,
				Sources:        ao.Sources,
			}

			fmt.Printf("TASK INFO ClientID: %q", taskInfo)

			notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)
		}
	}

	return nil
}

//getSourceListToLoadDB подготаваливаем список источников полученный от модуля
//NI для загрузки его в БД
func getSourceListToLoadDB(l interface{}) (*[]configure.MainOperatingParametersSource, error) {
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
		return nil, errors.New("type conversion error, function 'getSourceListToLoadDB'")
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
}
