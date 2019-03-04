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

//HandlerStatusSourceFromAPI обработчик сообщений об изменении информации по источникам
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
		IDClientAPI:     idClientAPI,
		AdvancedOptions: advancedOptions,
	}

	//передается ПОЛНЫЙ список источников
	if msg.SourceListIsExist {
		sendMsg.DataType = "sources_list"

		chanToDB <- sendMsg

		return nil
	}

	//передается только список источников данные о которых были изменены
	sendMsg.DataType = "change_status_source"

	return nil
}
