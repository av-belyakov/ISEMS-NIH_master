package configure

import (
	"errors"
	"fmt"
	"time"
)

//QueueTaskStorage содержит очереди ожидающих и выполняющихся задач
// StorageList - очередь задач, где
// ключ первого отображения - ID источника,
// ключ второго отображения - уникальный ID задачи
// ChannelReq - канал для запросов
type QueueTaskStorage struct {
	StorageList map[int]map[string]*QueueTaskInformation
	ChannelReq  chan chanRequest
}

//QueueTaskInformation подробная информация о задаче в очереди
// IDClientAPI - уникальный идентификатор клиента
// TaskIDClientAPI - идентификатор задачи полученный от клиента
// TaskStatus - статус задачи 'wait', 'execution', 'complite'
// TaskType - тип задачи 'filtration', 'download'
// CheckingStatusItems - проверка пунктов
// TaskParameters - параметры задачи
type QueueTaskInformation struct {
	CommonTaskInfo
	TaskStatus          string
	CheckingStatusItems StatusItems
	TaskParameters      DescriptionParametersReceivedFromUser
}

//DescriptionParametersReceivedFromUser описание параметров задачи
// FilterationParameters - параметры фильтрации
// DownloadList - список файлов полученный от пользователя
// ConfirmedListFiles - подтвержденный список файлов полученный из БД и прошедший сравнение
// с пользовательским (если нужно)
type DescriptionParametersReceivedFromUser struct {
	FilterationParameters FilteringOption
	DownloadList          []string
	ConfirmedListFiles    []*DetailedFileInformation
}

//StatusItems пункты состояния задачи или источника
// AvailabilityConnection - наличие подключения
// AvailabilityFilesDownload - наличие файлов для скачивания
type StatusItems struct {
	AvailabilityConnection, AvailabilityFilesDownload bool
}

//chanRequest канал для передачи запросов
// Action - действие ('add source', 'delete source', 'change task status')
// SourceID - ID источника
// TaskID - ID задачи
// TaskType - тип задачи ('filtration', 'download')
// NewStatus - новый статус задачи ('execution', 'complite')
// AdditionalOption - дополнительные опции для каждого типа задачи
type chanRequest struct {
	CommonTaskInfo
	Action, NewStatus, TaskID string
	SourceID                  int
	AdditionalOption          *DescriptionParametersReceivedFromUser
	ChanRes                   chan chanResponse
}

//chanResponse канал для приема ответов
// SourceID - ID источника
// TaskID - ID задачи
// TaskType - тип задачи ('filtration', 'download')
// TaskStatus - состояние задачи ('wait', 'execution')
// Settings - дополнительные опции для каждого типа задачи
type chanResponse struct {
	CommonTaskInfo
	TaskID, TaskStatus  string
	SourceID            int
	CheckingStatusItems StatusItems
	Settings            *DescriptionParametersReceivedFromUser
	ErrorDescription    error
}

//CommonTaskInfo общая информация о задаче
type CommonTaskInfo struct {
	IDClientAPI, TaskIDClientAPI, TaskType string
}

func checkTaskID(qts *QueueTaskStorage, sourceID int, taskID string) bool {
	if _, ok := qts.StorageList[sourceID]; !ok {
		return false
	}

	if _, ok := qts.StorageList[sourceID][taskID]; !ok {
		return false
	}

	return true
}

