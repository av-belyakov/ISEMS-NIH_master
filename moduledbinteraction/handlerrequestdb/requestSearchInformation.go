package handlerrequestdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ISEMS-NIH_master/configure"

	"github.com/mongodb/mongo-go-driver/bson"
)

//SearchShortInformationAboutTasks поиск ОБЩЕЙ информации по задачам
func SearchShortInformationAboutTasks(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	tssq *configure.TemporaryStorageSearchQueries,
	qp QueryParameters) {

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

	/*
		fmt.Println("func 'SearchShortInformationAboutTasks', START...")
		fmt.Println("take information from module cashe")
		fmt.Printf("task ID: %v\n", req.TaskID)
		fmt.Println(info)
	*/

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

	/*
		fmt.Println("func 'requestSearchInformation'")
		fmt.Println(*(listShortTaskInfo[0]))
	*/

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

	chanIn <- &msgRes
}

//SearchFullInformationAboutTasks поиск ПОЛНОЙ информации по задачам
func SearchFullInformationAboutTasks(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	const maxCountFiles = 50

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    req.MsgRecipient,
		MsgRecipient:    req.MsgGenerator,
		MsgSection:      "information search control",
		Instruction:     "information by task ID",
		IDClientAPI:     req.IDClientAPI,
		TaskID:          req.TaskID,
		TaskIDClientAPI: req.TaskIDClientAPI,
	}

	cur, err := qp.Find(bson.D{bson.E{Key: "task_id", Value: req.TaskID}})
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

	liat := []*configure.InformationAboutTask{}
	for cur.Next(context.Background()) {
		var model configure.InformationAboutTask
		if err := cur.Decode(&model); err != nil {
			msgRes.MsgSection = "error notification"
			msgRes.AdvancedOptions = configure.ErrorNotification{
				SourceReport:          "DB module",
				HumanDescriptionError: "search for information in the database is not possible, internal error when processing the DB response",
				ErrorBody:             err,
			}

			chanIn <- &msgRes

			return
		}

		liat = append(liat, &model)
	}

	if err := cur.Err(); err != nil {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "search for information in the database is not possible, internal error when processing the DB response",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	cur.Close(context.Background())

	if len(liat) == 0 {
		msgRes.AdvancedOptions = configure.ResponseInformationByTaskID{Status: "task not found"}

		chanIn <- &msgRes

		return
	}

	numFiles := len(liat[0].ListFilesResultTaskExecution)

	maxListSize := numFiles
	var filesList []configure.FileInformation
	if numFiles > 0 {
		if numFiles > maxCountFiles {
			maxListSize = maxCountFiles
		}

		filesList = make([]configure.FileInformation, 0, maxListSize)

		for i := 0; i < maxListSize; i++ {
			filesList = append(filesList, configure.FileInformation{
				Name:     liat[0].ListFilesResultTaskExecution[i].FileName,
				Size:     liat[0].ListFilesResultTaskExecution[i].FileSize,
				IsLoaded: liat[0].ListFilesResultTaskExecution[i].FileLoaded,
			})
		}
	}

	rtp := configure.ResponseTaskParameter{
		TaskID:                           liat[0].TaskID,
		ClientTaskID:                     liat[0].ClientTaskID,
		SourceID:                         liat[0].SourceID,
		UserInitiatedFilteringProcess:    liat[0].UserInitiatedFilteringProcess,
		UserInitiatedFileDownloadProcess: liat[0].UserInitiatedFileDownloadProcess,
		GeneralInformationAboutTask: configure.GeneralInformationAboutTask{
			TaskProcessed:     liat[0].GeneralInformationAboutTask.TaskProcessed,
			DateTimeProcessed: liat[0].GeneralInformationAboutTask.DateTimeProcessed,
			ClientIDIP:        liat[0].GeneralInformationAboutTask.ClientID,
			DetailDescription: configure.DetailDescription{
				UserNameClosedProcess:        liat[0].GeneralInformationAboutTask.DetailDescription.UserNameProcessed,
				DescriptionProcessingResults: liat[0].GeneralInformationAboutTask.DetailDescription.DescriptionProcessingResults,
			},
		},
		FilteringOption: configure.TaskFilteringOption{
			DateTime: configure.DateTimeParameters{
				Start: liat[0].FilteringOption.DateTime.Start,
				End:   liat[0].FilteringOption.DateTime.End,
			},
			Protocol: liat[0].FilteringOption.Protocol,
			Filters: configure.FiltrationControlParametersNetworkFilters{
				IP: configure.FiltrationControlIPorNetorPortParameters{
					Any: liat[0].FilteringOption.Filters.IP.Any,
					Src: liat[0].FilteringOption.Filters.IP.Src,
					Dst: liat[0].FilteringOption.Filters.IP.Dst,
				},
				Port: configure.FiltrationControlIPorNetorPortParameters{
					Any: liat[0].FilteringOption.Filters.Port.Any,
					Src: liat[0].FilteringOption.Filters.Port.Src,
					Dst: liat[0].FilteringOption.Filters.Port.Dst,
				},
				Network: configure.FiltrationControlIPorNetorPortParameters{
					Any: liat[0].FilteringOption.Filters.Network.Any,
					Src: liat[0].FilteringOption.Filters.Network.Src,
					Dst: liat[0].FilteringOption.Filters.Network.Dst,
				},
			},
		},
		DetailedInformationOnFiltering: configure.InformationOnFiltering{
			TaskStatus: liat[0].DetailedInformationOnFiltering.TaskStatus,
			TimeIntervalTaskExecution: configure.DateTimeParameters{
				Start: liat[0].DetailedInformationOnFiltering.TimeIntervalTaskExecution.Start,
				End:   liat[0].DetailedInformationOnFiltering.TimeIntervalTaskExecution.End,
			},
			WasIndexUsed:                    liat[0].DetailedInformationOnFiltering.WasIndexUsed,
			NumberProcessedFiles:            liat[0].DetailedInformationOnFiltering.NumberProcessedFiles,
			NumberDirectoryFiltartion:       liat[0].DetailedInformationOnFiltering.NumberDirectoryFiltartion,
			NumberErrorProcessedFiles:       liat[0].DetailedInformationOnFiltering.NumberErrorProcessedFiles,
			NumberFilesMeetFilterParameters: liat[0].DetailedInformationOnFiltering.NumberFilesMeetFilterParameters,
			NumberFilesFoundResultFiltering: liat[0].DetailedInformationOnFiltering.NumberFilesFoundResultFiltering,
			SizeFilesMeetFilterParameters:   liat[0].DetailedInformationOnFiltering.SizeFilesMeetFilterParameters,
			SizeFilesFoundResultFiltering:   liat[0].DetailedInformationOnFiltering.SizeFilesFoundResultFiltering,
			PathDirectoryForFilteredFiles:   liat[0].DetailedInformationOnFiltering.PathDirectoryForFilteredFiles,
		},
		DetailedInformationOnDownloading: configure.InformationOnDownloading{
			TaskStatus: liat[0].DetailedInformationOnDownloading.TaskStatus,
			TimeIntervalTaskExecution: configure.DateTimeParameters{
				Start: liat[0].DetailedInformationOnDownloading.TimeIntervalTaskExecution.Start,
				End:   liat[0].DetailedInformationOnDownloading.TimeIntervalTaskExecution.End,
			},
			NumberFilesTotal:                    liat[0].DetailedInformationOnDownloading.NumberFilesTotal,
			NumberFilesDownloaded:               liat[0].DetailedInformationOnDownloading.NumberFilesDownloaded,
			NumberFilesDownloadedError:          liat[0].DetailedInformationOnDownloading.NumberFilesDownloadedError,
			PathDirectoryStorageDownloadedFiles: liat[0].DetailedInformationOnDownloading.PathDirectoryStorageDownloadedFiles,
		},
		DetailedInformationListFiles: filesList,
	}

	msgRes.AdvancedOptions = configure.ResponseInformationByTaskID{
		Status:        "complete",
		TaskParameter: rtp,
	}

	chanIn <- &msgRes
}

