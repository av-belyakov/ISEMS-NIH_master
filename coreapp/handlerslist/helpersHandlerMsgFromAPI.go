package handlerslist

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//handlerFiltrationControlTypeStart обработчик запроса на фильтрацию
func handlerFiltrationControlTypeStart(
	fcts *configure.FiltrationControlTypeStart,
	hsm HandlersStoringMemory,
	clientID string,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) {

	funcName := "handlerFiltrationControlTypeStart"

	var sriga bool
	if fcts.MsgOption.UserName == "" {
		sriga = true
	}

	emt := ErrorMessageType{
		SourceID:                              fcts.MsgOption.ID,
		TaskIDClientAPI:                       fcts.MsgCommon.ClientTaskID,
		IDClientAPI:                           clientID,
		Section:                               "filtration control",
		Instruction:                           "task processing",
		MsgType:                               "danger",
		SearchRequestIsGeneratedAutomatically: sriga,
		ChanToAPI:                             chanToAPI,
	}

	//проверяем параметры фильтрации
	if msg, ok := сheckParametersFiltration(&fcts.MsgOption); !ok {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "incorrect parameters for filtering are set",
			FuncName:    funcName,
		})

		emt.MsgHuman = msg

		//сообщение о том что задача была отклонена
		if err := ErrorMessage(emt); err != nil {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    funcName,
			})
		}

		return
	}

	//проверяем состояние подключения источника
	connectionStatus, err := hsm.ISL.GetSourceConnectionStatus(fcts.MsgOption.ID)
	if err != nil || !connectionStatus {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprintf("source with ID %v not connected", fcts.MsgOption.ID),
			FuncName:    funcName,
		})

		emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID:   fcts.MsgOption.ID,
			TaskType:   "фильтрация",
			TaskAction: "задача отклонена",
			Message:    "источник не подключен",
		})

		//сообщение о том что задача была отклонена
		if err := ErrorMessage(emt); err != nil {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    funcName,
			})
		}

		return
	}

	//получаем новый идентификатор задачи
	taskID := common.GetUniqIDFormatMD5(clientID + "_" + fcts.ClientTaskID)

	//добавляем новую задачу в очередь задач
	hsm.QTS.AddQueueTaskStorage(taskID, fcts.MsgOption.ID, configure.CommonTaskInfo{
		IDClientAPI:     clientID,
		TaskIDClientAPI: fcts.ClientTaskID,
		TaskType:        "filtration control",
		UserName:        fcts.MsgOption.UserName,
	}, &configure.DescriptionParametersReceivedFromUser{
		FilterationParameters: configure.FilteringOption{
			DateTime: configure.TimeInterval{
				Start: fcts.MsgOption.DateTime.Start,
				End:   fcts.MsgOption.DateTime.End,
			},
			Protocol: fcts.MsgOption.Protocol,
			Filters: configure.FilteringExpressions{
				IP: configure.FilteringNetworkParameters{
					Any: fcts.MsgOption.Filters.IP.Any,
					Src: fcts.MsgOption.Filters.IP.Src,
					Dst: fcts.MsgOption.Filters.IP.Dst,
				},
				Port: configure.FilteringNetworkParameters{
					Any: fcts.MsgOption.Filters.Port.Any,
					Src: fcts.MsgOption.Filters.Port.Src,
					Dst: fcts.MsgOption.Filters.Port.Dst,
				},
				Network: configure.FilteringNetworkParameters{
					Any: fcts.MsgOption.Filters.Network.Any,
					Src: fcts.MsgOption.Filters.Network.Src,
					Dst: fcts.MsgOption.Filters.Network.Dst,
				},
			},
		},
		DownloadList: []string{},
	})

	//устанавливаем проверочный статус источника для данной задачи как подключен
	if err := hsm.QTS.ChangeAvailabilityConnectionOnConnection(fcts.MsgOption.ID, taskID); err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})
	}

	//информационное сообщение отправляем только если задача была сгенерирована пользователем
	if sriga {
		return
	}

	//информационное сообщение о том что задача добавлена в очередь
	notifications.SendNotificationToClientAPI(
		chanToAPI,
		notifications.NotificationSettingsToClientAPI{
			MsgType: "success",
			MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   fcts.MsgOption.ID,
				TaskType:   "фильтрация",
				TaskAction: "задача добавлена в очередь",
			}),
			Sources: []int{fcts.MsgOption.ID},
		},
		fcts.ClientTaskID,
		clientID)
}

