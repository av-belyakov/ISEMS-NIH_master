package configure

/*
* Описание типа в котором хранятся параметры для клиентов подключенных к API
* */

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//ClientSettings параметры подключения клиента
// IP: ip адрес клиента
// IsAllowed: разрешен ли доступ
type ClientSettings struct {
	IP         string
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
}

//NewRepository создаем новый репозиторий
func (smapi *StoringMemoryAPI) NewRepository() *StoringMemoryAPI {
	smapi.clientSettings = map[string]*ClientSettings{}

	return smapi
}

//AddNewClient добавляет нового клиента
func (smapi *StoringMemoryAPI) AddNewClient(clientIP string) string {
	currentTime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, clientIP+"_"+strconv.FormatInt(currentTime, 10))

	hsum := hex.EncodeToString(h.Sum(nil))
	smapi.clientSettings[hsum] = &ClientSettings{
		IP:        clientIP,
		IsAllowed: true,
	}

	return hsum
}

//SearchClientForIP поиск информации о клиенте по его ip адресу
func (smapi *StoringMemoryAPI) SearchClientForIP(ip string) (string, *ClientSettings, bool) {
	for id, client := range smapi.clientSettings {
		if client.IP == ip {
			return id, smapi.clientSettings[id], true
		}
	}

	return "", &ClientSettings{}, false
}

//GetClientSettings получить все настройки клиента
func (smapi *StoringMemoryAPI) GetClientSettings(id string) (*ClientSettings, bool) {
	if _, ok := smapi.clientSettings[id]; !ok {
		return &ClientSettings{}, false
	}

	return smapi.clientSettings[id], true
}

//SaveWssClientConnection сохранить линк соединения с клиентом
func (smapi *StoringMemoryAPI) SaveWssClientConnection(id string, conn *websocket.Conn) error {
	if err := smapi.searchID(id); err != nil {
		return err
	}

	smapi.clientSettings[id].Connection = conn

	return nil
}

//GetWssClientConnection получить линк wss соединения
func (smapi *StoringMemoryAPI) GetWssClientConnection(id string) (*websocket.Conn, error) {
	if err := smapi.searchID(id); err != nil {
		return nil, err
	}

	return smapi.clientSettings[id].Connection, nil
}

//DelClientAPI удалить всю информацию о клиенте
func (smapi *StoringMemoryAPI) DelClientAPI(id string) {
	delete(smapi.clientSettings, id)
}

func (smapi *StoringMemoryAPI) searchID(id string) error {
	if _, ok := smapi.clientSettings[id]; !ok {
		return errors.New("client with specified id not found")
	}

	return nil
}