//GetListFoundFiles получить список найденных в результате фильтрации файлов
func GetListFoundFiles(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    req.MsgRecipient,
		MsgRecipient:    req.MsgGenerator,
		MsgSection:      "information search control",
		Instruction:     "list files by task ID",
		IDClientAPI:     req.IDClientAPI,
		TaskID:          req.TaskID,
		TaskIDClientAPI: req.TaskIDClientAPI,
	}

	//приняты некорректные параметры запроса
	errMsg := "invalid request parameters were accepted"

	lffro, ok := req.AdvancedOptions.(configure.GetListFoundFilesRequestOption)
	if !ok {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             errors.New(errMsg),
		}

		chanIn <- &msgRes

		return
	}

	lfi := make([]*configure.FilesInformation, 0, lffro.PartSize)
	cur, err := qp.Find(bson.D{bson.E{Key: "task_id", Value: lffro.RequestTaskID}})
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

	liat := []*configure.InformationAboutTask{}
	for cur.Next(context.Background()) {
		var model configure.InformationAboutTask
		if err := cur.Decode(&model); err != nil {
			msgRes.MsgSection = "error notification"
			msgRes.AdvancedOptions = configure.ErrorNotification{
				SourceReport:          "DB module",
				HumanDescriptionError: "search for information in the database is not possible, internal error when processing the DB response",
				ErrorBody:             err,
			}

			chanIn <- &msgRes

			return
		}

		liat = append(liat, &model)
	}

	if err := cur.Err(); err != nil {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "search for information in the database is not possible, internal error when processing the DB response",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	cur.Close(context.Background())

	// информация по задаче не найдена
	if len(liat) == 0 {
		msgRes.AdvancedOptions = configure.ListFoundFilesResponseOption{Status: "task not found"}

		chanIn <- &msgRes

		return
	}

	partSize := lffro.PartSize
	if partSize > 250 {
		partSize = 250
	}

	commonPartSize := (partSize + lffro.OffsetListParts)
	numFoundFiles := len(liat[0].ListFilesResultTaskExecution)
	if numFoundFiles < (lffro.OffsetListParts + 1) {
		// общее количество найденных файлов, меньше чем количество файлов, на которое нужно выполнить смещение
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "the total number of files found is less than the number of files to offset",
			ErrorBody:             fmt.Errorf("the total number of files found for the issue with ID %q is less than the number of files to offset", lffro.RequestTaskID),
		}

		chanIn <- &msgRes

		return
	}

	if numFoundFiles <= commonPartSize {
		lfi = append(lfi, liat[0].ListFilesResultTaskExecution[lffro.OffsetListParts:]...)
	} else {
		lfi = append(lfi, liat[0].ListFilesResultTaskExecution[lffro.OffsetListParts:commonPartSize]...)
	}

	lf := configure.ListFoundFilesResponseOption{
		Status:          "complete",
		TaskID:          lffro.RequestTaskID,
		ClientTaskID:    liat[0].ClientTaskID,
		SourceID:        liat[0].SourceID,
		FullListSize:    numFoundFiles,
		RequestPartSize: partSize,
		OffsetListParts: lffro.OffsetListParts,
		ListFiles:       lfi,
	}
	msgRes.AdvancedOptions = lf

	chanIn <- &msgRes
}

