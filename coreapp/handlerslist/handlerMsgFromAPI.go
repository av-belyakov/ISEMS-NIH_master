package handlerslist

import (
	"encoding/json"
	"fmt"
	"time"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromAPI обработчик сообщений приходящих от модуля API
func HandlerMsgFromAPI(
	chanToNI chan<- *configure.MsgBetweenCoreAndNI,
	msg *configure.MsgBetweenCoreAndAPI,
	smt *configure.StoringMemoryTask,
	chanToDB chan<- *configure.MsgBetweenCoreAndDB,
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) {

	fmt.Println("--- START function 'HandlerMsgFromAPI'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	funcName := ", function 'HeaderMsgFromAPI'"
	msgc := configure.MsgCommon{}

	nsErrJSON := notifications.NotificationSettingsToClientAPI{
		MsgType:        "danger",
		MsgDescription: "получен некорректный формат JSON сообщения",
	}

	//	errBody := errors.New("received incorrect JSON messages, function 'HeaderMsgFromAPI'")

	msgJSON, ok := msg.MsgJSON.([]byte)
	if !ok {
		notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, "", msg.IDClientAPI)
		_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

		return
	}

	if err := json.Unmarshal(msgJSON, &msgc); err != nil {
		notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, "", msg.IDClientAPI)
		_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

		return
	}

	//логируем запросы клиентов
	_ = saveMessageApp.LogMessage("requests", "client name: '"+msg.ClientName+"' ("+msg.ClientIP+"), request: type = "+msgc.MsgType+", section = "+msgc.MsgSection+", instruction = "+msgc.MsgInsturction+", client task ID = "+msgc.ClientTaskID)

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInsturction == "send new source list" {
				var scmo configure.SourceControlMsgOptions
				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				//добавляем новую задачу
				taskID := smt.AddStoringMemoryTask(configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				chanToNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					ClientName:      msg.ClientName,
					Section:         "source control",
					Command:         "load list",
					AdvancedOptions: scmo.MsgOptions,
				}

				return
			}

			notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
			_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgSection'"+funcName)

			return
		}

		notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
		_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgSection'"+funcName)

		return
	}

	if msgc.MsgType == "command" {
		switch msgc.MsgSection {
		case "source control":
			//получить актуальный список источников
			if msgc.MsgInsturction == "get an updated list of sources" {

				//добавляем новую задачу
				taskID := smt.AddStoringMemoryTask(configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				chanToDB <- &configure.MsgBetweenCoreAndDB{
					MsgGenerator: "API module",
					MsgRecipient: "DB module",
					MsgSection:   "source control",
					Instruction:  "find_all",
					TaskID:       taskID,
				}

				return
			}

			//выполнить какие либо действия над источниками
			if msgc.MsgInsturction == "performing an action" {
				var scmo configure.SourceControlMsgOptions
				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, "", msg.IDClientAPI)
					_ = saveMessageApp.LogMessage("error", "bad cast type JSON messages"+funcName)

					return
				}

				//добавляем новую задачу
				taskID := smt.AddStoringMemoryTask(configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				chanToNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					ClientName:      msg.ClientName,
					Section:         "source control",
					Command:         "perform actions on sources",
					AdvancedOptions: scmo,
				}

				return
			}

			notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
			_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgInstruction'"+funcName)

			return

		case "filtration control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'filtration control'")

			return

		case "download control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'download control'")

			return

		case "information search control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'information search control'")

			return

		default:
			notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)
			_ = saveMessageApp.LogMessage("error", "in the json message is not found the right option for 'MsgInstruction'"+funcName)

			return
		}
	}
}
