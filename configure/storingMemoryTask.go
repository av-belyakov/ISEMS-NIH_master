package configure

import (
	"fmt"
	"time"
)

//StoringMemoryTask описание типа в котором храняться описание и ID выполняемых задач
// ключом отображения является уникальный идентификатор задачи
type StoringMemoryTask struct {
	tasks      map[string]*TaskDescription
	channelReq chan ChanStoringMemoryTask
}

//TaskDescription описание задачи
// ClientID - уникальный идентификатор клиента
// ClientTaskID - идентификатор задачи полученный от клиента
// TaskType - тип выполняемой задачи ("filtration control", "download control")
// TaskStatus - статус задачи, false выполняется, true завершена
// ModuleThatSetTask - модуль от которого поступила задача
// ModuleResponsibleImplementation - модуль который должен выполнить обработку
// TimeUpdate - время последнего обновления в формате Unix
// TimeInsertDB - время последней вставки в БД
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
	TimeInsertDB                    int64
	TimeInterval                    TimeIntervalTaskExecution
	TaskParameter                   DescriptionTaskParameters
	ChanStopTransferListFiles       chan struct{}
}

//TimeIntervalTaskExecution временной интервал выполнения задачи
type TimeIntervalTaskExecution struct {
	Start, End int64
}

//DescriptionTaskParameters описание параметров задачи
type DescriptionTaskParameters struct {
	FiltrationTask FiltrationTaskParameters
	DownloadTask   DownloadTaskParameters
}

//DownloadTaskParameters параметры задачи по скачиванию файлов
// ID - уникальный цифровой идентификатор источника
// Status - статус задачи 'wait'/'refused'/'execute'/'complete'/'stop' ('ожидает' / 'отклонена' / 'выполняется' / 'завершена' / 'остановлена')
// NumberFilesTotal - всего файлов предназначенных для скачивания
// NumberFilesDownloaded - кол-во загруженных файлов
// NumberFilesDownloadedError - кол-во загруженных с ошибкой файлов
// PathDirectoryStorageDownloadedFiles - директория в которую осуществляется скачивание файлов
// FileInformation - подробная информация о передаваемом файле
// DownloadingFilesInformation - информация о скачиваемых файлах, ключ - имя файла
type DownloadTaskParameters struct {
	ID                                  int
	Status                              string
	NumberFilesTotal                    int
	NumberFilesDownloaded               int
	NumberFilesDownloadedError          int
	PathDirectoryStorageDownloadedFiles string
	FileInformation                     DetailedFileInformation
	DownloadingFilesInformation         map[string]*DownloadFilesInformation
}

//DownloadFilesInformation подробная информация о скачиваемых файлах
type DownloadFilesInformation struct {
	FoundFilesInformation
	IsLoaded     bool
	TimeDownload int64
}

//FoundFilesInformation подробная информация о файлах
type FoundFilesInformation struct {
	Size int64
	Hex  string
}

//FiltrationTaskParameters параметры задачи по фильтрации файлов
// ID - уникальный цифровой идентификатор источника
// Status - статус задачи 'wait'/'refused'/'execute'/'complete'/'stop' ('ожидает' / 'отклонена' / 'выполняется' / 'завершена' / 'остановлена')
// UseIndex - используется ли индекс для поиска файлов
// NumberFilesMeetFilterParameters - кол-во файлов удовлетворяющих параметрам фильтрации
// NumberProcessedFiles - кол-во обработанных файлов
// NumberFilesFoundResultFiltering - кол-во найденных, в результате фильтрации, файлов
// NumberDirectoryFiltartion - кол-во директорий по которым выполняется фильтрация
// NumberErrorProcessedFiles - кол-во не обработанных файлов или файлов обработанных с ошибками
// SizeFilesMeetFilterParameters - общий размер файлов (в байтах) удовлетворяющих параметрам фильтрации
// SizeFilesFoundResultFiltering - общий размер найденных, в результате фильтрации, файлов (в байтах)
// PathStorageSource — путь до директории в которой сохраняются файлы
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

