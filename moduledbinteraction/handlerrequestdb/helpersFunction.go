package handlerrequestdb

import (
	"context"

	"github.com/mongodb/mongo-go-driver/bson"

	"ISEMS-NIH_master/configure"
)

//getInfoTaskForID получить информацию о найденной по параметру 'task_id' задаче
func getInfoTaskForID(qp QueryParameters, taskID string) (*[]configure.InformationAboutTask, error) {
	itf := []configure.InformationAboutTask{}

	cur, err := qp.Find(bson.D{bson.E{Key: "task_id", Value: taskID}})
	if err != nil {
		return &itf, err
	}

	for cur.Next(context.TODO()) {
		var model configure.InformationAboutTask
		err := cur.Decode(&model)
		if err != nil {
			return &itf, err
		}

		itf = append(itf, model)
	}

	if err := cur.Err(); err != nil {
		return &itf, err
	}

	cur.Close(context.TODO())

	return &itf, nil
}

//getShortInformation получить краткую информацию об найденных задачах
func getShortInformation(qp QueryParameters, sp *configure.SearchParameters) ([]*configure.BriefTaskInformation, error) {
	getQueryTmpNetParams := func(fcp configure.FiltrationControlParametersNetworkFilters, queryType string) (bson.E, bson.D) {
		listQueryType := map[string]struct {
			e string
			o configure.FiltrationControlIPorNetorPortParameters
		}{
			"ip":      {e: "ip", o: fcp.IP},
			"port":    {e: "port", o: fcp.Port},
			"network": {e: "network", o: fcp.Network},
		}

		numAny := len(listQueryType[queryType].o.Any)
		numSrc := len(listQueryType[queryType].o.Src)
		numDst := len(listQueryType[queryType].o.Dst)

		if numAny == 0 && numSrc == 0 && numDst == 0 {
			return bson.E{}, bson.D{}
		}

		if numAny > 0 && numSrc == 0 && numDst == 0 {
			be := bson.E{Key: "$or", Value: bson.A{
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".any", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Any}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Any}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Any}}}},
			}}

			bd := bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".any", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Any}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Any}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Any}}}},
			}}}

			return be, bd
		}

		if numSrc > 0 && numAny == 0 && numDst == 0 {
			be := bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Src}}}
			bd := bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Src}}}}

			return be, bd
		}

		if numDst > 0 && numAny == 0 && numSrc == 0 {
			be := bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Dst}}}
			bd := bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Dst}}}}

			return be, bd
		}

		if (numSrc > 0 && numDst > 0) && numAny == 0 {
			be := bson.E{Key: "$and", Value: bson.A{
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Src}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Dst}}}},
			}}
			bd := bson.D{{Key: "$and", Value: bson.A{
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Src}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Dst}}}},
			}}}

			return be, bd
		}

		return bson.E{Key: "$or", Value: bson.A{
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".any", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Any}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Src}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Dst}}}},
			}}, bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".any", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Any}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Src}}}},
				bson.D{{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.D{{Key: "$in", Value: listQueryType[queryType].o.Dst}}}},
			}}}
	}

	checkParameterContainsValues := func(fcinpp configure.FiltrationControlIPorNetorPortParameters) bool {
		if len(fcinpp.Any) > 0 {
			return true
		}

		if len(fcinpp.Src) > 0 {
			return true
		}

		if len(fcinpp.Dst) > 0 {
			return true
		}

		return false
	}

	queryTemplate := map[string]bson.E{
		"sourceID":             (bson.E{Key: "source_id", Value: bson.D{{Key: "$eq", Value: sp.ID}}}),
		"filesIsFound":         (bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: bson.D{{Key: "$gt", Value: 0}}}),
		"taskProcessed":        (bson.E{Key: "general_information_about_task.task_processed", Value: bson.D{{Key: "$eq", Value: sp.TaskProcessed}}}),
		"filesIsDownloaded":    (bson.E{Key: "detailed_information_on_downloading.number_files_downloaded", Value: bson.D{{Key: "$gt", Value: 0}}}),
		"filesIsNotDownloaded": (bson.E{Key: "detailed_information_on_downloading.number_files_downloaded", Value: bson.D{{Key: "$eq", Value: 0}}}),
		"allFilesIsDownloaded": (bson.E{Key: "$expr", Value: bson.D{
			{Key: "$eq", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}}),
		"allFilesIsNotDownloaded": (bson.E{Key: "$expr", Value: bson.D{
			{Key: "$ne", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}}),
		"sizeAllFiles": (bson.E{Key: "detailed_information_on_filtering.size_files_found_result_filtering", Value: bson.D{
			{Key: "$gte", Value: sp.InformationAboutFiltering.SizeAllFilesMin},
			{Key: "$lte", Value: sp.InformationAboutFiltering.SizeAllFilesMax},
		}}),
		"countAllFiles": (bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: bson.D{
			{Key: "$gte", Value: sp.InformationAboutFiltering.CountAllFilesMin},
			{Key: "$lte", Value: sp.InformationAboutFiltering.CountAllFilesMax},
		}}),
		"dateTimeParameters": (bson.E{Key: "$and", Value: bson.A{
			bson.D{{Key: "filtering_option.date_time_interval.start", Value: bson.D{
				{Key: "$gte", Value: sp.InstalledFilteringOption.DateTime.Start}}}},
			bson.D{{Key: "filtering_option.date_time_interval.end", Value: bson.D{
				{Key: "$lte", Value: sp.InstalledFilteringOption.DateTime.End}}}},
		}}),
		"transportProtocol":      (bson.E{Key: "filtering_option.protocol", Value: sp.InstalledFilteringOption.Protocol}),
		"statusFilteringTask":    (bson.E{Key: "detailed_information_on_filtering.task_status", Value: sp.StatusFilteringTask}),
		"statusFileDownloadTask": (bson.E{Key: "detailed_information_on_downloading.task_status", Value: sp.StatusFileDownloadTask}),
	}

	var (
		querySourceID               bson.E
		queryFilesIsFound           bson.E
		querySizeAllFiles           bson.E
		queryCountAllFiles          bson.E
		queryTaskProcessed          bson.E
		queryFilesIsDownloaded      bson.E
		queryTransportProtocol      bson.E
		querydateTimeParameters     bson.E
		queryStatusFilteringTask    bson.E
		queryAllFilesIsDownloaded   bson.E
		queryNetworkParametersPort  bson.E
		queryNetworkParametersIPNet bson.E
		queryStatusFileDownloadTask bson.E
	)

	//поиск по ID источника
	if sp.ID > 0 {
		querySourceID = queryTemplate["sourceID"]
	}

	//была ли задача обработана
	if sp.ConsiderParameterTaskProcessed {
		queryTaskProcessed = queryTemplate["taskProcessed"]
	}

	//выполнялась ли выгрузка файлов
	if sp.ConsiderParameterFilesIsDownloaded {
		if sp.FilesIsDownloaded {
			queryFilesIsDownloaded = queryTemplate["filesIsDownloaded"]
		} else {
			queryFilesIsDownloaded = queryTemplate["filesIsNotDownloaded"]
		}
	}

	//все ли файлы были выгружены
	if sp.ConsiderParameterAllFilesIsDownloaded {
		if sp.AllFilesIsDownloaded {
			queryFilesIsDownloaded = queryTemplate["filesIsDownloaded"]
			queryAllFilesIsDownloaded = queryTemplate["allFilesIsDownloaded"]
		} else {
			queryAllFilesIsDownloaded = queryTemplate["allFilesIsNotDownloaded"]
		}
	}

	//были ли найденны какие либо файлы в результате фильтрации
	if sp.InformationAboutFiltering.FilesIsFound {
		queryFilesIsFound = queryTemplate["filesIsFound"]
	}

	//диапазон количества найденных файлов
	cafmin := sp.InformationAboutFiltering.CountAllFilesMin
	cafmax := sp.InformationAboutFiltering.CountAllFilesMax
	if (cafmax > 0) && (cafmax > cafmin) {
		queryCountAllFiles = queryTemplate["countAllFiles"]
	}

	//диапазон общего размера всех найденных файлов
	safmin := sp.InformationAboutFiltering.SizeAllFilesMin
	safmax := sp.InformationAboutFiltering.SizeAllFilesMax
	if (safmax > 0) && (safmax > safmin) {
		querySizeAllFiles = queryTemplate["sizeAllFiles"]
	}

	//временной диапазон фильтруемых данных
	dts := sp.InstalledFilteringOption.DateTime.Start
	dte := sp.InstalledFilteringOption.DateTime.End
	if (dts > 0) && (dte > 0) && (dts < dte) {
		querydateTimeParameters = queryTemplate["dateTimeParameters"]
	}

	//транспортный протокол
	if sp.InstalledFilteringOption.Protocol == "tcp" || sp.InstalledFilteringOption.Protocol == "udp" {
		queryTransportProtocol = queryTemplate["transportProtocol"]
	}

	//статус задачи по фильтрации
	if (len(sp.StatusFilteringTask) > 0) && (sp.StatusFilteringTask != "any") {
		queryStatusFilteringTask = queryTemplate["statusFilteringTask"]
	}

	//статус задачи по скачиванию файлов
	if (len(sp.StatusFileDownloadTask) > 0) && (sp.StatusFileDownloadTask != "any") {
		queryStatusFileDownloadTask = queryTemplate["statusFileDownloadTask"]
	}

	isContainsValueIP := checkParameterContainsValues(sp.InstalledFilteringOption.NetworkFilters.IP)
	isContainsValuePort := checkParameterContainsValues(sp.InstalledFilteringOption.NetworkFilters.Port)
	isContainsValueNetwork := checkParameterContainsValues(sp.InstalledFilteringOption.NetworkFilters.Network)

	if isContainsValuePort {
		queryNetworkParametersPort, _ = getQueryTmpNetParams(sp.InstalledFilteringOption.NetworkFilters, "port")
	}

	if isContainsValueIP && !isContainsValueNetwork {
		queryNetworkParametersIPNet, _ = getQueryTmpNetParams(sp.InstalledFilteringOption.NetworkFilters, "ip")
	}

	if isContainsValueNetwork && !isContainsValueIP {
		queryNetworkParametersIPNet, _ = getQueryTmpNetParams(sp.InstalledFilteringOption.NetworkFilters, "network")
	}

	if isContainsValueIP && isContainsValueNetwork {
		_, bdIP := getQueryTmpNetParams(sp.InstalledFilteringOption.NetworkFilters, "ip")
		_, bdNetwork := getQueryTmpNetParams(sp.InstalledFilteringOption.NetworkFilters, "network")

		queryNetworkParametersIPNet = bson.E{Key: "$or", Value: bson.A{bdIP, bdNetwork}}
	}

	lbti := []*configure.BriefTaskInformation{}

	cur, err := qp.Find(bson.D{
		querySourceID,
		queryTaskProcessed,
		queryFilesIsDownloaded,
		queryAllFilesIsDownloaded,
		queryFilesIsFound,
		queryCountAllFiles,
		querySizeAllFiles,
		querydateTimeParameters,
		queryTransportProtocol,
		queryStatusFilteringTask,
		queryStatusFileDownloadTask,
		queryNetworkParametersPort,
		queryNetworkParametersIPNet})
	if err != nil {
		return lbti, err
	}

	for cur.Next(context.Background()) {
		var model configure.InformationAboutTask
		err := cur.Decode(&model)
		if err != nil {
			return lbti, err
		}

		bti := configure.BriefTaskInformation{
			TaskID:                 model.TaskID,
			ClientTaskID:           model.ClientTaskID,
			SourceID:               model.SourceID,
			StartTimeTaskExecution: model.DetailedInformationOnFiltering.TimeIntervalTaskExecution.Start,
			ParametersFiltration: configure.ParametersFiltrationOptions{
				DateTime: configure.DateTimeParameters{
					Start: model.FilteringOption.DateTime.Start,
					End:   model.FilteringOption.DateTime.End,
				},
				Protocol: model.FilteringOption.Protocol,
				Filters: configure.FiltrationControlParametersNetworkFilters{
					IP: configure.FiltrationControlIPorNetorPortParameters{
						Any: model.FilteringOption.Filters.IP.Any,
						Src: model.FilteringOption.Filters.IP.Src,
						Dst: model.FilteringOption.Filters.IP.Dst,
					},
					Port: configure.FiltrationControlIPorNetorPortParameters{
						Any: model.FilteringOption.Filters.Port.Any,
						Src: model.FilteringOption.Filters.Port.Src,
						Dst: model.FilteringOption.Filters.Port.Dst,
					},
					Network: configure.FiltrationControlIPorNetorPortParameters{
						Any: model.FilteringOption.Filters.Network.Any,
						Src: model.FilteringOption.Filters.Network.Src,
						Dst: model.FilteringOption.Filters.Network.Dst,
					},
				},
			},
			FilteringTaskStatus:                  model.DetailedInformationOnFiltering.TaskStatus,
			FileDownloadTaskStatus:               model.DetailedInformationOnDownloading.TaskStatus,
			NumberFilesFoundAsResultFiltering:    model.DetailedInformationOnFiltering.NumberFilesFoundResultFiltering,
			TotalSizeFilesFoundAsResultFiltering: model.DetailedInformationOnFiltering.SizeFilesFoundResultFiltering,
			NumberFilesDownloaded:                model.DetailedInformationOnDownloading.NumberFilesDownloaded,
		}

		lbti = append(lbti, &bti)
	}

	if err := cur.Err(); err != nil {
		return lbti, err
	}

	cur.Close(context.Background())

	return lbti, nil
}
