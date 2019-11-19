package moduledbinteraction

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduledbinteraction/handlerrequestdb"
	"ISEMS-NIH_master/savemessageapp"

	"github.com/mongodb/mongo-go-driver/mongo"
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
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
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
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
		}
	}
}