//DetailedFileInformation подробная информация о передаваемом файле
// Name - имя файла
// Hex - хеш сумма
// FullSizeByte - полный размер файла в байтах
// AcceptedSizeByte - принятый размер файла в байтах
// AcceptedSizePercent - принятый размер файла в процентах
// NumChunk - общее кол-во кусочков для передачи
// ChunkSize - размер передоваемого кусочка
// NumAcceptedChunk - кол-во принятых кусочков
type DetailedFileInformation struct {
	Name                string
	Hex                 string
	FullSizeByte        int64
	AcceptedSizeByte    int64
	AcceptedSizePercent int
	NumChunk            int
	ChunkSize           int
	NumAcceptedChunk    int
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
				smt.tasks[msg.TaskID].TaskStatus = false
				smt.tasks[msg.TaskID].TaskParameter.FiltrationTask.FoundFilesInformation = map[string]*FoundFilesInformation{}

				if msg.Description.TaskParameter.DownloadTask.ID == 0 {
					smt.tasks[msg.TaskID].TaskParameter.DownloadTask = DownloadTaskParameters{
						Status: "not executed",
					}
				} else {
					smt.tasks[msg.TaskID].TaskParameter.DownloadTask = msg.Description.TaskParameter.DownloadTask
				}

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "recover":
				smt.tasks[msg.TaskID] = msg.Description
				smt.tasks[msg.TaskID].TaskParameter.FiltrationTask.FoundFilesInformation = map[string]*FoundFilesInformation{}
				smt.tasks[msg.TaskID].TaskParameter.DownloadTask = DownloadTaskParameters{
					Status: "not executed",
				}

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "complete":
				if _, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TaskStatus = true
				}

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "timer update":
				if _, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TimeUpdate = time.Now().Unix()
				}

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "timer insert DB":
				if _, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TimeInsertDB = time.Now().Unix()
				}

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "delete":
				delete(smt.tasks, msg.TaskID)

			case "update task filtration all parameters":
				smt.updateTaskFiltrationAllParameters(msg.TaskID, msg.Description)

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "update task download all parameters":
				smt.updateTaskDownloadAllParameters(msg.TaskID, msg.Description)

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "update task download file is loaded":
				smt.updateTaskDownloadFileIsLoaded(msg.TaskID, msg.Description)

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "increment number files downloaded":
				if ti, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TaskParameter.DownloadTask.NumberFilesDownloaded = ti.TaskParameter.DownloadTask.NumberFilesDownloaded + 1
				}

			case "increment number files downloaded error":
				if ti, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TaskParameter.DownloadTask.NumberFilesDownloadedError = ti.TaskParameter.DownloadTask.NumberFilesDownloadedError + 1
				}

			}
		}
	}()

	return &smt
}

//AddStoringMemoryTask добавить задачу
func (smt StoringMemoryTask) AddStoringMemoryTask(taskID string, td TaskDescription) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType:  "add",
		TaskID:      taskID,
		ChannelRes:  chanRes,
		Description: &td,
	}

	<-chanRes
}

//RecoverStoringMemoryTask восстанавливает всю информацию о выполяемой задаче
func (smt StoringMemoryTask) RecoverStoringMemoryTask(td TaskDescription, taskID string) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType:  "recover",
		TaskID:      taskID,
		ChannelRes:  chanRes,
		Description: &td,
	}

	<-chanRes
}

//CompleteStoringMemoryTask установить статус 'выполненно' для задачи
func (smt *StoringMemoryTask) CompleteStoringMemoryTask(taskID string) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "complete",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	<-chanRes
}

//TimerUpdateStoringMemoryTask обновить значение таймера для задачи
func (smt *StoringMemoryTask) TimerUpdateStoringMemoryTask(taskID string) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "timer update",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	<-chanRes
}

//TimerUpdateTaskInsertDB обновить значение таймера для задачи выполняемой в БД
func (smt *StoringMemoryTask) TimerUpdateTaskInsertDB(taskID string) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "timer insert DB",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	<-chanRes
}

//GetStoringMemoryTask получить информацию о задаче по ее ID
func (smt StoringMemoryTask) GetStoringMemoryTask(taskID string) (*TaskDescription, bool) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "get task info",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	info := <-chanRes

	return info.Description, info.IsExist
}

//GetAllStoringMemoryTask получить все ID задач для выбранного клиента
/*func (smt StoringMemoryTask) GetAllStoringMemoryTask(clientID string) []string {
	foundTask := make([]string, 0, len(smt.tasks))

	for tid, v := range smt.tasks {
		if clientID == v.ClientID {
			foundTask = append(foundTask, tid)
		}
	}

	return foundTask
}

//GetStoringMemoryTaskForClientID получить всю инофрмацию о задаче по ID клиента и taskID клиента
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
}*/

//IncrementNumberFilesDownloaded увеличить кол-во успешно скаченных файлов на 1
func (smt StoringMemoryTask) IncrementNumberFilesDownloaded(taskID string) {
	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "increment number files downloaded",
		TaskID:     taskID,
	}
}

//IncrementNumberFilesDownloadedError увеличить кол-во успешно скаченных файлов на 1
func (smt StoringMemoryTask) IncrementNumberFilesDownloadedError(taskID string) {
	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "increment number files downloaded error",
		TaskID:     taskID,
	}
}

//UpdateTaskFiltrationAllParameters управление задачами по фильтрации
func (smt *StoringMemoryTask) UpdateTaskFiltrationAllParameters(taskID string, ftp FiltrationTaskParameters) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

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

	<-chanRes
}

//UpdateTaskDownloadAllParameters обновление параметров скачивания файлов
func (smt *StoringMemoryTask) UpdateTaskDownloadAllParameters(taskID string, dtp DownloadTaskParameters) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "update task download all parameters",
		TaskID:     taskID,
		Description: &TaskDescription{
			TaskParameter: DescriptionTaskParameters{
				DownloadTask: dtp,
			},
		},
		ChannelRes: chanRes,
	}

	<-chanRes
}

