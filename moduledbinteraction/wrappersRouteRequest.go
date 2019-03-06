package moduledbinteraction

import (
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduledbinteraction/handlerrequestdb"
	"fmt"

	"github.com/mongodb/mongo-go-driver/mongo"
)

/*
* Набор функций оберток для обработки запросов к БД
*
* Версия 0.1, дата релиза 05.03.2019
* */

//WrappersRouteRequest обертки для запросов
type WrappersRouteRequest struct {
	NameDB    string
	ConnectDB *mongo.Client
	ChanIn    chan configure.MsgBetweenCoreAndDB
}

//WrapperFuncSourceControl обработка запросов об источниках
func (wr *WrappersRouteRequest) WrapperFuncSourceControl(msg *configure.MsgBetweenCoreAndDB) {
	fmt.Printf("%v", msg)

	qcs := handlerrequestdb.QueryCollectionSources{
		NameDB:         wr.NameDB,
		CollectionName: "sources_list",
		ConnectDB:      wr.ConnectDB,
	}

	if msg.Instruction == "find_all" {
		qcs.GetAllSourcesList(wr.ChanIn, msg)
	}
	if msg.Instruction == "insert" {
		qcs.InserListSources(wr.ChanIn, msg)
	}
	if msg.Instruction == "update" {

	}

	if msg.Instruction == "delete" {

	}
}

//WrapperFuncFiltration обработка запросов по фильтрации
func (wr *WrappersRouteRequest) WrapperFuncFiltration(msg *configure.MsgBetweenCoreAndDB) {

}

//WrapperFuncDownload обработка запросов по скачиванию файлов
func (wr *WrappersRouteRequest) WrapperFuncDownload(msg *configure.MsgBetweenCoreAndDB) {

}
