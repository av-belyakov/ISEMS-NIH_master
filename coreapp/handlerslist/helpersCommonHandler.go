package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//ErrorMessageType параметры для отработки сообщений об ошибках
// SourceID - ID источника
// TaskID - ID задачи
// TaskIDClientAPI - ID задачи клиента API
// IDClientAPI - ID клиента API
// Section - секция (filtering, download and etc.)
// Instruction - инструкция
// MsgType - тип сообщения (danger, warning, success, info)
// MsgHuman - описание сообщения
// Sources - список источников
// SearchRequestIsGeneratedAutomatically — был ли запрос на поиск сгенерирован автоматически (TRUE — да, FALSE - нет)
type ErrorMessageType struct {
	SourceID                              int
	TaskID                                string
	TaskIDClientAPI                       string
	IDClientAPI                           string
	Section                               string
	Instruction                           string
	MsgType                               string
	MsgHuman                              string
	Sources                               []int
	SearchRequestIsGeneratedAutomatically bool
	ChanToAPI                             chan<- *configure.MsgBetweenCoreAndAPI
}

//ErrorMessage формирует и отправляет клиенту API два сообщения, информационное сообщение и сообщение с откланенным статусом задачи
func ErrorMessage(emt ErrorMessageType) error {
	//если запрос не был сгенерирован автоматически
	if !emt.SearchRequestIsGeneratedAutomatically {
		//отправляем информационное сообщение
		notifications.SendNotificationToClientAPI(
			emt.ChanToAPI,
			notifications.NotificationSettingsToClientAPI{
				MsgType:        emt.MsgType,
				MsgDescription: emt.MsgHuman,
				Sources:        emt.Sources,
			},
			emt.TaskIDClientAPI,
			emt.IDClientAPI)
	}

	//отправляем сообщение о том что задача была отклонена
	resMsg := configure.DownloadControlTypeInfo{
		MsgOption: configure.DownloadControlMsgTypeInfo{
			ID:        emt.SourceID,
			TaskIDApp: emt.TaskID,
			Status:    "refused",
		},
	}

	resMsg.MsgType = "information"
	resMsg.MsgSection = emt.Section
	resMsg.MsgInstruction = emt.Instruction
	resMsg.ClientTaskID = emt.TaskIDClientAPI

	msgJSON, err := json.Marshal(resMsg)
	if err != nil {
		return err
	}

	emt.ChanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  emt.IDClientAPI,
		MsgJSON:      msgJSON,
	}

	return nil
}

//HandlerAutomaticDownloadFiles обработчик автоматической загрузки файлов
func HandlerAutomaticDownloadFiles(
	taskID string,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	maxTotalSizeDownloadFiles int64,
	outCoreChanAPI chan<- *configure.MsgBetweenCoreAndAPI) error {

	funcName := "HandlerAutomaticDownloadFiles"

	//fmt.Printf("func '%v', START \n", funcName)

	ts, ok := smt.GetTaskStatusStoringMemoryTask(taskID, "filtration")
	if !ok {

		//fmt.Printf("func '%v' task with %v not found\n", funcName, taskID)

		return fmt.Errorf("task with %v not found%v", taskID, funcName)
	}

	if ts.Status != "complete" {

		//fmt.Printf("func '%v' task status not equal 'complete', task status is '%v'\n", funcName, ts.Status)

		return nil
	}

	taskInfo, taskIsExist := smt.GetStoringMemoryTask(taskID)
	if !taskIsExist {
		return fmt.Errorf("task with %v not found%v", taskID, funcName)
	}

	moreThanMax := taskInfo.TaskParameter.FiltrationTask.SizeFilesFoundResultFiltering > maxTotalSizeDownloadFiles
	sizeFilesFoundIsZero := taskInfo.TaskParameter.FiltrationTask.SizeFilesFoundResultFiltering == 0

	if moreThanMax || sizeFilesFoundIsZero {

		//fmt.Printf("func '%v' value 'moreThanMax' == '%v', value 'sizeFilesFoundIsZero' == '%v'\n", funcName, moreThanMax, sizeFilesFoundIsZero)

		//отмечаем выполняемую задачу как завершенную
		smt.CompleteStoringMemoryTask(taskID)

		//отмечаем задачу, в списке очередей, как завершенную в списке очередей и предотвращаем запуск автоматического скачивания файлов
		if err := qts.ChangeTaskStatusQueueTask(taskInfo.TaskParameter.FiltrationTask.ID, taskID, "complete"); err != nil {
			return err
		}

		return nil
	}

	//fmt.Printf("func '%v' begin automatic download files... \n", funcName)

	sourceID := taskInfo.TaskParameter.FiltrationTask.ID

	//получаем параметры фильтрации
	qti, err := qts.GetQueueTaskStorage(sourceID, taskID)
	if err != nil {
		return err
	}

	//получаем список проверенных файлов
	listDetailedFilesInformation, ok := smt.GetListFilesDetailedInformation(taskID)
	if !ok {
		return fmt.Errorf("the list of files intended for uploading was not found, task ID '%v' (%v)", taskID, funcName)
	}

	//добавляем задачу в очередь
	qts.AddQueueTaskStorage(taskID, sourceID, configure.CommonTaskInfo{
		IDClientAPI:     taskInfo.ClientID,
		TaskIDClientAPI: taskInfo.ClientTaskID,
		TaskType:        "download control",
	}, &configure.DescriptionParametersReceivedFromUser{
		FilterationParameters:         qti.TaskParameters.FilterationParameters,
		PathDirectoryForFilteredFiles: taskInfo.TaskParameter.FiltrationTask.PathStorageSource,
	})

	//информационное сообщение о том что задача добавлена в очередь
	notifications.SendNotificationToClientAPI(
		outCoreChanAPI,
		notifications.NotificationSettingsToClientAPI{
			MsgType: "success",
			MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   sourceID,
				TaskType:   "скачивание файлов",
				TaskAction: "задача автоматически добавлена в очередь",
			}),
			Sources: []int{sourceID},
		},
		taskInfo.ClientTaskID,
		taskInfo.ClientID)

	//устанавливаем проверочный статус источника для данной задачи как подключен
	if err := qts.ChangeAvailabilityConnectionOnConnection(sourceID, taskID); err != nil {
		return err
	}

	//добавляем подтвержденный список файлов для скачивания
	if err := qts.AddConfirmedListFiles(sourceID, taskID, listDetailedFilesInformation); err != nil {
		return err
	}

	//изменяем статус наличия файлов для скачивания
	if err := qts.ChangeAvailabilityFilesDownload(sourceID, taskID); err != nil {
		return err
	}

	return nil
}
