package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//getConfirmActionSourceListForAPI формирует список источников с выполненными
//над ними действиями и статусом успешности
func getConfirmActionSourceListForAPI(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	res *configure.MsgBetweenCoreAndNI,
	smt *configure.StoringMemoryTask) {

	fmt.Println("START function 'getConfirmActionSourceListForAPI'")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'getConfirmActionSourceListForAPI'"

	listSource, ok := res.AdvancedOptions.(*[]configure.ActionTypeListSources)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "type conversion error section type 'error notification'"+funcName)

		return
	}

	//получаем ID клиента API
	st, ok := smt.GetStoringMemoryTask(res.TaskID)
	if !ok {
		_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with %v not found", res.TaskID))

		return
	}

	msg := configure.SourceControlConfirmActionSource{
		MsgOptions: configure.SourceControlMsgTypeToAPI{
			TaskInfo: configure.MsgTaskInfo{
				State: "end",
			},
			SourceList: *listSource,
		},
	}
	msg.MsgType = "information"
	msg.MsgSection = "source control"
	msg.MsgInsturction = "confirm the action"
	msg.ClientTaskID = st.ClientTaskID

	msgjson, _ := json.Marshal(&msg)

	if err := senderMsgToAPI(chanToAPI, smt, res.TaskID, st.ClientID, msgjson); err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
	}
}

//sendChanStatusSourceForAPI формирование информационного сообщения
//об изменении статуса соединения источника
func sendChanStatusSourceForAPI(chanToAPI chan<- *configure.MsgBetweenCoreAndAPI, res *configure.MsgBetweenCoreAndNI) {
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'sendChanStatusSourceForAPI'"

	s, ok := res.AdvancedOptions.(configure.SettingsChangeConnectionStatusSource)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "type conversion error section type 'error notification'"+funcName)

		return
	}

	sl := []configure.ActionTypeListSources{
		configure.ActionTypeListSources{
			ID:         s.ID,
			Status:     s.Status,
			ActionType: "none",
			IsSuccess:  true,
		},
	}

	msg := configure.SourceControlActionsTakenSources{
		MsgOptions: configure.SourceControlMsgTypeToAPI{
			SourceList: sl,
		},
	}

	msg.MsgType = "information"
	msg.MsgSection = "source control"
	msg.MsgInsturction = "change status source"

	msgjson, _ := json.Marshal(&msg)

	//отправляем данные клиенту
	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  "",
		MsgJSON:      msgjson,
	}
}

//sendInformationFiltrationTask отправляет информационное сообщение о ходе фильтрации
func sendInformationFiltrationTask(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	taskInfo *configure.TaskDescription,
	msg *configure.MsgBetweenCoreAndNI) {

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	ti := taskInfo.TaskParameter.FiltrationTask
	resMsg := configure.FiltrationControlTypeInfo{
		MsgOption: configure.FiltrationControlMsgTypeInfo{
			ID:                              msg.SourceID,
			Status:                          ti.Status,
			NumberFilesMeetFilterParameters: ti.NumberFilesMeetFilterParameters,
			NumberProcessedFiles:            ti.NumberProcessedFiles,
			NumberFilesFoundResultFiltering: ti.NumberFilesFoundResultFiltering,
			NumberErrorProcessedFiles:       ti.NumberErrorProcessedFiles,
			NumberDirectoryFiltartion:       ti.NumberDirectoryFiltartion,
			SizeFilesMeetFilterParameters:   ti.SizeFilesMeetFilterParameters,
			SizeFilesFoundResultFiltering:   ti.SizeFilesFoundResultFiltering,
			PathStorageSource:               ti.PathStorageSource,
		},
	}

	resMsg.MsgType = "information"
	resMsg.MsgSection = "filtration control"
	resMsg.MsgInsturction = "task processing"
	resMsg.ClientTaskID = taskInfo.ClientTaskID

	if ti.Status == "execute" {
		nffi := make(map[string]*configure.InputFilesInformation, len(ti.FoundFilesInformation))
		for n, v := range ti.FoundFilesInformation {
			nffi[n] = &configure.InputFilesInformation{
				Size: v.Size,
				Hex:  v.Hex,
			}
		}

		resMsg.MsgOption.FoundFilesInformation = nffi
	}

	msgJSON, err := json.Marshal(resMsg)
	if err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

		return
	}

	if ti.Status == "complite" {
		fmt.Println("-+-++--+--+-+ SEND FILTRATION MSG 'COMPLITE'")
	}

	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  taskInfo.ClientID,
		MsgJSON:      msgJSON,
	}
}