//handlerInformationSearchControlTypeSearchCommanInformation обработчик запроса по поиску общей информации о задачах
func handlerInformationSearchControlTypeSearchCommanInformation(
	siatr *configure.SearchInformationAboutTasksRequest,
	hsm HandlersStoringMemory,
	clientID string,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	chanToDB chan<- *configure.MsgBetweenCoreAndDB) {

	funcName := "handlerInformationSearchControlTypeSearchCommanInformation"

	emt := ErrorMessageType{
		TaskIDClientAPI:                       siatr.MsgCommon.ClientTaskID,
		IDClientAPI:                           clientID,
		Section:                               "information search control",
		Instruction:                           "task processing",
		MsgType:                               "danger",
		SearchRequestIsGeneratedAutomatically: siatr.MsgOption.SearchRequestIsGeneratedAutomatically,
		ChanToAPI:                             chanToAPI,
	}

	//проверяем параметры необходимые для поиска общей информации по задачам
	if msg, ok := CheckParametersSearchCommonInformation(&siatr.MsgOption); !ok {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "incorrect search parameters are set",
			FuncName:    funcName,
		})

		emt.MsgHuman = msg

		//сообщение о том что задача была отклонена
		if err := ErrorMessage(emt); err != nil {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    funcName,
			})
		}

		return
	}

	//добавляем информацию о задаче в кеширующий модуль
	taskID, _, err := hsm.TSSQ.CreateNewSearchTask(clientID, &configure.SearchParameters{
		SearchRequestIsGeneratedAutomatically: siatr.MsgOption.SearchRequestIsGeneratedAutomatically,
		ID:                                    siatr.MsgOption.ID,
		ConsiderParameterTaskProcessed:        siatr.MsgOption.ConsiderParameterTaskProcessed,
		TaskProcessed:                         siatr.MsgOption.TaskProcessed,
		ConsiderParameterFilesIsDownloaded:    siatr.MsgOption.ConsiderParameterFilesIsDownloaded,
		FilesIsDownloaded:                     siatr.MsgOption.FilesIsDownloaded,
		ConsiderParameterAllFilesIsDownloaded: siatr.MsgOption.ConsiderParameterAllFilesIsDownloaded,
		AllFilesIsDownloaded:                  siatr.MsgOption.AllFilesIsDownloaded,
		InformationAboutFiltering:             siatr.MsgOption.InformationAboutFiltering,
		InstalledFilteringOption:              siatr.MsgOption.InstalledFilteringOption,
	})
	if err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})

		emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
			TaskType:   "поиск информации о задаче",
			TaskAction: "задача отклонена",
			Message:    "невозможно выполнить поиск, внутренняя ошибка приложения",
		})

		//сообщение о том что задача была отклонена
		if err := ErrorMessage(emt); err != nil {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    funcName,
			})
		}

		return
	}

	chanToDB <- &configure.MsgBetweenCoreAndDB{
		MsgGenerator:    "Core module",
		MsgRecipient:    "DB module",
		MsgSection:      "information search control",
		Instruction:     "search common information",
		IDClientAPI:     clientID,
		TaskID:          taskID,
		TaskIDClientAPI: siatr.ClientTaskID,
	}
}

