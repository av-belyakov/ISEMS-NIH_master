package configure

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"

	"ISEMS-NIH_master/common"
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

//NewRepositorySMAPI создание нового репозитория
func NewRepositorySMAPI() *StoringMemoryAPI {
	smapi := StoringMemoryAPI{}
	smapi.clientSettings = map[string]*ClientSettings{}

	return &smapi
}

//AddNewClient добавляет нового клиента
func (smapi *StoringMemoryAPI) AddNewClient(clientIP string) string {
	hsum := common.GetUniqIDFormatMD5(clientIP)

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

	return "", nil, false
}

//GetClientSettings получить все настройки клиента
func (smapi *StoringMemoryAPI) GetClientSettings(id string) (*ClientSettings, bool) {
	if _, ok := smapi.clientSettings[id]; !ok {
		return nil, false
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

//DescriptionTaskParameters описание параметров задачи
type DescriptionTaskParameters struct{}

//TaskDescription описание задачи
// ClientID - уникальный идентификатор клиента
// TaskSection - секция к которой относится задача
// TaskType - тип выполняемой задачи
// ModuleThatSetTask - модуль от которого поступила задача
// ModuleResponsibleImplementation - модуль который должен выполнить обработку
// TimeUpdate - время последнего обновления в формате Unix
// TaskParameter - дополнительные параметры
type TaskDescription struct {
	ClientID                        string
	TaskType                        string
	ModuleThatSetTask               string
	ModuleResponsibleImplementation string
	TimeUpdate                      int64
	TaskParameter                   DescriptionTaskParameters
}

//StoringMemoryTask описание типа в котором храняться описание и ID выполняемых задач
// ключом отображения является уникальный идентификатор задачи
type StoringMemoryTask struct {
	tasks map[string]*TaskDescription
}

//NewRepositorySMT создание нового рапозитория для хранения выполняемых задач
func NewRepositorySMT() *StoringMemoryTask {
	smt := StoringMemoryTask{}
	smt.tasks = map[string]*TaskDescription{}

	return &smt
}

//AddStoringMemoryTask добавить задачу
// если задачи с заданным ID нет, то в ответ TRUE, если есть то задача не
// изменяется, а в ответ приходит FALSE
func (smt *StoringMemoryTask) AddStoringMemoryTask(td TaskDescription) string {
	taskID := common.GetUniqIDFormatMD5(td.ClientID)

	smt.tasks[taskID] = &td

	return taskID
}

//DelStoringMemoryTask удалить задачу
func (smt *StoringMemoryTask) DelStoringMemoryTask(taskID string) {
	delete(smt.tasks, taskID)
}

//GetStoringMemoryTask получить информацию о задаче по ее ID
func (smt StoringMemoryTask) GetStoringMemoryTask(taskID string) (*TaskDescription, bool) {
	if _, ok := smt.tasks[taskID]; ok {
		return smt.tasks[taskID], ok
	}

	return nil, false
}

//GetAllStoringMemoryTask получить все ID задач для выбранного клиента
func (smt StoringMemoryTask) GetAllStoringMemoryTask(clientID string) []string {
	foundTask := make([]string, 0, len(smt.tasks))

	for tid, v := range smt.tasks {
		if clientID == v.ClientID {
			foundTask = append(foundTask, tid)
		}
	}

	return foundTask
}

//TimeUpdateStoringMemoryTask обновить значение таймера в задачи
func (smt *StoringMemoryTask) TimeUpdateStoringMemoryTask(taskID string, time int64) bool {
	if _, ok := smt.GetStoringMemoryTask(taskID); !ok {
		return false
	}

	smt.tasks[taskID].TimeUpdate = time

	return true
}

//CounterCheckTimeUpdateStoringMemoryTask счетчик проверяющий время обнавления
// задачи и отправляющий, через канал, ID задачи которая устарела
func (smt StoringMemoryTask) CounterCheckTimeUpdateStoringMemoryTask() chan string {
	chanOut := make(chan string)

	/*
		!!! НЕ ДОПИСАН !!!
	*/

	return chanOut
}
