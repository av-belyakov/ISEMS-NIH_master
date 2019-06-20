package moduledbinteraction

/*
* Модуль взаимодействия с БД MongoDB
*
* Версия 0.11, дата релиза 05.06.2019
* */

import (
	"ISEMS-NIH_master/configure"
)

//MainDBInteraction обработка запросов к БД
func MainDBInteraction(
	nameDB string,
	linkConnection *configure.MongoDBConnect,
	smt *configure.StoringMemoryTask) (chanOut, chanIn chan *configure.MsgBetweenCoreAndDB) {

	//инициализируем каналы для обмена данными между БД м ядром приложения
	chanOut = make(chan *configure.MsgBetweenCoreAndDB) //->БД
	chanIn = make(chan *configure.MsgBetweenCoreAndDB)  //<-БД

	go RouteRequest(chanIn, nameDB, linkConnection, smt, chanOut)

	return chanOut, chanIn
}
