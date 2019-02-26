package handlerrequestdb

import (
	"context"
	"fmt"
	"sort"

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

//InserListSourcesTMP тестовая вставка
func (qcs *QueryCollectionSources) InserListSourcesTMP(list []interface{}) (bool, error) {
	//получаем список источников
	listSources, err := qcs.FindAll()
	if err != nil {
		fmt.Println(err)

		return false, err
	}

	fmt.Println(listSources)

	if len(listSources) == 0 {
		return qcs.InsertData(list)
	}

	fmt.Println("--- Требуются доп. вычисления, поиск уикальных значений")
	fmt.Println(listSources)

	listSourcesInsert := []interface{}{}
	for _, itemAddList := range list {
		if sourceAddList, ok := itemAddList.(configure.InformationAboutSource); ok {
			isExist := false

			for _, itemFindList := range listSources {
				if sourceFindList, ok := itemFindList.(configure.InformationAboutSource); ok {
					if sourceFindList.ID == sourceAddList.ID {
						isExist = true
						break
					}
				}
			}

			if !isExist {
				listSourcesInsert = append(listSourcesInsert, sourceAddList)
			}
		}
	}

	fmt.Println("listSourceInser = ", listSourcesInsert)

	return qcs.InsertData(listSourcesInsert)
}

//FindAll найти всю информацию об источниках
func (qcs *QueryCollectionSources) FindAll() ([]interface{}, error) {
	collection := qcs.ConnectDB.Database(qcs.NameDB).Collection(qcs.CollectionName)
	options := options.Find()

	cur, err := collection.Find(context.TODO(), bson.D{{}}, options)
	if err != nil {
		return nil, err
	}

	listSources := []interface{}{}
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
func (qcs *QueryCollectionSources) InsertData(list []interface{}) (bool, error) {
	fmt.Println("===== INSERT DATA ======")
	collection := qcs.ConnectDB.Database(qcs.NameDB).Collection(qcs.CollectionName)
	_, err := collection.InsertMany(context.TODO(), list)
	if err != nil {
		return false, err
	}

	return true, nil
}

//InsertListSource добавляет список источников !!! ТЕСТ !!!
func (qcs *QueryCollectionSources) InsertListSource() (bool, error) {
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
	cur, err := collection.Find(context.TODO(), bson.D{{}} /*bson.D{{"ip", "192.168.0.10"}}*/, options)
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
				insertListSource = append(insertListSource, configure.InformationAboutSource{im.ID, im.MaxCountProcessFiltering, im.IP, im.Token, im.AsServer})
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
/*func InserListSourcesTMPFinaly(ism *configure.InformationStoringMemory, appConfig *configure.AppConfig) (bool, error) {
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
