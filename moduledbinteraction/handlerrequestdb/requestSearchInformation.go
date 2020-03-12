package handlerrequestdb

import (
	"context"

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
		msgRes.Instruction = "task not found"
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
		TaskID:       liat[0].TaskID,
		ClientTaskID: liat[0].ClientTaskID,
		SourceID:     liat[0].SourceID,
		GeneralInformationAboutTask: configure.GeneralInformationAboutTask{
			TaskProcessed:     liat[0].GeneralInformationAboutTask.TaskProcessed,
			DateTimeProcessed: liat[0].GeneralInformationAboutTask.DateTimeProcessed,
			ClientIDIP:        liat[0].GeneralInformationAboutTask.ClientID, // <ID_client:IP_client>
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

	msgRes.AdvancedOptions = rtp

	chanIn <- &msgRes
}
