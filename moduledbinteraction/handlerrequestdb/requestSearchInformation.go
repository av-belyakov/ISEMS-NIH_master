package handlerrequestdb

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	//	"github.com/mongodb/mongo-go-driver/bson"
)

//SearchShortInformationAboutTasks поиск ОБЩЕЙ информации по задачам
func SearchShortInformationAboutTasks(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	tssq *configure.TemporaryStorageSearchQueries,
	qp QueryParameters) {

	fmt.Println("func 'SearchShortInformationAboutTasks', START...")

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    req.MsgRecipient,
		MsgRecipient:    req.MsgGenerator,
		MsgSection:      "information search control",
		Instruction:     "short search result",
		IDClientAPI:     req.IDClientAPI,
		TaskID:          req.TaskID,
		TaskIDClientAPI: req.TaskIDClientAPI,
	}

	//получаем информацию о задаче
	info, err := tssq.GetInformationAboutSearchTask(req.TaskID)
	if err != nil {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "the data required to search for information about the task was not found by the passed ID",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	fmt.Printf("forming query for DB to search parameters: %v\n", info)

	listShortTaskInfo, err := getShortInformation(qp, &info.SearchParameters)
	if err != nil {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "search for information in the database is not possible, error processing the request to the database",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	fmt.Printf("func 'SearchShortInformationAboutTasks', SEARCH INFO: '%v'\n", listShortTaskInfo)

	//добавляем найденную информацию в TemporaryStorageSearchQueries
	if err := tssq.AddInformationFoundSearchResult(req.TaskID, listShortTaskInfo); err != nil {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "you cannot add information to the object 'TemporaryStorageSearchQueries' is not found corresponding to ID",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	fmt.Printf("func 'SearchShortInformationAboutTasks', msgRes: '%v'\n", msgRes)

	chanIn <- &msgRes
}

//SearchFullInformationAboutTasks поиск ПОЛНОЙ информации по задачам
func SearchFullInformationAboutTasks() {
	fmt.Println("func 'SearchFullInformationAboutTasks', START...")
}
