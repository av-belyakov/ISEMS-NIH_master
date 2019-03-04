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

/*
--- ДОЛГОВРЕМЕННОЕ ХРАНЕНИЕ ВРЕМЕННЫХ ФАЙЛОВ ---
*/

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
	Settings SourceServiceSettings
}

//SourceServiceSettings настройки влияющие на обработку данных на стороне источника
type SourceServiceSettings struct {
	EnableTelemetry          bool
	MaxCountProcessfiltration int
}

//WssConnection дескриптор соединения по протоколу websocket
type WssConnection struct {
	Link *websocket.Conn
	//mu   sync.Mutex
}

//sourcesListSetting настройки источников
type sourcesListSetting map[string]SourceSetting

//sourcesListConnection дескрипторы соединения с источниками по протоколу websocket
type sourcesListConnection map[string]WssConnection

//InformationStoringMemory часто используемые параметры
type InformationStoringMemory struct {
	sourcesListSetting
	sourcesListConnection
}

//AddSourceSettings добавить настройки источника
func (ism *InformationStoringMemory) AddSourceSettings(host string, settings SourceSetting) {
	ism.sourcesListSetting[host] = settings
}

//SearchSourceToken поиск id источника по его токену и ip
func (ism *InformationStoringMemory) SearchSourceToken(host, token string) (int, bool) {
	if s, ok := ism.sourcesListSetting[host]; ok {
		if s.Token == token {
			//разрешаем соединение с данным источником
			s.AccessIsAllowed = true

			return s.ID, true
		}

	}

	return 0, false
}

//GetSourceSetting получить все настройки источника по его ip
func (ism *InformationStoringMemory) GetSourceSetting(host string) (SourceSetting, bool) {
	if s, ok := ism.sourcesListSetting[host]; ok {
		return s, true
	}

	return SourceSetting{}, false
}

//ChangeSourceConnectionStatus изменить состояние источника
func (ism *InformationStoringMemory) ChangeSourceConnectionStatus(host string) bool {
	if s, ok := ism.sourcesListSetting[host]; ok {
		s.ConnectionStatus = !s.ConnectionStatus

		if s.ConnectionStatus {
			s.DateLastConnected = time.Now().Unix()
		} else {
			s.AccessIsAllowed = false
		}
		ism.sourcesListSetting[host] = s

		return true
	}

	return false
}

//GetAccessIsAllowed возвращает значение подтверждающее или отклоняющее права доступа источника
func (ism *InformationStoringMemory) GetAccessIsAllowed(host string) bool {
	if s, ok := ism.sourcesListSetting[host]; ok {
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
func (ism *InformationStoringMemory) GetSourcesListConnection() map[string]WssConnection {
	return ism.sourcesListConnection
}

//AddLinkWebsocketConnect добавить линк соединения по websocket
func (ism *InformationStoringMemory) AddLinkWebsocketConnect(host string, lwsc *websocket.Conn) {
	ism.sourcesListConnection[host] = WssConnection{
		Link: lwsc,
	}
}

//DelLinkWebsocketConnection удаление дескриптора соединения при отключении источника
func (ism *InformationStoringMemory) DelLinkWebsocketConnection(host string) {
	delete(ism.sourcesListConnection, host)
	/*if _, ok := ism.SourcesListConnection[host]; ok {
		ism.SourcesListConnection[host] = WssConnection{
			Link: nil,
		}
	}*/
}

//GetLinkWebsocketConnect получить линк соединения по websocket
func (ism *InformationStoringMemory) GetLinkWebsocketConnect(host string) (*WssConnection, bool) {
	if conn, ok := ism.sourcesListConnection[host]; ok {
		return &conn, true
	}

	return nil, false
}
