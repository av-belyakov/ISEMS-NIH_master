package handlerrequestdb

import (
	"ISEMS-NIH_master/configure"
	"fmt"
)

//FindingInformationAboutTask поиск в БД информации по ID задачи
func FindingInformationAboutTask(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    "DB module",
		MsgRecipient:    "Core module",
		MsgSection:      "download control",
		Instruction:     "all information about task",
		IDClientAPI:     req.IDClientAPI,
		TaskID:          req.TaskID,
		TaskIDClientAPI: req.TaskIDClientAPI,
	}

	//восстанавливаем задачу по ее ID
	taskInfo, err := getInfoTaskForID(qp, req.TaskID)
	if err != nil {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "error reading information on the task in the database",
			ErrorBody:             err,
		}

		return
	}

	msgRes.AdvancedOptions = taskInfo

	chanIn <- &msgRes
}

//UpdateFinishedInformationAboutTask запись информации по задаче (задача завершена)
func UpdateFinishedInformationAboutTask(
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters,
	smt *configure.StoringMemoryTask) error {

	ti, ok := smt.GetStoringMemoryTask(req.TaskID)
	if !ok {
		return fmt.Errorf("task with ID '%v' not found (DB module)", req.TaskID)
	}

	//записать инфорацию в БД

	//пометить задачу в StoringMemoryTask как выполненную
	smt.CompleteStoringMemoryTask(req.TaskID)

	return nil
}
