package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"
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

	//сообщение о том что задача была отклонена
	resMsg := configure.FiltrationControlTypeInfo{
		MsgOption: configure.FiltrationControlMsgTypeInfo{
			ID:     fcts.MsgOption.ID,
			Status: "refused",
		},
	}
	resMsg.MsgType = "information"
	resMsg.MsgSection = "filtration control"
	resMsg.MsgInstruction = "task processing"
	resMsg.ClientTaskID = fcts.ClientTaskID

	msgJSON, err := json.Marshal(resMsg)
	if err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})

		return
	}

	//проверяем параметры фильтрации
	if msg, ok := сheckParametersFiltration(&fcts.MsgOption); !ok {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "incorrect parameters for filtering are set",
			FuncName:    funcName,
		})

		//отправляем информационное сообщение
		notifications.SendNotificationToClientAPI(
			chanToAPI,
			notifications.NotificationSettingsToClientAPI{
				MsgType:        "danger",
				MsgDescription: msg,
				Sources:        []int{fcts.MsgOption.ID},
			},
			fcts.ClientTaskID,
			clientID)

		//отправляем сообщение что задача была отклонена
		chanToAPI <- &configure.MsgBetweenCoreAndAPI{
			MsgGenerator: "Core module",
			MsgRecipient: "API module",
			IDClientAPI:  clientID,
			MsgJSON:      msgJSON,
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

		//отправляем информационное сообщение
		notifications.SendNotificationToClientAPI(
			chanToAPI,
			notifications.NotificationSettingsToClientAPI{
				MsgType: "danger",
				MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   fcts.MsgOption.ID,
					TaskType:   "фильтрация",
					TaskAction: "задача отклонена",
					Message:    "источник не подключен",
				}),
			},
			fcts.ClientTaskID,
			clientID)

		//отправляем сообщение что задача была отклонена
		chanToAPI <- &configure.MsgBetweenCoreAndAPI{
			MsgGenerator: "Core module",
			MsgRecipient: "API module",
			IDClientAPI:  clientID,
			MsgJSON:      msgJSON,
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
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) {

	funcName := "handlerInformationSearchControlTypeSearchCommanInformation"

	//сообщение о том что задача была отклонена
	resMsg := configure.SearchInformationResponseCommanInfo{MsgOption: configure.SearchInformationResponseOptionCommanInfo{Status: "refused"}}
	resMsg.MsgType = "information"
	resMsg.MsgSection = "information search control"
	resMsg.MsgInstruction = "task processing"
	resMsg.ClientTaskID = siatr.ClientTaskID

	msgJSON, err := json.Marshal(resMsg)
	if err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})

		return
	}

	//проверяем параметры необходимые для поиска общей информации по задачам
	if msg, ok := CheckParametersSearchCommonInformation(&siatr.MsgOption); !ok {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "incorrect search parameters are set",
			FuncName:    funcName,
		})

		//отправляем информационное сообщение
		notifications.SendNotificationToClientAPI(
			chanToAPI,
			notifications.NotificationSettingsToClientAPI{
				MsgType:        "danger",
				MsgDescription: msg,
			},
			siatr.ClientTaskID,
			clientID)

		//отправляем сообщение что задача была отклонена
		chanToAPI <- &configure.MsgBetweenCoreAndAPI{
			MsgGenerator: "Core module",
			MsgRecipient: "API module",
			IDClientAPI:  clientID,
			MsgJSON:      msgJSON,
		}

		return
	}

	/*
		!!!
			добавляем задачу в очередь (или не использовать очередь а сразу отправлять в БД)
		!!!
	*/
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
				Message:    "начальное время для поиска информации не должно быть больше конечного времени",
			}), false
		}

		return "", true
	}

	//проверяем максимальное кол-во задач в возвращаемой части
	if (siatro.NumberTasksReturnedPart == 0) || (siatro.NumberTasksReturnedPart > 101) {
		siatro.NumberTasksReturnedPart = 35
	}

	//проверяем временной интервал
	if msgInfo, ok := checkDateTimeFiltering(siatro.InstalledFilteringOption.DateTime); !ok {
		return msgInfo, false
	}

	//проверяем тип сетевого протокола
	if isCorrectProtocol := checkCorrectProtocol(siatro.InstalledFilteringOption.Protocol); !isCorrectProtocol {
		siatro.InstalledFilteringOption.Protocol = "any"
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
