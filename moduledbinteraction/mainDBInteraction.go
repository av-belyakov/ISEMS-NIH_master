package moduledbinteraction

/*
* Модуль взаимодействия с БД MongoDB
* */

import (
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//MainDBInteraction обработка запросов к БД
func MainDBInteraction(
	nameDB string,
	linkConnection *configure.MongoDBConnect,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) (chanOut, chanIn chan *configure.MsgBetweenCoreAndDB) {

	//инициализируем каналы для обмена данными между БД м ядром приложения
	chanOut = make(chan *configure.MsgBetweenCoreAndDB) //->БД
	chanIn = make(chan *configure.MsgBetweenCoreAndDB)  //<-БД

	go RouteRequest(chanIn, nameDB, linkConnection, smt, qts, saveMessageApp, chanOut)

	return chanOut, chanIn
}
