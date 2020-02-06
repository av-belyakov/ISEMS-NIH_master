package configure

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
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
// ListFoundInformation - список найденной информации
type SearchTaskDescription struct {
	UpdateTimeInformation int64
	NotRelevance          bool
	TransmissionStatus    bool
	SearchParameters      SearchParameters
	ListFoundInformation  ListFoundInformation
}

//SearchParameters параметры поиска
type SearchParameters SearchInformationAboutTasksRequestOption

//ListFoundInformation список найденной информаци
type ListFoundInformation struct {
	List []*BriefTaskInformation
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
		boolStr(sp.FilesDownloaded.AllFilesIsDownloaded),
		boolStr(sp.FilesDownloaded.FilesIsDownloaded),
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

	if tr.TimeExpiration > 360 {
		tr.TimeExpiration = 360
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
				searchParameters, ok := msg.information.(SearchParameters)
				if !ok {
					msg.channelRes <- SearchChannelResponse{
						errMsg: fmt.Errorf("ActionType: %q, ERROR: type conversion error", msg.actionType),
					}

					close(msg.channelRes)

					break
				}

				//проверяем количество записей в кэше
				if len(tssq.tasks) > tssq.maxCacheSize {
					tssq.deleteOldestRecord()
				}

				if info, err := tssq.temporaryStorageSearchInfo(msg.searchTaskID); err != nil {
					//добавляем информацию
					tssq.tasks[msg.searchTaskID] = &SearchTaskDescription{
						SearchParameters: searchParameters,
					}

					tssq.tasks[msg.searchTaskID] = &SearchTaskDescription{}
				} else {
					//проверяем задачу на актуальность
					if info.NotRelevance {
						tssq.delInformationAboutSearchTask(msg.searchTaskID)

						tssq.tasks[msg.searchTaskID] = &SearchTaskDescription{
							SearchParameters: searchParameters,
						}
					}

					msg.channelRes <- SearchChannelResponse{
						errMsg:          nil,
						findInformation: info,
					}
				}

				close(msg.channelRes)

			case "get information about search task":
				info, err := tssq.temporaryStorageSearchInfo(msg.searchTaskID)

				msg.channelRes <- SearchChannelResponse{
					errMsg:          err,
					findInformation: info,
				}

				close(msg.channelRes)

			case "add information found search result":
				lbti, ok := msg.information.([]*BriefTaskInformation)
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
					tssq.tasks[msg.searchTaskID].ListFoundInformation.List = lbti
					tssq.tasks[msg.searchTaskID].UpdateTimeInformation = time.Now().Unix()

					msg.channelRes <- SearchChannelResponse{}
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
				tssq.changingStatusInformationRelevance()

				close(msg.channelRes)
			}
		}
	}()

	go checkTimeDeleteTemporaryStorageSearchQueries(&tssq)

	return &tssq
}

//CreateNewSearchTask создание новой временной записи о поисковой задаче
func (tssq *TemporaryStorageSearchQueries) CreateNewSearchTask(clientID string, sp *SearchParameters) (string, *SearchTaskDescription, error) {
	//fmt.Println("func 'CreateNewTemporaryStorage', START...")

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
	//fmt.Println("func 'GetInformationAboutSearchTask', START...")

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
	//fmt.Println("func 'AddInformationFoundSearchResult', START...")

	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "add information found search result",
		searchTaskID: taskID,
		information:  lbti,
		channelRes:   chanRes,
	}

	return (<-chanRes).errMsg
}

//ChangeStatusTransmissionTask изменение статуса передачи задачи (в данный момент информация по данной задаче передается или нет)
func (tssq *TemporaryStorageSearchQueries) ChangeStatusTransmissionTask(taskID string, transmissionStatus bool) error {
	//fmt.Println("func 'ChangeStatusTransmissionTask', START...")

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
func (tssq *TemporaryStorageSearchQueries) ChangingStatusInformationRelevance() {
	//fmt.Println("func 'ChangingStatusInformationRelevance', START...")

	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType: "change status information relevance",
		channelRes: chanRes,
	}
}