//checkParametersFiltration проверяет параметры фильтрации
func сheckParametersFiltration(fccpf *configure.FiltrationControlCommonParametersFiltration) (string, bool) {
	//проверяем наличие ID источника
	if fccpf.ID == 0 {
		return common.PatternUserMessage(&common.TypePatternUserMessage{
			TaskType:   "фильтрация",
			TaskAction: "задача отклонена",
			Message:    "отсутствует идентификатор источника",
		}), false
	}

	//проверяем временной интервал
	isZero := ((fccpf.DateTime.Start == 0) || (fccpf.DateTime.End == 0))
	if isZero || (fccpf.DateTime.Start > fccpf.DateTime.End) {
		return common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID:   fccpf.ID,
			TaskType:   "фильтрация",
			TaskAction: "задача отклонена",
			Message:    "задан неверный временной интервал",
		}), false
	}

	//проверяем тип сетевого протокола
	if isCorrectProtocol := checkCorrectProtocol(fccpf.Protocol); !isCorrectProtocol {
		fccpf.Protocol = "any"
	}

	filterParameters := map[string]map[string]*[]string{
		"IP": map[string]*[]string{
			"Any": &fccpf.Filters.IP.Any,
			"Src": &fccpf.Filters.IP.Src,
			"Dst": &fccpf.Filters.IP.Dst,
		},
		"Port": map[string]*[]string{
			"Any": &fccpf.Filters.Port.Any,
			"Src": &fccpf.Filters.Port.Src,
			"Dst": &fccpf.Filters.Port.Dst,
		},
		"Network": map[string]*[]string{
			"Any": &fccpf.Filters.Network.Any,
			"Src": &fccpf.Filters.Network.Src,
			"Dst": &fccpf.Filters.Network.Dst,
		},
	}

	if !checkNetworkParametersIsNotEmpty(filterParameters) {
		return common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID:   fccpf.ID,
			TaskType:   "фильтрация",
			TaskAction: "задача отклонена",
			Message:    "необходимо указать хотя бы один искомый ip адрес, порт или подсеть",
		}), false
	}

	//проверяем параметры сетевых фильтров
	if err := LoopHandler(LoopHandlerParameters{
		OptionsCheckFilterParameters{
			SourceID: fccpf.ID,
			TaskType: "фильтрация",
		},
		filterParameters,
		CheckIPPortNetwork,
	}); err != nil {
		return fmt.Sprint(err), false
	}

	return "", true
}

//CheckParametersSearchCommonInformation проверяет параметры запроса для поиска общей информации
func CheckParametersSearchCommonInformation(siatro *configure.SearchInformationAboutTasksRequestOption) (string, bool) {

	fmt.Println("func 'checkParametersSearchCommonInformation', START...")

	checkDateTimeFiltering := func(dtp configure.DateTimeParameters) (string, bool) {
		if dtp.Start == 0 && dtp.End != 0 {
			return common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   siatro.ID,
				TaskType:   "поиск информации",
				TaskAction: "задача отклонена",
				Message:    "не задано начальное время для поиска информации",
			}), false
		}
		if dtp.Start != 0 && dtp.End == 0 {
			return common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   siatro.ID,
				TaskType:   "поиск информации",
				TaskAction: "задача отклонена",
				Message:    "не задано конечное время для поиска информации",
			}), false
		}
		if dtp.Start > dtp.End {
			return common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   siatro.ID,
				TaskType:   "поиск информации",
				TaskAction: "задача отклонена",
				Message:    "начальное время, для поиска информации, не должно быть больше конечного",
			}), false
		}

		return "", true
	}

	//проверяем временной интервал
	if msgInfo, ok := checkDateTimeFiltering(siatro.InstalledFilteringOption.DateTime); !ok {
		return msgInfo, false
	}

	//проверяем тип сетевого протокола
	if isCorrectProtocol := checkCorrectProtocol(siatro.InstalledFilteringOption.Protocol); !isCorrectProtocol {
		siatro.InstalledFilteringOption.Protocol = "any"
	}

	//проверяем статус задачи по фильтрации
	if isCorrectStatus := checkCorrectStatusTask(siatro.StatusFilteringTask); !isCorrectStatus {
		siatro.StatusFilteringTask = "any"
	}

	//проверяем статус задачи по скачиванию файлов
	if isCorrectStatus := checkCorrectStatusTask(siatro.StatusFileDownloadTask); !isCorrectStatus {
		siatro.StatusFileDownloadTask = "any"
	}

	filterParameters := map[string]map[string]*[]string{
		"IP": map[string]*[]string{
			"Any": &siatro.InstalledFilteringOption.NetworkFilters.IP.Any,
			"Src": &siatro.InstalledFilteringOption.NetworkFilters.IP.Src,
			"Dst": &siatro.InstalledFilteringOption.NetworkFilters.IP.Dst,
		},
		"Port": map[string]*[]string{
			"Any": &siatro.InstalledFilteringOption.NetworkFilters.Port.Any,
			"Src": &siatro.InstalledFilteringOption.NetworkFilters.Port.Src,
			"Dst": &siatro.InstalledFilteringOption.NetworkFilters.Port.Dst,
		},
		"Network": map[string]*[]string{
			"Any": &siatro.InstalledFilteringOption.NetworkFilters.Network.Any,
			"Src": &siatro.InstalledFilteringOption.NetworkFilters.Network.Src,
			"Dst": &siatro.InstalledFilteringOption.NetworkFilters.Network.Dst,
		},
	}

	//проверяем параметры сетевых фильтров
	if err := LoopHandler(LoopHandlerParameters{
		OptionsCheckFilterParameters{
			SourceID: siatro.ID,
			TaskType: "поиск информации",
		},
		filterParameters,
		CheckIPPortNetwork,
	}); err != nil {
		return fmt.Sprint(err), false
	}

	return "", true
}

