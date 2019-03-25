package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromAPI обработчик сообщений приходящих от модуля API
func HandlerMsgFromAPI(
	chanToNI chan<- configure.MsgBetweenCoreAndNI,
	msg *configure.MsgBetweenCoreAndAPI,
	smt *configure.StoringMemoryTask,
	chanToDB chan<- configure.MsgBetweenCoreAndDB,
	chanToAPI chan<- configure.MsgBetweenCoreAndAPI) error {

	fmt.Println("--- START function 'HandlerMsgFromAPI'...")

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

		return errors.New("bad cast type JSON messages" + funcName)
	}

	if err := json.Unmarshal(msgJSON, &msgc); err != nil {
		notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, "", msg.IDClientAPI)

		return errors.New("bad cast type JSON messages" + funcName)
	}

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	//логируем запросы клиентов
	_ = saveMessageApp.LogMessage("requests", "client name: '"+msg.ClientName+"' ("+msg.ClientIP+"), request: type = "+msgc.MsgType+", section = "+msgc.MsgSection+", instruction = "+msgc.MsgInsturction+", client task ID = "+msgc.ClientTaskID)

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInsturction == "send new source list" {
				var scmo configure.SourceControlMsgOptions
				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, "", msg.IDClientAPI)

					return errors.New("bad cast type JSON messages" + funcName)
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

				chanToNI <- configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					ClientName:      msg.ClientName,
					Section:         "source control",
					Command:         "load list",
					AdvancedOptions: scmo.MsgOptions,
				}

				return nil
			}

			notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)

			return errors.New("in the json message is not found the right option for 'MsgSection'" + funcName)
		}

		notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)

		return errors.New("in the json message is not found the right option for 'MsgInsturction'" + funcName)
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

				chanToDB <- configure.MsgBetweenCoreAndDB{
					MsgGenerator: "API module",
					MsgRecipient: "DB module",
					MsgSection:   "source control",
					Instruction:  "find_all",
					TaskID:       taskID,
				}

				return nil
			}

			//выполнить какие либо действия над источниками
			if msgc.MsgInsturction == "performing an action" {
				var scmo configure.SourceControlMsgOptions
				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, "", msg.IDClientAPI)

					return errors.New("bad cast type JSON messages" + funcName)
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

				chanToNI <- configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					ClientName:      msg.ClientName,
					Section:         "source control",
					Command:         "perform actions on sources",
					AdvancedOptions: scmo,
				}

				return nil
			}

			notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)

			return errors.New("in the json message is not found the right option for 'MsgInstruction'" + funcName)

		case "filtration control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'filtration control'")

			return nil

		case "download control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'download control'")

			return nil

		case "information search control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'information search control'")

			return nil

		default:
			notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)

			return errors.New("in the json message is not found the right option for 'MsgSection'" + funcName)
		}
	}

	notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)

	return errors.New("in the json message is not found the right option for 'MsgType'" + funcName)
}