//NewRepositoryQTS создание нового репозитория для хранения очередей задач
func NewRepositoryQTS() *QueueTaskStorage {
	qts := QueueTaskStorage{}
	qts.StorageList = map[int]map[string]*QueueTaskInformation{}
	qts.ChannelReq = make(chan chanRequest)

	go func() {
		for msg := range qts.ChannelReq {
			msgRes := chanResponse{SourceID: msg.SourceID}
			msgRes.TaskID = msg.TaskID

			switch msg.Action {
			case "get information for task":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				settings, _ := qts.StorageList[msg.SourceID][msg.TaskID]

				msgRes.TaskType = settings.TaskType
				msgRes.TaskStatus = settings.TaskStatus
				msgRes.IDClientAPI = settings.IDClientAPI
				msgRes.TaskIDClientAPI = settings.TaskIDClientAPI

				msgRes.CheckingStatusItems = StatusItems{
					AvailabilityConnection:    settings.CheckingStatusItems.AvailabilityConnection,
					AvailabilityFilesDownload: settings.CheckingStatusItems.AvailabilityFilesDownload,
				}
				msgRes.Settings = &DescriptionParametersReceivedFromUser{}

				msgRes.Settings.FilterationParameters = settings.TaskParameters.FilterationParameters
				msgRes.Settings.DownloadList = settings.TaskParameters.DownloadList
				msgRes.Settings.ConfirmedListFiles = settings.TaskParameters.ConfirmedListFiles

				msg.ChanRes <- msgRes

			case "add task":
				ts := "wait"
				msgRes.TaskStatus = ts

				qts.StorageList[msg.SourceID] = map[string]*QueueTaskInformation{}
				qts.StorageList[msg.SourceID][msg.TaskID] = &QueueTaskInformation{TaskStatus: ts}
				qts.StorageList[msg.SourceID][msg.TaskID].TaskType = msg.TaskType
				qts.StorageList[msg.SourceID][msg.TaskID].IDClientAPI = msg.IDClientAPI
				qts.StorageList[msg.SourceID][msg.TaskID].TaskIDClientAPI = msg.TaskIDClientAPI

				if msg.TaskType == "filtration" {
					qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.FilterationParameters = msg.AdditionalOption.FilterationParameters
					msg.ChanRes <- msgRes

					break
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.DownloadList = msg.AdditionalOption.DownloadList

				msg.ChanRes <- msgRes

			case "add confirmed list of files":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.ConfirmedListFiles = msg.AdditionalOption.ConfirmedListFiles
				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.DownloadList = []string{}

				msg.ChanRes <- msgRes

			case "add information on the filter":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.FilterationParameters = msg.AdditionalOption.FilterationParameters

				msg.ChanRes <- msgRes

			case "delete task":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				//удаляеть можно только в том случае если задача в состоянии 'wait' или 'complite'
				if qts.StorageList[msg.SourceID][msg.TaskID].TaskStatus == "execution" {
					msgRes.ErrorDescription = fmt.Errorf("deleting is not possible, the task with ID %v is in progress", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				delete(qts.StorageList, msg.SourceID)

				msg.ChanRes <- msgRes

			case "change task status":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskStatus = msg.NewStatus

				msg.ChanRes <- msgRes

			case "change availability connection":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				qts.StorageList[msg.SourceID][msg.TaskID].CheckingStatusItems.AvailabilityConnection = true

				msg.ChanRes <- msgRes

			case "change availability files download":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				qts.StorageList[msg.SourceID][msg.TaskID].CheckingStatusItems.AvailabilityFilesDownload = true

				msg.ChanRes <- msgRes

			case "clear all file list":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", msg.SourceID)
					msg.ChanRes <- msgRes

					break
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.DownloadList = []string{}
				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.ConfirmedListFiles = []*DetailedFileInformation{}

				msg.ChanRes <- msgRes
			}
		}
	}()

	return &qts
}

//GetAllSourcesQueueTaskStorage получить информацию по всем источникам
func (qts *QueueTaskStorage) GetAllSourcesQueueTaskStorage() map[int]map[string]*QueueTaskInformation {
	return qts.StorageList
}

//GetAllTaskQueueTaskStorage получить все задачи выполняемые на выбранном источнике
func (qts *QueueTaskStorage) GetAllTaskQueueTaskStorage(sourceID int) (map[string]*QueueTaskInformation, bool) {
	i, ok := qts.StorageList[sourceID]

	return i, ok
}

//IsExistTaskDownloadQueueTaskStorage есть ли задачи по скачиванию файлов на выбранном источнике
func (qts QueueTaskStorage) IsExistTaskDownloadQueueTaskStorage(sourceID int) bool {
	list, ok := qts.GetAllTaskQueueTaskStorage(sourceID)
	if !ok {
		return false
	}

	for _, info := range list {
		if info.TaskType == "download" {
			return true
		}
	}

	return false
}

//SearchTaskForIDQueueTaskStorage поиск информации по ID задачи
func (qts *QueueTaskStorage) SearchTaskForIDQueueTaskStorage(taskID string) (int, *QueueTaskInformation, error) {
	var sourceID int

	sourceList := qts.GetAllSourcesQueueTaskStorage()
	if len(sourceList) == 0 {
		return sourceID, nil, errors.New("error, empty queue of pending tasks")
	}

	chanRes := make(chan chanResponse)
	defer close(chanRes)

DONE:
	for sID, tasks := range sourceList {
		for tID := range tasks {
			if tID == taskID {
				sourceID = sID

				break DONE
			}
		}
	}

	if sourceID == 0 {
		return sourceID, nil, fmt.Errorf("error, task ID %v not found", taskID)
	}

	qts.ChannelReq <- chanRequest{
		Action:   "get information for task",
		SourceID: sourceID,
		TaskID:   taskID,
		ChanRes:  chanRes,
	}

	msgRes := <-chanRes
	qti := QueueTaskInformation{
		TaskStatus:          msgRes.TaskStatus,
		CheckingStatusItems: msgRes.CheckingStatusItems,
		TaskParameters:      *msgRes.Settings,
	}
	qti.TaskType = msgRes.TaskType
	qti.IDClientAPI = msgRes.IDClientAPI
	qti.TaskIDClientAPI = msgRes.TaskIDClientAPI

	return sourceID, &qti, nil
}

//GetQueueTaskStorage получить информацию по задаче
func (qts *QueueTaskStorage) GetQueueTaskStorage(sourceID int, taskID string) (*QueueTaskInformation, error) {
	if !checkTaskID(qts, sourceID, taskID) {
		return nil, fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID", sourceID)
	}

	chanRes := make(chan chanResponse)
	defer close(chanRes)

	qts.ChannelReq <- chanRequest{
		Action:   "get information for task",
		SourceID: sourceID,
		TaskID:   taskID,
		ChanRes:  chanRes,
	}

	msgRes := <-chanRes
	qti := QueueTaskInformation{
		TaskStatus:          msgRes.TaskStatus,
		CheckingStatusItems: msgRes.CheckingStatusItems,
		TaskParameters:      *msgRes.Settings,
	}
	qti.TaskType = msgRes.TaskType
	qti.IDClientAPI = msgRes.IDClientAPI
	qti.TaskIDClientAPI = msgRes.TaskIDClientAPI

	return &qti, nil
}

//AddQueueTaskStorage добавить информацию по задаче
func (qts *QueueTaskStorage) AddQueueTaskStorage(
	taskID string,
	sourceID int,
	cti CommonTaskInfo,
	options *DescriptionParametersReceivedFromUser) {

	chanRes := make(chan chanResponse)
	defer close(chanRes)

	cr := chanRequest{
		Action:           "add task",
		SourceID:         sourceID,
		TaskID:           taskID,
		AdditionalOption: options,
		ChanRes:          chanRes,
	}
	cr.TaskType = cti.TaskType
	cr.IDClientAPI = cti.IDClientAPI
	cr.TaskIDClientAPI = cti.TaskIDClientAPI

	qts.ChannelReq <- cr

	<-chanRes
}

//AddFiltrationParametersQueueTaskstorage добавляет параметры по фильтрации в существующую задачу
func (qts *QueueTaskStorage) AddFiltrationParametersQueueTaskstorage(sourceID int, taskID string, fp *FilteringOption) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	options := &DescriptionParametersReceivedFromUser{
		FilterationParameters: *fp,
	}

	cr := chanRequest{
		Action:           "add information on the filter",
		SourceID:         sourceID,
		TaskID:           taskID,
		AdditionalOption: options,
		ChanRes:          chanRes,
	}

	qts.ChannelReq <- cr

	return (<-chanRes).ErrorDescription
}

//DelQueueTaskStorage удалить задачу из очереди
func (qts *QueueTaskStorage) DelQueueTaskStorage(sourceID int, taskID string) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	qts.ChannelReq <- chanRequest{
		Action:   "delete task",
		SourceID: sourceID,
		TaskID:   taskID,
		ChanRes:  chanRes,
	}

	return (<-chanRes).ErrorDescription
}

//ChangeTaskStatusQueueTask изменить статус задачи
func (qts *QueueTaskStorage) ChangeTaskStatusQueueTask(sourceID int, taskID, newStatus string) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	qts.ChannelReq <- chanRequest{
		Action:    "change task status",
		SourceID:  sourceID,
		TaskID:    taskID,
		NewStatus: newStatus,
		ChanRes:   chanRes,
	}

	return (<-chanRes).ErrorDescription
}

