package handlerrequestdb

import (
	"context"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

//QueryProcessor интерфейс обработчик запросов к БД
type QueryProcessor interface {
	InsertData([]interface{}) (bool, error)
	UpdateOne(interface{}, interface{}) error
	DeleteOneData(interface{}) error
	DeleteManyData([]interface{}) error
	Find(interface{}) (*mongo.Cursor, error)
	FindAlltoCollection() (*mongo.Cursor, error)
}

//QueryParameters параметры для работы с коллекциями БД
type QueryParameters struct {
	NameDB, CollectionName string
	ConnectDB              *mongo.Client
}

//InsertData добавляет все данные
func (qp *QueryParameters) InsertData(list []interface{}) (bool, error) {
	fmt.Println("===== INSERT DATA ======")
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.InsertMany(context.TODO(), list); err != nil {
		return false, err
	}

	return true, nil
}

//DeleteOneData удаляет елемент
func (qp *QueryParameters) DeleteOneData(elem interface{}) error {
	//fmt.Println("===== DELETE DATA ONE ======")
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.DeleteOne(context.TODO(), elem); err != nil {
		return err
	}

	return nil
}

//DeleteManyData удаляет группу элементов
func (qp *QueryParameters) DeleteManyData(list []interface{}) error {
	fmt.Println("===== DELETE DATA MANY ======")
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.DeleteMany(context.TODO(), list); err != nil {
		return err
	}

	return nil
}

//UpdateOne обновляет параметры в элементе
func (qp *QueryParameters) UpdateOne(searchElem, update interface{}) error {
	fmt.Println("===== UPDATE ONE ======")
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.UpdateOne(context.TODO(), searchElem, update); err != nil {
		return err
	}

	return nil
}

//Find найти всю информацию по заданному элементы
func (qp QueryParameters) Find(elem interface{}) (*mongo.Cursor, error) {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	options := options.Find()

	return collection.Find(context.TODO(), elem, options)
}

//FindAlltoCollection найти всю информацию в коллекции
func (qp QueryParameters) FindAlltoCollection() (*mongo.Cursor, error) {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	options := options.Find()

	return collection.Find(context.TODO(), bson.D{{}}, options)
}