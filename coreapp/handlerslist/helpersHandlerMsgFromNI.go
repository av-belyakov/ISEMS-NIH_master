package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

type handlerDownloadTaskStatusCompleteType struct {
	SourceID       int
	TaskID         string
	ClientID       string
	ClientTaskID   string
	QTS            *configure.QueueTaskStorage
	SMT            *configure.StoringMemoryTask
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
	clientID, clientTaskID string) error {

	funcName := ", function 'getConfirmActionSourceListForAPI'"

	listSource, ok := res.AdvancedOptions.(*[]configure.ActionTypeListSources)
	if !ok {
		return fmt.Errorf("type conversion error section type 'error notification'%v", funcName)
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
	msg.ClientTaskID = clientTaskID

	msgjson, err := json.Marshal(&msg)
	if err != nil {
		return err
	}

	//отправляем данные клиенту
	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  clientID,
		MsgJSON:      msgjson,
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
		(configure.ActionTypeListSources{
			ID:         s.ID,
			Status:     s.Status,
			ActionType: "none",
			IsSuccess:  true,
		}),
	}

	msg := configure.SourceControlActionsTakenSources{
		MsgOptions: configure.SourceControlMsgTypeToAPI{
			SourceList: sl,
		},
	}
	msg.MsgType = "information"
	msg.MsgSection = "source control"
	msg.MsgInstruction = "change status source"

	msgjson, err := json.Marshal(&msg)
	if err != nil {
		return err
	}

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
	smt *configure.StoringMemoryTask,
	taskInfo *configure.TaskDescription,
	ldfi map[string]*configure.DetailedFilesInformation,
	sourceID int,
	taskID string) error {

	//fmt.Println("func 'sendInformationFiltrationTask', send to client API ---> --->")

	st, ok := smt.GetTaskStatusStoringMemoryTask(taskID, "filtration")
	if !ok {
		return fmt.Errorf("func 'sendInformationFiltrationTask', task with ID '%v' not found", taskID)
	}

	ti := taskInfo.TaskParameter.FiltrationTask
	resMsg := configure.FiltrationControlTypeInfo{
		MsgOption: configure.FiltrationControlMsgTypeInfo{
			ID:                              sourceID,
			TaskIDApp:                       taskID,
			Status:                          st.Status,
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

	nffi := make(map[string]*configure.InputFilesInformation, len(ldfi))

	if st.Status == "execute" {
		for n, v := range ldfi {
			nffi[n] = &configure.InputFilesInformation{
				Size: v.Size,
				Hex:  v.Hex,
			}
		}
	}
	resMsg.MsgOption.FoundFilesInformation = nffi

	msgJSON, err := json.Marshal(&resMsg)
	if err != nil {
		return err
	}

	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  taskInfo.ClientID,
		MsgJSON:      msgJSON,
	}

	if (st.Status == "complete") || (st.Status == "stop") {
		ns := notifications.NotificationSettingsToClientAPI{
			MsgType: "success",
			MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   ti.ID,
				TaskType:   "фильтрация",
				TaskAction: "задача успешно выполнена",
			}),
			Sources: []int{ti.ID},
		}

		if st.Status == "stop" {
			ns.MsgDescription = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   ti.ID,
				TaskType:   "фильтрация",
				TaskAction: "задача успешно остановлена",
			})
		}

		notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, taskInfo.ClientID)
	}

	return nil
}

func handlerDownloadTaskStatusComplete(hdtsct handlerDownloadTaskStatusCompleteType) error {
	//обновление статуса задачи по скачиванию файлов
	if err := setStatusCompleteDownloadTask(hdtsct.TaskID, hdtsct.SMT); err != nil {
		return err
	}

	//устанавливаем общий статус задачи в StoringMemoryTask как завершенный (то есть TRUE)
	hdtsct.SMT.CompleteStoringMemoryTask(hdtsct.TaskID)

	//записываем информацию в БД
	hdtsct.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
		MsgGenerator:    "NI module",
		MsgRecipient:    "DB module",
		MsgSection:      "download control",
		Instruction:     "update",
		TaskID:          hdtsct.TaskID,
		AdvancedOptions: "task complete",
	}

	//отправляем информацию по задаче клиенту API
	msgJSONInfo, err := json.Marshal(&hdtsct.ResMsgInfo)
	if err != nil {
		return err
	}

	hdtsct.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  hdtsct.ClientID,
		MsgJSON:      msgJSONInfo,
	}

	//отправляем информационное сообщение клиенту API
	notifications.SendNotificationToClientAPI(hdtsct.OutCoreChanAPI, hdtsct.NS, hdtsct.ClientTaskID, hdtsct.ClientID)

	//изменяем статус задачи в storingMemoryQueueTask на 'complete'
	//(ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА функцией 'CheckTimeQueueTaskStorage')
	if err := hdtsct.QTS.ChangeTaskStatusQueueTask(hdtsct.SourceID, hdtsct.TaskID, "complete"); err != nil {
		return err
	}

	return nil
}

func setStatusCompleteDownloadTask(taskID string, smt *configure.StoringMemoryTask) error {
	ti, ok := smt.GetStoringMemoryTask(taskID)
	if !ok {
		return fmt.Errorf("task with ID %v not found", taskID)
	}

	ti.TaskParameter.DownloadTask.Status = "complete"
	//обновление статуса задачи
	smt.UpdateTaskDownloadAllParameters(taskID, ti.TaskParameter.DownloadTask)

	return nil
}
