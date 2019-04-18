package moduledbinteraction

/*
* Набор функций оберток для обработки запросов к БД
*
* Версия 0.2, дата релиза 27.03.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduledbinteraction/handlerrequestdb"

	"github.com/mongodb/mongo-go-driver/mongo"
)

//WrappersRouteRequest обертки для запросов
type WrappersRouteRequest struct {
	NameDB    string
	ConnectDB *mongo.Client
	ChanIn    chan *configure.MsgBetweenCoreAndDB
}

//WrapperFuncSourceControl обработка запросов об источниках
func (wr *WrappersRouteRequest) WrapperFuncSourceControl(msg *configure.MsgBetweenCoreAndDB) {
	qp := handlerrequestdb.QueryParameters{
		NameDB:         wr.NameDB,
		CollectionName: "sources_list",
		ConnectDB:      wr.ConnectDB,
	}

	switch msg.Instruction {
	case "find_all":
		handlerrequestdb.GetAllSourcesList(wr.ChanIn, msg, qp)

	case "insert":
		fmt.Println("func 'WrapperFuncSourceControl' RESIVED COMMAND 'INSERT'")

		handlerrequestdb.InsertListSources(wr.ChanIn, msg, qp)

	case "update":
		fmt.Println("func 'WrapperFuncSourceControl' RESIVED COMMAND 'UPDATE'")

		fmt.Printf("%v\n", msg.AdvancedOptions)

		handlerrequestdb.UpdateSourceToSourcesList(wr.ChanIn, msg, qp)

	case "delete":
		fmt.Println("func 'WrapperFuncSourceControl' RESIVED COMMAND 'DELETE'")

		handlerrequestdb.DelSourceToSourcesList(wr.ChanIn, msg, qp)

	}
}

//WrapperFuncFiltration обработка запросов по фильтрации
func (wr *WrappersRouteRequest) WrapperFuncFiltration(msg *configure.MsgBetweenCoreAndDB) {
	qp := handlerrequestdb.QueryParameters{
		NameDB:         wr.NameDB,
		CollectionName: "filter_task_list",
		ConnectDB:      wr.ConnectDB,
	}

	switch msg.Instruction {
	case "insert":
		fmt.Println("func 'WrapperFuncFiltration' RESIVED COMMAND 'INSERT'")

		handlerrequestdb.CreateNewFiltrationTask(wr.ChanIn, msg, qp)

	case "find":
		fmt.Println("func 'WrapperFuncFiltration' RESIVED COMMAND 'FIND'")

	case "find_all":
		fmt.Println("func 'WrapperFuncFiltration' RESIVED COMMAND 'FIND_ALL'")

	case "update":
		fmt.Println("func 'WrapperFuncFiltration' RESIVED COMMAND 'UPDATE'")

	}
}

//WrapperFuncDownload обработка запросов по скачиванию файлов
func (wr *WrappersRouteRequest) WrapperFuncDownload(msg *configure.MsgBetweenCoreAndDB) {

}
