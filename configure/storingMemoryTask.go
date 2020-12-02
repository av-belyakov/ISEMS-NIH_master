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
// UserName - имя пользователя инициировавшего задачу (если поле пустое, то
//  считается что выполнение задачи было инициировано автоматически)
// TaskType - тип выполняемой задачи ("filtration control", "download control")
// TaskStatus - статус задачи, false выполняется, true завершена
// IsSlowDown - останавливается ли задача
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
	UserName                        string
	TaskType                        string
	TaskStatus                      bool
	IsSlowDown                      bool
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
	FiltrationTask               *FiltrationTaskParameters
	DownloadTask                 *DownloadTaskParameters
	ListFilesDetailedInformation map[string]*DetailedFilesInformation
}

//DownloadTaskParameters параметры задачи по скачиванию файлов
// ID - уникальный цифровой идентификатор источника
// Status - статус задачи 'wait'/'refused'/'execute'/'complete'/'stop'
// ('ожидает' / 'отклонена' / 'выполняется' / 'завершена' / 'остановлена')
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
}

//DetailedFilesInformation подробная информация о файлах
type DetailedFilesInformation struct {
	Hex          string
	Size         int64
	IsLoaded     bool
	TimeDownload int64
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
	ActionType, TaskID       string
	Description              *TaskDescription
	DetailedFilesInformation map[string]*DetailedFilesInformation
	ChannelRes               chan channelResSettings
}

//ChannelResSettings параметры канала ответа
type channelResSettings struct {
	IsExist                  bool
	TaskID                   string
	Description              *TaskDescription
	DetailedFilesInformation map[string]*DetailedFilesInformation
}

//NewRepositorySMT создание нового репозитория для хранения выполняемых задач
func NewRepositorySMT() *StoringMemoryTask {
	smt := StoringMemoryTask{
		tasks:      map[string]*TaskDescription{},
		channelReq: make(chan ChanStoringMemoryTask),
	}

	/*
		//------------ Memory Dump (START) ---------------
		const logFileName = "memdumpfile"

		fl, err := os.Create(logFileName)
		if err != nil {
			fmt.Printf("Create file %v, error: %v\n", logFileName, fmt.Sprint(err))
		}
		defer fl.Close()
		pprof.Lookup("heap").WriteTo(fl, 0)
		//------------ Memory Dump (END) ---------------
	*/

	go func() {
		for msg := range smt.channelReq {
			switch msg.ActionType {
			case "get task info":
				mr := channelResSettings{
					IsExist: false,
					TaskID:  msg.TaskID,
				}

				task, ok := smt.tasks[msg.TaskID]
				if ok {
					mr.IsExist = true
					mr.Description = task
				}

				msg.ChannelRes <- mr

			case "get list files detailed information":
				mr := channelResSettings{
					IsExist: false,
					TaskID:  msg.TaskID,
				}

				if task, ok := smt.tasks[msg.TaskID]; ok {
					mr.IsExist = true
					lfdi := make(map[string]*DetailedFilesInformation, len((*task).TaskParameter.ListFilesDetailedInformation))
					for fn, info := range task.TaskParameter.ListFilesDetailedInformation {
						lfdi[fn] = info
					}

					mr.DetailedFilesInformation = lfdi
				}

				msg.ChannelRes <- mr

			case "check task is exist":
				_, ok := smt.tasks[msg.TaskID]

				msg.ChannelRes <- channelResSettings{
					IsExist: ok,
					TaskID:  msg.TaskID,
				}

			case "add":
				smt.tasks[msg.TaskID] = msg.Description
				smt.tasks[msg.TaskID].TaskStatus = false
				smt.tasks[msg.TaskID].TaskParameter.ListFilesDetailedInformation = map[string]*DetailedFilesInformation{}

				if msg.Description.TaskParameter.DownloadTask.ID == 0 {
					smt.tasks[msg.TaskID].TaskParameter.DownloadTask = &DownloadTaskParameters{
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
				smt.tasks[msg.TaskID].TaskParameter.ListFilesDetailedInformation = map[string]*DetailedFilesInformation{}

				smt.tasks[msg.TaskID].TaskParameter.DownloadTask = &DownloadTaskParameters{
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

			case "is slow down":
				if _, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].IsSlowDown = true
				}

				msg.ChannelRes <- channelResSettings{
					TaskID: msg.TaskID,
				}

			case "timer update":
				if _, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TimeUpdate = time.Now().Unix()
				}

				msg.ChannelRes <- channelResSettings{}

			case "timer insert DB":
				if _, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TimeInsertDB = time.Now().Unix()
				}

				msg.ChannelRes <- channelResSettings{}

			case "delete":
				delete(smt.tasks, msg.TaskID)

				msg.ChannelRes <- channelResSettings{}

			case "update list files detailed information":
				smt.updateListFilesDetailedInformation(msg.TaskID, msg.DetailedFilesInformation)

				msg.ChannelRes <- channelResSettings{}

			case "update task filtration all parameters":
				smt.updateTaskFiltrationAllParameters(msg.TaskID, msg.Description)

				msg.ChannelRes <- channelResSettings{}

			case "update task download all parameters":
				smt.updateTaskDownloadAllParameters(msg.TaskID, msg.Description)

				msg.ChannelRes <- channelResSettings{}

			case "update list files detailed information file is loaded":
				smt.updateListFilesDetailedInformationFileIsLoaded(msg.TaskID, msg.DetailedFilesInformation)

				msg.ChannelRes <- channelResSettings{}

			case "increment number files downloaded":
				if ti, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TaskParameter.DownloadTask.NumberFilesDownloaded = ti.TaskParameter.DownloadTask.NumberFilesDownloaded + 1
				}

				msg.ChannelRes <- channelResSettings{}

			case "increment number files downloaded error":
				if ti, ok := smt.tasks[msg.TaskID]; ok {
					smt.tasks[msg.TaskID].TaskParameter.DownloadTask.NumberFilesDownloadedError = ti.TaskParameter.DownloadTask.NumberFilesDownloadedError + 1
				}

				msg.ChannelRes <- channelResSettings{}

			}
		}
	}()

	return &smt
}

