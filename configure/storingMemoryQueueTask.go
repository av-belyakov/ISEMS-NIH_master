package configure

import (
	"ISEMS-NIH_master/savemessageapp"
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
// TaskStatus - статус задачи 'wait', 'execution', 'complete', 'pause'
// UserName - имя пользователя инициировавшего задачу (если поле пустое,
//  то считается что выполнение задачи было инициировано автоматически)
// TimeUpdate - время последнего обновления задачи (используется для
//  удаления 'подвисших' задач)
// TaskType - тип задачи ('filtration control', 'download control')
// CheckingStatusItems - проверка пунктов
// TaskParameters - параметры задачи
type QueueTaskInformation struct {
	CommonTaskInfo
	TaskStatus          string
	TimeUpdate          int64
	CheckingStatusItems StatusItems
	TaskParameters      DescriptionParametersReceivedFromUser
}

//DescriptionParametersReceivedFromUser описание параметров задачи
// FilterationParameters - параметры фильтрации
// PathDirectoryForFilteredFiles - путь до директории с отфильтрованными файлами на источнике
// DownloadList - список файлов полученный от пользователя
// ConfirmedListFiles - подтвержденный список файлов полученный из БД и прошедший сравнение
// с пользовательским (если нужно)
type DescriptionParametersReceivedFromUser struct {
	FilterationParameters         FilteringOption
	PathDirectoryForFilteredFiles string
	DownloadList                  []string
	ConfirmedListFiles            map[string]*DetailedFilesInformation
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
// NewStatus - новый статус задачи ('execution', 'complete')
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
	Settings            DescriptionParametersReceivedFromUser
	ErrorDescription    error
}

//CommonTaskInfo общая информация о задаче
// IDClientAPI - ID клиента API
// TaskIDClientAPI - ID задачи клиента API
// TaskType - тип задачи
type CommonTaskInfo struct {
	IDClientAPI, TaskIDClientAPI, TaskType, UserName string
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
func NewRepositoryQTS(saveMessageApp *savemessageapp.PathDirLocationLogFiles) *QueueTaskStorage {
	qts := QueueTaskStorage{
		StorageList: map[int]map[string]*QueueTaskInformation{},
		ChannelReq:  make(chan chanRequest),
	}

	go func() {
		for msg := range qts.ChannelReq {
			msgRes := chanResponse{SourceID: msg.SourceID}
			msgRes.TaskID = msg.TaskID

			switch msg.Action {
			case "get information for task":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'get information for task'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				settings := qts.StorageList[msg.SourceID][msg.TaskID]

				msgRes.TaskType = settings.TaskType
				msgRes.TaskStatus = settings.TaskStatus
				msgRes.IDClientAPI = settings.IDClientAPI
				msgRes.TaskIDClientAPI = settings.TaskIDClientAPI
				msgRes.UserName = settings.UserName

				msgRes.CheckingStatusItems = StatusItems{
					AvailabilityConnection:    settings.CheckingStatusItems.AvailabilityConnection,
					AvailabilityFilesDownload: settings.CheckingStatusItems.AvailabilityFilesDownload,
				}

				msgRes.Settings = settings.TaskParameters

				msg.ChanRes <- msgRes

			case "add task":
				ts := "wait"
				msgRes.TaskStatus = ts

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: добавляем информацию о задаче ID '%s', типа: '%s' для источника: '%d' в хранилище очередей задач", msg.TaskID, msg.TaskType, msg.SourceID),
					FuncName:    "NewRepositoryQTS",
				})

				if len(qts.StorageList[msg.SourceID]) == 0 {

					/*** Временное логирование событий ***/
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						TypeMessage: "info",
						Description: "TEST Message: список источников пуст. Инициализируем список.",
						FuncName:    "NewRepositoryQTS",
					})

					qts.StorageList[msg.SourceID] = map[string]*QueueTaskInformation{}
				}

				qts.StorageList[msg.SourceID][msg.TaskID] = &QueueTaskInformation{
					CommonTaskInfo: CommonTaskInfo{
						TaskType:        msg.TaskType,
						IDClientAPI:     msg.IDClientAPI,
						TaskIDClientAPI: msg.TaskIDClientAPI,
						UserName:        msg.UserName,
					},
					TaskStatus: ts,
					TimeUpdate: time.Now().Unix(),
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.FilterationParameters = msg.AdditionalOption.FilterationParameters
				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.PathDirectoryForFilteredFiles = msg.AdditionalOption.PathDirectoryForFilteredFiles

				if msg.TaskType == "filtration control" {

					/*** Временное логирование событий ***/
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						TypeMessage: "info",
						Description: fmt.Sprintf("TEST Message: выполненно добавление задачи по фильтрации, список задач: %v", qts.StorageList[msg.SourceID][msg.TaskID]),
						FuncName:    "NewRepositoryQTS",
					})

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.DownloadList = msg.AdditionalOption.DownloadList

				msg.ChanRes <- msgRes

			case "add confirmed list of files":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'add confirmed list of files'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.ConfirmedListFiles = msg.AdditionalOption.ConfirmedListFiles
				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.DownloadList = []string{}

				msg.ChanRes <- msgRes

			case "add information on the filter":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'add information on the filter'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.FilterationParameters = msg.AdditionalOption.FilterationParameters

				msg.ChanRes <- msgRes

			case "add path directory for filtered files":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'add path directory for filtered files'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.PathDirectoryForFilteredFiles = msg.AdditionalOption.PathDirectoryForFilteredFiles

				msg.ChanRes <- msgRes

			case "delete task":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'delete task'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: выполнение удаление задачи ID: '%s', ТИП: '%s' для источика SourceID: '%d', список задач ДО: %v", msg.TaskID, msg.TaskType, msg.SourceID, qts.StorageList[msg.SourceID][msg.TaskID]),
					FuncName:    "NewRepositoryQTS",
				})

				//удаляеть можно только в том случае если задача в состоянии 'wait' или 'complete'
				if qts.StorageList[msg.SourceID][msg.TaskID].TaskStatus == "execution" {
					msgRes.ErrorDescription = fmt.Errorf("deleting is not possible, the task with ID %v is in progress", msg.SourceID)
				} else {
					delete(qts.StorageList[msg.SourceID], msg.TaskID)

					if len(qts.StorageList[msg.SourceID]) == 0 {
						delete(qts.StorageList, msg.SourceID)
					}
				}

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: выполненно удаление задачи ID: '%s', ТИП: '%s' для источика SourceID: '%d', список задач ПОСЛЕ: %v", msg.TaskID, msg.TaskType, msg.SourceID, qts.StorageList[msg.SourceID][msg.TaskID]),
					FuncName:    "NewRepositoryQTS",
				})

				msg.ChanRes <- msgRes

			case "change task status":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'change task status'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskStatus = msg.NewStatus
				qts.StorageList[msg.SourceID][msg.TaskID].TimeUpdate = time.Now().Unix()

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: изменение статуса '%s' для задачи ID: '%s', ТИП: '%s' для источика SourceID: '%d'", msg.NewStatus, msg.TaskID, msg.TaskType, msg.SourceID),
					FuncName:    "NewRepositoryQTS",
				})

				msg.ChanRes <- msgRes

			case "change availability connection on connection":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'change availability connection on connection'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].CheckingStatusItems.AvailabilityConnection = true

				msg.ChanRes <- msgRes

			case "change availability connection on disconnection":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'change availability connection on disconnection'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].CheckingStatusItems.AvailabilityConnection = false

				msg.ChanRes <- msgRes

			case "change availability files download":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'change availability files download'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].CheckingStatusItems.AvailabilityFilesDownload = true

				msg.ChanRes <- msgRes

			case "clear all file list":
				if !checkTaskID(&qts, msg.SourceID, msg.TaskID) {
					msgRes.ErrorDescription = fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID, 'clear all file list'", msg.SourceID)

					msg.ChanRes <- msgRes

					continue
				}

				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.DownloadList = []string{}
				qts.StorageList[msg.SourceID][msg.TaskID].TaskParameters.ConfirmedListFiles = map[string]*DetailedFilesInformation{}

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
		if info.TaskType == "download control" {
			return true
		}
	}

	return false
}

