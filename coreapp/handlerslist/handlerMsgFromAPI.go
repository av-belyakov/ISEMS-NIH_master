package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
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

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInsturction == "send new source list" {
				var scmo configure.SourceControlMsgOptions
				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, "", msg.IDClientAPI)

					return errors.New("bad cast type JSON messages" + funcName)
				}

				//				fmt.Printf("From API resived msg %v\n", scmo)

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

			notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)

			return errors.New("in the json message is not found the right option for 'MsgInsturction'" + funcName)
		}

		notifications.SendNotificationToClientAPI(chanToAPI, nsErrJSON, msgc.ClientTaskID, msg.IDClientAPI)

		return errors.New("in the json message is not found the right option for 'MsgSection'" + funcName)
	}

	if msgc.MsgType == "command" {
		//добавляем новую задачу
		taskID := smt.AddStoringMemoryTask(configure.TaskDescription{})

		fmt.Println("task ID =", taskID)

		switch msgc.MsgSection {
		case "source control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'source control'")

			/*mo, ok := msgc.MsgOptions.(configure.SourceControlMsgTypeFromAPI)
			if !ok {
				return msgc.ClientTaskID, nsErrJSON, errors.New("received incorrect JSON messages, function 'HeaderMsgFromAPI'")
			}

			chanToNI <- configure.MsgBetweenCoreAndNI{
				TaskID:          taskID,
				Section:         "source control",
				Command:         "load list",
				AdvancedOptions: mo,
			}
			*/

			return nil

		case "filtration control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'filtration control'")

			return nil

		case "download control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'download control'")

			return nil

		case "information search control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'information search control'")

			return nil
		}
	}

	panic("unattainable horses of the function are obtained")
}
