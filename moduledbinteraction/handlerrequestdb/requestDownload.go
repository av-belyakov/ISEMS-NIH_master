package handlerrequestdb

import (
	"fmt"
	"time"

	"ISEMS-NIH_master/configure"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo/options"
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

//UpdateInformationAboutTask запись информации по задаче
func UpdateInformationAboutTask(
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters,
	smt *configure.StoringMemoryTask) error {

	const timeUpdate = 30

	ti, ok := smt.GetStoringMemoryTask(req.TaskID)
	if !ok {
		return fmt.Errorf("task with ID '%v' not found (DB module)", req.TaskID)
	}

	taskStatus := ti.TaskParameter.DownloadTask.Status

	//выполнять обновление информации в БД для сообщения типа 'complete' всегда,
	// для сообщения типа 'execute' только раз 31 секунду
	if (taskStatus == "execute") && ((time.Now().Unix() - ti.TimeInsertDB) < timeUpdate) {
		return nil
	}

	//обновление основной информации
	commonValueUpdate := bson.D{
		bson.E{Key: "$set", Value: bson.D{
			bson.E{Key: "user_initiated_file_download_process", Value: ti.UserName},
			bson.E{Key: "detailed_information_on_downloading.task_status", Value: taskStatus},
			bson.E{Key: "detailed_information_on_downloading.time_interval_task_execution.start", Value: ti.TimeInterval.Start},
			bson.E{Key: "detailed_information_on_downloading.path_directory_storage_downloaded_files", Value: ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles},
		}}}

	var arrayFiles []interface{}
	isTaskComplete := false
	if verificationStatus, ok := req.AdvancedOptions.(string); ok {
		if verificationStatus == "task complete" {
			isTaskComplete = true
		}
	} else {
		if ti.TaskParameter.DownloadTask.Status == "complete" {
			isTaskComplete = true
		}
	}

	if isTaskComplete {
		taskInfo, _ := getInfoTaskForID(qp, req.TaskID)

		if len(*taskInfo) > 0 {
			commonValueUpdate = bson.D{
				bson.E{Key: "$set", Value: bson.D{
					bson.E{Key: "detailed_information_on_downloading.task_status", Value: taskStatus},
					bson.E{Key: "detailed_information_on_downloading.time_interval_task_execution.start", Value: ti.TimeInterval.Start},
					bson.E{Key: "detailed_information_on_downloading.time_interval_task_execution.end", Value: time.Now().Unix()},
					bson.E{Key: "detailed_information_on_downloading.number_files_downloaded", Value: (*taskInfo)[0].DetailedInformationOnDownloading.NumberFilesDownloaded + ti.TaskParameter.DownloadTask.NumberFilesDownloaded},
					bson.E{Key: "detailed_information_on_downloading.number_files_downloaded_error", Value: (*taskInfo)[0].DetailedInformationOnDownloading.NumberFilesDownloadedError + ti.TaskParameter.DownloadTask.NumberFilesDownloadedError},
					bson.E{Key: "detailed_information_on_downloading.path_directory_storage_downloaded_files", Value: ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles},
				}}}

		}

		for fn, fi := range ti.TaskParameter.DownloadTask.DownloadingFilesInformation {
			if fi.IsLoaded {
				arrayFiles = append(arrayFiles, bson.D{bson.E{Key: "elem.file_name", Value: fn}})
			}
		}
	} else {
		for fn, fi := range ti.TaskParameter.DownloadTask.DownloadingFilesInformation {
			if fi.IsLoaded && (fi.TimeDownload > time.Now().Unix()-(timeUpdate*2)) {
				arrayFiles = append(arrayFiles, bson.D{bson.E{Key: "elem.file_name", Value: fn}})
			}
		}
	}

	//обновляем детальную информацию о ходе скачивания файлов
	if err := qp.UpdateOne(bson.D{bson.E{Key: "task_id", Value: req.TaskID}}, commonValueUpdate); err != nil {
		return err
	}

	if len(arrayFiles) == 0 {
		return nil
	}

	//обновляем информацию по загруженным файлам
	if err := qp.UpdateOneArrayFilters(bson.D{
		bson.E{Key: "task_id", Value: req.TaskID}},
		bson.D{
			bson.E{Key: "$set", Value: bson.D{
				bson.E{Key: "list_files_result_task_execution.$[elem].file_loaded", Value: true},
			}}},
		&options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: []interface{}{bson.D{bson.E{
					Key: "$or", Value: arrayFiles,
				}}},
			},
		}); err != nil {

		return err
	}

	return nil
}
