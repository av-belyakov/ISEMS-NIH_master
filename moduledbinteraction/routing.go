package moduledbinteraction

/*
* Маршрутизация запросов получаемых через канал
*
* Версия 0.1, дата релиза 05.03.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//RouteRequest маршрутизатор запросов
func RouteRequest(chanIn chan configure.MsgBetweenCoreAndDB, nameDB string, linkConnection *configure.MongoDBConnect, chanOut <-chan configure.MsgBetweenCoreAndDB) {
	fmt.Println("START module 'RouteRequest' (module DBInteraction)...")

	/*
			   configure.MsgBetweenCoreAndDB{
			   		MsgGenerator:    "Core module",
			   		MsgRecipient:    "DB module",
			   		MsgType:    "request",
					   DataType:		 "sources_list/change_status_sources",
					   Insturction: insert
		//  - find
		//  - update
		//  - delete
			   		IDClientAPI:     idClientAPI,
			   		AdvancedOptions: advancedOptions,
			   	}


	*/
	wrapperFunc := WrappersRouteRequest{
		NameDB:    nameDB,
		ConnectDB: linkConnection.Connect,
		ChanIn:    chanIn,
	}

	for msg := range chanOut {
		switch msg.MsgSection {
		case "source control":
			go wrapperFunc.WrapperFuncSourceControl(&msg)
		case "source telemetry":

		case "filtration":

		case "dawnload":

		case "error notification":

		case "information search":

		}
	}

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