func (tssq *TemporaryStorageSearchQueries) temporaryStorageSearchInfo(taskID string) (*SearchTaskDescription, error) {
	/*
		fmt.Println("func 'temporaryStorageSearchInfo' ---- START ----")
		fmt.Println(tssq.tasks)
		info, ok := tssq.tasks[taskID]
		fmt.Println(taskID)
		fmt.Println(info)
		fmt.Println("func 'temporaryStorageSearchInfo' ---- END ----")
	*/
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
func (tssq *TemporaryStorageSearchQueries) changingStatusInformationRelevance() {
	//все задачи по которым ранее был поиск становятся не актуальными и подлежат удалению
	for id, info := range tssq.tasks {
		if info.UpdateTimeInformation == 0 {
			continue
		}

		tssq.tasks[id].NotRelevance = true
	}
}

//delInformationAboutSearchTask удалить информацию о задаче
func (tssq *TemporaryStorageSearchQueries) delInformationAboutSearchTask(taskID string) {
	//fmt.Println("func 'delInformationAboutSearchTask', START...")

	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:   "del information about search task",
		searchTaskID: taskID,
		channelRes:   chanRes,
	}

	<-chanRes
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

//DeleteOldestRecord удалить самую старую запись
func (tssq *TemporaryStorageSearchQueries) deleteOldestRecord() {
	//fmt.Println("func 'deleteOldestRecord', START...")

	ls := make(listTask, 0, len(tssq.tasks))

	for key, value := range tssq.tasks {
		ls = append(ls, listTaskInfo{
			id:           key,
			time:         value.UpdateTimeInformation,
			transmission: value.TransmissionStatus,
		})
	}

	sort.Sort(ls)

	//удаляем самую старую запись (если информация по задаче не передается)
	for _, info := range ls {
		if info.transmission {
			continue
		}

		//		fmt.Printf("DELETE the oldest record: %q\n", info.id)

		tssq.delInformationAboutSearchTask(info.id)
	}
}

//checkTimeDeleteTemporaryStorageSearchQueries очистка информации о задаче по истечении определенного времени или неактуальности данных
func checkTimeDeleteTemporaryStorageSearchQueries(tssq *TemporaryStorageSearchQueries) chan string {

	fmt.Println("func 'CheckTimeDeleteTemporaryStorageSearchQueries', START...")

	testChan := make(chan string)

	go func(tssq *TemporaryStorageSearchQueries, testChan chan<- string) {
		ticker := time.NewTicker(time.Duration(tssq.tickerSec) * time.Second)

		for range ticker.C {
			if len(tssq.tasks) == 0 {
				continue
			}

			timeNow := time.Now().Unix()

			for id, t := range tssq.tasks {

				fmt.Printf("func 'CheckTimeDeleteTemporaryStorageSearchQueries', CHECK TASK ID: %q\n", id)

				//если задача не актуальна, информация о задаче не передается
				if t.NotRelevance && !t.TransmissionStatus {

					fmt.Println("func 'CheckTimeDeleteTemporaryStorageSearchQueries', удалить так как задача не актуальна")

					tssq.delInformationAboutSearchTask(id)

					testChan <- "func 'CheckTimeDeleteTemporaryStorageSearchQueries', DELETE. удалить так как задача не актуальна"
				}

				//если время устаревания кэша выставлено в 0, информация о кэше считается безсрочной по времени
				if tssq.timeExpiration == 0 {
					continue
				}

				//если задача устарела по времени и информация о ней не передается
				if ((t.UpdateTimeInformation + int64(tssq.timeExpiration)) < timeNow) && !t.TransmissionStatus {

					fmt.Println("func 'CheckTimeDeleteTemporaryStorageSearchQueries', удалить так как задача просрочена")

					tssq.delInformationAboutSearchTask(id)

					testChan <- "func 'CheckTimeDeleteTemporaryStorageSearchQueries', DELETE. удалить так как задача просрочена"
				}
			}
		}
	}(tssq, testChan)

	return testChan
}
