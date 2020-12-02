package configure

import (
	"ISEMS-NIH_master/common"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

//ClientSettings параметры подключения клиента
// IP - IP адрес клиента
// Port - сетевой порт клиента
// ClientName - имя клиента из config.json
// IsAllowed: разрешен ли доступ
// Connection - дескриптор соединения через websocket
type ClientSettings struct {
	IP         string
	Token      string
	ClientName string
	IsAllowed  bool
	mu         sync.Mutex
	Connection *websocket.Conn
}

//SendWsMessage используется для отправки сообщений через протокол websocket (применяется Mutex)
func (cs *ClientSettings) SendWsMessage(t int, v []byte) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return cs.Connection.WriteMessage(t, v)
}

//StoringMemoryAPI используется для хранения параметров клиентов
// clientSettings: КЛЮЧ уникальный идентификатор клиента
type StoringMemoryAPI struct {
	clientSettings map[string]*ClientSettings
	chanReqSetting chan typeChanReqSetting
}

type typeReqResCommonSetting struct {
	clientSetting *ClientSettings
	connect       *websocket.Conn
}

type typeChanReqSetting struct {
	actionType string
	clientIP   string
	token      string
	clientID   string
	clientName string
	chanRes    chan typeChanResSetting
	typeReqResCommonSetting
}

type typeChanResSetting struct {
	msgErr     error
	clientID   string
	clientList map[string]*ClientSettings
	typeReqResCommonSetting
}

//NewRepositorySMAPI создание нового репозитория
func NewRepositorySMAPI() *StoringMemoryAPI {
	smapi := StoringMemoryAPI{}
	smapi.clientSettings = map[string]*ClientSettings{}
	smapi.chanReqSetting = make(chan typeChanReqSetting)

	go func() {
		for msg := range smapi.chanReqSetting {
			switch msg.actionType {
			case "add new client":
				smapi.clientSettings[msg.clientID] = &ClientSettings{
					IP:         msg.clientIP,
					Token:      msg.token,
					ClientName: msg.clientName,
					IsAllowed:  true,
				}

				msg.chanRes <- typeChanResSetting{}

			case "get client list":
				msg.chanRes <- typeChanResSetting{
					clientList: smapi.clientSettings,
				}

			case "get client setting":
				if err := smapi.searchID(msg.clientID); err != nil {
					msg.chanRes <- typeChanResSetting{
						msgErr: err,
					}
				} else {
					res := typeChanResSetting{}
					res.clientSetting = smapi.clientSettings[msg.clientID]

					msg.chanRes <- res
				}

			case "search client for ip":
				if len(smapi.clientSettings) == 0 {
					msg.chanRes <- typeChanResSetting{
						msgErr: fmt.Errorf("the client list is empty"),
					}
				} else {
					res := typeChanResSetting{}
					for clientID, setting := range smapi.clientSettings {
						if (msg.clientIP == setting.IP) && (msg.token == setting.Token) {
							res.clientID = clientID
							res.clientSetting = setting
						}
					}

					msg.chanRes <- res
				}

			case "save client connection":
				if err := smapi.searchID(msg.clientID); err != nil {
					msg.chanRes <- typeChanResSetting{
						msgErr: err,
					}
				} else {
					smapi.clientSettings[msg.clientID].Connection = msg.connect

					msg.chanRes <- typeChanResSetting{}
				}

			case "get client connection":
				if err := smapi.searchID(msg.clientID); err != nil {
					msg.chanRes <- typeChanResSetting{
						msgErr: err,
					}
				} else {
					res := typeChanResSetting{}
					res.connect = smapi.clientSettings[msg.clientID].Connection

					msg.chanRes <- res
				}

			case "del client":
				delete(smapi.clientSettings, msg.clientID)

				msg.chanRes <- typeChanResSetting{}

			}
		}
	}()

	return &smapi
}

func (smapi *StoringMemoryAPI) searchID(clientID string) error {
	if _, ok := smapi.clientSettings[clientID]; !ok {
		return fmt.Errorf("client with specified ID %v not found", clientID)
	}

	return nil
}

//AddNewClient добавляет нового клиента
func (smapi *StoringMemoryAPI) AddNewClient(clientIP, port, clientName, token string) string {
	hsum := common.GetUniqIDFormatMD5(clientIP + "_" + port + "_client API")

	cr := make(chan typeChanResSetting)
	defer close(cr)

	smapi.chanReqSetting <- typeChanReqSetting{
		actionType: "add new client",
		clientIP:   clientIP,
		token:      token,
		clientID:   hsum,
		clientName: clientName,
		chanRes:    cr,
	}

	<-cr

	return hsum
}

//SearchClientForIP поиск информации о клиенте по его ip адресу
func (smapi *StoringMemoryAPI) SearchClientForIP(clientIP, token string) (string, *ClientSettings, bool) {
	cr := make(chan typeChanResSetting)
	defer close(cr)

	smapi.chanReqSetting <- typeChanReqSetting{
		actionType: "search client for ip",
		clientIP:   clientIP,
		token:      token,
		chanRes:    cr,
	}

	res := <-cr

	if (res.clientID == "") || (res.msgErr != nil) {
		return "", nil, false
	}

	return res.clientID, res.clientSetting, true
}

//GetClientSettings получить все настройки клиента
func (smapi *StoringMemoryAPI) GetClientSettings(clientID string) (*ClientSettings, error) {
	cr := make(chan typeChanResSetting)
	defer close(cr)

	smapi.chanReqSetting <- typeChanReqSetting{
		actionType: "get client setting",
		clientID:   clientID,
		chanRes:    cr,
	}

	res := <-cr

	return res.clientSetting, res.msgErr
}

//GetClientList получить весь список клиентов
func (smapi *StoringMemoryAPI) GetClientList() map[string]*ClientSettings {
	cr := make(chan typeChanResSetting)
	defer close(cr)

	smapi.chanReqSetting <- typeChanReqSetting{
		actionType: "get client list",
		chanRes:    cr,
	}

	return (<-cr).clientList
}

//SaveWssClientConnection сохранить линк соединения с клиентом
func (smapi *StoringMemoryAPI) SaveWssClientConnection(clientID string, conn *websocket.Conn) error {
	cr := make(chan typeChanResSetting)
	defer close(cr)

	req := typeChanReqSetting{
		actionType: "save client connection",
		clientID:   clientID,
		chanRes:    cr,
	}
	req.connect = conn

	smapi.chanReqSetting <- req

	return (<-cr).msgErr
}

//GetWssClientConnection получить линк wss соединения
func (smapi *StoringMemoryAPI) GetWssClientConnection(clientID string) (*websocket.Conn, error) {
	cr := make(chan typeChanResSetting)
	defer close(cr)

	smapi.chanReqSetting <- typeChanReqSetting{
		actionType: "get client connection",
		clientID:   clientID,
		chanRes:    cr,
	}

	res := <-cr

	return res.connect, res.msgErr
}

//DelClientAPI удалить всю информацию о клиенте
func (smapi *StoringMemoryAPI) DelClientAPI(clientID string) {
	cr := make(chan typeChanResSetting)
	defer close(cr)

	smapi.chanReqSetting <- typeChanReqSetting{
		actionType: "del client",
		clientID:   clientID,
		chanRes:    cr,
	}

	<-cr
}
