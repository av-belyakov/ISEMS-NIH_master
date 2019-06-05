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

//StoringMemoryTask описание типа в котором храняться описание и ID выполняемых задач
// ключом отображения является уникальный идентификатор задачи
type StoringMemoryTask struct {
	tasks      map[string]*TaskDescription
	channelReq chan ChanStoringMemoryTask
}

//TaskDescription описание задачи
// ClientID - уникальный идентификатор клиента
// ClientTaskID - идентификатор задачи полученный от клиента
// TaskType - тип выполняемой задачи
// TaskStatus - статус задачи, false выполняется, true завершена
// ModuleThatSetTask - модуль от которого поступила задача
// ModuleResponsibleImplementation - модуль который должен выполнить обработку
// TimeUpdate - время последнего обновления в формате Unix
// TimeInterval - интервал времени выполнения задачи
// TaskParameter - дополнительные параметры
// ChanStopTransferListFiles - останов передачи списка файлов (полученных в
// результате поиска по индексам или фильтрации)
type TaskDescription struct {
	ClientID                        string
	ClientTaskID                    string
	TaskType                        string
	TaskStatus                      bool
	ModuleThatSetTask               string
	ModuleResponsibleImplementation string
	TimeUpdate                      int64
	TimeInterval                    TimeIntervalTaskExecution
	TaskParameter                   DescriptionTaskParameters
	ChanStopTransferListFiles       chan struct{}
}

//DescriptionTaskParameters описание параметров задачи
type DescriptionTaskParameters struct {
	FiltrationTask FiltrationTaskParameters
	DownloadTask   DownloadTaskParameters
}

//TimeIntervalTaskExecution временной интервал выполнения задачи
type TimeIntervalTaskExecution struct {
	Start, End int64
}

//FoundFilesInformation подробная информация о файлах
type FoundFilesInformation struct {
	Size     int64
	Hex      string
	IsLoaded bool
}

//FiltrationTaskParameters параметры задачи по фильтрации файлов
// ID - уникальный цифровой идентификатор источника
// Status - статус задачи 'wait'/'refused'/'execute'/'completed'/'stop' ('ожидает' / 'отклонена' / 'выполняется' / 'завершена' / 'остановлена')
// UseIndex - используется ли индекс для поиска файлов
// NumberFilesMeetFilterParameters - кол-во файлов удовлетворяющих параметрам фильтрации
// NumberProcessedFiles - кол-во обработанных файлов
// NumberFilesFoundResultFiltering - кол-во найденных, в результате фильтрации, файлов
// NumberDirectoryFiltartion - кол-во директорий по которым выполняется фильтрация
// NumberErrorProcessedFiles - кол-во не обработанных файлов или файлов обработанных с ошибками
// SizeFilesMeetFilterParameters - общий размер файлов (в байтах) удовлетворяющих параметрам фильтрации
// SizeFilesFoundResultFiltering - общий размер найденных, в результате фильтрации, файлов (в байтах)
// PathStorageSource — путь до директории в которой сохраняются файлы при
// FoundFilesInformation - информация о файлах, ключ - имя файла
type FiltrationTaskParameters struct {
	ID                              int
	Status                          string
	UseIndex                        bool
	NumberFilesMeetFilterParameters int
	NumberProcessedFiles            int
	NumberFilesFoundResultFiltering int
	NumberDirectoryFiltartion       int
	NumberErrorProcessedFiles       int
	SizeFilesMeetFilterParameters   int64
	SizeFilesFoundResultFiltering   int64
	PathStorageSource               string
	FoundFilesInformation           map[string]*FoundFilesInformation
}

//DownloadTaskParameters параметры задачи по скачиванию файлов
type DownloadTaskParameters struct {
}

//ChanStoringMemoryTask описание информации передаваемой через канал
type ChanStoringMemoryTask struct {
	ActionType, TaskID string
	Description        *TaskDescription
	ChannelRes         chan channelResSettings
}

//ChannelResSettings параметры канала ответа
type channelResSettings struct {
	IsExist     bool
	TaskID      string
	Description *TaskDescription
}

//NewRepositorySMT создание нового рапозитория для хранения выполняемых задач
func NewRepositorySMT() *StoringMemoryTask {
	smt := StoringMemoryTask{}
	smt.tasks = map[string]*TaskDescription{}
	smt.channelReq = make(chan ChanStoringMemoryTask)

	go func() {
		for msg := range smt.channelReq {
			switch msg.ActionType {
			case "get task info":
				task, ok := smt.tasks[msg.TaskID]

				msg.ChannelRes <- channelResSettings{
					IsExist:     ok,
					TaskID:      msg.TaskID,
					Description: task,
				}

			case "add":
				smt.tasks[msg.TaskID] = msg.Description
				smt.tasks[msg.TaskID].TaskParameter.FiltrationTask = FiltrationTaskParameters{
					FoundFilesInformation: map[string]*FoundFilesInformation{},
				}
				smt.tasks[msg.TaskID].TaskParameter.DownloadTask = DownloadTaskParameters{}

			case "complete":
				if _, ok := smt.GetStoringMemoryTask(msg.TaskID); ok {
					smt.tasks[msg.TaskID].TaskStatus = true
				}

			case "timer update":
				if _, ok := smt.GetStoringMemoryTask(msg.TaskID); ok {
					smt.tasks[msg.TaskID].TimeUpdate = time.Now().Unix()
				}

			case "delete":
				delete(smt.tasks, msg.TaskID)

			case "update task filtration all parameters":
				smt.updateTaskFiltrationAllParameters(msg.TaskID, msg.Description)

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}
				/*case "update task filtration files list":
				smt.updateTaskFiltrationFilesList(msg.TaskID, msg.Description)
				*/
			}
		}
	}()

	return &smt
}

