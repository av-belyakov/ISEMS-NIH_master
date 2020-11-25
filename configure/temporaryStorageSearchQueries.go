package configure

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

//TemporaryStorageSearchQueries временное хранилище задач поиска информации в БД
// tasks - список задач поиска
// tickerSec - интервал обработки информации в кэше
// maxCacheSize - максимальный размер кэша
// timeExpiration - время устаревания кэша
// channelReq - канал для передачи информации о запросе внутри хранилища
type TemporaryStorageSearchQueries struct {
	tasks          map[string]*SearchTaskDescription
	tickerSec      int
	maxCacheSize   int
	timeExpiration int
	channelReq     chan SearchChannelRequest
}

//SearchTaskDescription описание задачи по поиску информации в БД
// UpdateTimeInformation - время обновления информации
// NotRelevance - статус актуальности информации (false - актуальна)
// TransmissionStatus - передается ли найденная информация клиенту API (актуально когда
//  найденной информации много и она передается клиенту API частями) true - передается
// SearchParameters - описание параметров поискового запроса
// SummarySearchQueryProcessingResults - краткие результаты обработки поискового запроса
// ListFoundInformation - список найденной информации
type SearchTaskDescription struct {
	UpdateTimeInformation               int64
	NotRelevance                        bool
	TransmissionStatus                  bool
	SearchParameters                    SearchParameters
	SummarySearchQueryProcessingResults SummarySearchQueryProcessingResultsDetailed
	ListFoundInformation                ListFoundInformation
}

//SearchParameters параметры поиска
type SearchParameters SearchInformationAboutTasksRequestOption

//ListFoundInformation список найденной информаци
type ListFoundInformation struct {
	List []*BriefTaskInformation
}

//SummarySearchQueryProcessingResultsDetailed краткие результаты обработки поискового запроса
// NumFoundTasks - количество найденных задач
type SummarySearchQueryProcessingResultsDetailed struct {
	NumFoundTasks int64
}

//SearchChannelRequest параметры канала запроса
// actionType - тип действия
// searchTaskID - уникальный идентификатор задачи
// information - информация
// channelRes - описание канала ответа
type SearchChannelRequest struct {
	actionType   string
	searchTaskID string
	information  interface{}
	channelRes   chan SearchChannelResponse
}

//SearchChannelResponse параметры канала ответа
// errMsg - содержит ошибку
// taskID - уникальный идентификатор задачи
// findInformation - найденная по заданному ID информации
type SearchChannelResponse struct {
	errMsg          error
	findInformation *SearchTaskDescription
}

//TemporaryStorageSearcher интерфейс TemporaryStorageSearcher
type TemporaryStorageSearcher interface {
	CreateNewSearchTask(string, *SearchParameters) (string, *SearchTaskDescription)
	GetInformationAboutSearchTask(string) (*SearchTaskDescription, error)
}

type listTask []listTaskInfo

type listTaskInfo struct {
	time         int64
	id           string
	transmission bool
}

func (lt listTask) Len() int {
	return len(lt)
}

func (lt listTask) Swap(i, j int) {
	lt[i], lt[j] = lt[j], lt[i]
}

func (lt listTask) Less(i, j int) bool {
	return lt[i].time < lt[j].time
}

//CreateTmpStorageID генерирует идентификатор задачи поиска в БД по заданным параметрам
func CreateTmpStorageID(clientID string, sp *SearchParameters) string {
	boolStr := func(b bool) string {
		if b {
			return "true"
		}

		return "false"
	}

	nf := sp.InstalledFilteringOption.NetworkFilters

	s := []string{
		boolStr(sp.TaskProcessed),
		strconv.Itoa(sp.ID),
		sp.StatusFilteringTask,
		sp.StatusFileDownloadTask,
		boolStr(sp.ConsiderParameterFilesIsDownloaded),
		boolStr(sp.FilesIsDownloaded),
		boolStr(sp.ConsiderParameterAllFilesIsDownloaded),
		boolStr(sp.AllFilesIsDownloaded),
		boolStr(sp.ConsiderParameterTaskProcessed),
		boolStr(sp.TaskProcessed),
		boolStr(sp.InformationAboutFiltering.FilesIsFound),
		strconv.Itoa(sp.InformationAboutFiltering.CountAllFilesMin),
		strconv.Itoa(sp.InformationAboutFiltering.CountAllFilesMax),
		strconv.FormatInt(sp.InformationAboutFiltering.SizeAllFilesMin, 10),
		strconv.FormatInt(sp.InformationAboutFiltering.SizeAllFilesMax, 10),
		strconv.FormatInt(sp.InstalledFilteringOption.DateTime.Start, 10),
		strconv.FormatInt(sp.InstalledFilteringOption.DateTime.End, 10),
		sp.InstalledFilteringOption.Protocol,
	}
	s = append(s, nf.IP.Any...)
	s = append(s, nf.IP.Dst...)
	s = append(s, nf.IP.Src...)
	s = append(s, nf.Port.Any...)
	s = append(s, nf.Port.Dst...)
	s = append(s, nf.Port.Src...)
	s = append(s, nf.Network.Any...)
	s = append(s, nf.Network.Dst...)
	s = append(s, nf.Network.Src...)
	s = append(s, strconv.FormatInt(time.Now().Unix(), 10))

	h := sha1.New()
	io.WriteString(h, strings.Join(s, "_"))

	return hex.EncodeToString(h.Sum(nil))
}