//ChangeAvailabilityConnection проверить наличие соединение с источником
func (qts *QueueTaskStorage) ChangeAvailabilityConnection(sourceID int, taskID string) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	qts.ChannelReq <- chanRequest{
		Action:   "change availability connection",
		SourceID: sourceID,
		TaskID:   taskID,
		ChanRes:  chanRes,
	}

	return (<-chanRes).ErrorDescription
}

//ChangeAvailabilityFilesDownload проверить наличие файлов для скачивания
func (qts *QueueTaskStorage) ChangeAvailabilityFilesDownload(sourceID int, taskID string) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	qts.ChannelReq <- chanRequest{
		Action:   "change availability files download",
		SourceID: sourceID,
		TaskID:   taskID,
		ChanRes:  chanRes,
	}

	return (<-chanRes).ErrorDescription
}

//AddConfirmedListFiles добавляет проверенный список файлов предназначенных для скачивания и удаляет список переданный клиентом API (если есть)
func (qts *QueueTaskStorage) AddConfirmedListFiles(sourceID int, taskID string, clf []*DetailedFileInformation) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	options := &DescriptionParametersReceivedFromUser{ConfirmedListFiles: clf}

	qts.ChannelReq <- chanRequest{
		Action:           "add confirmed list of files",
		SourceID:         sourceID,
		TaskID:           taskID,
		AdditionalOption: options,
		ChanRes:          chanRes,
	}

	return (<-chanRes).ErrorDescription
}

