package handlerrequestdb

import (
	"errors"
	"fmt"

	//"github.com/mongodb/mongo-go-driver/mongo/options"
	//"github.com/mongodb/mongo-go-driver/bson"

	"ISEMS-NIH_master/configure"
)

//CreateNewFiltrationTask запись информации о новой фильтрации
//обрабатывается при получении запроса на фильтрацию
func CreateNewFiltrationTask(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	fmt.Println("START function 'CreateNewFiltrationTask'...")

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

	itf := configure.InformationAboutTaskFiltration{
		TaskID:       req.TaskID,
		ClientID:     req.IDClientAPI,
		ClientTaskID: req.TaskIDClientAPI,
		FilteringOption: configure.FiletringOption{
			ID: tf.ID,
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
			TaskStatus: "wait",
		},
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

	itf.DetailedInformationOnFiltering.WasIndexUsed = isFound

	insertData := make([]interface{}, 0, 1)
	insertData = append(insertData, itf)

	fmt.Printf("------- %v --------\n", insertData)

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
	qp QueryParameters) {

	fmt.Println("START function 'UpdateParametersFiltrationTask'...")
	/*

	   !!! ВНИМАНИЕ !!!
	   Это примерный шаблон, наверное подробно стоит писать, когда
	   приложение начнет принимать данные о фильтрации от
	   приложения ISEMS-NIH_slave

	   Но перед этим нужно сделать на ISEMS-NIH_slave:
	   1. Подготовить функцию для формирование шаблона для фильтрации
	   + 2. Описать JSON тип для передачи информации о ходе фильтрации
	   3. Функцию выполнения фильтрации
	   4. Функцию передачи данных при разрывесоединения и его последующего
	   востановления
	   ______________________________________________________________

	   	msgRes := configure.MsgBetweenCoreAndDB{
	   		MsgGenerator: req.MsgRecipient,
	   		MsgRecipient: req.MsgGenerator,
	   		MsgSection:   "filtration control",
	   		IDClientAPI:  req.IDClientAPI,
	   		TaskID:       req.TaskID,
	   	}

	   	//приведение типа
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

	   	valueUpdate := bson.D{
	   		bson.E{Key: "$set", Value: bson.D{
	   			bson.E{Key: "detailed_information_on_filtering.task_status", Value: ftp.Status},
	   			bson.E{Key: "detailed_information_on_filtering.was_index_used", Value: ftp.UseIndex},
	   			bson.E{Key: "detailed_information_on_filtering.time_interval_task_execution.end", Value: time.Now().Unix()},
	   			bson.E{Key: "detailed_information_on_filtering.number_files_meet_filter_parameters", Value: ftp.NumberFilesToBeFiltered},
	   			bson.E{Key: "detailed_information_on_filtering.number_processed_files", Value: ftp.NumberFilesProcessed},
	   			bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: ftp.NumberFilesFound},
	   			bson.E{Key: "detailed_information_on_filtering.number_directory_filtartion", Value: ftp.CountDirectoryFiltartion},
	   			bson.E{Key: "detailed_information_on_filtering.size_files_meet_filter_parameters", Value: ftp.SizeFilesToBeFiltered},
	   			bson.E{Key: "detailed_information_on_filtering.size_files_found_result_filtering", Value: ftp.SizeFilesFound},
	   			bson.E{Key: "detailed_information_on_filtering.path_directory_for_filtered_files", Value: ftp.PathStorageSource},
	   			bson.E{Key: "detailed_information_on_filtering.list_files_found_result_filtering", Value: ftp.FoundFilesInformation},
	   		}}}

	   	//запись информации по задачи фильтрации в коллекцию 'filter_task_list'
	   	if _, err := qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: req.TaskID}}, valueUpdate); err != nil {
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
	*/
}
