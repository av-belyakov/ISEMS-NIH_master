package moduledbinteraction

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduledbinteraction/handlerrequestdb"
	"ISEMS-NIH_master/savemessageapp"

	"go.mongodb.org/mongo-driver/mongo"
)

//WrappersRouteRequest обертки для запросов
type WrappersRouteRequest struct {
	NameDB    string
	ConnectDB *mongo.Client
	ChanIn    chan *configure.MsgBetweenCoreAndDB
}

//WrapperFuncSourceControl обработка запросов об источниках
func (wr *WrappersRouteRequest) WrapperFuncSourceControl(msg *configure.MsgBetweenCoreAndDB, saveMessageApp *savemessageapp.PathDirLocationLogFiles) {
	qp := handlerrequestdb.QueryParameters{
		NameDB:         wr.NameDB,
		CollectionName: "sources_list",
		ConnectDB:      wr.ConnectDB,
	}

	switch msg.Instruction {
	case "find_all":
		handlerrequestdb.GetAllSourcesList(wr.ChanIn, msg, qp)

	case "find_all_for_client_API":
		handlerrequestdb.GetAllSourcesListForClientAPI(wr.ChanIn, msg, qp)

	case "insert":
		handlerrequestdb.InsertListSources(wr.ChanIn, msg, qp)

	case "update":
		handlerrequestdb.UpdateSourceToSourcesList(wr.ChanIn, msg, qp)

	case "delete":
		handlerrequestdb.DelSourceToSourcesList(wr.ChanIn, msg, qp)

	}
}

//WrapperFuncFiltration обработка запросов по фильтрации
func (wr *WrappersRouteRequest) WrapperFuncFiltration(
	msg *configure.MsgBetweenCoreAndDB,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) {

	qp := handlerrequestdb.QueryParameters{
		NameDB:         wr.NameDB,
		CollectionName: "task_list",
		ConnectDB:      wr.ConnectDB,
	}

	switch msg.Instruction {
	case "insert":
		handlerrequestdb.CreateNewFiltrationTask(wr.ChanIn, msg, qp, qts)

	case "find":
		fmt.Println("func 'WrapperFuncFiltration' RESIVED COMMAND 'FIND'")

	case "find_all":
		fmt.Println("func 'WrapperFuncFiltration' RESIVED COMMAND 'FIND_ALL'")

	case "update":
		if err := handlerrequestdb.UpdateParametersFiltrationTask(wr.ChanIn, msg, qp, smt); err != nil {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    "WrapperFuncFiltration",
			})
		}

	case "restore":
		if err := handlerrequestdb.RestoreParametersFiltrationTask(wr.ChanIn, msg, qp, smt); err != nil {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    "WrapperFuncFiltration",
			})
		}
	}
}

//WrapperFuncDownload обработка запросов по скачиванию файлов
func (wr *WrappersRouteRequest) WrapperFuncDownload(
	msg *configure.MsgBetweenCoreAndDB,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) {

	qp := handlerrequestdb.QueryParameters{
		NameDB:         wr.NameDB,
		CollectionName: "task_list",
		ConnectDB:      wr.ConnectDB,
	}

	switch msg.Instruction {
	case "finding information about a task":
		handlerrequestdb.FindingInformationAboutTask(wr.ChanIn, msg, qp)

	case "update":
		if err := handlerrequestdb.UpdateInformationAboutTask(msg, qp, smt); err != nil {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    "WrapperFuncDownload",
			})
		}
	}
}

//WrapperFuncSearch обработка запросов поиска
func (wr *WrappersRouteRequest) WrapperFuncSearch(
	msg *configure.MsgBetweenCoreAndDB,
	tssq *configure.TemporaryStorageSearchQueries,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) {

	qp := handlerrequestdb.QueryParameters{
		NameDB:         wr.NameDB,
		CollectionName: "task_list",
		ConnectDB:      wr.ConnectDB,
	}

	switch msg.Instruction {
	case "search common information":
		handlerrequestdb.SearchShortInformationAboutTasks(wr.ChanIn, msg, tssq, qp)

	case "search full information by task ID":
		handlerrequestdb.SearchFullInformationAboutTasks(wr.ChanIn, msg, qp)

	case "get part of the list files":
		handlerrequestdb.GetListFoundFiles(wr.ChanIn, msg, qp)

	case "get info from mark an task as completed":
		handlerrequestdb.GetInfoTaskFromMarkTaskCompleteProcess(wr.ChanIn, msg, qp)

	case "mark an task as completed":
		handlerrequestdb.MarkTaskCompleteProcess(wr.ChanIn, msg, qp)

	case "delete all information about a task":
		handlerrequestdb.DeleteInformationAboutTask(wr.ChanIn, msg, qp)

	case "get common analytics information about task ID":
		handlerrequestdb.CommonAnalyticsInformationAboutTaskID(wr.ChanIn, msg, qp)

	}
}
