package handlerrequestdb

import (
	"errors"
	"fmt"
	"time"

	"ISEMS-NIH_master/configure"

	"go.mongodb.org/mongo-driver/bson"
)

//CreateNewFiltrationTask запись информации о новой фильтрации
//обрабатывается при получении запроса на фильтрацию
func CreateNewFiltrationTask(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters,
	qts *configure.QueueTaskStorage) {

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgSection:   "filtration control",
		IDClientAPI:  req.IDClientAPI,
		TaskID:       req.TaskID,
	}

	errMsg := "taken incorrect settings for task filtering"

	sourceID, ok := req.AdvancedOptions.(int)
	if !ok {
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

	tf, err := qts.GetQueueTaskStorage(sourceID, req.TaskID)
	if err != nil {
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

	//поиск индексов
	isFound, index, err := searchIndexFromFiltration("index_filtration", sourceID, &tf, qp)
	if err != nil {
		msgRes.MsgRecipient = "Core module"
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "error when searching of the index information files of network traffic",
			ErrorBody:             err,
		}

		chanIn <- &msgRes
	}

	itf := configure.InformationAboutTask{
		TaskID:                        req.TaskID,
		ClientID:                      req.IDClientAPI,
		ClientTaskID:                  req.TaskIDClientAPI,
		SourceID:                      sourceID,
		UserInitiatedFilteringProcess: tf.UserName,
		FilteringOption:               tf.TaskParameters.FilterationParameters,
		DetailedInformationOnFiltering: configure.DetailedInformationFiltering{
			TaskStatus:   "wait",
			WasIndexUsed: isFound,
			TimeIntervalTaskExecution: configure.TimeInterval{
				Start: time.Now().Unix(),
			},
		},
		DetailedInformationOnDownloading: configure.DetailedInformationDownloading{
			TaskStatus: "not executed",
		},
		ListFilesResultTaskExecution: []*configure.FilesInformation{},
	}

	insertData := make([]interface{}, 0, 1)
	insertData = append(insertData, itf)

	//запись информации по задачи фильтрации в коллекцию 'filter_task_list'
	if _, err := qp.InsertData(insertData); err != nil {
		msgRes.MsgRecipient = "Core module"
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "error writing tasks for filtering in the application database",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	msgRes.MsgRecipient = "NI module"
	msgRes.AdvancedOptions = configure.TypeFiltrationMsgFoundIndex{
		FilteringOption: configure.FiltrationControlCommonParametersFiltration{
			ID: sourceID,
			DateTime: configure.DateTimeParameters{
				Start: tf.TaskParameters.FilterationParameters.DateTime.Start,
				End:   tf.TaskParameters.FilterationParameters.DateTime.End,
			},
			Protocol: tf.TaskParameters.FilterationParameters.Protocol,
			Filters: configure.FiltrationControlParametersNetworkFilters{
				IP: configure.FiltrationControlIPorNetorPortParameters{
					Any: tf.TaskParameters.FilterationParameters.Filters.IP.Any,
					Src: tf.TaskParameters.FilterationParameters.Filters.IP.Src,
					Dst: tf.TaskParameters.FilterationParameters.Filters.IP.Dst,
				},
				Port: configure.FiltrationControlIPorNetorPortParameters{
					Any: tf.TaskParameters.FilterationParameters.Filters.Port.Any,
					Src: tf.TaskParameters.FilterationParameters.Filters.Port.Src,
					Dst: tf.TaskParameters.FilterationParameters.Filters.Port.Dst,
				},
				Network: configure.FiltrationControlIPorNetorPortParameters{
					Any: tf.TaskParameters.FilterationParameters.Filters.Network.Any,
					Src: tf.TaskParameters.FilterationParameters.Filters.Network.Src,
					Dst: tf.TaskParameters.FilterationParameters.Filters.Network.Dst,
				},
			},
		},
		IndexIsFound: isFound,
		IndexData:    *index,
	}

	//отправляем в ядро сообщение о возможности продолжения обработки запроса на фильтрацию
	chanIn <- &msgRes
}

//UpdateParametersFiltrationTask изменение параметров задачи по фильтрации
func UpdateParametersFiltrationTask(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters,
	smt *configure.StoringMemoryTask) error {

	var err error
	funcName := "UpdateParametersFiltrationTask"

	ts, ok := smt.GetTaskStatusStoringMemoryTask(req.TaskID, "filtration")
	if !ok {
		//если задачи с указанным ID не существует
		ao, ok := req.AdvancedOptions.(configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus)
		if ok {
			//обновляем детальную информацию о ходе фильтрации
			err = qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: req.TaskID}}, bson.D{
				bson.E{Key: "$set", Value: bson.D{
					bson.E{Key: "detailed_information_on_filtering.task_status", Value: ao.TaskStatus},
					bson.E{Key: "detailed_information_on_filtering.time_interval_task_execution.end", Value: time.Now().Unix()},
					bson.E{Key: "detailed_information_on_filtering.number_files_meet_filter_parameters", Value: ao.NumberFilesMeetFilterParameters},
					bson.E{Key: "detailed_information_on_filtering.number_processed_files", Value: ao.NumberProcessedFiles},
					bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: ao.NumberFilesFoundResultFiltering},
					bson.E{Key: "detailed_information_on_filtering.number_directory_filtartion", Value: ao.NumberDirectoryFiltartion},
					bson.E{Key: "detailed_information_on_filtering.number_error_processed_files", Value: ao.NumberErrorProcessedFiles},
					bson.E{Key: "detailed_information_on_filtering.size_files_meet_filter_parameters", Value: ao.SizeFilesMeetFilterParameters},
					bson.E{Key: "detailed_information_on_filtering.size_files_found_result_filtering", Value: ao.SizeFilesFoundResultFiltering},
					bson.E{Key: "detailed_information_on_filtering.path_directory_for_filtered_files", Value: ao.PathStorageSource},
					bson.E{Key: "detailed_information_on_downloading.number_files_total", Value: ao.NumberFilesFoundResultFiltering},
				}}})
		}

		return fmt.Errorf("task with ID '%v', (%v)", req.TaskID, funcName)
	}

	if (ts.Status == "execute") && ((time.Now().Unix() - ts.TimeLastUpdate) < 31) {
		return nil
	}

	taskInfo, ok := smt.GetStoringMemoryTask(req.TaskID)
	if !ok {
		return fmt.Errorf("task with ID '%v', (%v)", req.TaskID, funcName)
	}

	ti := taskInfo.TaskParameter.FiltrationTask

	//обновление основной информации
	commonValueUpdate := bson.D{
		bson.E{Key: "$set", Value: bson.D{
			bson.E{Key: "detailed_information_on_filtering.task_status", Value: ts.Status},
			bson.E{Key: "detailed_information_on_filtering.time_interval_task_execution.end", Value: time.Now().Unix()},
			bson.E{Key: "detailed_information_on_filtering.number_files_meet_filter_parameters", Value: ti.NumberFilesMeetFilterParameters},
			bson.E{Key: "detailed_information_on_filtering.number_processed_files", Value: ti.NumberProcessedFiles},
			bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: ti.NumberFilesFoundResultFiltering},
			bson.E{Key: "detailed_information_on_filtering.number_directory_filtartion", Value: ti.NumberDirectoryFiltartion},
			bson.E{Key: "detailed_information_on_filtering.number_error_processed_files", Value: ti.NumberErrorProcessedFiles},
			bson.E{Key: "detailed_information_on_filtering.size_files_meet_filter_parameters", Value: ti.SizeFilesMeetFilterParameters},
			bson.E{Key: "detailed_information_on_filtering.size_files_found_result_filtering", Value: ti.SizeFilesFoundResultFiltering},
			bson.E{Key: "detailed_information_on_filtering.path_directory_for_filtered_files", Value: ti.PathStorageSource},
			bson.E{Key: "detailed_information_on_downloading.number_files_total", Value: ti.NumberFilesFoundResultFiltering},
		}}}

	//обновляем детальную информацию о ходе фильтрации
	err = qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: req.TaskID}}, commonValueUpdate)

	lfdi, ok := smt.GetListFilesDetailedInformation(req.TaskID)
	if !ok {
		return nil
	}

	//очищаем отображение с файлами чтобы предотвратить утечки памяти
	defer func(lfdi map[string]configure.DetailedFilesInformation) {
		lfdi = make(map[string]configure.DetailedFilesInformation)
	}(lfdi)

	arr := []interface{}{}
	for n, v := range lfdi {
		arr = append(arr, bson.D{
			bson.E{Key: "file_name", Value: n},
			bson.E{Key: "file_size", Value: v.Size},
			bson.E{Key: "file_hex", Value: v.Hex},
			bson.E{Key: "file_loaded", Value: false},
		})
	}

	arrayValueUpdate := bson.D{
		bson.E{
			Key: "$addToSet", Value: bson.D{
				bson.E{
					Key: "list_files_result_task_execution",
					Value: bson.D{
						bson.E{
							Key:   "$each",
							Value: arr,
						},
					},
				},
			},
		},
	}

	//обновление информации об отфильтрованном файле
	err = qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: req.TaskID}}, arrayValueUpdate)

	//обновление таймера вставки информации в БД
	smt.TimerUpdateTaskInsertDB(req.TaskID)

	return err
}