//UpdateTaskDownloadFileIsLoaded обновление параметра - файл загружен
func (smt *StoringMemoryTask) UpdateTaskDownloadFileIsLoaded(taskID string, dtp DownloadTaskParameters) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "update task download file is loaded",
		TaskID:     taskID,
		Description: &TaskDescription{
			TaskParameter: DescriptionTaskParameters{
				DownloadTask: dtp,
			},
		},
		ChannelRes: chanRes,
	}

	<-chanRes
}

//delStoringMemoryTask удалить задачу
func (smt StoringMemoryTask) delStoringMemoryTask(taskID string) {
	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "delete",
		TaskID:     taskID,
	}
}

func (smt *StoringMemoryTask) updateTaskFiltrationAllParameters(taskID string, td *TaskDescription) {
	if _, ok := smt.tasks[taskID]; !ok {
		return
	}

	//изменяем время окончания задачи
	smt.tasks[taskID].TimeInterval.End = time.Now().Unix()

	ft := smt.tasks[taskID].TaskParameter.FiltrationTask
	nft := td.TaskParameter.FiltrationTask

	for fn, fi := range nft.FoundFilesInformation {
		ft.FoundFilesInformation[fn] = fi
	}

	ft.Status = nft.Status
	ft.NumberFilesMeetFilterParameters = nft.NumberFilesMeetFilterParameters
	ft.NumberFilesFoundResultFiltering = nft.NumberFilesFoundResultFiltering
	ft.NumberErrorProcessedFiles = nft.NumberErrorProcessedFiles
	ft.NumberDirectoryFiltartion = nft.NumberDirectoryFiltartion
	ft.NumberProcessedFiles = nft.NumberProcessedFiles
	ft.SizeFilesMeetFilterParameters = nft.SizeFilesMeetFilterParameters
	ft.SizeFilesFoundResultFiltering = nft.SizeFilesFoundResultFiltering
	ft.PathStorageSource = nft.PathStorageSource

	smt.tasks[taskID].TaskParameter.FiltrationTask = ft
}

func (smt *StoringMemoryTask) updateTaskDownloadAllParameters(taskID string, td *TaskDescription) {
	if _, ok := smt.tasks[taskID]; !ok {
		return
	}

	//изменяем время окончания задачи
	smt.tasks[taskID].TimeInterval.End = time.Now().Unix()

	dt := smt.tasks[taskID].TaskParameter.DownloadTask
	ndt := td.TaskParameter.DownloadTask

	dt.Status = ndt.Status
	dt.NumberFilesTotal = ndt.NumberFilesTotal
	dt.NumberFilesDownloaded = ndt.NumberFilesDownloaded
	dt.NumberFilesDownloadedError = ndt.NumberFilesDownloadedError
	dt.PathDirectoryStorageDownloadedFiles = ndt.PathDirectoryStorageDownloadedFiles
	dt.FileInformation = ndt.FileInformation

	smt.tasks[taskID].TaskParameter.DownloadTask = dt
}

func (smt *StoringMemoryTask) updateTaskDownloadFileIsLoaded(taskID string, td *TaskDescription) {
	if _, ok := smt.tasks[taskID]; !ok {
		return
	}

	for fn := range td.TaskParameter.DownloadTask.DownloadingFilesInformation {
		if _, ok := smt.tasks[taskID].TaskParameter.DownloadTask.DownloadingFilesInformation[fn]; ok {
			smt.tasks[taskID].TaskParameter.DownloadTask.DownloadingFilesInformation[fn].IsLoaded = true
			smt.tasks[taskID].TaskParameter.DownloadTask.DownloadingFilesInformation[fn].TimeDownload = time.Now().Unix()
		}
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
			if len(smt.tasks) == 0 {
				continue
			}

			timeNow := time.Now().Unix()

			//fmt.Println("task count =", len(smt.tasks))

			for id, t := range smt.tasks {

				//fmt.Printf("Next Tick %v\n task status:%v, time task:%v < time now:%v (%v)\n", time.Now(), t.TaskStatus, (t.TimeUpdate + 60), timeNow, ((t.TimeUpdate + 60) < timeNow))

				if t.TaskStatus && ((t.TimeUpdate + 60) < timeNow) {

					fmt.Printf("//////// func 'CheckTimeUpdateStoringMemoryTask' ****** delete task ID - %v\n", id)

					//если задача выполнена и прошло какое то время удаляем ее
					smt.delStoringMemoryTask(id)

					continue
				}

				if (t.TimeUpdate + 121) < timeNow {
					smt.CompleteStoringMemoryTask(id)

					chanOut <- MsgChanStoringMemoryTask{
						ID:          id,
						Type:        "warning",
						Description: fmt.Sprintf("информация по задаче с ID %v достаточно долго не обновлялась, возможно выполнение задачи было приостановленно", id),
					}
				}
			}
		}
	}()

	return chanOut
}