//TypeRepositoryTSSQ описание типа для NewRepositoryTSSQ
// TickerSec - интервал обработки информации в кэше
// TimeExpiration - через какое количество секунд кэша будет считаться устаревшим
// MaxCacheSize - максимальный размер кэша
type TypeRepositoryTSSQ struct {
	TickerSec      int
	TimeExpiration int
	MaxCacheSize   int
}

//NewRepositoryTSSQ создание нового репозитория для хранения информации о задачах поиска
func NewRepositoryTSSQ(tr TypeRepositoryTSSQ) *TemporaryStorageSearchQueries {
	if tr.MaxCacheSize <= 10 || tr.MaxCacheSize > 1000 {
		tr.MaxCacheSize = 100
	}

	if tr.TickerSec <= 1 || tr.TickerSec > 10 {
		tr.TickerSec = 5
	}

	if tr.TimeExpiration > 180 {
		tr.TimeExpiration = 180
	}

	tssq := TemporaryStorageSearchQueries{
		tickerSec:      tr.TickerSec,
		maxCacheSize:   tr.MaxCacheSize,
		timeExpiration: tr.TimeExpiration,
		tasks:          map[string]*SearchTaskDescription{},
		channelReq:     make(chan SearchChannelRequest),
	}

	go func() {
		for msg := range tssq.channelReq {
			switch msg.actionType {
			case "create new search task":

				fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', start")

				searchParameters, ok := msg.information.(SearchParameters)
				if !ok {

					fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', error 1")

					msg.channelRes <- SearchChannelResponse{
						errMsg: fmt.Errorf("ActionType: %q, ERROR: type conversion error", msg.actionType),
					}

					close(msg.channelRes)

					break
				}

				fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', 2222")

				fmt.Printf("count tasks: '%v', count max cache size: '%v'\n", len(tssq.tasks), tssq.maxCacheSize)

				//проверяем количество записей в кэше
				/*if len(tssq.tasks) > tssq.maxCacheSize-5 {

					fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', delete")

					//tssq.deleteOldestRecord()
					ls := make(listTask, 0, len(tssq.tasks))

					fmt.Println(ls)

					for key, value := range tssq.tasks {

						fmt.Printf("update time information: '%v'\n", value.UpdateTimeInformation)

						ls = append(ls, listTaskInfo{
							id:           key,
							time:         value.UpdateTimeInformation,
							transmission: value.TransmissionStatus,
						})
					}

					sort.Sort(ls)

					fmt.Println(ls)

					//удаляем самую старую запись (если информация по задаче не передается)
					for _, info := range ls {
						if info.transmission {
							continue
						}

						fmt.Println(info)

						delete(tssq.tasks, msg.searchTaskID)
					}
				}*/

				fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', 3333")

				if info, err := tssq.temporaryStorageSearchInfo(msg.searchTaskID); err != nil {

					fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', err 4444")

					//добавляем информацию
					tssq.tasks[msg.searchTaskID] = &SearchTaskDescription{
						UpdateTimeInformation: 0, //time.Now().Unix(),
						SearchParameters:      searchParameters,
					}
				} else {

					fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', 5555")

					//проверяем задачу на актуальность
					if info.NotRelevance {

						fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', 6666")

						tssq.delInformationAboutSearchTask(msg.searchTaskID)

						fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', 7777")

						tssq.tasks[msg.searchTaskID] = &SearchTaskDescription{
							UpdateTimeInformation: 0, //time.Now().Unix(),
							SearchParameters:      searchParameters,
						}
					}

					fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', 8888")

					msg.channelRes <- SearchChannelResponse{
						errMsg:          nil,
						findInformation: info,
					}
				}

				fmt.Println("NewRepositoryTSSQ, action type: 'create new search task', close channel")

				close(msg.channelRes)

			case "get information about search task":
				info, err := tssq.temporaryStorageSearchInfo(msg.searchTaskID)

				msg.channelRes <- SearchChannelResponse{
					errMsg:          err,
					findInformation: info,
				}

				close(msg.channelRes)

			case "add information found search result":

				fmt.Println("-----=== 'add information found search result' START ===------")

				lbti, ok := msg.information.([]*BriefTaskInformation)
				if !ok {

					fmt.Println("-----=== 'add information found search result' ERROR 111 ===------")

					msg.channelRes <- SearchChannelResponse{
						errMsg: fmt.Errorf("ActionType: %q, ERROR: type conversion error", msg.actionType),
					}

					close(msg.channelRes)

					break
				}

				if _, err := tssq.temporaryStorageSearchInfo(msg.searchTaskID); err != nil {

					fmt.Println("-----=== 'add information found search result' ERROR 222 ===------")

					msg.channelRes <- SearchChannelResponse{errMsg: err}
				} else {

					fmt.Println("-----=== 'add information found search result' 333 ===------")

					tssq.tasks[msg.searchTaskID].ListFoundInformation.List = lbti
					tssq.tasks[msg.searchTaskID].UpdateTimeInformation = time.Now().Unix()
				}

				fmt.Println("func 'NewRepositoryTSSQ', action type 'add information found search result'")

				fmt.Println("-----=== 'add information found search result' END ===------")

				close(msg.channelRes)

			case "add count document found search result":
				cd, ok := msg.information.(int64)
				if !ok {
					msg.channelRes <- SearchChannelResponse{
						errMsg: fmt.Errorf("ActionType: %q, ERROR: type conversion error", msg.actionType),
					}

					close(msg.channelRes)

					break
				}

				if _, err := tssq.temporaryStorageSearchInfo(msg.searchTaskID); err != nil {
					msg.channelRes <- SearchChannelResponse{errMsg: err}
				} else {
					tssq.tasks[msg.searchTaskID].UpdateTimeInformation = time.Now().Unix()
					tssq.tasks[msg.searchTaskID].SummarySearchQueryProcessingResults.NumFoundTasks = cd
				}

				close(msg.channelRes)

			case "del information about search task":
				delete(tssq.tasks, msg.searchTaskID)

				close(msg.channelRes)

			case "change status transmission task":
				status, ok := msg.information.(bool)
				if !ok {
					msg.channelRes <- SearchChannelResponse{
						errMsg: fmt.Errorf("ActionType: %q, ERROR: type conversion error", msg.actionType),
					}
				} else {
					msg.channelRes <- SearchChannelResponse{
						errMsg: tssq.changeStatusTransmissionTask(msg.searchTaskID, status),
					}
				}

				close(msg.channelRes)

			case "change status information relevance":
				ltid, ok := msg.information.([]string)
				if !ok {
					msg.channelRes <- SearchChannelResponse{
						errMsg: fmt.Errorf("ActionType: %q, ERROR: type conversion error", msg.actionType),
					}
				}

				tssq.changingStatusInformationRelevance(ltid)

				close(msg.channelRes)

			case "change update time information":
				if _, err := tssq.temporaryStorageSearchInfo(msg.searchTaskID); err != nil {
					msg.channelRes <- SearchChannelResponse{errMsg: err}
				} else {
					tssq.tasks[msg.searchTaskID].UpdateTimeInformation = time.Now().Unix()
				}

				close(msg.channelRes)
			}
		}
	}()

	go checkTimeDeleteTemporaryStorageSearchQueries(&tssq)

	return &tssq
}

