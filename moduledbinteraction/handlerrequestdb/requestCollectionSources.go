package handlerrequestdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"

	"ISEMS-NIH_master/configure"
)

//QueryCollectionSources значения для работы с коллекцией источников
type QueryCollectionSources struct {
	NameDB, CollectionName string
	ConnectDB              *mongo.Client
}

//GetAllSourcesList получить весь список источников
func (qcs *QueryCollectionSources) GetAllSourcesList(chanIn chan<- configure.MsgBetweenCoreAndDB, req *configure.MsgBetweenCoreAndDB) {
	msgResult := configure.MsgBetweenCoreAndDB{
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgDirection: "response",
	}

	sourcesList, err := qcs.findAll()
	if err != nil {
		msgResult.MsgSection = "error_notification"
		msgResult.AdvancedOptions = configure.ErrorNotification{
			SourceReport: "DB module",
			ErrorBody:    err,
		}

		chanIn <- msgResult

		return
	}

	msgResult.MsgSection = "source list"
	msgResult.AdvancedOptions = sourcesList

	//отправка списка источников маршрутизатору ядра приложения
	chanIn <- msgResult
}

//InsertListSources добавить информацию об источниках
//которых нет в БД или параметры по которым отличаются
func (qcs *QueryCollectionSources) InsertListSources(chanIn chan<- configure.MsgBetweenCoreAndDB, req *configure.MsgBetweenCoreAndDB) {
	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgDirection: "response",
		MsgSection:   "source control",
		IDClientAPI:  req.IDClientAPI,
	}

	fmt.Printf("func 'InsertListSources' resived request from Core module %v\n", req)

	//получаем список источников
	listSources, err := qcs.findAll()
	if err != nil {
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "an error occurred while processing request get the list of sources",
			ErrorBody:             err,
		}

		chanIn <- msgRes
	}

	ao, ok := req.AdvancedOptions.(configure.MsgInfoChangeStatusSource)
	if !ok {
		errMsg := "incorrect list of sources received"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             errors.New(errMsg),
		}

		chanIn <- msgRes
	}

	if !ao.SourceListIsExist {
		return
	}

	fmt.Printf("--- source list %v", listSources)

	list := *ao.SourceList

	insertData := make([]interface{}, 0, len(list))

	//если список источников в БД пуст, добавляем все что есть
	if len(listSources) == 0 {
		for _, v := range list {
			insertData = append(insertData, v)
		}

		qcs.insertData(insertData)

		return
	}

	fmt.Println("--- Требуются доп. вычисления, поиск уикальных значений")

	for _, itemAddList := range list {
		for _, itemFindList := range listSources {
			//если источник с таким ID существует, удаляем его и заменяем новым
			if itemFindList.ID == itemAddList.ID {
				_ = qcs.deleteOneData(bson.D{{"id", itemAddList.ID}})
			}
		}

		insertData = append(insertData, itemAddList)
	}

	fmt.Println("listSourceInser = ", insertData)

	qcs.insertData(insertData)
}

//AddSourceToSourcesList добавить новые источники
func AddSourceToSourcesList(chanIn chan<- configure.MsgBetweenCoreAndDB, req *configure.MsgBetweenCoreAndDB) {

}

//UpdateSourceToSourcesList обновить информацию об источниках
func UpdateSourceToSourcesList() {}

//DelSourceToSourcesList удалить источники
func DelSourceToSourcesList() {}

