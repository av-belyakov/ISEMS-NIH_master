package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

type handlerDownloadTaskStatusCompleteType struct {
	SourceID       int
	TaskID         string
	TI             *configure.TaskDescription
	QTS            *configure.QueueTaskStorage
	NS             notifications.NotificationSettingsToClientAPI
	ResMsgInfo     configure.DownloadControlTypeInfo
	OutCoreChanAPI chan<- *configure.MsgBetweenCoreAndAPI
	OutCoreChanDB  chan<- *configure.MsgBetweenCoreAndDB
}

//getConfirmActionSourceListForAPI формирует список источников с выполненными
//над ними действиями и статусом успешности
func getConfirmActionSourceListForAPI(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	res *configure.MsgBetweenCoreAndNI,
	smt *configure.StoringMemoryTask) error {

	funcName := ", function 'getConfirmActionSourceListForAPI'"

	listSource, ok := res.AdvancedOptions.(*[]configure.ActionTypeListSources)
	if !ok {
		return fmt.Errorf("type conversion error section type 'error notification'%v", funcName)
	}

	//получаем ID клиента API
	st, ok := smt.GetStoringMemoryTask(res.TaskID)
	if !ok {
		return fmt.Errorf("task with %v not found", res.TaskID)
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
	msg.MsgInstruction = "confirm the action"
	msg.ClientTaskID = st.ClientTaskID

	msgjson, _ := json.Marshal(&msg)

	if err := senderMsgToAPI(chanToAPI, smt, res.TaskID, st.ClientID, msgjson); err != nil {
		return err
	}

	return nil
}

//sendChanStatusSourceForAPI формирование информационного сообщения
//об изменении статуса соединения источника
func sendChanStatusSourceForAPI(chanToAPI chan<- *configure.MsgBetweenCoreAndAPI, res *configure.MsgBetweenCoreAndNI) error {
	funcName := ", function 'sendChanStatusSourceForAPI'"

	s, ok := res.AdvancedOptions.(configure.SettingsChangeConnectionStatusSource)
	if !ok {
		return fmt.Errorf("type conversion error section type 'error notification'%v", funcName)
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
	msg.MsgInstruction = "change status source"

	msgjson, _ := json.Marshal(&msg)

	//отправляем данные клиенту
	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  "",
		MsgJSON:      msgjson,
	}

	return nil
}

//sendInformationFiltrationTask отправляет информационное сообщение о ходе фильтрации
func sendInformationFiltrationTask(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	taskInfo *configure.TaskDescription,
	msg *configure.MsgBetweenCoreAndNI) error {

	ti := taskInfo.TaskParameter.FiltrationTask
	resMsg := configure.FiltrationControlTypeInfo{
		MsgOption: configure.FiltrationControlMsgTypeInfo{
			ID:                              msg.SourceID,
			TaskIDApp:                       msg.TaskID,
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
	resMsg.MsgInstruction = "task processing"
	resMsg.ClientTaskID = taskInfo.ClientTaskID

	if ti.Status == "execute" {
		if tfmffiats, ok := msg.AdvancedOptions.(configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus); ok {
			nffi := make(map[string]*configure.InputFilesInformation, len(tfmffiats.ListFoundFile))
			for n, v := range tfmffiats.ListFoundFile {
				nffi[n] = &configure.InputFilesInformation{
					Size: v.Size,
					Hex:  v.Hex,
				}
			}

			resMsg.MsgOption.FoundFilesInformation = nffi
		}
	}

	msgJSON, err := json.Marshal(resMsg)
	if err != nil {
		return err
	}

	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  taskInfo.ClientID,
		MsgJSON:      msgJSON,
	}

	if ti.Status == "complete" || ti.Status == "stop" {
		ns := notifications.NotificationSettingsToClientAPI{
			MsgType:        "success",
			MsgDescription: fmt.Sprintf("Задача по фильтрации сетевого трафика на источнике с ID %v, успешно завершена", ti.ID),
			Sources:        []int{ti.ID},
		}

		if ti.Status == "stop" {
			ns.MsgDescription = fmt.Sprintf("задача по фильтрации сетевого трафика на источнике с ID %v, была успешно остановлена", ti.ID)
		}

		notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)
	}

	return nil
}

func handlerDownloadTaskStatusComplete(hdtsct handlerDownloadTaskStatusCompleteType) error {
	//записываем информацию в БД
	hdtsct.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
		MsgGenerator: "NI module",
		MsgRecipient: "DB module",
		MsgSection:   "download control",
		Instruction:  "update",
		TaskID:       hdtsct.TaskID,
	}

	//отправляем информационное сообщение клиенту API
	notifications.SendNotificationToClientAPI(hdtsct.OutCoreChanAPI, hdtsct.NS, hdtsct.TaskID, hdtsct.TI.ClientID)

	//отправляем информацию по задаче клиенту API
	msgJSONInfo, err := json.Marshal(hdtsct.ResMsgInfo)
	if err != nil {
		return err
	}
	hdtsct.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  hdtsct.TI.ClientID,
		MsgJSON:      msgJSONInfo,
	}

	//изменяем статус задачи в storingMemoryQueueTask
	// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
	// функцией 'CheckTimeQueueTaskStorage')
	if err := hdtsct.QTS.ChangeTaskStatusQueueTask(hdtsct.SourceID, hdtsct.TaskID, "complete"); err != nil {
		return err
	}

	return nil
}
