package configure

/*
* Описание типа для хранения в памяти часто используемых параметров
* */

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
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
// ShortName - краткое название источника
// DateLastConnected - время последнего соединения (в формате unix timestamp)
// Token - токен для авторизации
// ClientName - имя клиента API (нужно для того чтобы контролировать управление определенным источником)
// AccessIsAllowed - разрешен ли доступ, по умолчанию false (при проверке токена ставится true если он верен)
// AsServer - false запуск как клиент, true запуск как сервер
type SourceSetting struct {
	ConnectionStatus  bool
	IP                string
	ShortName         string
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
	chanReq chan chanReqSetting
}

type chanReqSetting struct {
	actionType string
	id         int
	setting    SourceSetting
	link       *websocket.Conn
	chanRes    chan chanResSetting
}

type chanResSetting struct {
	err                   error
	id                    int
	setting               *SourceSetting
	additionalInformation interface{}
}

type sourceConnectDisconnectLists struct {
	listConnected, listDisconnected map[int]string
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
	isl := InformationSourcesList{
		sourcesListSetting:    sourcesListSetting{},
		sourcesListConnection: sourcesListConnection{},
		chanReq:               make(chan chanReqSetting),
	}

	go func() {
		for msg := range isl.chanReq {
			switch msg.actionType {
			case "add source settings":
				isl.sourcesListSetting[msg.id] = msg.setting

				msg.chanRes <- chanResSetting{}
				close(msg.chanRes)

			case "add link ws connection":
				isl.sourcesListConnection[msg.setting.IP] = WssConnection{Link: msg.link}

				msg.chanRes <- chanResSetting{}
				close(msg.chanRes)

			case "del info about source":
				delete(isl.sourcesListSetting, msg.id)

				msg.chanRes <- chanResSetting{}
				close(msg.chanRes)

			case "del link ws connection":
				delete(isl.sourcesListConnection, msg.setting.IP)

				msg.chanRes <- chanResSetting{}
				close(msg.chanRes)

			case "source authentication by ip and token":
				var sourceID int
				for id, s := range isl.sourcesListSetting {
					if s.IP == msg.setting.IP && s.Token == msg.setting.Token {
						//разрешаем соединение с данным источником
						s.AccessIsAllowed = true
						isl.sourcesListSetting[id] = s
						sourceID = id

						break
					}
				}

				msgRes := chanResSetting{id: sourceID}

				if sourceID == 0 {
					msgRes.err = fmt.Errorf("source with ip address %v not found", msg.setting.IP)
				}

				msg.chanRes <- msgRes
				close(msg.chanRes)

			case "get source list":
				sl := make(map[int]SourceSetting, len(isl.sourcesListSetting))

				for id, ss := range isl.sourcesListSetting {
					sl[id] = ss
				}

				msg.chanRes <- chanResSetting{additionalInformation: sl}
				close(msg.chanRes)

			case "get source setting by id":
				si, ok := isl.sourcesListSetting[msg.id]
				if ok {
					msg.chanRes <- chanResSetting{
						setting: &si,
					}
				} else {
					msg.chanRes <- chanResSetting{
						err: fmt.Errorf("source with ID %v not found", msg.id),
					}
				}

				close(msg.chanRes)

			case "get source id by ip":
				var sourceID int
				for id, s := range isl.sourcesListSetting {
					if s.IP == msg.setting.IP {
						sourceID = id

						break
					}
				}

				if sourceID == 0 {
					msg.chanRes <- chanResSetting{
						err: fmt.Errorf("source with ip address %v not found", msg.setting.IP),
					}
				} else {
					msg.chanRes <- chanResSetting{id: sourceID}
				}

				close(msg.chanRes)

			case "get source connection status":
				s, ok := isl.sourcesListSetting[msg.id]
				if !ok {
					msg.chanRes <- chanResSetting{
						err: fmt.Errorf("source with ID %v not found", msg.id),
					}

					close(msg.chanRes)

					break
				}

				msg.chanRes <- chanResSetting{setting: &SourceSetting{ConnectionStatus: s.ConnectionStatus}}
				close(msg.chanRes)

			case "get lists connected and disconnected sources":
				listConnected, listDisconnected := map[int]string{}, map[int]string{}

				for id, source := range isl.sourcesListSetting {
					if source.ConnectionStatus {
						listConnected[id] = source.IP
					} else {
						listDisconnected[id] = source.IP
					}
				}

				msg.chanRes <- chanResSetting{additionalInformation: sourceConnectDisconnectLists{
					listConnected:    listConnected,
					listDisconnected: listDisconnected,
				}}
				close(msg.chanRes)

			case "change source connection status":
				s, ok := isl.sourcesListSetting[msg.id]
				if !ok {
					msg.chanRes <- chanResSetting{
						err: fmt.Errorf("source with ID %v not found", msg.id),
					}

					close(msg.chanRes)

					break
				}
				s.ConnectionStatus = msg.setting.ConnectionStatus

				if msg.setting.ConnectionStatus {
					s.DateLastConnected = time.Now().Unix()
				} else {
					s.AccessIsAllowed = false
				}

				isl.sourcesListSetting[msg.id] = s

				close(msg.chanRes)

			case "get access is allowed":
				var aia bool
				for _, s := range isl.sourcesListSetting {
					if s.IP == msg.setting.IP {
						aia = s.AccessIsAllowed
					}
				}

				msg.chanRes <- chanResSetting{setting: &SourceSetting{AccessIsAllowed: aia}}
				close(msg.chanRes)

			case "set access is allowed":
				if s, ok := isl.sourcesListSetting[msg.id]; ok {
					s.AccessIsAllowed = true
					isl.sourcesListSetting[msg.id] = s
				}

				msg.chanRes <- chanResSetting{}
				close(msg.chanRes)
			}
		}
	}()

	return &isl
}

