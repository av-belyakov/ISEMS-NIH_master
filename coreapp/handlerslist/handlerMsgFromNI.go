package handlerslist

/*
* Обработчик запросов поступающих от модуля сетевого взаимодействия
*
* Версия 0.1, дата релиза 18.03.2019
* */

import (
	"errors"
	"fmt"

	"ISEMS-NIH_master/common/notifications"
	"ISEMS-NIH_master/configure"
)

//HandlerMsgFromNI обработчик запросов поступающих от модуля сетевого взаимодействия
func HandlerMsgFromNI(chanToAPI chan<- configure.MsgBetweenCoreAndAPI, msg *configure.MsgBetweenCoreAndNI, smt *configure.StoringMemoryTask, chanToDB chan<- configure.MsgBetweenCoreAndDB) error {
	fmt.Println("--- START function 'HandlerMsgFromNI'...")

	funcName := ", function 'HandlerMsgFromNI'"

	/*
		chanInCore <- configure.MsgBetweenCoreAndNI{
			TaskID:  msg.TaskID,
			Section: "message notification",
			Command: "send client API",
			AdvancedOptions: configure.MessageNotification{
				SourceReport:                 "NI module",
				Section:                      "source control",
				TypeActionPerformed:          "load list",
				HumanDescriptionNotification: fmt.Sprintf("На источнике (-ах) %q выполняются задачи, изменение настроек не доступно", notAddSourceList),
				Sources: notAddSourceList,
			},
		}
	*/

	taskInfo, ok := smt.GetStoringMemoryTask(msg.TaskID)
	if !ok {
		return errors.New("task with " + msg.TaskID + " not found")
	}
	switch msg.Section {
	case "source control":
		fmt.Println("func 'HandlerMsgFromNI', section SOURCE CONTROL")

	/*				chanToAPI <- configure.MsgBetweenCoreAndAPI{
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
				MsgType:        "info",
				MsgDescription: ao.HumanDescriptionNotification,
				Sources:        ao.Sources,
			}

			notifications.SendNotificationToClientAPI(chanToAPI, ns, "", taskInfo.ClientID)
		}
	}

	return nil
}
