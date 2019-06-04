package handlerrequestdb

import (
	"errors"
	"fmt"
	"time"

	//"github.com/mongodb/mongo-go-driver/mongo/options"
	//"github.com/mongodb/mongo-go-driver/bson"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"

	"github.com/mongodb/mongo-go-driver/bson"
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
			TaskStatus:                    "wait",
			ListFilesFoundResultFiltering: []*configure.InformationFilesFoundResultFiltering{},
			WasIndexUsed:                  isFound,
		},
	}

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
	qp QueryParameters,
	smt *configure.StoringMemoryTask) {

	fmt.Println("START function 'UpdateParametersFiltrationTask'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'UpdateParametersFiltrationTask'"

	ao, ok := req.AdvancedOptions.(configure.DetailInfoMsgFiltration)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "type conversion error"+funcName)

		return
	}

	commonValueUpdate := bson.D{
		bson.E{Key: "$set", Value: bson.D{
			bson.E{Key: "detailed_information_on_filtering.task_status", Value: ao.TaskStatus},
			bson.E{Key: "detailed_information_on_filtering.time_interval_task_execution.end", Value: time.Now().Unix()},
			bson.E{Key: "detailed_information_on_filtering.number_files_meet_filter_parameters", Value: ao.NumberFilesMeetFilterParameters},
			bson.E{Key: "detailed_information_on_filtering.number_processed_files", Value: ao.NumberProcessedFiles},
			bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: ao.NumberFilesFoundResultFiltering},
			bson.E{Key: "detailed_information_on_filtering.number_error_processed_files", Value: ao.NumberErrorProcessedFiles},
			bson.E{Key: "detailed_information_on_filtering.number_directory_filtartion", Value: ao.NumberDirectoryFiltartion},
			bson.E{Key: "detailed_information_on_filtering.size_files_meet_filter_parameters", Value: ao.SizeFilesMeetFilterParameters},
			bson.E{Key: "detailed_information_on_filtering.size_files_found_result_filtering", Value: ao.SizeFilesFoundResultFiltering},
			bson.E{Key: "detailed_information_on_filtering.path_directory_for_filtered_files", Value: ao.PathStorageSource},
		}}}

	//обновляем детальную информацию о ходе фильтрации
	if err := qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: ao.TaskID}}, commonValueUpdate); err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

		return
	}

	arr := []interface{}{}

	//если сообщение имеет статус 'execute'
	if ao.TaskStatus == "execute" {
		for fileName, v := range ao.FoundFilesInformation {
			arr = append(arr, bson.D{
				bson.E{Key: "file_name", Value: fileName},
				bson.E{Key: "file_size", Value: v.Size},
				bson.E{Key: "file_hax", Value: v.Hex},
			})
		}
	}

	//если сообщение имеет статусы 'stop' или 'complite'
	if ao.TaskStatus == "stop" || ao.TaskStatus == "complite" {

	}

	/*

	   td, ok := smt.GetStoringMemoryTask(resMsg.Info.TaskID)
	   			if !ok {
	   				_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with %v not found", resMsg.Info.TaskID))

	   				return
	   			}

	   		numFoundFiles := len(td.TaskParameter.FiltrationTask.FoundFilesInformation)
	   			//проверяем общее кол-во найденных файлов
	   			if numFoundFiles != resMsg.Info.NumberFilesFoundResultFiltering {
	   				_ = saveMessageApp.LogMessage("error", fmt.Sprintf("the number of files in the list does not match the total number of files found as a result of filtering (task ID %v)", resMsg.Info.TaskID))
	   			}
	   		}

	   //конвертируем список файлов в подходящий для записи в БД
	   newListFoundFiles := make(map[string]*configure.InputFilesInformation, numFoundFiles)
	   for n, v := range td.TaskParameter.FiltrationTask.FoundFilesInformation {
	   	newListFoundFiles[n] = &configure.InputFilesInformation{
	   		Size: v.Size,
	   		Hex: v.Hex,
	   	}
	*/

	arrayValueUpdate := bson.D{
		bson.E{
			Key: "$addToSet", Value: bson.D{
				bson.E{
					Key: "detailed_information_on_filtering.list_files_found_result_filtering",
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
	if err := qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: ao.TaskID}}, arrayValueUpdate); err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

		return
	}

}
