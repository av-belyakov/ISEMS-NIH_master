package configure

/*
* Описание типа для хранения в памяти часто используемых параметров
*
* Версия 0.1, дата релиза 18.02.2019
* */

import (
	"context"

	"github.com/mongodb/mongo-go-driver/mongo"
)

//ChanReguestDatabase содержит запросы для модуля обеспечивающего доступ к БД
type ChanReguestDatabase struct {
}

//ChanResponseDatabase содержит ответы от модуля обеспечивающего доступ к БД
type ChanResponseDatabase struct {
}

//MessageAPI параметры для взаимодействия с API
type MessageAPI struct {
	MsgID, MsgType string
	MsgDate        int
}

/*
//ChanMessageToAPI запросы к API
type ChanMessageToAPI struct {
	MessageAPI
}

//ChanMessageFromAPI запросы от API
type ChanMessageFromAPI struct {
	MessageAPI
}
*/
//mongoConnection параметры соединения с БД
type mongoConnection struct {
	Connect *mongo.Client
	CTX     context.Context
}

//channelCollection набор каналов
type channelCollection struct {
	ChanMessageToAPI   MessageAPI
	ChanMessageFromAPI MessageAPI
}

//InformationStoringMemory часто используемые параметры
type InformationStoringMemory struct {
	MongoConnect      mongoConnection
	ChannelCollection channelCollection
}

/*
QuerySelect выполняет запросы к БД
id: <ID элемента> (может быть object, number, string или undefined),
     *                            isMany: <true/false/undefined>,
     *                            query: <object / undefined>,
     *                            select: <object / string / undefined>,
     *                            options: <object / undefined>
*/
/*func (connect *mongoConnection.Connect) QuerySelect() {
	fmt.Println("FUNC querySelect")
}

func (connect *mongoConnection.Connect) QueryCreate() {
	fmt.Println("FUNC queryCreate")
}

func (connect *mongoConnection.Connect) QueryUpdate() {
	fmt.Println("FUNC queryUpdate")
}

func (connect *mongoConnection.Connect) QueryDelete() {
	fmt.Println("FUNC queryDelete")

}*/