//AddStoringMemoryTask добавить задачу
func (smt *StoringMemoryTask) AddStoringMemoryTask(taskID string, td TaskDescription) {
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

//CheckIsExistMemoryTask проверяет наличие задачи по ее ID
func (smt *StoringMemoryTask) CheckIsExistMemoryTask(taskID string) bool {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "check task is exist",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	info := <-chanRes

	return info.IsExist
}

//RecoverStoringMemoryTask восстанавливает всю информацию о выполяемой задаче
func (smt *StoringMemoryTask) RecoverStoringMemoryTask(td TaskDescription, taskID string) {
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

//IsSlowDownStoringMemoryTask отмечает задачу как находящуюся в процессе останова
func (smt *StoringMemoryTask) IsSlowDownStoringMemoryTask(taskID string) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "is slow down",
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
func (smt *StoringMemoryTask) GetStoringMemoryTask(taskID string) (*TaskDescription, bool) {
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

//GetListFilesDetailedInformation получить информацию со списком найденных в результате фильтрации файлах
func (smt *StoringMemoryTask) GetListFilesDetailedInformation(taskID string) (map[string]*DetailedFilesInformation, bool) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "get list files detailed information",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	info := <-chanRes

	return info.DetailedFilesInformation, info.IsExist
}

//IncrementNumberFilesDownloaded увеличить кол-во успешно скаченных файлов на 1
func (smt *StoringMemoryTask) IncrementNumberFilesDownloaded(taskID string) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "increment number files downloaded",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	<-chanRes
}

//IncrementNumberFilesDownloadedError увеличить кол-во успешно скаченных файлов на 1
func (smt *StoringMemoryTask) IncrementNumberFilesDownloadedError(taskID string) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "increment number files downloaded error",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	<-chanRes
}