//RestoreParametersFiltrationTask восстановление информации о задачи по фильтрации
func RestoreParametersFiltrationTask(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters,
	smt *configure.StoringMemoryTask) error {

	//fmt.Println("func 'RestoreParametersFiltrationTask', START...")

	//восстанавливаем задачу по ее ID
	taskInfoFromDB, err := getInfoTaskForID(qp, req.TaskID)
	if err != nil {

		//fmt.Println("func 'RestoreParametersFiltrationTask', ERROR 111")

		return err
	}

	if len(*taskInfoFromDB) == 0 {

		//fmt.Println("func 'RestoreParametersFiltrationTask', ERROR 222")

		return fmt.Errorf("the task ID information cannot be restored because the ID information was not found")
	}

	itd := (*taskInfoFromDB)[0]

	taskStatusRecovery := itd.DetailedInformationOnFiltering.TaskStatus

	tfmffiats, ok := req.AdvancedOptions.(configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus)
	if !ok {
		return fmt.Errorf("error when converting types")
	}

	taskStatusRecovery = tfmffiats.TaskStatus

	smt.RecoverStoringMemoryTask(configure.TaskDescription{
		ClientID:                        itd.ClientID,
		ClientTaskID:                    itd.ClientTaskID,
		TaskType:                        "filtration control",
		ModuleThatSetTask:               "API module",
		ModuleResponsibleImplementation: "NI module",
		TimeUpdate:                      time.Now().Unix(),
		TimeInterval: configure.TimeIntervalTaskExecution{
			Start: itd.DetailedInformationOnFiltering.TimeIntervalTaskExecution.Start,
			End:   itd.DetailedInformationOnFiltering.TimeIntervalTaskExecution.End,
		},
		TaskParameter: configure.DescriptionTaskParameters{
			FiltrationTask: &configure.FiltrationTaskParameters{
				ID:                              itd.SourceID,
				Status:                          taskStatusRecovery,
				UseIndex:                        itd.DetailedInformationOnFiltering.WasIndexUsed,
				NumberFilesMeetFilterParameters: itd.DetailedInformationOnFiltering.NumberFilesMeetFilterParameters,
				NumberProcessedFiles:            itd.DetailedInformationOnFiltering.NumberProcessedFiles,
				NumberFilesFoundResultFiltering: itd.DetailedInformationOnFiltering.NumberFilesFoundResultFiltering,
				NumberDirectoryFiltartion:       itd.DetailedInformationOnFiltering.NumberDirectoryFiltartion,
				NumberErrorProcessedFiles:       itd.DetailedInformationOnFiltering.NumberErrorProcessedFiles,
				SizeFilesMeetFilterParameters:   itd.DetailedInformationOnFiltering.SizeFilesMeetFilterParameters,
				SizeFilesFoundResultFiltering:   itd.DetailedInformationOnFiltering.SizeFilesFoundResultFiltering,
				PathStorageSource:               itd.DetailedInformationOnFiltering.PathDirectoryForFilteredFiles,
			},
			DownloadTask:                 &configure.DownloadTaskParameters{},
			ListFilesDetailedInformation: map[string]*configure.DetailedFilesInformation{},
		},
	}, req.TaskID)

	//fmt.Println("func 'RestoreParametersFiltrationTask', recovered task")

	//если по восстановленной задаче было получено последнее сообщение то нужно проинформировать пользователя
	if taskStatusRecovery == "complete" || taskStatusRecovery == "stop" {
		chanIn <- &configure.MsgBetweenCoreAndDB{
			MsgGenerator:    "DB module",
			MsgRecipient:    "API module",
			MsgSection:      "filtration control",
			Instruction:     "filtration complete",
			TaskID:          req.TaskID,
			TaskIDClientAPI: itd.ClientTaskID,
			AdvancedOptions: configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus{
				TaskStatus:    taskStatusRecovery,
				ListFoundFile: map[string]*configure.DetailedFilesInformation{},
			},
		}
	}

	return nil
}