//SearchTaskForIDQueueTaskStorage поиск информации по ID задачи (внутренний task ID приложения)
func (qts *QueueTaskStorage) SearchTaskForIDQueueTaskStorage(taskID string) (int, *QueueTaskInformation, error) {
	var sourceID int
	sourceList := qts.GetAllSourcesQueueTaskStorage()

	if len(sourceList) == 0 {
		return sourceID, nil, errors.New("error, empty queue of pending tasks (SearchTaskForIDQueueTaskStorage)")
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
		return sourceID, nil, fmt.Errorf("error, task ID %v not found (SearchTaskForIDQueueTaskStorage)", taskID)
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
		TaskParameters:      msgRes.Settings,
	}
	qti.TaskType = msgRes.TaskType
	qti.IDClientAPI = msgRes.IDClientAPI
	qti.TaskIDClientAPI = msgRes.TaskIDClientAPI
	qti.UserName = msgRes.UserName

	return sourceID, &qti, nil
}

//SearchTaskForClientIDQueueTaskStorage поиск информации по ID задачи клиента API
func (qts *QueueTaskStorage) SearchTaskForClientIDQueueTaskStorage(clientTaskID string) (int, string, error) {
	var sourceID int
	var taskID string
	errMsg := fmt.Errorf("error, client task ID %v not found (SearchTaskForClientIDQueueTaskStorage)", clientTaskID)

	sourceList := qts.GetAllSourcesQueueTaskStorage()

	if len(sourceList) == 0 {
		return sourceID, taskID, errMsg
	}

	chanRes := make(chan chanResponse)
	defer close(chanRes)

DONE:
	for sID, tasks := range sourceList {
		for tID := range tasks {
			qts.ChannelReq <- chanRequest{
				Action:   "get information for task",
				SourceID: sID,
				TaskID:   tID,
				ChanRes:  chanRes,
			}

			msgRes := <-chanRes
			if msgRes.TaskIDClientAPI == clientTaskID {
				sourceID = sID
				taskID = tID
				errMsg = nil

				break DONE
			}
		}
	}

	return sourceID, taskID, errMsg
}

