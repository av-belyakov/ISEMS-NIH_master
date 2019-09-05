package handlerrequestdb

import (
	"errors"
	"fmt"
	"time"

	//"github.com/mongodb/mongo-go-driver/mongo/options"
	//"github.com/mongodb/mongo-go-driver/bson"

	"ISEMS-NIH_master/configure"

	"github.com/mongodb/mongo-go-driver/bson"
)

//CreateNewFiltrationTask запись информации о новой фильтрации
//обрабатывается при получении запроса на фильтрацию
func CreateNewFiltrationTask(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgSection:   "filtration control",
		IDClientAPI:  req.IDClientAPI,
		TaskID:       req.TaskID,
	}

	tf, ok := req.AdvancedOptions.(configure.FiltrationControlCommonParametersFiltration)
	if !ok {
		errMsg := "taken incorrect settings for task filtering"

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
	isFound, index, err := searchIndexFormFiltration("index_filtration", &tf, qp)
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
		TaskID:       req.TaskID,
		ClientID:     req.IDClientAPI,
		ClientTaskID: req.TaskIDClientAPI,
		SourceID:     tf.ID,
		FilteringOption: configure.FilteringOption{
			DateTime: configure.TimeInterval{
				Start: tf.DateTime.Start,
				End:   tf.DateTime.End,
			},
			Protocol: tf.Protocol,
			Filters: configure.FilteringExpressions{
				IP: configure.FilteringNetworkParameters{
					Any: tf.Filters.IP.Any,
					Src: tf.Filters.IP.Src,
					Dst: tf.Filters.IP.Dst,
				},
				Port: configure.FilteringNetworkParameters{
					Any: tf.Filters.Port.Any,
					Src: tf.Filters.Port.Src,
					Dst: tf.Filters.Port.Dst,
				},
				Network: configure.FilteringNetworkParameters{
					Any: tf.Filters.Network.Any,
					Src: tf.Filters.Network.Src,
					Dst: tf.Filters.Network.Dst,
				},
			},
		},
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
		FilteringOption: tf,
		IndexIsFound:    isFound,
		IndexData:       *index,
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

	//получаем всю информацию по выполняемой задаче
	taskInfo, ok := smt.GetStoringMemoryTask(req.TaskID)
	if !ok {
		fmt.Println("\tвосстанавливаем задачу по ее ID")

		//восстанавливаем задачу по ее ID
		taskInfoFromDB, err := getInfoTaskForID(qp, req.TaskID)
		if err != nil {
			return err
		}

		if len(*taskInfoFromDB) == 0 {
			return err
		}

		itd := (*taskInfoFromDB)[0]
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
				FiltrationTask: configure.FiltrationTaskParameters{
					ID:                              itd.SourceID,
					Status:                          itd.DetailedInformationOnFiltering.TaskStatus,
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
			},
		}, req.TaskID)

		return fmt.Errorf("task with ID '%v' not found (DB module)", req.TaskID)
	}

	ti := taskInfo.TaskParameter.FiltrationTask

	//выполнять обновление информации в БД для сообщения типа 'complete' всегда,
	// для сообщения типа 'execute' только раз 31 секунду
	if (ti.Status == "execute") && ((time.Now().Unix() - taskInfo.TimeInsertDB) < 30) {
		return nil
	}

	//обновление основной информации
	commonValueUpdate := bson.D{
		bson.E{Key: "$set", Value: bson.D{
			bson.E{Key: "detailed_information_on_filtering.task_status", Value: ti.Status},
			bson.E{Key: "detailed_information_on_filtering.time_interval_task_execution.end", Value: time.Now().Unix()},
			bson.E{Key: "detailed_information_on_filtering.number_files_meet_filter_parameters", Value: ti.NumberFilesMeetFilterParameters},
			bson.E{Key: "detailed_information_on_filtering.number_processed_files", Value: ti.NumberProcessedFiles},
			bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: ti.NumberFilesFoundResultFiltering},
			bson.E{Key: "detailed_information_on_filtering.number_directory_filtartion", Value: ti.NumberDirectoryFiltartion},
			bson.E{Key: "detailed_information_on_filtering.number_error_processed_files", Value: ti.NumberErrorProcessedFiles},
			bson.E{Key: "detailed_information_on_filtering.size_files_meet_filter_parameters", Value: ti.SizeFilesMeetFilterParameters},
			bson.E{Key: "detailed_information_on_filtering.size_files_found_result_filtering", Value: ti.SizeFilesFoundResultFiltering},
			bson.E{Key: "detailed_information_on_filtering.path_directory_for_filtered_files", Value: ti.PathStorageSource},
		}}}

	//обновляем детальную информацию о ходе фильтрации
	err = qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: req.TaskID}}, commonValueUpdate)

	arr := []interface{}{}

	for n, v := range ti.FoundFilesInformation {
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

	infoMsg := configure.MsgBetweenCoreAndDB{
		MsgGenerator:    "DB module",
		MsgRecipient:    "API module",
		MsgSection:      "filtration control",
		Instruction:     "filtration complete",
		IDClientAPI:     req.IDClientAPI,
		TaskID:          req.TaskID,
		TaskIDClientAPI: taskInfo.ClientTaskID,
	}

	//если статус задачи "stop" или "complete" через ядро останавливаем задачу и оповещаем пользователя
	if ti.Status == "stop" || ti.Status == "complete" {
		chanIn <- &infoMsg
	}

	//если статус задачи "refused" то есть, задача была отклонена
	if ti.Status == "refused" {
		infoMsg.Instruction = "filtration refused"

		chanIn <- &infoMsg
	}

	return err
}
