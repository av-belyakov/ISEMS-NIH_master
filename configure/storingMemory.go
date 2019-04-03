package configure

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"ISEMS-NIH_master/common"
)

//ClientSettings параметры подключения клиента
// IP: ip адрес клиента
// IsAllowed: разрешен ли доступ
type ClientSettings struct {
	IP         string
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
}

//NewRepositorySMAPI создание нового репозитория
func NewRepositorySMAPI() *StoringMemoryAPI {
	smapi := StoringMemoryAPI{}
	smapi.clientSettings = map[string]*ClientSettings{}

	return &smapi
}

//AddNewClient добавляет нового клиента
func (smapi *StoringMemoryAPI) AddNewClient(clientIP, clientName string) string {
	hsum := common.GetUniqIDFormatMD5(clientIP)

	smapi.clientSettings[hsum] = &ClientSettings{
		IP:         clientIP,
		ClientName: clientName,
		IsAllowed:  true,
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

//GetClientList получить весь список клиентов
func (smapi *StoringMemoryAPI) GetClientList() map[string]*ClientSettings {
	return smapi.clientSettings
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
// ClientTaskID - идентификатор задачи полученный от клиента
// TaskType - тип выполняемой задачи
// TaskStatus - статус задачи, false выполняется, true завершена
// ModuleThatSetTask - модуль от которого поступила задача
// ModuleResponsibleImplementation - модуль который должен выполнить обработку
// TimeUpdate - время последнего обновления в формате Unix
// TaskParameter - дополнительные параметры
type TaskDescription struct {
	ClientID                        string
	ClientTaskID                    string
	TaskType                        string
	TaskStatus                      bool
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

//delStoringMemoryTask удалить задачу
func (smt *StoringMemoryTask) delStoringMemoryTask(taskID string) {
	delete(smt.tasks, taskID)
}

//StoringMemoryTaskComplete установить статус выполненно для задачи
func (smt *StoringMemoryTask) StoringMemoryTaskComplete(taskID string) {
	if _, ok := smt.tasks[taskID]; ok {
		smt.tasks[taskID].TaskStatus = true
	}
}

//GetStoringMemoryTask получить информацию о задаче по ее ID
func (smt StoringMemoryTask) GetStoringMemoryTask(taskID string) (*TaskDescription, bool) {
	if task, ok := smt.tasks[taskID]; ok {
		return task, ok
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
func (smt *StoringMemoryTask) TimeUpdateStoringMemoryTask(taskID string) {
	if _, ok := smt.GetStoringMemoryTask(taskID); ok {
		smt.tasks[taskID].TimeUpdate = time.Now().Unix()
	}
}

//MsgChanStoringMemoryTask информация о подвисшей задачи
type MsgChanStoringMemoryTask struct {
	ID, Type, Description string
}

//CheckTimeUpdateStoringMemoryTask проверка времени выполнения задач
// если обновление задачи было больше заданного времени то проверяется
// если задача была выполнена то она удаляется, если нет то отправляется сообщение
// о подвисшей задачи и счетчик увеличивается на 1 до 3, потом задача удаляется
func (smt *StoringMemoryTask) CheckTimeUpdateStoringMemoryTask(sec int) chan MsgChanStoringMemoryTask {
	chanOut := make(chan MsgChanStoringMemoryTask)

	ticker := time.NewTicker(time.Duration(sec) * time.Second)

	go func() {
		for range ticker.C {
			if len(smt.tasks) > 0 {
				timeNow := time.Now().Unix()

				fmt.Println("task count =", len(smt.tasks))

				for id, t := range smt.tasks {

					fmt.Printf("Next Tick %v\n task status:%v, time:%v < %v (%v)\n", time.Now(), t.TaskStatus, (t.TimeUpdate + 60), timeNow, ((t.TimeUpdate + 60) < timeNow))

					if t.TaskStatus && ((t.TimeUpdate + 60) < timeNow) {

						fmt.Println("delete task ID -", id)
						//если задача выполнена и прошло какое то время удаляем ее
						smt.delStoringMemoryTask(id)
					} else {
						if (t.TimeUpdate + 60) < timeNow {
							chanOut <- MsgChanStoringMemoryTask{
								ID:          id,
								Type:        "warning",
								Description: "информация по задаче с ID " + id + " достаточно долго не обновлялась, возможно выполнение задачи было приостановленно",
							}
						} else if (t.TimeUpdate + 180) < timeNow {
							chanOut <- MsgChanStoringMemoryTask{
								ID:          id,
								Type:        "danger",
								Description: "обработка задачи с ID " + id + " была прервана",
							}

							smt.StoringMemoryTaskComplete(id)
						}
					}
				}
			}
		}
	}()

	return chanOut
}