//GetQueueTaskStorage получить информацию по задаче
func (qts *QueueTaskStorage) GetQueueTaskStorage(sourceID int, taskID string) (QueueTaskInformation, error) {
	if !checkTaskID(qts, sourceID, taskID) {
		return QueueTaskInformation{}, fmt.Errorf("problem with ID %v not found, not correct sourceID or taskID (GetQueueTaskStorage)", sourceID)
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
		TaskParameters:      msgRes.Settings,
	}
	qti.TaskType = msgRes.TaskType
	qti.IDClientAPI = msgRes.IDClientAPI
	qti.TaskIDClientAPI = msgRes.TaskIDClientAPI
	qti.UserName = msgRes.UserName

	return qti, nil
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
	cr.UserName = cti.UserName

	qts.ChannelReq <- cr

	<-chanRes
}

//AddFiltrationParametersQueueTaskStorage добавляет параметры по фильтрации в существующую задачу
func (qts *QueueTaskStorage) AddFiltrationParametersQueueTaskStorage(sourceID int, taskID string, fp *FilteringOption) error {
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

//AddPathDirectoryFilteredFiles добавляет путь к директории на источнике в которой хранятся отфильтрованные файлы
func (qts *QueueTaskStorage) AddPathDirectoryFilteredFiles(sourceID int, taskID, pathDir string) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	options := &DescriptionParametersReceivedFromUser{
		PathDirectoryForFilteredFiles: pathDir,
	}

	cr := chanRequest{
		Action:           "add path directory for filtered files",
		SourceID:         sourceID,
		TaskID:           taskID,
		AdditionalOption: options,
		ChanRes:          chanRes,
	}

	qts.ChannelReq <- cr

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

//ChangeAvailabilityConnectionOnConnection изменить статус соединения с источником
func (qts *QueueTaskStorage) ChangeAvailabilityConnectionOnConnection(sourceID int, taskID string) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	qts.ChannelReq <- chanRequest{
		Action:   "change availability connection on connection",
		SourceID: sourceID,
		TaskID:   taskID,
		ChanRes:  chanRes,
	}

	return (<-chanRes).ErrorDescription
}

//ChangeAvailabilityConnectionOnDisconnection изменить статус соединения с источником
func (qts *QueueTaskStorage) ChangeAvailabilityConnectionOnDisconnection(sourceID int, taskID string) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	qts.ChannelReq <- chanRequest{
		Action:   "change availability connection on disconnection",
		SourceID: sourceID,
		TaskID:   taskID,
		ChanRes:  chanRes,
	}

	return (<-chanRes).ErrorDescription
}