//GetInfoTaskFromMarkTaskCompleteProcess обрабатываем запрос на получения информации при
// выполнении команды 'отметить задачу как обработанную'
func GetInfoTaskFromMarkTaskCompleteProcess(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	fmt.Println("func 'GetInfoTaskFromMarkTaskCompleteProcess', START...")

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    req.MsgRecipient,
		MsgRecipient:    req.MsgGenerator,
		MsgSection:      "information search control",
		Instruction:     "mark an task as completed processed",
		IDClientAPI:     req.IDClientAPI,
		TaskID:          req.TaskID,
		TaskIDClientAPI: req.TaskIDClientAPI,
	}

	fmt.Println("func 'GetInfoTaskFromMarkTaskCompleteProcess', get information about task ID")

	errMsg := fmt.Sprintf("It is not possible to mark a task with ID %q as successfully completed. Internal application error.", req.TaskIDClientAPI)

	listTaskInfo, err := getInfoTaskForID(qp, req.TaskID)
	if err != nil {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	//проверяем найдена ли задача
	if len(*listTaskInfo) == 0 {
		msgRes.AdvancedOptions = configure.TypeGetInfoTaskFromMarkTaskCompleteProcess{}

		chanIn <- &msgRes

		return
	}

	ti := (*listTaskInfo)[0]

	fmt.Println(ti)

	mtcro, ok := req.AdvancedOptions.(configure.MarkTaskCompletedRequestOption)
	if !ok {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             errors.New(errMsg),
		}

		chanIn <- &msgRes

		return
	}

	isDownloaded := ti.DetailedInformationOnDownloading.NumberFilesDownloaded > 0
	taskStatus := ti.DetailedInformationOnFiltering.TaskStatus == "complete"

	msgRes.AdvancedOptions = configure.TypeGetInfoTaskFromMarkTaskCompleteProcess{
		TaskIsExist:          true,
		UserName:             mtcro.UserName,
		Description:          mtcro.Description,
		FiltrationTaskStatus: taskStatus,
		FilesDownloaded:      isDownloaded,
	}

	fmt.Println("func 'GetInfoTaskFromMarkTaskCompleteProcess', send message to Core")
	fmt.Println(msgRes)

	chanIn <- &msgRes

}

