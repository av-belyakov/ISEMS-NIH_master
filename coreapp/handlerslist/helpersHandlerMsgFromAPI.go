package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//checkParametersFiltration проверяет параметры фильтрации
func сheckParametersFiltration(fccpf *configure.FiltrationControlCommonParametersFiltration) (string, bool) {
	//проверяем наличие ID источника
	if fccpf.ID == 0 {
		return "отсутствует идентификатор источника", false
	}

	//проверяем временной интервал
	isZero := ((fccpf.DateTime.Start == 0) || (fccpf.DateTime.End == 0))
	if isZero || (fccpf.DateTime.Start > fccpf.DateTime.End) {
		return "задан неверный временной интервал", false
	}

	//проверяем тип протокола
	if strings.EqualFold(fccpf.Protocol, "") {
		fccpf.Protocol = "any"
	}

	isProtoTCP := strings.EqualFold(fccpf.Protocol, "tcp")
	isProtoUDP := strings.EqualFold(fccpf.Protocol, "udp")
	isProtoANY := strings.EqualFold(fccpf.Protocol, "any")

	if !isProtoTCP && !isProtoUDP && !isProtoANY {
		return "задан неверный идентификатор транспортного протокола", false
	}

	isEmpty := true

	circle := func(fp map[string]map[string]*[]string, f func(string, *[]string) error) error {
		for pn, pv := range fp {
			var err error

			for _, v := range pv {
				if err = f(pn, v); err != nil {
					return err
				}
			}

			if err != nil {
				return err
			}
		}

		return nil
	}

	checkIPOrPortOrNetwork := func(paramType string, param *[]string) error {
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
				return errors.New("неверные параметры фильтрации, один или более переданных пользователем IP адресов имеет некорректное значение")
			}

		case "Port":
			if ok := iteration(param, checkPort); !ok {
				return errors.New("неверные параметры фильтрации, один или более из заданных пользователем портов имеет некорректное значение")
			}

		case "Network":
			if ok := iteration(param, checkNetwork); !ok {
				return errors.New("неверные параметры фильтрации, некорректное значение маски подсети заданное пользователем")
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
	if err := circle(filterParameters, checkIPOrPortOrNetwork); err != nil {
		return fmt.Sprint(err), false
	}

	//проверяем параметры свойства 'Filters' на пустоту
	if isEmpty {
		return "невозможно начать фильтрацию, необходимо указать хотя бы один искомый ip адрес, порт или подсеть", false
	}

	return "", true
}

//handlerFiltrationControlTypeStart обработчик запроса на фильтрацию
func handlerFiltrationControlTypeStart(
	chanToDB chan<- *configure.MsgBetweenCoreAndDB,
	fcts *configure.FiltrationControlTypeStart,
	hsm HandlersStoringMemory,
	clientID string,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) {

	funcName := ", function 'handlerFiltrationControlTypeStart'"

	if msg, ok := сheckParametersFiltration(&fcts.MsgOption); !ok {
		_ = saveMessageApp.LogMessage("error", "incorrect parameters for filtering are set"+funcName)

		//отправляем информационное сообщение
		notifications.SendNotificationToClientAPI(
			chanToAPI,
			notifications.NotificationSettingsToClientAPI{
				MsgType:        "danger",
				MsgDescription: msg,
			},
			fcts.ClientTaskID,
			clientID)

		//отправляем сообщение о том что задача была отклонена
		resMsg := configure.FiltrationControlTypeInfo{
			MsgOption: configure.FiltrationControlMsgTypeInfo{
				ID:     fcts.MsgOption.ID,
				Status: "refused",
			},
		}
		resMsg.MsgType = "information"
		resMsg.MsgSection = "filtration control"
		resMsg.MsgInsturction = "task processing"
		resMsg.ClientTaskID = fcts.ClientTaskID

		msgJSON, err := json.Marshal(resMsg)
		if err != nil {
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

			return
		}

		chanToAPI <- &configure.MsgBetweenCoreAndAPI{
			MsgGenerator: "Core module",
			MsgRecipient: "API module",
			IDClientAPI:  clientID,
			MsgJSON:      msgJSON,
		}

		return
	}

	/*
			   После разработки и апробации StoringMemoryQueueTask
			   в место
		<<<<<<< HEAD
			   taskID := smt.AddStoringMemoryTask(configure.TaskDescription...
		=======
			   smt.AddStoringMemoryTask(configure.TaskDescription...
		>>>>>>> ISEMS-NIH_master 06.08.2019
			   	chanToDB <- &configure.MsgBetweenCoreAndDB...

			   	нужно добавить

			   	qts.AddQueueTaskStorage(
			   		fcts.MsgOption.ID, //это sourceID
			   		configure.CommonTaskInfo{
			   			IDClientAPI:     clientID,
			   			TaskIDClientAPI: fcts.ClientTaskID,
			   			TaskType:        "filtration",
			   		},
			   		&configure.DescriptionParametersReceivedFromUser{
			   			FiltrationParameters: FiltrationOptions,
			   		})
	*/

	taskID := common.GetUniqIDFormatMD5(clientID)

	//добавляем новую задачу
	hsm.SMT.AddStoringMemoryTask(taskID, configure.TaskDescription{
		ClientID:                        clientID,
		ClientTaskID:                    fcts.ClientTaskID,
		TaskType:                        fcts.MsgSection,
		ModuleThatSetTask:               "API module",
		ModuleResponsibleImplementation: "NI module",
		TimeUpdate:                      time.Now().Unix(),
		TimeInterval: configure.TimeIntervalTaskExecution{
			Start: time.Now().Unix(),
			End:   time.Now().Unix(),
		},
		TaskParameter: configure.DescriptionTaskParameters{
			FiltrationTask: configure.FiltrationTaskParameters{
				ID:     fcts.MsgOption.ID,
				Status: "wait",
			},
		},
	})

	//сохранение параметров задачи в БД
	chanToDB <- &configure.MsgBetweenCoreAndDB{
		MsgGenerator:    "Core module",
		MsgRecipient:    "DB module",
		MsgSection:      "filtration control",
		Instruction:     "insert",
		IDClientAPI:     clientID,
		TaskID:          taskID,
		TaskIDClientAPI: fcts.ClientTaskID,
		AdvancedOptions: fcts.MsgOption,
	}
}