//findAll найти всю информацию по всем источникам
func (qcs *QueryCollectionSources) findAll() ([]configure.InformationAboutSource, error) {
	collection := qcs.ConnectDB.Database(qcs.NameDB).Collection(qcs.CollectionName)
	options := options.Find()

	cur, err := collection.Find(context.TODO(), bson.D{{}}, options)
	if err != nil {
		return nil, err
	}

	listSources := []configure.InformationAboutSource{} //[]interface{}{}
	//получаем все ID источников
	for cur.Next(context.TODO()) {
		var model configure.InformationAboutSource
		err := cur.Decode(&model)
		if err != nil {
			return nil, err
		}

		listSources = append(listSources, model)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	cur.Close(context.TODO())

	return listSources, nil
}

//InsertData добавляет все данные
func (qcs *QueryCollectionSources) insertData(list []interface{}) (bool, error) {
	fmt.Println("===== INSERT DATA ======")
	collection := qcs.ConnectDB.Database(qcs.NameDB).Collection(qcs.CollectionName)
	if _, err := collection.InsertMany(context.TODO(), list); err != nil {
		return false, err
	}

	return true, nil
}

func (qcs *QueryCollectionSources) deleteOneData(elem interface{}) error {
	fmt.Println("===== DELETE DATA ONE ======")
	collection := qcs.ConnectDB.Database(qcs.NameDB).Collection(qcs.CollectionName)
	if _, err := collection.DeleteOne(context.TODO(), elem); err != nil {
		return err
	}

	return nil
}

func (qcs *QueryCollectionSources) deleteManyData(list interface{}) (bool, error) {
	fmt.Println("===== DELETE DATA MANY ======")
	collection := qcs.ConnectDB.Database(qcs.NameDB).Collection(qcs.CollectionName)
	if _, err := collection.DeleteMany(context.TODO(), list); err != nil {
		return false, err
	}

	return true, nil
}

//InsertListSource добавляет список источников !!! ТЕСТ !!!
/*func (qcs *QueryCollectionSources) InsertListSource() (bool, error) {
	fmt.Println("START func InserListSourcesTMPFinaly...")

	listSources := []interface{}{
		configure.InformationAboutSource{9, 3, "127.0.0.1", "fmdif3o444fdf344k0fiif", true},
		configure.InformationAboutSource{10, 3, "192.168.0.10", "fmdif3o444fdf344k0fiif", true},
		configure.InformationAboutSource{11, 3, "192.168.0.11", "ttrr9gr9r9e9f9fadx94", false},
		configure.InformationAboutSource{12, 3, "192.168.0.12", "2n3n3iixcxcc3444xfg0222", false},
		configure.InformationAboutSource{13, 3, "192.168.0.13", "osdsoc9c933cc9cn939f9f33", false},
	}

	funcInserMany := func(collection *mongo.Collection, insertListSource []interface{}) (bool, error) {
		fmt.Println("===== INSERT DATA ======")
		_, err := collection.InsertMany(context.TODO(), insertListSource)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	//ищем все источники
	collection := qcs.ConnectDB.Database(qcs.NameDB).Collection(qcs.CollectionName)
	options := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{{}} , options)
	//bson.D{{"ip", "192.168.0.10"}}, options)
	if err != nil {
		fmt.Println("ERROR FIND", err)
	}

	insertListSource := []interface{}{}
	listSourcesID := []int{}

	//получаем все ID источников
	for cur.Next(context.TODO()) {
		var im configure.InformationAboutSource
		err := cur.Decode(&im)
		if err != nil {
			fmt.Println(err)
		}

		listSourcesID = append(listSourcesID, im.ID)
	}

	if err := cur.Err(); err != nil {
		fmt.Println(err)
	}

	cur.Close(context.TODO())

	fmt.Println("---------", "LIST SOURCES FROM DB", listSourcesID, "---------")

	if len(listSourcesID) == 0 {
		return funcInserMany(collection, listSources)
	}

	//готовим insertListSources
	for _, value := range listSources {
		//контролируемое привидение типов и получаем срез id
		if im, ok := value.(configure.InformationAboutSource); ok {

			fmt.Println(sort.SearchInts(listSourcesID, im.ID))

			i := sort.Search(len(listSourcesID), func(i int) bool {
				return listSourcesID[i] >= im.ID
			})
			if i < len(listSourcesID) && listSourcesID[i] == im.ID {
				fmt.Println("ID", im.ID, "listSourcesID[i] == im.ID", listSourcesID[i] == im.ID)
			}

			if sort.SearchInts(listSourcesID, im.ID) == -1 {
				insertListSource = append(insertListSource, configure.InformationAboutSource{im.ID, im.MaxCountProcessfiltration, im.IP, im.Token, im.AsServer})
			}
		}
	}

	fmt.Println("count isnert sources=", len(insertListSource), insertListSource)

	if len(insertListSource) == 0 {
		return false, nil
	}

	fmt.Println("===== INSERT DATA ======")

	return funcInserMany(collection, insertListSource)
}

//InserListSourcesTMPFinaly вставляем список источников ДЛЯ ТЕСТА
func InserListSourcesTMPFinaly(ism *configure.InformationStoringMemory, appConfig *configure.AppConfig) (bool, error) {
	fmt.Println("START func InserListSourcesTMPFinaly...")

	type infoMsg struct {
		ID        int
		IP, Token string
	}

	listSources := []interface{}{
		infoMsg{9, "127.0.0.1", "fmdif3o444fdf344k0fiif"},
		infoMsg{10, "192.168.0.10", "fmdif3o444fdf344k0fiif"},
		infoMsg{11, "192.168.0.11", "ttrr9gr9r9e9f9fadx94"},
		infoMsg{12, "192.168.0.12", "2n3n3iixcxcc3444xfg0222"},
		infoMsg{13, "192.168.0.13", "osdsoc9c933cc9cn939f9f33"},
	}

	funcInserMany := func(collection *mongo.Collection, insertListSource []interface{}) (bool, error) {
		fmt.Println("===== INSERT DATA ======")
		_, err := collection.InsertMany(context.TODO(), insertListSource)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	//ищем все источники
	collection := ism.MongoConnect.Connect.Database(appConfig.ConnectionDB.NameDB).Collection("sources_list")
	options := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{{}}
	//bson.D{{"ip", "192.168.0.10"}}, options)
	if err != nil {
		fmt.Println("ERROR FIND", err)
	}

	insertListSource := []interface{}{}
	listSourcesID := []int{}

	//получаем все ID источников
	for cur.Next(context.TODO()) {
		var im infoMsg
		err := cur.Decode(&im)
		if err != nil {
			fmt.Println(err)
		}

		listSourcesID = append(listSourcesID, im.ID)
	}

	if err := cur.Err(); err != nil {
		fmt.Println(err)
	}

	cur.Close(context.TODO())

	fmt.Println("---------", "LIST SOURCES FROM DB", listSourcesID, "---------")

	if len(listSourcesID) == 0 {
		return funcInserMany(collection, listSources)
	}

	//готовим insertListSources
	for _, value := range listSources {
		//контролируемое привидение типов и получаем срез id
		if im, ok := value.(infoMsg); ok {

			fmt.Println(sort.SearchInts(listSourcesID, im.ID))

			i := sort.Search(len(listSourcesID), func(i int) bool {
				return listSourcesID[i] >= im.ID
			})
			if i < len(listSourcesID) && listSourcesID[i] == im.ID {
				fmt.Println("ID", im.ID, "listSourcesID[i] == im.ID", listSourcesID[i] == im.ID)
			}

			if sort.SearchInts(listSourcesID, im.ID) == -1 {
				insertListSource = append(insertListSource, infoMsg{im.ID, im.IP, im.Token})
			}
		}
	}

	fmt.Println("count isnert sources=", len(insertListSource), insertListSource)

	if len(insertListSource) == 0 {
		return false, nil
	}

	fmt.Println("===== INSERT DATA ======")

	return funcInserMany(collection, insertListSource)
}

//InsertTestData тест
func InsertTestData(ism *configure.InformationStoringMemory, appConfig *configure.AppConfig) (*mongo.InsertOneResult, error) {
	type testType struct {
		name, city, country string
		age                 int
	}

	collection := ism.MongoConnect.Connect.Database(appConfig.ConnectionDB.NameDB).Collection("tcollection")
	res, err := collection.InsertOne(context.Background(), bson.M{
		"name":    "Mariya",
		"city":    "Moscow",
		"country": "Russia",
		"ago":     34,
	})
	if err != nil {
		fmt.Println("ERROR:", err)

		return nil, err
	}

	return res, nil
}
*/
