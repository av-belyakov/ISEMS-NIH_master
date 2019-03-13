package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"

	"ISEMS-NIH_master/configure"
)

//HandlerMsgFromAPI обработчик сообщений приходящих от модуля API
func HandlerMsgFromAPI(chanToNI chan<- configure.MsgBetweenCoreAndNI, msg *configure.MsgBetweenCoreAndAPI, smt *configure.StoringMemoryTask, chanToDB chan<- configure.MsgBetweenCoreAndDB) (string, error) {
	fmt.Println("--- START function 'HandlerMsgFromAPI'...")

	fmt.Println("RESIVED MSG from:", msg.MsgGenerator)
	fmt.Println("resipent MSG:", msg.MsgRecipient)
	fmt.Println("ID client API", msg.IDClientAPI)
	/*
				MsgGenerator: "API module",
				MsgRecipient: "Core module",
				IDClientAPI:  clientID,
				MsgJSON:      message,


				MsgType        string `json:"msgType"`
			MsgSection     string `json:"msgSection"`
			MsgInsturction string `json:"msgInsturction"`
			MsgOptions     []byte `json:"msgOptions"`

			msgType: 'information',
		msgSection: 'source control',
		msgInsturction: 'send new source list',
		msgOptions: {
			sourceList: [
				{
					id: <уникальный цифровой идентификатор источника>,
					actionType: 'none',
					argument: {
						ip: <ip адрес>,
						token: <уникальный строковый токен для аутентификации>,
						settings: {
							enableTelemetry: true/false,
							maxCountProcessFiltering: <число 1-10>,
							storageFolders: [] //список директорий с файлами сет. трафика
							}
						}
				}
			]
		}
	*/

	msgc := configure.MsgCommon{}

	m := "получен некорректный формат JSON сообщения"
	if err := json.Unmarshal(msg.MsgJSON, &msgc); err != nil {

		fmt.Println("ERROR parse json msg", err)

		return m, err
	}

	//добавляем новую задачу
	taskID, ok := smt.AddStoringMemoryTask(configure.TaskDescription{})
	if !ok {
		me := "задача с заданным идентификатором существует, добавление задачи невозможно"
		return me, errors.New(me)
	}

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInsturction == "send new source list" {
				/*jsonMsg := configure.SourceControlMsgTypeFromAPI{}
				if err := json.Unmarshal(msgc.MsgOptions, &jsonMsg); err != nil {
					return m, err
				}*/

				mo, ok := msgc.MsgOptions.(configure.SourceControlMsgTypeFromAPI)
				if !ok {
					fmt.Println("написать что некорректный JSON объект")
				}

				chanToNI <- configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					Section:         "source control",
					Command:         "load list",
					AdvancedOptions: mo,
				}
			}
		}
	}

	if msgc.MsgType == "command" {
		switch msgc.MsgSection {
		case "source control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'source control'")

		case "filtration control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'filtration control'")

		case "download control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'download control'")

		case "information search control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'information search control'")

		}
	}

	return "", nil
}