//MarkTaskCompleteProcess отметить задачу как завершенную
func MarkTaskCompleteProcess(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	fmt.Println("func 'MarkTaskCompleteProcess', START...")

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    req.MsgRecipient,
		MsgRecipient:    req.MsgGenerator,
		MsgSection:      "information search control",
		Instruction:     "mark an task as completed",
		IDClientAPI:     req.IDClientAPI,
		TaskID:          req.TaskID,
		TaskIDClientAPI: req.TaskIDClientAPI,
	}

	errMsg := fmt.Sprintf("It is not possible to mark a task with ID %q as successfully completed. Internal application error.", req.TaskIDClientAPI)
	tgitfmtcp, ok := req.AdvancedOptions.(configure.TypeGetInfoTaskFromMarkTaskCompleteProcess)
	if !ok {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             errors.New(errMsg),
		}

		chanIn <- &msgRes

		return
	}

	fmt.Println("func 'MarkTaskCompleteProcess', отмечаем задачу как завершенную")

	//отмечаем задачу как завершенную
	if err := qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: req.TaskID}},
		bson.D{
			bson.E{Key: "$set", Value: bson.D{
				bson.E{Key: "general_information_about_task.task_processed", Value: true},
				bson.E{Key: "general_information_about_task.date_time_processed", Value: time.Now().Unix()},
				bson.E{Key: "general_information_about_task.client_id", Value: req.IDClientAPI},
				bson.E{Key: "general_information_about_task.detail_description_general_information_about_task.user_name_processed", Value: tgitfmtcp.UserName},
				bson.E{Key: "general_information_about_task.detail_description_general_information_about_task.description_processing_results", Value: tgitfmtcp.Description},
			}},
		}); err != nil {

		fmt.Println("func 'MarkTaskCompleteProcess', ERROR")
		fmt.Println(err)

		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "search for information in the database is not possible, error processing the request to the database",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	fmt.Println("func 'MarkTaskCompleteProcess', запись в БД завершена")

	chanIn <- &msgRes
}

//DeleteInformationAboutTask удалить всю информацию о выбранных задачах
func DeleteInformationAboutTask(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	fmt.Println("func 'DeleteInformationAboutTask', START...")

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    req.MsgRecipient,
		MsgRecipient:    req.MsgGenerator,
		MsgSection:      "information search control",
		Instruction:     "delete all information about a task",
		IDClientAPI:     req.IDClientAPI,
		TaskIDClientAPI: req.TaskIDClientAPI,
	}

	l, ok := req.AdvancedOptions.(*configure.DeleteInformationListTaskCompletedRequestOption)
	if !ok {

		fmt.Println("an incorrect list of issues that should be deleted was accepted")

		errMsg := "an incorrect list of issues that should be deleted was accepted"

		msgRes.MsgRecipient = "Core module"
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             errors.New(errMsg),
		}

		chanIn <- &msgRes

		return
	}

	fmt.Printf("func 'DeleteInformationAboutTask', ------- delete information about task -------\n %v \n", (*l).ListTaskID)

	for _, id := range (*l).ListTaskID {
		_ = qp.DeleteOneData(bson.D{bson.E{Key: "task_id", Value: id}})
	}

	chanIn <- &msgRes
}