//AddStoringMemoryTask добавить задачу
// если задачи с заданным ID нет, то в ответ TRUE, если есть то задача не
// изменяется, а в ответ приходит FALSE
func (smt StoringMemoryTask) AddStoringMemoryTask(td TaskDescription) string {
	taskID := common.GetUniqIDFormatMD5(td.ClientID)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType:  "add",
		TaskID:      taskID,
		Description: &td,
	}

	return taskID
}

//delStoringMemoryTask удалить задачу
func (smt StoringMemoryTask) delStoringMemoryTask(taskID string) {
	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "delete",
		TaskID:     taskID,
	}
}

//CompleteStoringMemoryTask установить статус выполненно для задачи
func (smt *StoringMemoryTask) CompleteStoringMemoryTask(taskID string) {
	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "complete",
		TaskID:     taskID,
	}
}

//TimerUpdateStoringMemoryTask обновить значение таймера в задачи
func (smt *StoringMemoryTask) TimerUpdateStoringMemoryTask(taskID string) {
	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "timer update",
		TaskID:     taskID,
	}
}

//GetStoringMemoryTask получить информацию о задаче по ее ID
func (smt StoringMemoryTask) GetStoringMemoryTask(taskID string) (*TaskDescription, bool) {
	chanRes := make(chan channelResSettings)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "get task info",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	info := <-chanRes

	return info.Description, info.IsExist
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

//GetStoringMemoryTaskForClientID получить всю инофрмацию о задаче по ID клиента
func (smt StoringMemoryTask) GetStoringMemoryTaskForClientID(clientID, ClientTaskID string) (string, *TaskDescription, bool) {
	listTask := smt.GetAllStoringMemoryTask(clientID)
	if len(listTask) == 0 {
		return "", nil, false
	}

	for _, id := range listTask {
		info, ok := smt.GetStoringMemoryTask(id)
		if !ok {
			continue
		}

		if info.ClientTaskID == ClientTaskID {
			return id, info, true
		}
	}

	return "", nil, false
}

//UpdateTaskFiltrationAllParameters управление задачами по фильтрации
func (smt *StoringMemoryTask) UpdateTaskFiltrationAllParameters(taskID string, ftp FiltrationTaskParameters) {
	chanRes := make(chan channelResSettings)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "update task filtration all parameters",
		TaskID:     taskID,
		Description: &TaskDescription{
			TaskParameter: DescriptionTaskParameters{
				FiltrationTask: ftp,
			},
		},
		ChannelRes: chanRes,
	}

	for task := range chanRes {
		if task.TaskID == taskID {
			break
		}
	}
}

//UpdateTaskFiltrationFilesList обновление списка файлов полученных в результате фильтрации
/*func (smt *StoringMemoryTask) UpdateTaskFiltrationFilesList(taskID string, filesList map[string]*FoundFilesInformation) {
	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "update task filtration files list",
		TaskID:     taskID,
		Description: &TaskDescription{
			TaskParameter: DescriptionTaskParameters{
				FiltrationTask: FiltrationTaskParameters{
					FoundFilesInformation: filesList,
				},
			},
		},
	}
}*/

func (smt *StoringMemoryTask) updateTaskFiltrationAllParameters(taskID string, td *TaskDescription) {
	if _, ok := smt.tasks[taskID]; !ok {
		return
	}

	//изменяем время окончания задачи
	smt.tasks[taskID].TimeInterval.End = time.Now().Unix()

	ft := smt.tasks[taskID].TaskParameter.FiltrationTask
	nft := td.TaskParameter.FiltrationTask

	for n, v := range nft.FoundFilesInformation {
		ft.FoundFilesInformation[n] = &FoundFilesInformation{
			Size: v.Size,
			Hex:  v.Hex,
		}
	}

	ft.NumberFilesMeetFilterParameters = nft.NumberFilesMeetFilterParameters
	ft.NumberFilesFoundResultFiltering = nft.NumberFilesFoundResultFiltering
	ft.NumberErrorProcessedFiles = nft.NumberErrorProcessedFiles
	ft.NumberDirectoryFiltartion = nft.NumberDirectoryFiltartion
	ft.NumberProcessedFiles = nft.NumberProcessedFiles
	ft.SizeFilesMeetFilterParameters = nft.SizeFilesMeetFilterParameters
	ft.SizeFilesFoundResultFiltering = nft.SizeFilesFoundResultFiltering
	ft.PathStorageSource = nft.PathStorageSource
	ft.Status = nft.Status
	ft.ID = nft.ID

	smt.tasks[taskID].TaskParameter.FiltrationTask = ft
}

/*func (smt *StoringMemoryTask) updateTaskFiltrationFilesList(taskID string, td *TaskDescription) {
	if _, ok := smt.tasks[taskID]; !ok {
		return
	}

	//получаем список файлов
	//		listFoundFiles := smt.tasks[taskID].TaskParameter.FiltrationTask.FoundFilesInformation
	filesList := td.TaskParameter.FiltrationTask.FoundFilesInformation

	for n, v := range filesList {
		smt.tasks[taskID].TaskParameter.FiltrationTask.FoundFilesInformation[n] = v
	}
}*/

/* управление задачами по скачиванию файлов */

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

				//fmt.Println("task count =", len(smt.tasks))

				for id, t := range smt.tasks {

					//fmt.Printf("Next Tick %v\n task status:%v, time:%v < %v (%v)\n", time.Now(), t.TaskStatus, (t.TimeUpdate + 60), timeNow, ((t.TimeUpdate + 60) < timeNow))

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

							smt.CompleteStoringMemoryTask(id)
						}
					}
				}
			}
		}
	}()

	return chanOut
}
