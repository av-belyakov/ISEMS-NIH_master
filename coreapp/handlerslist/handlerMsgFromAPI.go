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
func HandlerMsgFromAPI(chanToNI chan<- configure.MsgBetweenCoreAndNI, msg *configure.MsgBetweenCoreAndAPI, smt *configure.StoringMemoryTask, chanToDB chan<- configure.MsgBetweenCoreAndDB) (string, notifications.NotificationSettingsToClientAPI, error) {
	fmt.Println("--- START function 'HandlerMsgFromAPI'...")

	funcName := ", function 'HeaderMsgFromAPI'"
	msgc := configure.MsgCommon{}

	nsSuccess := notifications.NotificationSettingsToClientAPI{}

	nsErrJSON := notifications.NotificationSettingsToClientAPI{
		MsgType:        "danger",
		MsgDescription: "получен некорректный формат JSON сообщения",
	}

	//	errBody := errors.New("received incorrect JSON messages, function 'HeaderMsgFromAPI'")

	msgJSON, ok := msg.MsgJSON.([]byte)
	if !ok {
		return msgc.ClientTaskID, nsErrJSON, errors.New("bad cast type JSON messages" + funcName)
	}

	if err := json.Unmarshal(msgJSON, &msgc); err != nil {
		return msgc.ClientTaskID, nsErrJSON, err
	}

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInsturction == "send new source list" {

				var scmo configure.SourceControlMsgOptions
				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					return msgc.ClientTaskID, nsErrJSON, err
				}

				fmt.Printf("From API resived msg %q", msg)

				//добавляем новую задачу
				taskID := smt.AddStoringMemoryTask(configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				chanToNI <- configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					Section:         "source control",
					Command:         "load list",
					AdvancedOptions: scmo,
				}

				return "", nsSuccess, nil
			}

			return "", nsErrJSON, errors.New("in the json message is not found the right option for 'MsgInsturction'" + funcName)
		}

		return "", nsErrJSON, errors.New("in the json message is not found the right option for 'MsgSection'" + funcName)
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
			return "", nsSuccess, nil

		case "filtration control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'filtration control'")

			return "", nsSuccess, nil

		case "download control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'download control'")

			return "", nsSuccess, nil

		case "information search control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'information search control'")

			return "", nsSuccess, nil
		}
	}

	panic("unattainable horses of the function are obtained")
}
