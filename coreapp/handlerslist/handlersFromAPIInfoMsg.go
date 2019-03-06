package handlerslist

/*
* Модуль обработки информационных сообщений поступающих от API App
*
* Версия 0.1, дата релиза 04.03.2019
* */

import (
	"errors"

	"ISEMS-NIH_master/configure"
)

//HandlerStatusSourceFromAPI обработчик для добавления нового списка источников
func HandlerStatusSourceFromAPI(chanToDB chan<- configure.MsgBetweenCoreAndDB, idClientAPI string, advancedOptions interface{}) error {
	msg, ok := advancedOptions.(configure.MsgInfoChangeStatusSource)
	if !ok {
		err := errors.New("it is impossible to convert the type in a function 'HandlerStatusSourceFromAPI'")

		return err
	}

	sendMsg := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    "Core module",
		MsgRecipient:    "DB module",
		MsgDirection:    "request",
		MsgSection:      "source control",
		IDClientAPI:     idClientAPI,
		AdvancedOptions: advancedOptions,
	}

	if msg.SourceListIsExist {
		sendMsg.Instruction = "insert"
	}

	chanToDB <- sendMsg

	return nil
}
