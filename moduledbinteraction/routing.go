package moduledbinteraction

/*
* Маршрутизация запросов получаемых через канал
*
* Версия 0.1, дата релиза 04.03.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduledbinteraction/handlerrequestdb"
)

func getSourcesList(chanIn chan<- configure.MsgBetweenCoreAndDB, qcs handlerrequestdb.QueryCollectionSources) {
	msgResult := configure.MsgBetweenCoreAndDB{
		MsgGenerator: "DB module",
		MsgRecipient: "Core module",
		MsgDirection: "response",
	}

	sourcesList, err := qcs.FindAll()
	if err != nil {
		msgResult.DataType = "error_notification"
		msgResult.AdvancedOptions = configure.ErrorNotification{
			SourceReport: "DB module",
			ErrorBody:    err,
		}

		chanIn <- msgResult

		return
	}

	fmt.Println("request from DB processed success")
	fmt.Printf("%v", sourcesList)

	msgResult.DataType = "sources_list"
	msgResult.AdvancedOptions = sourcesList

	//отправка списка источников маршрутизатору ядра приложения
	chanIn <- msgResult
}

//RouteRequest маршрутизатор запросов
func RouteRequest(chanIn chan<- configure.MsgBetweenCoreAndDB, nameDB string, linkConnection *configure.MongoDBConnect, ism *configure.InformationStoringMemory, chanOut <-chan configure.MsgBetweenCoreAndDB) {
	fmt.Println("START module 'RouteRequest' (module DBInteraction)...")

	qcs := handlerrequestdb.QueryCollectionSources{
		NameDB:         nameDB,
		CollectionName: "sources_list",
		ConnectDB:      linkConnection.Connect,
	}

	//при старте запрашиваем список источников с которыми устанавливаются подключения
	go getSourcesList(chanIn, qcs)

	//handlersRequest := map[string]

	for msg := range chanOut {
		fmt.Println("resived request to DB")
		fmt.Printf("%v", msg)

		/* ОБРАБОТКА ЗАПРОСОВ К БД ПОЛУЧАЕМЫХ от CoreRoute */

		//типы запросов
		//	- sources_list (sources_list collection)
		//  - change_status_source (sources_list collection)
		//  - source_telemetry (source_telemetry collection)
		//  - filtration (filtartion collection)
		//  - download (download collection)
		//  - error_notification (error_notification collection)
		//  - information_search (all collections)
	}
}