//UpdateListFilesDetailedInformation обновление информации о файлах
func (smt *StoringMemoryTask) UpdateListFilesDetailedInformation(taskID string, lfdi map[string]*DetailedFilesInformation) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType:               "update list files detailed information",
		TaskID:                   taskID,
		DetailedFilesInformation: lfdi,
		ChannelRes:               chanRes,
	}

	<-chanRes
}

//UpdateTaskFiltrationAllParameters обновление параметров выполнения задачи по фильтрации
func (smt *StoringMemoryTask) UpdateTaskFiltrationAllParameters(taskID string, ftp *FiltrationTaskParameters) {
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
func (smt *StoringMemoryTask) UpdateTaskDownloadAllParameters(taskID string, dtp *DownloadTaskParameters) {
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

//UpdateListFilesDetailedInformationFileIsLoaded обновление информации о файлах
func (smt *StoringMemoryTask) UpdateListFilesDetailedInformationFileIsLoaded(taskID string, lfdi map[string]*DetailedFilesInformation) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType:               "update list files detailed information file is loaded",
		TaskID:                   taskID,
		DetailedFilesInformation: lfdi,
		ChannelRes:               chanRes,
	}

	<-chanRes
}

//delStoringMemoryTask удалить задачу
func (smt StoringMemoryTask) delStoringMemoryTask(taskID string) {
	chanRes := make(chan channelResSettings)
	defer close(chanRes)

	smt.channelReq <- ChanStoringMemoryTask{
		ActionType: "delete",
		TaskID:     taskID,
		ChannelRes: chanRes,
	}

	<-chanRes
}

//updateListFilesDetailedInformation обновляем информацию о файлах (при фильтрации)
func (smt *StoringMemoryTask) updateListFilesDetailedInformation(taskID string, dfi map[string]*DetailedFilesInformation) {
	//	fmt.Println("StoringMemoryTask - func 'updateListFilesDetailedInformation', START...")

	if _, ok := smt.tasks[taskID]; !ok {
		return
	}

	for fn, fi := range dfi {
		smt.tasks[taskID].TaskParameter.ListFilesDetailedInformation[fn] = fi
	}
}

//updateListFilesDetailedInformationFileIsLoaded обновляем информацию о файлах (при скачивании)
func (smt *StoringMemoryTask) updateListFilesDetailedInformationFileIsLoaded(taskID string, dfi map[string]*DetailedFilesInformation) {
	if _, ok := smt.tasks[taskID]; !ok {
		return
	}

	for fn, fi := range dfi {
		if _, ok := smt.tasks[taskID].TaskParameter.ListFilesDetailedInformation[fn]; ok {
			smt.tasks[taskID].TaskParameter.ListFilesDetailedInformation[fn].Hex = fi.Hex
			smt.tasks[taskID].TaskParameter.ListFilesDetailedInformation[fn].Size = fi.Size
			smt.tasks[taskID].TaskParameter.ListFilesDetailedInformation[fn].IsLoaded = fi.IsLoaded
			smt.tasks[taskID].TaskParameter.ListFilesDetailedInformation[fn].TimeDownload = fi.TimeDownload
		}
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

			for id, t := range smt.tasks {
				if t.TaskStatus && ((t.TimeUpdate + 180) < timeNow) {
					//если задача выполнена и прошло какое то время удаляем ее
					smt.delStoringMemoryTask(id)

					continue
				}

				if (t.TimeUpdate + 481) < timeNow {
					smt.CompleteStoringMemoryTask(id)

					chanOut <- MsgChanStoringMemoryTask{
						ID:          id,
						Type:        "warning",
						Description: fmt.Sprintf("информация по задаче с ID %q достаточно долго не обновлялась, возможно выполнение задачи было приостановленно", id),
					}
				}
			}
		}
	}()

	return chanOut
}
