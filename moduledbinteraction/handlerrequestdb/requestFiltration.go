package handlerrequestdb

import (
	"fmt"

	"github.com/mongodb/mongo-go-driver/mongo"
	//"github.com/mongodb/mongo-go-driver/mongo/options"
	//"github.com/mongodb/mongo-go-driver/bson"

	"ISEMS-NIH_master/configure"
)

//QueryCollectionFiltration значения для работы с коллекцией источников
type QueryCollectionFiltration struct {
	NameDB, CollectionName string
	ConnectDB              *mongo.Client
}

//CreateNewFiltrationTask запись информации о новой фильтрации
//обрабатывается при получении запроса на фильтрацию
func (qcs *QueryCollectionFiltration) CreateNewFiltrationTask(chanIn chan<- *configure.MsgBetweenCoreAndDB, req *configure.MsgBetweenCoreAndDB) {
	fmt.Println("START function 'CreateNewFiltrationTask'...")

}
