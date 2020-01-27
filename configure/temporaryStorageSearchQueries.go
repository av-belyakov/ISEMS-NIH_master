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
// relevanceStatus - статус актуальности инофрмации
// transmissionStatus - статус передачи информации (передается ли информация в настоящий момент)
// searchParameters - описание параметров поискового запроса
// listFindInformation - список найденной информации
type SearchTaskDescription struct {
	addedDataTask       int64
	relevanceStatus     bool
	transmissionStatus  bool
	searchParameters    SearchParameters
	listFindInformation ListFindInformation
}

//SearchParameters параметры поиска
type SearchParameters SearchInformationAboutTasksRequestOption

//ListFindInformation список найденной информации
type ListFindInformation []*BriefTaskInformation

//SearchChannelRequest параметры канала запроса
// actionType - тип действия
// searchTaskID - уникальный идентификатор задачи
// searchParameters - параметры поиска
// listFindInformation - список найденной информации
// channelRes - описание канала ответа
type SearchChannelRequest struct {
	actionType          string
	searchTaskID        string
	searchParameters    SearchParameters
	listFindInformation ListFindInformation
	channelRes          chan SearchChannelResponse
}

//SearchChannelResponse параметры канала ответа
// listFindInformation - список найденной информации
type SearchChannelResponse struct {
	listFindInformation ListFindInformation
}

//CreateTmpStorageID генерирует идентификатор задачи поиска в БД по заданным параметрам
func CreateTmpStorageID(clientID string, sp *SearchParameters) string {
	/*
		type SearchInformationAboutTasksRequestOption struct {
			TaskProcessed             bool                             `json:"tp"`
			ID                        int                              `json:"id"`
			NumberTasksReturnedPart   int                              `json:"ntrp"`
			FilesDownloaded           FilesDownloadedOptions           `json:"fd"`
			InformationAboutFiltering InformationAboutFilteringOptions `json:"iaf"`
			InstalledFilteringOption  SearchFilteringOptions           `json:"ifo"`
		}

		!!! ПРОТЕСТИРОВАТЬ ЭТУ ФУНКЦИЮ !!!
	*/

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
		strconv.Itoa(sp.NumberTasksReturnedPart),
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

	stringParameters := strings.Join(s, "_")

	fmt.Println(stringParameters)

	h := sha1.New()
	io.WriteString(h, stringParameters)

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

			case "get information about search task":

			case "add information found search result":

			case "del information about search task":

			case "change status transmission task":

			case "changing status information relevance":

			}
		}
	}()

	return &tssq
}

//CreateNewSearchTask создание новой временной записи о пиосковой задаче
func (tssq *TemporaryStorageSearchQueries) CreateNewSearchTask(clientID string, sp *SearchParameters) (string, error) {
	fmt.Println("func 'CreateNewTemporaryStorage', START...")

	taskID := ""

	return taskID, nil
}

//GetInformationAboutSearchTask вывод информации из кэша
func (tssq *TemporaryStorageSearchQueries) GetInformationAboutSearchTask(taskID string) error {
	fmt.Println("func 'GetInformationAboutSearchTask', START...")

	return nil
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