//CreateNewSearchTask создание новой временной записи о поисковой задаче
func (tssq *TemporaryStorageSearchQueries) CreateNewSearchTask(clientID string, sp *SearchParameters) (string, *SearchTaskDescription, error) {
	taskID := CreateTmpStorageID(clientID, sp)

	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "create new search task",
		searchTaskID: taskID,
		information:  *sp,
		channelRes:   chanRes,
	}

	info := <-chanRes

	return taskID, info.findInformation, info.errMsg
}

//GetInformationAboutSearchTask вывод всей найденной информации из кэша
func (tssq *TemporaryStorageSearchQueries) GetInformationAboutSearchTask(taskID string) (*SearchTaskDescription, error) {
	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "get information about search task",
		searchTaskID: taskID,
		channelRes:   chanRes,
	}

	info := <-chanRes

	return info.findInformation, info.errMsg
}

//AddInformationFoundSearchResult добавление результата поиска в БД к информации о задаче
func (tssq *TemporaryStorageSearchQueries) AddInformationFoundSearchResult(taskID string, lbti []*BriefTaskInformation) error {
	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "add information found search result",
		searchTaskID: taskID,
		information:  lbti,
		channelRes:   chanRes,
	}

	return (<-chanRes).errMsg
}

//AddCountDocumentFoundSearchResult добавление информации о количестве найденных, в результате поиска, документов
func (tssq *TemporaryStorageSearchQueries) AddCountDocumentFoundSearchResult(taskID string, cd int64) error {
	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "add count document found search result",
		searchTaskID: taskID,
		information:  cd,
		channelRes:   chanRes,
	}

	return (<-chanRes).errMsg
}

