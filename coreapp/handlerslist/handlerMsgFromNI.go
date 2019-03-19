package handlerslist

/*
* Обработчик запросов поступающих от модуля сетевого взаимодействия
*
* Версия 0.1, дата релиза 18.03.2019
* */

import (
	"errors"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//HandlerMsgFromNI обработчик запросов поступающих от модуля сетевого взаимодействия
func HandlerMsgFromNI(chanToAPI chan<- configure.MsgBetweenCoreAndAPI, msg *configure.MsgBetweenCoreAndNI, smt *configure.StoringMemoryTask, chanToDB chan<- configure.MsgBetweenCoreAndDB) error {
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

			//			fmt.Printf("to Database %T %v\n\n", msg.AdvancedOptions, msg.AdvancedOptions)

			sourceList, err := getSourceListToLoadDB(msg.AdvancedOptions)
			if err != nil {
				return err
			}

			fmt.Println("*-**********", sourceList, "---********")

			chanToDB <- configure.MsgBetweenCoreAndDB{
				MsgGenerator: "NI module",
				MsgRecipient: "DB module",
				MsgDirection: "request",
				MsgSection:   "source control",
				Instruction:  "insert",
				AdvancedOptions: configure.MsgInfoChangeStatusSource{
					SourceListIsExist: true,
					SourceList:        sourceList,
				},
			}
		} else if msg.Command == "send list sources to client api" {
			//отправить список источников ТОЛЬКО клиенту API

		}

	/*
		//чтении из памяти нового списка для сохранения в БД
					chanInCore <- configure.MsgBetweenCoreAndNI{
						TaskID:  msg.TaskID,
						Section: "source control",
						Command: "keep list sources in database",
					}

					//чтение из памяти для клиента API
					chanInCore <- configure.MsgBetweenCoreAndNI{
						TaskID:  msg.TaskID,
						Section: "source control",
						Command: "send list sources to client api",
					}

		chanToAPI <- configure.MsgBetweenCoreAndAPI{
							MsgGenerator: taskInfo.ModuleResponsibleImplementation,
							MsgRecipient: taskInfo.ModuleThatSetTask,
							IDClientAPI:  taskInfo.ClientID,
							//MsgJSON:
						}
					}*/

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

			notifications.SendNotificationToClientAPI(chanToAPI, ns, "", taskInfo.ClientID)
		}
	}

	return nil
}

//подготаваливаем список источников полученный от модуля
//NI для загрузки его в БД
func getSourceListToLoadDB(l interface{}) (*[]configure.MainOperatingParametersSource, error) {
	ls, ok := l.(*map[int]configure.SourceSetting)
	if !ok {
		return nil, errors.New("type conversion error, function 'getSourceListToLoadDB'")
	}

	list := make([]configure.MainOperatingParametersSource, 0, 0) //len(ls))

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
