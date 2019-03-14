package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"

	"ISEMS-NIH_master/common/notifications"
	"ISEMS-NIH_master/configure"
)

//HandlerMsgFromAPI обработчик сообщений приходящих от модуля API
func HandlerMsgFromAPI(chanToNI chan<- configure.MsgBetweenCoreAndNI, msg *configure.MsgBetweenCoreAndAPI, smt *configure.StoringMemoryTask, chanToDB chan<- configure.MsgBetweenCoreAndDB) (string, notifications.NotificationSettingsToClientAPI, error) {
	fmt.Println("--- START function 'HandlerMsgFromAPI'...")

	msgc := configure.MsgCommon{}

	nsSuccess := notifications.NotificationSettingsToClientAPI{}

	nsErrJSON := notifications.NotificationSettingsToClientAPI{
		MsgType:        "danger",
		MsgDescription: "получен некорректный формат JSON сообщения",
	}

	//	errBody := errors.New("received incorrect JSON messages, function 'HeaderMsgFromAPI'")

	if err := json.Unmarshal(msg.MsgJSON.([]byte), &msgc); err != nil {
		return msgc.ClientTaskID, nsErrJSON, err
	}

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInsturction == "send new source list" {

				var scmo configure.SourceControlMsgOptions
				if msgjson, ok := msg.MsgJSON.([]byte); ok {
					if err := json.Unmarshal(msgjson, &scmo); err != nil {
						return msgc.ClientTaskID, nsErrJSON, err
					}

					fmt.Println("______ load source list from client API _________")
					fmt.Printf("%v\n", scmo)
					fmt.Println("____________________________________")

					chanToNI <- configure.MsgBetweenCoreAndNI{
						Section:         "source control",
						Command:         "load list",
						AdvancedOptions: scmo,
					}

					return "", nsSuccess, nil
				}
			}
			fmt.Println("*** 1111 ***")

			nsErrJSON.MsgDescription = "msg 1111"
			return "", nsErrJSON, errors.New("in the json message is not found the right option for 'MsgInsturction', function 'HeaderMsgFromAPI'")
		}
		fmt.Println("***** 2222 ****")

		nsErrJSON.MsgDescription = "msg 2222"
		return "", nsErrJSON, errors.New("in the json message is not found the right option for 'MsgSection', function 'HeaderMsgFromAPI'")
	}

	if msgc.MsgType == "command" {
		//добавляем новую задачу
		taskID, ok := smt.AddStoringMemoryTask(configure.TaskDescription{})
		if !ok {
			nsErrJSON.MsgDescription = "задача с заданным идентификатором существует, добавление задачи невозможно"

			return msgc.ClientTaskID, nsErrJSON, errors.New("a task with the specified ID exists, and you cannot add a task, function 'HeaderMsgFromAPI'")
		}

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