//ClearAllListFiles очищает все списки файлов
func (qts *QueueTaskStorage) ClearAllListFiles(sourceID int, taskID string) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	qts.ChannelReq <- chanRequest{
		Action:   "clear all file list",
		SourceID: sourceID,
		TaskID:   taskID,
		ChanRes:  chanRes,
	}

	return (<-chanRes).ErrorDescription
}

//MessageInformationQueueTaskStorage краткая информация о задаче
type MessageInformationQueueTaskStorage struct {
	SourceID int
	TaskID   string
}

//CheckTimeQueueTaskStorage подпрограмма для отслеживания очередности выполнения задач
func (qts *QueueTaskStorage) CheckTimeQueueTaskStorage(isl *InformationSourcesList, sec int) chan MessageInformationQueueTaskStorage {
	type executionTasks struct {
		filtrationTask, downloadTask []string
	}

	chanMsgInfoQueueTaskStorage := make(chan MessageInformationQueueTaskStorage)

	ticker := time.NewTicker(time.Duration(sec) * time.Second)

	//поиск и контроль количества задач на выполнение
	searchForTasksPerform := func() {
		for range ticker.C {
			//весь список источников
			storageList := qts.GetAllSourcesQueueTaskStorage()
			if len(storageList) == 0 {
				continue
			}

			for sourceID, tasks := range storageList {
				if len(tasks) == 0 {
					continue
				}

				//получаем максимальное количество одновременно запущенных задач по фильтрации
				sourceSettings, sourceIsExist := isl.GetSourceSetting(sourceID)
				if !sourceIsExist {
					continue
				}

				maxProcessFiltration := int(sourceSettings.Settings.MaxCountProcessFiltration)

				et := executionTasks{
					filtrationTask: make([]string, 0, maxProcessFiltration),
					downloadTask:   make([]string, 0, 1),
				}

				for taskID, taskInfo := range tasks {
					//если задача помечена как выполненная удаляем ее
					if taskInfo.TaskStatus == "complite" {
						_ = qts.DelQueueTaskStorage(sourceID, taskID)
					}

					if taskInfo.TaskType == "download" {
						//выполняется ли задача
						if len(et.downloadTask) > 0 {
							continue
						}

						//если задача не выполнялась и источник подключен и есть файлы для скачивания
						if (taskInfo.TaskStatus == "wait") && taskInfo.CheckingStatusItems.AvailabilityConnection && taskInfo.CheckingStatusItems.AvailabilityFilesDownload {
							if err := qts.ChangeTaskStatusQueueTask(sourceID, taskID, "execution"); err == nil {
								et.downloadTask = append(et.downloadTask, taskID)

								//запускаем выполнение задачи
								chanMsgInfoQueueTaskStorage <- MessageInformationQueueTaskStorage{
									SourceID: sourceID,
									TaskID:   taskID,
								}
							}
						}
					}

					if taskInfo.TaskType == "filtration" {
						if len(et.filtrationTask) == maxProcessFiltration {
							continue
						}

						//если задача не выполнялась и источник подключен и есть файлы для скачивания
						if (taskInfo.TaskStatus == "wait") && taskInfo.CheckingStatusItems.AvailabilityConnection {
							if err := qts.ChangeTaskStatusQueueTask(sourceID, taskID, "execution"); err == nil {
								et.filtrationTask = append(et.filtrationTask, taskID)

								//запускаем выполнение задачи
								chanMsgInfoQueueTaskStorage <- MessageInformationQueueTaskStorage{
									SourceID: sourceID,
									TaskID:   taskID,
								}
							}
						}
					}
				}
			}
		}
	}

	//поиск и контроль количества задач на выполнения
	go searchForTasksPerform()

	return chanMsgInfoQueueTaskStorage
}
