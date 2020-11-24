package handlerrequestdb

import (
	"context"

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
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.InsertMany(context.TODO(), list); err != nil {
		return false, err
	}

	return true, nil
}

//DeleteOneData удаляет елемент
func (qp *QueryParameters) DeleteOneData(elem interface{}) error {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.DeleteOne(context.TODO(), elem); err != nil {
		return err
	}

	return nil
}

//DeleteManyData удаляет группу элементов
func (qp *QueryParameters) DeleteManyData(list []interface{}) error {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.DeleteMany(context.TODO(), list); err != nil {
		return err
	}

	return nil
}

//UpdateOne обновляет параметры в элементе
func (qp *QueryParameters) UpdateOne(searchElem, update interface{}) error {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.UpdateOne(context.TODO(), searchElem, update); err != nil {
		return err
	}

	return nil
}

//UpdateMany обновляет множественные параметры в элементе
func (qp *QueryParameters) UpdateMany(searchElem, update []interface{}) error {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.UpdateMany(context.TODO(), searchElem, update); err != nil {
		return err
	}

	return nil
}

//UpdateOneArrayFilters обновляет множественные параметры в массиве
func (qp *QueryParameters) UpdateOneArrayFilters(filter, update interface{}, uo *options.UpdateOptions) error {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.UpdateOne(context.TODO(), filter, update, uo); err != nil {
		return err
	}

	return nil
}

//Find найти всю информацию по заданному элементу
func (qp QueryParameters) Find(elem interface{}) (*mongo.Cursor, error) {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	options := options.Find().SetSort(bson.D{{Key: "detailed_information_on_filtering.time_interval_task_execution.start", Value: -1}})

	return collection.Find(context.TODO(), elem, options)
}

//FindOne найти информацию по заданному элементу
func (qp QueryParameters) FindOne(elem interface{}) *mongo.SingleResult {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	options := options.FindOne().SetSort(bson.D{{Key: "detailed_information_on_filtering.time_interval_task_execution.start", Value: -1}})

	return collection.FindOne(context.TODO(), elem, options)
}

//FindAlltoCollection найти всю информацию в коллекции
func (qp QueryParameters) FindAlltoCollection() (*mongo.Cursor, error) {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	options := options.Find()

	return collection.Find(context.TODO(), bson.D{{}}, options)
}

//CountDocuments подсчитать количество документов в коллекции
func (qp QueryParameters) CountDocuments(filter interface{}) (int64, error) {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	options := options.Count()

	return collection.CountDocuments(context.TODO(), filter, options)
}

//Indexes возвращает представление индекса для этой коллекции
func (qp QueryParameters) Indexes() mongo.IndexView {
	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)

	return collection.Indexes()
}
