package handlerslist

/*
* Модуль содержит набор обработчиков результатов запросов к БД
*
* Версия 0.1, дата релиза 04.03.2019
* */

import (
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
	"errors"
	"fmt"
)

//HandlerSourcesControlFromDB обработчик сообщения 'sources_list'
func HandlerSourcesControlFromDB(chanOutAPI chan<- configure.MsgBetweenCoreAndAPI, chanOutNI chan<- configure.MsgBetweenCoreAndNI, msgFromDB configure.MsgBetweenCoreAndDB) error {

	return nil
}

//HandlerChangeStatusSourceFromDB обработчик сообщения типа 'change_status_source'
func HandlerChangeStatusSourceFromDB(chanOutAPI chan<- configure.MsgBetweenCoreAndAPI, chanOutNI chan<- configure.MsgBetweenCoreAndNI, msgFromDB configure.MsgBetweenCoreAndDB) error {

	return nil
}

//HandlerFiltrationFromDB обработчик сообщения типа 'filtration'
func HandlerFiltrationFromDB(chanOutAPI chan<- configure.MsgBetweenCoreAndAPI, chanOutNI chan<- configure.MsgBetweenCoreAndNI, msgFromDB configure.MsgBetweenCoreAndDB) error {

	return nil
}

//HandlerDownloadFromDB обработчик сообщения типа 'download'
func HandlerDownloadFromDB(chanOutAPI chan<- configure.MsgBetweenCoreAndAPI, chanOutNI chan<- configure.MsgBetweenCoreAndNI, msgFromDB configure.MsgBetweenCoreAndDB) error {

	return nil
}

//HandlerMsgInfoSearchResultsFromDB обработчик сообщения типа 'information_search_results'
func HandlerMsgInfoSearchResultsFromDB(chanOutAPI chan<- configure.MsgBetweenCoreAndAPI, chanOutNI chan<- configure.MsgBetweenCoreAndNI, msgFromDB configure.MsgBetweenCoreAndDB) error {

	return nil
}

//HandlerErrorNotificationFromDB обработчик сообщения типа 'error_notification'
func HandlerErrorNotificationFromDB(chanOutAPI chan<- configure.MsgBetweenCoreAndAPI, chanOutNI chan<- configure.MsgBetweenCoreAndNI, msgFromDB configure.MsgBetweenCoreAndDB) error {
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	errorMsg, ok := msgFromDB.AdvancedOptions.(configure.ErrorNotification)
	if !ok {
		err := errors.New("it is impossible to convert the type in a function 'HandlerErrorNotificationFromDB'")
		_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))

		return err
	}
	_ = saveMessageApp.LogMessage("err", fmt.Sprint(errorMsg.ErrorBody))

	if msgFromDB.MsgDirection == "API module" {
		chanOutAPI <- configure.MsgBetweenCoreAndAPI{
			MsgGenerator: "Core module",
			MsgType:      "information",
			MsgSection:   "error notification",
			IDClientAPI:  msgFromDB.IDClientAPI,
			AdvancedOptions: configure.ErrorNotification{
				SourceReport:          errorMsg.SourceReport,
				HumanDescriptionError: errorMsg.HumanDescriptionError,
			},
		}
	}

	return nil
}

//HandlerSourceTelemetryFromDB обработчик сообщения типа 'source_telemetry'
func HandlerSourceTelemetryFromDB(chanOutAPI chan<- configure.MsgBetweenCoreAndAPI, chanOutNI chan<- configure.MsgBetweenCoreAndNI, msgFromDB configure.MsgBetweenCoreAndDB) error {

	return nil
}
