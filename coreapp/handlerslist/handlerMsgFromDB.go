package handlerslist

import (
	"errors"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromDB обработчик сообщений приходящих от модуля взаимодействия с базой данных
func HandlerMsgFromDB(
	chanToAPI chan<- configure.MsgBetweenCoreAndAPI,
	res *configure.MsgBetweenCoreAndDB,
	smt *configure.StoringMemoryTask,
	chanToNI chan<- configure.MsgBetweenCoreAndNI) error {

	fmt.Println("START function 'HandlerMsgFromDB' module coreapp")
	fmt.Printf("%v", res)

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'HandlerMsgFromDB'"

	taskInfo, ok := smt.GetStoringMemoryTask(res.TaskID)
	if !ok {
		return errors.New("task with " + res.TaskID + " not found")
	}

	if res.MsgRecipient == "API module" {
		switch res.MsgSection {
		case "source control":

		case "source telemetry":

		case "filtration":

		case "download":

		case "information search results":

		case "error notification":
			en, ok := res.AdvancedOptions.(configure.ErrorNotification)
			if !ok {
				return errors.New("type conversion error section type 'error notification'" + funcName)
			}

			_ = saveMessageApp.LogMessage("error", fmt.Sprint(en.ErrorBody))

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType:        "danger",
				MsgDescription: "ошибка при обработке запроса к базе данных",
			}

			notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, res.IDClientAPI)

		case "message notification":
			mn, ok := res.AdvancedOptions.(configure.MessageNotification)
			if !ok {
				return errors.New("type conversion error section type 'message notification'" + funcName)
			}

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType:        mn.CriticalityMessage,
				MsgDescription: mn.HumanDescriptionNotification,
				Sources:        mn.Sources,
			}

			notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, res.IDClientAPI)
		}
	} else if res.MsgRecipient == "NI module" {
		switch res.MsgSection {
		case "source list":
			chanToNI <- configure.MsgBetweenCoreAndNI{
				Section:         "source control",
				Command:         "create list",
				AdvancedOptions: res.AdvancedOptions,
			}

		case "source control":

		case "filtration":

		case "download":
		}
	} else if res.MsgRecipient == "Core module" {
		fmt.Printf("RESIPENT MSG FOR CORE %v", res)

		if res.MsgSection == "error notification" {
			//если сообщение об ошибке только для ядра приложения
			if en, ok := res.AdvancedOptions.(configure.ErrorNotification); ok {
				return en.ErrorBody
			}
		}
	}

	return nil
}
