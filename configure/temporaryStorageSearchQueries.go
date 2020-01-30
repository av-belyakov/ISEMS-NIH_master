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
// channelReq - канал для передачи информации о запросе внутри хранилища
type TemporaryStorageSearchQueries struct {
	tasks      map[string]*SearchTaskDescription
	channelReq chan SearchChannelRequest
}

//SearchTaskDescription описание задачи по поиску информации в БД
// addedDataTask - дата добавления задачи
// relevanceStatus - статус актуальности информации
// transmitionToClientStatus - передается ли найденная информация клиенту API (актуально когда
// найденной информации много и она передается клиенту API частями)
// transmissionStatus - статус передачи информации (передается ли информация в настоящий момент)
// searchParameters - описание параметров поискового запроса
// listFindInformation - список найденной информации
type SearchTaskDescription struct {
	addedDataTask             int64
	relevanceStatus           bool
	transmissionStatus        bool
	transmitionToClientStatus bool
	searchParameters          SearchParameters
	listFindInformation       ListFindInformation
}

//SearchParameters параметры поиска
type SearchParameters SearchInformationAboutTasksRequestOption

//ListFindInformation список найденной информации
type ListFindInformation []*BriefTaskInformation

//SearchChannelRequest параметры канала запроса
// actionType - тип действия
// clientID - идентификатор клиента
// searchTaskID - уникальный идентификатор задачи
// searchParameters - параметры поиска
// listFindInformation - список найденной информации
// channelRes - описание канала ответа
type SearchChannelRequest struct {
	actionType          string
	searchTaskID        string
	searchParameters    *SearchParameters
	listFindInformation ListFindInformation
	channelRes          chan SearchChannelResponse
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

//NewRepositoryTSSQ создание нового репозитория для хранения информации о задачах поиска
func NewRepositoryTSSQ() *TemporaryStorageSearchQueries {
	tssq := TemporaryStorageSearchQueries{}
	tssq.tasks = map[string]*SearchTaskDescription{}
	tssq.channelReq = make(chan SearchChannelRequest)

	go func() {
		for msg := range tssq.channelReq {
			switch msg.actionType {
			case "create new search task":
				info, err := tssq.temporaryStorageSearchInfo(msg.searchTaskID)

				if err != nil {
					//добавляем информацию
					tssq.tasks[msg.searchTaskID] = &SearchTaskDescription{
						searchParameters: *msg.searchParameters,
					}

					msg.channelRes <- SearchChannelResponse{}
				} else {
					msg.channelRes <- SearchChannelResponse{
						errMsg:          err,
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

			case "del information about search task":

			case "change status transmission task":

			case "changing status information relevance":

			}
		}
	}()

	return &tssq
}

//CreateNewSearchTask создание новой временной записи о поисковой задаче
func (tssq *TemporaryStorageSearchQueries) CreateNewSearchTask(clientID string, sp *SearchParameters) (string, *SearchTaskDescription) {
	//fmt.Println("func 'CreateNewTemporaryStorage', START...")

	taskID := CreateTmpStorageID(clientID, sp)

	chanRes := make(chan SearchChannelResponse)

	tssq.channelReq <- SearchChannelRequest{
		actionType:       "create new search task",
		searchTaskID:     taskID,
		searchParameters: sp,
		channelRes:       chanRes,
	}

	info := <-chanRes

	return taskID, info.findInformation
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
func (tssq *TemporaryStorageSearchQueries) AddInformationFoundSearchResult() error {
	fmt.Println("func 'AddInformationFoundSearchResult', START...")

	return nil
}

//ChangeStatusTransmissionTask изменение статуса передачи задачи (в данный момент информация по данной задаче передается или нет)
func (tssq *TemporaryStorageSearchQueries) ChangeStatusTransmissionTask(taskID string, transmissionStatus bool) error {
	fmt.Println("func 'ChangeStatusTransmissionTask', START...")

	return nil
}

//ChangingStatusInformationRelevance изменение статуса актуальности задачи
func (tssq *TemporaryStorageSearchQueries) ChangingStatusInformationRelevance() {
	fmt.Println("func 'ChangingStatusInformationRelevance', START...")
}

//ChangingStatusTransmitionToClientStatus изменение статуса задачи при передачи найденной информации клиенту API
func (tssq *TemporaryStorageSearchQueries) ChangingStatusTransmitionToClientStatus(taskID string, transmitionStatus bool) error {
	fmt.Println("func 'ChangingStatusTransmitionToClientStatus', START...")

	//ошибка если информация не найдена
	return nil
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

func (tssq *TemporaryStorageSearchQueries) delInformationAboutSearchTask(taskID string) {
	fmt.Println("func 'delInformationAboutSearchTask', START...")
}

//CheckTimeDeleteTemporaryStorageSearchQueries очистка информации о задаче по истечении определенного времени или неактуальности данных
func (tssq *TemporaryStorageSearchQueries) CheckTimeDeleteTemporaryStorageSearchQueries(sec int) chan string {

	fmt.Println("func 'CheckTimeDeleteTemporaryStorageSearchQueries', START...")

	ticker := time.NewTicker(time.Duration(sec) * time.Second)

	testChan := make(chan string)

	go func() {
		for range ticker.C {
			if len(tssq.tasks) == 0 {
				continue
			}

			timeNow := time.Now().Unix()

			for id, t := range tssq.tasks {
				//если задача не актуальна и информация о задаче не передается
				if t.relevanceStatus && !t.transmissionStatus {

					fmt.Println("func 'CheckTimeDeleteTemporaryStorageSearchQueries', удалить так как задача не актуальна")

					tssq.delInformationAboutSearchTask(id)

					testChan <- "func 'CheckTimeDeleteTemporaryStorageSearchQueries', DELETE. удалить так как задача не актуальна"
				}

				//если задача устарела по времени и информация о ней не передается
				if ((t.addedDataTask + 5) < timeNow) && !t.transmissionStatus {

					fmt.Println("func 'CheckTimeDeleteTemporaryStorageSearchQueries', удалить так как задача просрочена")

					tssq.delInformationAboutSearchTask(id)

					testChan <- "func 'CheckTimeDeleteTemporaryStorageSearchQueries', DELETE. удалить так как задача просрочена"
				}
			}
		}
	}()

	return testChan
}