//ChangeStatusTransmissionTask изменение статуса передачи задачи (в данный момент информация по данной задаче передается или нет)
func (tssq *TemporaryStorageSearchQueries) ChangeStatusTransmissionTask(taskID string, transmissionStatus bool) error {
	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "change status transmission task",
		searchTaskID: taskID,
		information:  transmissionStatus,
		channelRes:   chanRes,
	}

	return (<-chanRes).errMsg
}

//ChangingStatusInformationRelevance изменение статуса актуальности задачи
func (tssq *TemporaryStorageSearchQueries) ChangingStatusInformationRelevance(listTaskID []string) {
	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:  "change status information relevance",
		information: listTaskID,
		channelRes:  chanRes,
	}
}

//ChangeUpdateTimeInformation изменение времени обновления информации
func (tssq *TemporaryStorageSearchQueries) ChangeUpdateTimeInformation(taskID string) error {
	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "change update time information",
		searchTaskID: taskID,
		channelRes:   chanRes,
	}

	return (<-chanRes).errMsg
}

func (tssq *TemporaryStorageSearchQueries) temporaryStorageSearchInfo(taskID string) (*SearchTaskDescription, error) {
	info, ok := tssq.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task with ID %q not found", taskID)
	}

	return info, nil
}

func (tssq *TemporaryStorageSearchQueries) changeStatusTransmissionTask(taskID string, transmissionStatus bool) error {
	if _, ok := tssq.tasks[taskID]; !ok {
		return fmt.Errorf("task with ID %q not found", taskID)
	}

	tssq.tasks[taskID].TransmissionStatus = transmissionStatus

	return nil
}

//changingStatusInformationRelevance изменение статуса актуальности всех задач у которых updateTimeInformation != 0
func (tssq *TemporaryStorageSearchQueries) changingStatusInformationRelevance(ltid []string) {
	//все задачи по которым ранее был поиск становятся не актуальными и подлежат удалению
	for id, info := range tssq.tasks {
		if info.UpdateTimeInformation == 0 {
			continue
		}

		for _, tid := range ltid {
			if id == tid {
				tssq.tasks[id].NotRelevance = true
			}
		}
	}
}

//delInformationAboutSearchTask удалить информацию о задаче
func (tssq *TemporaryStorageSearchQueries) delInformationAboutSearchTask(taskID string) {
	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "del information about search task",
		searchTaskID: taskID,
		channelRes:   chanRes,
	}

	<-chanRes
}

//DeleteOldestRecord удалить самую старую запись
/*func (tssq *TemporaryStorageSearchQueries) deleteOldestRecord() {
	fmt.Println("func 'deleteOldestRecord', START...")

	ls := make(listTask, 0, len(tssq.tasks))

	fmt.Println(ls)

	for key, value := range tssq.tasks {
		ls = append(ls, listTaskInfo{
			id:           key,
			time:         value.UpdateTimeInformation,
			transmission: value.TransmissionStatus,
		})
	}

	sort.Sort(ls)

	fmt.Println(ls)

	//удаляем самую старую запись (если информация по задаче не передается)
	for _, info := range ls {
		if info.transmission {
			continue
		}

		fmt.Println(info)

		//tssq.delInformationAboutSearchTask(info.id)
	}
}*/

//checkTimeDeleteTemporaryStorageSearchQueries очистка информации о задаче по истечении определенного времени или неактуальности данных
func checkTimeDeleteTemporaryStorageSearchQueries(tssq *TemporaryStorageSearchQueries) {
	ticker := time.NewTicker(time.Duration(tssq.tickerSec) * time.Second)

	for range ticker.C {
		if len(tssq.tasks) == 0 {
			continue
		}

		timeNow := time.Now().Unix()

		for id, t := range tssq.tasks {
			//если задача не актуальна и информация о задаче не передается
			if t.NotRelevance && !t.TransmissionStatus {
				tssq.delInformationAboutSearchTask(id)
			}

			//если параметр t.UpdateTimeInformation равен 0, значит поиск в БД по данной задаче ещё не выполнялся
			if t.UpdateTimeInformation == 0 {
				continue
			}

			//если задача устарела по времени и информация о ней не передается
			if ((t.UpdateTimeInformation + int64(tssq.timeExpiration)) < timeNow) && !t.TransmissionStatus {
				tssq.delInformationAboutSearchTask(id)
			}
		}
	}
}
