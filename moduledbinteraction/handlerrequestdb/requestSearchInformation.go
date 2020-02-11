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
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgSection:   "information search control",
		IDClientAPI:  req.IDClientAPI,
		TaskID:       req.TaskID,
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

	/*
		!!! Строим поисковый запрос к БД и выполняем поиск
		эту функцию надо написать и потестировать !!!
	*/

	fmt.Printf("forming query for DB to search parameters: %v\n", info)
}

//SearchFullInformationAboutTasks поиск ПОЛНОЙ информации по задачам
func SearchFullInformationAboutTasks() {
	fmt.Println("func 'SearchFullInformationAboutTasks', START...")
}
