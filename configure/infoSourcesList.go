package configure

/*
* Описание типа для хранения в памяти часто используемых параметров
*
* Версия 0.13, дата релиза 28.05.2019
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

//sourcesListSetting настройки источников, ключ ID источника
type sourcesListSetting map[int]SourceSetting

//SourceSetting параметры источника
// ConnectionStatus - статус соединения с источником
// IP - ip адрес источника
// DateLastConnected - время последнего соединения (в формате unix)
// Token - токен для автризации
// ClientName - имя клиента API (нужно для того чтобы контролировать управление определенным источником)
// AccessIsAllowed - разрешен ли доступ, по умолчанию false (при проверке токена ставится true если он верен)
// AsServer - false запуск как клиент, true запуск как сервер
type SourceSetting struct {
	ConnectionStatus  bool
	IP                string
	DateLastConnected int64
	Token             string
	ClientName        string
	AccessIsAllowed   bool
	AsServer          bool
	Settings          InfoServiceSettings
}

//InformationSourcesList информация об источниках
type InformationSourcesList struct {
	sourcesListSetting
	sourcesListConnection
}

//sourcesListConnection дескрипторы соединения с источниками по протоколу websocket
type sourcesListConnection map[string]WssConnection

//WssConnection дескриптор соединения по протоколу websocket
type WssConnection struct {
	Link *websocket.Conn
	//mu   sync.Mutex
}

//NewRepositoryISL инициализация хранилища
func NewRepositoryISL() *InformationSourcesList {
	isl := InformationSourcesList{}
	isl.sourcesListSetting = sourcesListSetting{}
	isl.sourcesListConnection = sourcesListConnection{}

	return &isl
}

//AddSourceSettings добавить настройки источника
func (isl *InformationSourcesList) AddSourceSettings(id int, settings SourceSetting) {
	isl.sourcesListSetting[id] = settings
}

//DelSourceSettings удаление информации об источнике
func (isl *InformationSourcesList) DelSourceSettings(id int) {
	delete(isl.sourcesListSetting, id)
}

//SearchSourceIPAndToken поиск id источника по его ip и токену
func (isl *InformationSourcesList) SearchSourceIPAndToken(ip, token string) (int, bool) {
	for id, s := range isl.sourcesListSetting {
		if s.IP == ip && s.Token == token {
			//разрешаем соединение с данным источником
			s.AccessIsAllowed = true

			return id, true
		}
	}

	return 0, false
}

//GetSourceIDOnIP получить ID источника по его IP
func (isl *InformationSourcesList) GetSourceIDOnIP(ip string) (int, bool) {
	for id, s := range isl.sourcesListSetting {
		if s.IP == ip {
			return id, true
		}
	}

	return 0, false
}

//GetSourceSetting получить все настройки источника по его ID
func (isl *InformationSourcesList) GetSourceSetting(id int) (*SourceSetting, bool) {
	if s, ok := isl.sourcesListSetting[id]; ok {
		return &s, true
	}

	return nil, false
}

//GetSourceList возвращает список источников
func (isl *InformationSourcesList) GetSourceList() *map[int]SourceSetting {
	sl := map[int]SourceSetting{}

	for id, ss := range isl.sourcesListSetting {
		sl[id] = ss
	}

	return &sl
}

//ChangeSourceConnectionStatus изменить состояние источника
func (isl *InformationSourcesList) ChangeSourceConnectionStatus(id int, status bool) bool {
	if s, ok := isl.sourcesListSetting[id]; ok {
		s.ConnectionStatus = status

		if status {
			s.DateLastConnected = time.Now().Unix()
		} else {
			s.AccessIsAllowed = false
		}

		isl.sourcesListSetting[id] = s

		return true
	}

	return false
}

//GetAccessIsAllowed возвращает значение подтверждающее или отклоняющее права доступа источника
func (isl *InformationSourcesList) GetAccessIsAllowed(ip string) bool {
	for _, s := range isl.sourcesListSetting {
		if s.IP == ip {
			return s.AccessIsAllowed
		}
	}

	return false
}

//SetAccessIsAllowed устанавливает статус позволяющий продолжать wss соединение
func (isl *InformationSourcesList) SetAccessIsAllowed(id int) {
	if s, ok := isl.sourcesListSetting[id]; ok {
		s.AccessIsAllowed = true
		isl.sourcesListSetting[id] = s
	}
}

//GetCountSources возвращает общее количество источников
func (isl InformationSourcesList) GetCountSources() int {
	return len(isl.sourcesListSetting)
}

//GetListsConnectedAndDisconnectedSources возвращает списки источников подключенных и не подключенных
func (isl InformationSourcesList) GetListsConnectedAndDisconnectedSources() (listConnected, listDisconnected map[int]string) {
	listConnected, listDisconnected = map[int]string{}, map[int]string{}

	for id, source := range isl.sourcesListSetting {
		if source.ConnectionStatus {
			listConnected[id] = source.IP
		} else {
			listDisconnected[id] = source.IP
		}
	}

	return listConnected, listDisconnected
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
