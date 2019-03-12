package configure

/*
* Описание типа для хранения в памяти часто используемых параметров
*
* Версия 0.12, дата релиза 26.02.2019
* */

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mongodb/mongo-go-driver/mongo"
)

//MongoDBConnect содержит дискриптор соединения с БД
type MongoDBConnect struct {
	Connect *mongo.Client
	CTX     context.Context
}

//SourceSetting параметры источника
type SourceSetting struct {
	ConnectionStatus  bool //true/false
	ID                int
	DateLastConnected int64 //Unix time
	Token             string
	AccessIsAllowed   bool              //разрешен ли доступ, по умолчанию false (при проверке токена ставится true если он верен)
	AsServer          bool              //false - как клиент, true - как сервер
	CurrentTasks      map[string]string // задачи для данного источника,
	//key - ID задачи, value - ее тип 'in queuq' или 'in process'
	Settings InfoServiceSettings
}

//WssConnection дескриптор соединения по протоколу websocket
type WssConnection struct {
	Link *websocket.Conn
	//mu   sync.Mutex
}

//sourcesListSetting настройки источников, ключ IP
type sourcesListSetting map[string]SourceSetting

//sourcesListConnection дескрипторы соединения с источниками по протоколу websocket
type sourcesListConnection map[string]WssConnection

//InformationSourcesList информация об источниках
type InformationSourcesList struct {
	sourcesListSetting
	sourcesListConnection
}

//NewRepositoryISL инициализация хранилища
func NewRepositoryISL() *InformationSourcesList {
	isl := InformationSourcesList{}
	isl.sourcesListSetting = sourcesListSetting{}
	isl.sourcesListConnection = sourcesListConnection{}

	return &isl
}

//AddSourceSettings добавить настройки источника
func (isl *InformationSourcesList) AddSourceSettings(host string, settings SourceSetting) {
	isl.sourcesListSetting[host] = settings
}

//SearchSourceToken поиск id источника по его токену и ip
func (isl *InformationSourcesList) SearchSourceToken(host, token string) (int, bool) {
	if s, ok := isl.sourcesListSetting[host]; ok {
		if s.Token == token {
			//разрешаем соединение с данным источником
			s.AccessIsAllowed = true

			return s.ID, true
		}

	}

	return 0, false
}

//GetSourceSetting получить все настройки источника по его ip
func (isl *InformationSourcesList) GetSourceSetting(host string) (SourceSetting, bool) {
	if s, ok := isl.sourcesListSetting[host]; ok {
		return s, true
	}

	return SourceSetting{}, false
}

//ChangeSourceConnectionStatus изменить состояние источника
func (isl *InformationSourcesList) ChangeSourceConnectionStatus(host string) bool {
	if s, ok := isl.sourcesListSetting[host]; ok {
		s.ConnectionStatus = !s.ConnectionStatus

		if s.ConnectionStatus {
			s.DateLastConnected = time.Now().Unix()
		} else {
			s.AccessIsAllowed = false
		}
		isl.sourcesListSetting[host] = s

		return true
	}

	return false
}

//GetAccessIsAllowed возвращает значение подтверждающее или отклоняющее права доступа источника
func (isl *InformationSourcesList) GetAccessIsAllowed(host string) bool {
	if s, ok := isl.sourcesListSetting[host]; ok {
		return s.AccessIsAllowed
	}

	return false
}

//SendWsMessage используется для отправки сообщений через протокол websocket (применяется Mutex)
func (wssc *WssConnection) SendWsMessage(t int, v []byte) error {
	/*wssc.mu.Lock()
	defer wssc.mu.Unlock()*/

	return wssc.Link.WriteMessage(t, v)
}

//GetSourcesListConnection получить список всех соединений
func (isl *InformationSourcesList) GetSourcesListConnection() map[string]WssConnection {
	return isl.sourcesListConnection
}

//AddLinkWebsocketConnect добавить линк соединения по websocket
func (isl *InformationSourcesList) AddLinkWebsocketConnect(host string, lwsc *websocket.Conn) {
	isl.sourcesListConnection[host] = WssConnection{
		Link: lwsc,
	}
}

//DelLinkWebsocketConnection удаление дескриптора соединения при отключении источника
func (isl *InformationSourcesList) DelLinkWebsocketConnection(host string) {
	delete(isl.sourcesListConnection, host)
	/*if _, ok := ism.SourcesListConnection[host]; ok {
		ism.SourcesListConnection[host] = WssConnection{
			Link: nil,
		}
	}*/
}

//GetLinkWebsocketConnect получить линк соединения по websocket
func (isl *InformationSourcesList) GetLinkWebsocketConnect(host string) (*WssConnection, bool) {
	if conn, ok := isl.sourcesListConnection[host]; ok {
		return &conn, true
	}

	return nil, false
}