//CommonAnalyticsInformationAboutTaskID получить общую информацию о задаче и загруженном, и разобранном сетевом трафике
func CommonAnalyticsInformationAboutTaskID(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	fmt.Println("func 'CommonAnalyticsInformationAboutTaskID', START...")

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    req.MsgRecipient,
		MsgRecipient:    req.MsgGenerator,
		MsgSection:      "information search control",
		Instruction:     "get common analytics information about task ID",
		IDClientAPI:     req.IDClientAPI,
		TaskIDClientAPI: req.TaskIDClientAPI,
	}

	errMsg := fmt.Sprintf("It is not possible to mark a task with ID %q as successfully completed. Internal application error.", req.TaskIDClientAPI)

	/*
		Ищем информацию о задаче в коллекции 'task_list' где хранится
		ОБЩАЯ информация по задаче фильтрации и скачивания и список найденных
		и загруженных файлов. Это се наполняет структуры:
			- GeneralInformationAboutTask
			- InstalledFilteringOption
			- CommonInformationAboutReceivedFiles
	*/
	listTaskInfo, err := getInfoTaskForID(qp, req.TaskID)
	if err != nil {
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	//проверяем найдена ли задача
	if len(*listTaskInfo) == 0 {
		msgRes.AdvancedOptions = configure.TypeGetInfoTaskFromMarkTaskCompleteProcess{}

		chanIn <- &msgRes

		return
	}

	ti := (*listTaskInfo)[0]

	pcai := configure.ParametersCommonAnalyticsInformation{
		TaskID:   ti.TaskID,
		SourceID: ti.SourceID,
		GeneralInformationAboutTask: configure.GeneralInformationAboutTask{
			TaskProcessed:     ti.GeneralInformationAboutTask.TaskProcessed,
			DateTimeProcessed: ti.GeneralInformationAboutTask.DateTimeProcessed,
			ClientIDIP:        ti.GeneralInformationAboutTask.ClientID,
			DetailDescription: configure.DetailDescription{
				UserNameClosedProcess:        ti.GeneralInformationAboutTask.DetailDescription.UserNameProcessed,
				DescriptionProcessingResults: ti.GeneralInformationAboutTask.DetailDescription.DescriptionProcessingResults,
			},
		},
		InstalledFilteringOption: configure.SearchFilteringOptions{
			DateTime: configure.DateTimeParameters{
				Start: ti.FilteringOption.DateTime.Start,
				End:   ti.FilteringOption.DateTime.End,
			},
			Protocol: ti.FilteringOption.Protocol,
			NetworkFilters: configure.FiltrationControlParametersNetworkFilters{
				IP: configure.FiltrationControlIPorNetorPortParameters{
					Any: ti.FilteringOption.Filters.IP.Any,
					Src: ti.FilteringOption.Filters.IP.Src,
					Dst: ti.FilteringOption.Filters.IP.Dst,
				},
				Port: configure.FiltrationControlIPorNetorPortParameters{
					Any: ti.FilteringOption.Filters.Port.Any,
					Src: ti.FilteringOption.Filters.Port.Src,
					Dst: ti.FilteringOption.Filters.Port.Dst,
				},
				Network: configure.FiltrationControlIPorNetorPortParameters{
					Any: ti.FilteringOption.Filters.Network.Any,
					Src: ti.FilteringOption.Filters.Network.Src,
					Dst: ti.FilteringOption.Filters.Network.Dst,
				},
			},
		},
		CommonInformationAboutReceivedFiles: configure.CommonInformationAboutReceivedFilesDescription{
			NumberFilesTotal:                    ti.DetailedInformationOnDownloading.NumberFilesTotal,
			DownloadTaskStatus:                  ti.DetailedInformationOnDownloading.TaskStatus,
			FilteringTaskStatus:                 ti.DetailedInformationOnFiltering.TaskStatus,
			NumberFilesDownloaded:               ti.DetailedInformationOnDownloading.NumberFilesDownloaded,
			SizeFilesFoundResultFiltering:       ti.DetailedInformationOnFiltering.SizeFilesFoundResultFiltering,
			PathDirectoryStorageDownloadedFiles: ti.DetailedInformationOnDownloading.PathDirectoryStorageDownloadedFiles,
		},
		DetailedInformationAboutReceivedFiles: configure.DetailedInformationAboutReceivedFilesDescription{},
	}

	/*
	   Наполнение информацией структуры DetailedInformationAboutReceivedFiles поке не реализованно
	   (нет описания самой структуры)
	*/

	fmt.Printf("func 'CommonAnalyticsInformationAboutTaskID', FOUND INFORMATION: \n ********** %v ********** \n", pcai)

	msgRes.AdvancedOptions = configure.CommonAnalyticsInformationAboutTaskResponsOption{
		Status:        "complete",
		TaskParameter: pcai,
	}

	chanIn <- &msgRes
}
