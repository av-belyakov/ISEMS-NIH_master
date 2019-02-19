package handlerrequestdb

import (
	"context"
	"fmt"
	"sort"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

//QueryCollectionSources значения для работы с коллекцией источников
type QueryCollectionSources struct {
	NameDB, CollectionName string
	ConnectDB              *mongo.Client
}

//InsertListSource добавляет список источников ТЕСТ
func (qcs *QueryCollectionSources) InsertListSource() (bool, error) {
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