//OptionsCheckFilterParameters общие опции
type OptionsCheckFilterParameters struct {
	SourceID int
	TaskType string
}

//LoopHandlerParameters параметра обработчика цикла
type LoopHandlerParameters struct {
	OptionsCheckFilterParameters
	FilterParameters map[string]map[string]*[]string
	Function         func(CheckIPPortNetworkParameters) error
}

//CheckIPPortNetworkParameters параметры для функции поиска по сетевым фильтрам
type CheckIPPortNetworkParameters struct {
	OptionsCheckFilterParameters
	paramType string
	paramList *[]string
}

//LoopHandler обработчик циклов
func LoopHandler(lhp LoopHandlerParameters) error {
	for paramType, paramValue := range lhp.FilterParameters {
		for _, paramList := range paramValue {
			if err := lhp.Function(CheckIPPortNetworkParameters{
				OptionsCheckFilterParameters{
					SourceID: lhp.SourceID,
					TaskType: lhp.TaskType,
				},
				paramType,
				paramList,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

//CheckIPPortNetwork функция проверки сетевых фильтров
func CheckIPPortNetwork(checkParameters CheckIPPortNetworkParameters) error {
	checkIP := func(item string) bool {
		ok, _ := common.CheckStringIP(item)

		return ok
	}

	checkPort := func(item string) bool {
		p, err := strconv.Atoi(item)
		if err != nil {
			return false
		}

		if p == 0 || p > 65536 {
			return false
		}

		return true
	}

	checkNetwork := func(item string) bool {
		ok, _ := common.CheckStringNetwork(item)

		return ok
	}

	iteration := func(param *[]string, f func(string) bool) bool {
		if len(*param) == 0 {
			return true
		}

		for _, v := range *param {
			if ok := f(v); !ok {
				return false
			}
		}

		return true
	}

	switch checkParameters.paramType {
	case "IP":
		if ok := iteration(checkParameters.paramList, checkIP); !ok {
			return errors.New(common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   checkParameters.SourceID,
				TaskType:   checkParameters.TaskType,
				TaskAction: "задача отклонена",
				Message:    "неверные параметры фильтрации, один или более переданных пользователем IP адресов имеет некорректное значение",
			}))
		}

	case "Port":
		if ok := iteration(checkParameters.paramList, checkPort); !ok {
			return errors.New(common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   checkParameters.SourceID,
				TaskType:   checkParameters.TaskType,
				TaskAction: "задача отклонена",
				Message:    "неверные параметры фильтрации, один или более из заданных пользователем портов имеет некорректное значение",
			}))
		}

	case "Network":
		if ok := iteration(checkParameters.paramList, checkNetwork); !ok {
			return errors.New(common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   checkParameters.SourceID,
				TaskType:   checkParameters.TaskType,
				TaskAction: "задача отклонена",
				Message:    "неверные параметры фильтрации, получен неверный сетевой диапазон",
			}))
		}

	}

	return nil
}

func checkCorrectStatusTask(status string) bool {
	var isCorrectStatus bool
	listStatusName := []string{"wait", "refused", "execute", "not executed", "complete", "stop"}

	for _, s := range listStatusName {
		if status == s {
			isCorrectStatus = true

			break
		}
	}

	return isCorrectStatus
}

func checkCorrectProtocol(proto string) bool {
	var isCorrectProtocol bool
	listProto := []string{"any", "tcp", "udp"}

	for _, p := range listProto {
		if proto == p {
			isCorrectProtocol = true

			break
		}
	}

	return isCorrectProtocol
}

func checkNetworkParametersIsNotEmpty(np map[string]map[string]*[]string) bool {
	for _, v := range np {
		for _, item := range v {
			if len(*item) != 0 {
				return true
			}
		}
	}

	return false
}

func checkValidtaskID(taskID string) bool {
	rx := regexp.MustCompile(`^(\w)+$`)
	ok := rx.MatchString(taskID)

	return ok
}
