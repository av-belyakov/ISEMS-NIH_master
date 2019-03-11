package moduledbinteraction

/*
* Модуль взаимодействия с БД MongoDB
*
* Версия 0.1, дата релиза 04.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduledbinteraction/handlerrequestdb"
)

//MainDBInteraction обработка запросов к БД
func MainDBInteraction(nameDB string, linkConnection *configure.MongoDBConnect, ism *configure.InformationStoringMemory) (chanOut, chanIn chan configure.MsgBetweenCoreAndDB) {
	fmt.Println("START module 'MainDBInteraction'...")

	//инициализируем каналы для обмена данными между БД м ядром приложения
	chanOut = make(chan configure.MsgBetweenCoreAndDB)
	chanIn = make(chan configure.MsgBetweenCoreAndDB)

	qcs := handlerrequestdb.QueryCollectionSources{
		NameDB:         nameDB,
		CollectionName: "sources_list",
		ConnectDB:      linkConnection.Connect,
	}

	//при старте запрашиваем список источников с которыми устанавливаются подключения
	go qcs.GetAllSourcesList(chanIn, &configure.MsgBetweenCoreAndDB{
		MsgGenerator: "Core module",
		MsgRecipient: "DB module",
		MsgDirection: "request",
		DataType:     "sources_control",
	})

	go RouteRequest(chanIn, nameDB, linkConnection, ism, chanOut)

	return chanOut, chanIn
}