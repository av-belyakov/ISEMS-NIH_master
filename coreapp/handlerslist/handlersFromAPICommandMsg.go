package handlerslist

/*
* Модуль обработки команд поступающих через API App
*
* Версия 0.1, дата релиза 04.03.2019
* */

import (
	"ISEMS-NIH_master/configure"
)

//HandlerSourceControlFromAPI обработчик команды типа 'source_control' принятой от модуля API
func HandlerSourceControlFromAPI(chanToDB chan<- configure.MsgBetweenCoreAndDB, idClientAPI string, advancedOptions interface{}) error {

	return nil
}

//HandlerFiltrationFromAPI обработчик команды типа 'filtration' принятой от модуля API
func HandlerFiltrationFromAPI(chanToDB chan<- configure.MsgBetweenCoreAndDB, idClientAPI string, advancedOptions interface{}) error {

	return nil
}

//HandlerDownloadFromAPI обработчик команды типа 'download' принятой от модуля API
func HandlerDownloadFromAPI(chanToDB chan<- configure.MsgBetweenCoreAndDB, idClientAPI string, advancedOptions interface{}) error {

	return nil
}

//HandlerInformationSearchFromAPI обработчик команды типа 'information_search' принятой от модуля API
func HandlerInformationSearchFromAPI(chanToDB chan<- configure.MsgBetweenCoreAndDB, idClientAPI string, advancedOptions interface{}) error {

	return nil
}
