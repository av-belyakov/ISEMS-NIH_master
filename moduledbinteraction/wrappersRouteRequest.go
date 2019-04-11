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
	qcs := handlerrequestdb.QueryCollectionSources{
		NameDB:         wr.NameDB,
		CollectionName: "sources_list",
		ConnectDB:      wr.ConnectDB,
	}

	switch msg.Instruction {
	case "find_all":
		qcs.GetAllSourcesList(wr.ChanIn, msg)

	case "insert":
		fmt.Println("func 'WrapperFuncSourceControl' RESIVED COMMAND 'INSERT'")

		qcs.InsertListSources(wr.ChanIn, msg)

	case "update":
		fmt.Println("func 'WrapperFuncSourceControl' RESIVED COMMAND 'UPDATE'")

		fmt.Printf("%v\n", msg.AdvancedOptions)

		qcs.UpdateSourceToSourcesList(wr.ChanIn, msg)

	case "delete":
		fmt.Println("func 'WrapperFuncSourceControl' RESIVED COMMAND 'DELETE'")

		qcs.DelSourceToSourcesList(wr.ChanIn, msg)

	}
}

//WrapperFuncFiltration обработка запросов по фильтрации
func (wr *WrappersRouteRequest) WrapperFuncFiltration(msg *configure.MsgBetweenCoreAndDB) {

}

//WrapperFuncDownload обработка запросов по скачиванию файлов
func (wr *WrappersRouteRequest) WrapperFuncDownload(msg *configure.MsgBetweenCoreAndDB) {

}
