package moduledbinteraction

import (
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//RouteRequest маршрутизатор запросов
func RouteRequest(
	chanIn chan *configure.MsgBetweenCoreAndDB,
	nameDB string,
	linkConnection *configure.MongoDBConnect,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanOut <-chan *configure.MsgBetweenCoreAndDB) {

	wrapperFunc := WrappersRouteRequest{
		NameDB:    nameDB,
		ConnectDB: linkConnection.Connect,
		ChanIn:    chanIn,
	}

	for msg := range chanOut {
		switch msg.MsgSection {
		case "source control":
			go wrapperFunc.WrapperFuncSourceControl(msg, saveMessageApp)

		case "source telemetry":

		case "filtration control":
			go wrapperFunc.WrapperFuncFiltration(msg, smt, qts, saveMessageApp)

		case "download control":
			go wrapperFunc.WrapperFuncDownload(msg, smt, qts, saveMessageApp)

		case "error notification":

		case "information search":

		}
	}
}