//AddSourceSettings добавляет настройки источника
func (isl *InformationSourcesList) AddSourceSettings(id int, settings SourceSetting) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "add source settings",
		id:         id,
		setting:    settings,
		chanRes:    chanRes,
	}

	<-chanRes
}

//DelSourceSettings удаляет информацию об источнике
func (isl *InformationSourcesList) DelSourceSettings(id int) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "del info about source",
		id:         id,
		chanRes:    chanRes,
	}

	<-chanRes
}

//SourceAuthenticationByIPAndToken поиск id источника по его ip и токену
func (isl *InformationSourcesList) SourceAuthenticationByIPAndToken(ip, token string) (int, bool) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "source authentication by ip and token",
		setting: SourceSetting{
			IP:    ip,
			Token: token,
		},
		chanRes: chanRes,
	}

	resMsg := <-chanRes

	if resMsg.err != nil {
		return 0, false
	}

	return resMsg.id, true
}

//GetSourceSetting возвращает все настройки источника по его ID
func (isl *InformationSourcesList) GetSourceSetting(id int) (*SourceSetting, bool) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "get source setting by id",
		id:         id,
		chanRes:    chanRes,
	}

	resMsg := <-chanRes

	if resMsg.err != nil {
		return nil, false
	}

	return resMsg.setting, true
}

//GetSourceIDOnIP возвращает ID источника по его IP
func (isl *InformationSourcesList) GetSourceIDOnIP(ip string) (int, bool) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "get source id by ip",
		setting:    SourceSetting{IP: ip},
		chanRes:    chanRes,
	}

	resMsg := <-chanRes

	var isExist bool
	if resMsg.err == nil {
		isExist = true
	}

	return resMsg.id, isExist
}

//GetSourceList возвращает список источников
func (isl *InformationSourcesList) GetSourceList() *map[int]SourceSetting {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "get source list",
		chanRes:    chanRes,
	}

	if sl, ok := (<-chanRes).additionalInformation.(map[int]SourceSetting); ok {
		return &sl
	}

	return &map[int]SourceSetting{}
}

//GetSourceConnectionStatus возвращает состояние соединения с источником
func (isl *InformationSourcesList) GetSourceConnectionStatus(id int) (bool, error) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "get source connection status",
		id:         id,
		chanRes:    chanRes,
	}

	resMsg := <-chanRes

	return resMsg.setting.ConnectionStatus, resMsg.err
}

//ChangeSourceConnectionStatus изменяет состояние источника
func (isl *InformationSourcesList) ChangeSourceConnectionStatus(id int, status bool) bool {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "change source connection status",
		id:         id,
		setting:    SourceSetting{ConnectionStatus: status},
		chanRes:    chanRes,
	}

	if (<-chanRes).err != nil {
		return false
	}

	return true
}

//GetAccessIsAllowed возвращает значение подтверждающее или отклоняющее права доступа источника
func (isl *InformationSourcesList) GetAccessIsAllowed(ip string) bool {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "get access is allowed",
		setting:    SourceSetting{IP: ip},
		chanRes:    chanRes,
	}

	return (<-chanRes).setting.AccessIsAllowed
}

//SetAccessIsAllowed устанавливает статус позволяющий продолжать wss соединение
func (isl *InformationSourcesList) SetAccessIsAllowed(id int) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "set access is allowed",
		setting:    SourceSetting{AccessIsAllowed: true},
		chanRes:    chanRes,
	}

	<-chanRes
}

//GetCountSources возвращает общее количество источников
func (isl InformationSourcesList) GetCountSources() int {
	return len(isl.sourcesListSetting)
}

//GetListsConnectedAndDisconnectedSources возвращает списки источников подключенных и не подключенных
func (isl InformationSourcesList) GetListsConnectedAndDisconnectedSources() (listConnected, listDisconnected map[int]string) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "get lists connected and disconnected sources",
		setting:    SourceSetting{AccessIsAllowed: true},
		chanRes:    chanRes,
	}

	if lists, ok := (<-chanRes).additionalInformation.(sourceConnectDisconnectLists); ok {
		return lists.listConnected, lists.listDisconnected
	}

	return map[int]string{}, map[int]string{}
}

//SendWsMessage используется для отправки сообщений через протокол websocket
func (wssc *WssConnection) SendWsMessage(t int, v []byte) error {
	/*wssc.mu.Lock()
	defer wssc.mu.Unlock()*/

	return wssc.Link.WriteMessage(t, v)
}

//GetSourcesListConnection возвращает список всех соединений
func (isl *InformationSourcesList) GetSourcesListConnection() map[string]WssConnection {
	return isl.sourcesListConnection
}

//AddLinkWebsocketConnect добавляет линк соединения по websocket
func (isl *InformationSourcesList) AddLinkWebsocketConnect(ip string, lwsc *websocket.Conn) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "add link ws connection",
		link:       lwsc,
		setting:    SourceSetting{IP: ip},
		chanRes:    chanRes,
	}

	<-chanRes
}

//DelLinkWebsocketConnection удаляет дескриптор соединения при отключении источника
func (isl *InformationSourcesList) DelLinkWebsocketConnection(ip string) {
	chanRes := make(chan chanResSetting)

	isl.chanReq <- chanReqSetting{
		actionType: "del link ws connection",
		setting:    SourceSetting{IP: ip},
		chanRes:    chanRes,
	}

	<-chanRes
}

//GetLinkWebsocketConnect возвращает линк соединения по websocket
func (isl *InformationSourcesList) GetLinkWebsocketConnect(ip string) (*WssConnection, bool) {
	if conn, ok := isl.sourcesListConnection[ip]; ok {
		return &conn, true
	}

	return nil, false
}
