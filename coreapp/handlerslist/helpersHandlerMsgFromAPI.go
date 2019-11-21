package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

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

	//проверяем тип протокола
	if strings.EqualFold(fccpf.Protocol, "") {
		fccpf.Protocol = "any"
	}

	isProtoTCP := strings.EqualFold(fccpf.Protocol, "tcp")
	isProtoUDP := strings.EqualFold(fccpf.Protocol, "udp")
	isProtoANY := strings.EqualFold(fccpf.Protocol, "any")

	if !isProtoTCP && !isProtoUDP && !isProtoANY {
		return common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID:   fccpf.ID,
			TaskType:   "фильтрация",
			TaskAction: "задача отклонена",
			Message:    "задан неверный идентификатор транспортного протокола",
		}), false
	}

	isEmpty := true

	circle := func(sID int, fp map[string]map[string]*[]string, f func(int, string, *[]string) error) error {
		for pn, pv := range fp {
			var err error

			for _, v := range pv {
				if err = f(sID, pn, v); err != nil {
					return err
				}
			}

			if err != nil {
				return err
			}
		}

		return nil
	}

	checkIPOrPortOrNetwork := func(sourceID int, paramType string, param *[]string) error {
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

			isEmpty = false

			for _, v := range *param {
				if ok := f(v); !ok {
					return false
				}
			}

			return true
		}

		switch paramType {
		case "IP":
			if ok := iteration(param, checkIP); !ok {
				return errors.New(common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   sourceID,
					TaskType:   "фильтрация",
					TaskAction: "задача отклонена",
					Message:    "неверные параметры фильтрации, один или более переданных пользователем IP адресов имеет некорректное значение",
				}))
			}

		case "Port":
			if ok := iteration(param, checkPort); !ok {
				return errors.New(common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   sourceID,
					TaskType:   "фильтрация",
					TaskAction: "задача отклонена",
					Message:    "неверные параметры фильтрации, один или более из заданных пользователем портов имеет некорректное значение",
				}))
			}

		case "Network":
			if ok := iteration(param, checkNetwork); !ok {
				return errors.New(common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   sourceID,
					TaskType:   "фильтрация",
					TaskAction: "задача отклонена",
					Message:    "неверные параметры фильтрации, некорректное значение маски подсети заданное пользователем",
				}))
			}

		}

		return nil
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

	//проверка ip адресов, портов и подсетей
	if err := circle(fccpf.ID, filterParameters, checkIPOrPortOrNetwork); err != nil {
		return fmt.Sprint(err), false
	}

	//проверяем параметры свойства 'Filters' на пустоту
	if isEmpty {
		return common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID:   fccpf.ID,
			TaskType:   "фильтрация",
			TaskAction: "задача отклонена",
			Message:    "необходимо указать хотя бы один искомый ip адрес, порт или подсеть",
		}), false
	}

	return "", true
}

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
		_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})

		return
	}

	fmt.Println("function 'handlerFiltrationControlTypeStart' - decode JSON")

	//проверяем параметры фильтрации
	if msg, ok := сheckParametersFiltration(&fcts.MsgOption); !ok {
		_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
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

	fmt.Println("function 'handlerFiltrationControlTypeStart' - check parameters filtration")

	//проверяем состояние подключения источника
	connectionStatus, err := hsm.ISL.GetSourceConnectionStatus(fcts.MsgOption.ID)
	if err != nil || !connectionStatus {
		_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
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

	fmt.Println("function 'handlerFiltrationControlTypeStart' - check connection status")

	//получаем новый идентификатор задачи
	taskID := common.GetUniqIDFormatMD5(clientID + "_" + fcts.ClientTaskID)

	fmt.Printf("func 'handlerMsgFromAPI', ADD NEW TASK, ClientTaskID: '%v', APP TASK ID: '%v'\n", fcts.ClientTaskID, taskID)
	fmt.Printf("function 'handlerFiltrationControlTypeStart' - add task QueueTaskStorage with ID:'%v'\n", taskID)

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
		_ = saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
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