//ChangeAvailabilityFilesDownload изменить статус наличия файлов для скачивания
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
func (qts *QueueTaskStorage) AddConfirmedListFiles(sourceID int, taskID string, clf map[string]DetailedFilesInformation) error {
	chanRes := make(chan chanResponse)
	defer close(chanRes)

	lf := make(map[string]*DetailedFilesInformation, len(clf))
	for f, v := range clf {
		lf[f] = &DetailedFilesInformation{
			Hex:          v.Hex,
			Size:         v.Size,
			IsLoaded:     v.IsLoaded,
			TimeDownload: v.TimeDownload,
		}
	}

	options := &DescriptionParametersReceivedFromUser{ConfirmedListFiles: lf}

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

//delQueueTaskStorage удалить задачу из очереди
func (qts *QueueTaskStorage) delQueueTaskStorage(sourceID int, taskID string) error {
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

//MessageInformationQueueTaskStorage краткая информация о задаче
type MessageInformationQueueTaskStorage struct {
	SourceID int
	TaskID   string
}

//CheckTimeQueueTaskStorage подпрограмма для отслеживания очередности выполнения задач
func (qts *QueueTaskStorage) CheckTimeQueueTaskStorage(isl *InformationSourcesList, sec int, saveMessageApp *savemessageapp.PathDirLocationLogFiles) chan MessageInformationQueueTaskStorage {
	type executionTasks struct {
		filtrationTask, downloadTask []string
	}

	et := executionTasks{
		filtrationTask: make([]string, 0, 10),
		downloadTask:   make([]string, 0, 1),
	}

	chanMsgInfoQueueTaskStorage := make(chan MessageInformationQueueTaskStorage)

	handlerTaskInfo := func(maxProcessFiltration, sourceID int, taskID string, taskInfo *QueueTaskInformation) {
		if (taskInfo.TaskType == "download control") && (!taskInfo.CheckingStatusItems.AvailabilityConnection) {
			et.downloadTask = make([]string, 0, 1)

			qts.ChangeTaskStatusQueueTask(sourceID, taskID, "wait")
		}

		/* удалаем задачи которые были успешно выполненны */
		if taskInfo.TaskStatus == "complete" {

			/*** Временное логирование событий ***/
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				TypeMessage: "info",
				Description: fmt.Sprintf("TEST Message: задача '%s' типа '%s', для источника '%d', помечена как выполненная, удалаем", taskID, taskInfo.TaskType, sourceID),
				FuncName:    "NewRepositoryQTS",
			})

			_ = qts.delQueueTaskStorage(sourceID, taskID)

			//удаляем задачу из списка отслеживания кол-ва выполняемых задач
			if taskInfo.TaskType == "download control" {
				et.downloadTask = make([]string, 0, 1)

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: удаляем задачу '%s', типа '%s', для источника '%d' из списка отслеживания кол-ва выполняемых задач по скачиванию файлов", taskID, taskInfo.TaskType, sourceID),
					FuncName:    "NewRepositoryQTS",
				})
			}

			for key, tID := range et.filtrationTask {
				if tID != taskID {
					continue
				}

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: удаляем задачу по фильтрации с ID '%s' из et.filtrationTask, список задач по фильтрации et.filtrationTask ДО удаления: '%v'", taskID, et.filtrationTask),
					FuncName:    "NewRepositoryQTS",
				})

				et.filtrationTask = append(et.filtrationTask[:key], et.filtrationTask[key+1:]...)

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: список задач по фильтрации et.filtrationTask ПОСЛЕ удаления: '%v'", et.filtrationTask),
					FuncName:    "NewRepositoryQTS",
				})
			}
		}

		/* удаляем 'старые' или 'подвисшие' задачи */

		//удаляем задачу по ФИЛЬТРАЦИИ, находящуюся в очереди более ПОЛУТРА ЧАСОВ
		if (taskInfo.TaskType == "filtration control") && (taskInfo.TaskStatus == "wait") && (time.Now().Unix() > (taskInfo.TimeUpdate + 54000)) {

			/*** Временное логирование событий ***/
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				TypeMessage: "info",
				Description: "TEST Message: удаляем задачу по ФИЛЬТРАЦИИ, находящуюся в очереди более полутра часов",
				FuncName:    "NewRepositoryQTS",
			})

			_ = qts.delQueueTaskStorage(sourceID, taskID)

			for key, tID := range et.filtrationTask {
				if tID != taskID {
					continue
				}

				et.filtrationTask = append(et.filtrationTask[:key], et.filtrationTask[key+1:]...)
			}
		}

		//удаляем задачу по СКАЧИВАНИЮ, находящуюся в очереди более ТРЕХ МИНУТ
		if (taskInfo.TaskType == "download control") && (taskInfo.TaskStatus == "wait") && (time.Now().Unix() > (taskInfo.TimeUpdate + 180)) {

			/*** Временное логирование событий ***/
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				TypeMessage: "info",
				Description: "TEST Message: удаляем задачу по СКАЧИВАНИЮ, находящуюся в очереди более 3 минут",
				FuncName:    "NewRepositoryQTS",
			})

			_ = qts.delQueueTaskStorage(sourceID, taskID)

			et.downloadTask = make([]string, 0, 1)
		}

		/* ОБРАБОТКА задач по фильтрации файлов */
		if taskInfo.TaskType == "filtration control" {
			if len(et.filtrationTask) == maxProcessFiltration {

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: колличество задач по фильтрации достигло максимального значения, пропускаем задачу '%s', для источника SourceID '%d'", taskID, sourceID),
					FuncName:    "NewRepositoryQTS",
				})

			}

			if (len(et.filtrationTask) <= maxProcessFiltration) && (taskInfo.TaskStatus == "wait") && taskInfo.CheckingStatusItems.AvailabilityConnection {
				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: "TEST Message: задача не выполнялась и источник подключен",
					FuncName:    "NewRepositoryQTS",
				})

				if err := qts.ChangeTaskStatusQueueTask(sourceID, taskID, "execution"); err == nil {
					//добавляем в массив выполняющихся задач
					et.filtrationTask = append(et.filtrationTask, taskID)

					/*** Временное логирование событий ***/
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						TypeMessage: "info",
						Description: fmt.Sprintf("TEST Message: ЗАПУСКАЕМ ВЫПОЛНЕНИЕ ЗАДАЧИ ID '%s' ПО ФИЛЬТРАЦИИ, ДЛЯ ИСТОЧНИКА '%d'", taskID, sourceID),
						FuncName:    "NewRepositoryQTS",
					})

					//запускаем выполнение задачи
					chanMsgInfoQueueTaskStorage <- MessageInformationQueueTaskStorage{
						SourceID: sourceID,
						TaskID:   taskID,
					}
				}
			}
		}

		/* ОБРАБОТКА задач по скачиванию файлов */
		if taskInfo.TaskType == "download control" {
			//выполняется ли задача
			if len(et.downloadTask) > 0 {

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: НЕВОЗМОЖНО запустить задачу ID '%s' по скачиванию файлов для источника '%d', список выполняемых задач не пуст '%v'", taskID, sourceID, et.downloadTask),
					FuncName:    "NewRepositoryQTS",
				})

				return
			}

			//если задача не выполнялась, источник подключен и есть файлы для скачивания
			if (taskInfo.TaskStatus == "wait") && taskInfo.CheckingStatusItems.AvailabilityConnection && taskInfo.CheckingStatusItems.AvailabilityFilesDownload {
				if err := qts.ChangeTaskStatusQueueTask(sourceID, taskID, "execution"); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						TypeMessage: "error",
						Description: fmt.Sprint(err),
						FuncName:    "NewRepositoryQTS",
					})
				}

				//добавляем в массив выполняющихся задач
				listTask := et.downloadTask
				listTask = append(listTask, taskID)
				et.downloadTask = listTask

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: ЗАПУСКАЕМ ВЫПОЛНЕНИЕ ЗАДАЧИ ID '%s' ПО СКАЧИВАНИЮ, ДЛЯ ИСТОЧНИКА '%d'", taskID, sourceID),
					FuncName:    "NewRepositoryQTS",
				})

				//запускаем выполнение задачи
				chanMsgInfoQueueTaskStorage <- MessageInformationQueueTaskStorage{
					SourceID: sourceID,
					TaskID:   taskID,
				}
			}
		}
	}

	//поиск и контроль количества задач на выполнение
	searchForTasksPerform := func(storageList map[int]map[string]*QueueTaskInformation) {
		for sourceID, tasks := range storageList {
			if len(tasks) == 0 {

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: никаких задач для источника с ID '%d' не обнаружено", sourceID),
					FuncName:    "NewRepositoryQTS",
				})

				continue
			}

			//получаем максимальное количество одновременно запущенных задач по фильтрации
			sourceSettings, sourceIsExist := isl.GetSourceSetting(sourceID)
			if !sourceIsExist {

				/*** Временное логирование событий ***/
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: fmt.Sprintf("TEST Message: источник '%d' не существует", sourceID),
					FuncName:    "NewRepositoryQTS",
				})

				continue
			}

			maxProcessFiltration := int(sourceSettings.Settings.MaxCountProcessFiltration)

			for taskID, taskInfo := range tasks {
				handlerTaskInfo(maxProcessFiltration, sourceID, taskID, taskInfo)
			}
		}
	}

	go func() {
		ticker := time.NewTicker(time.Duration(sec) * time.Second)
		for range ticker.C {
			//весь список источников
			storageList := qts.GetAllSourcesQueueTaskStorage()

			if len(storageList) == 0 {

				/*** Временное логирование событий ***/
				/*saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					TypeMessage: "info",
					Description: "TEST Message: отсутствует список источников",
					FuncName:    "NewRepositoryQTS",
				})*/

				continue
			}

			//поиск и контроль количества задач на выполнения
			searchForTasksPerform(storageList)
		}
	}()

	return chanMsgInfoQueueTaskStorage
}
