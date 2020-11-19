package processresponse

import (
	"encoding/json"

	"ISEMS-NIH_master/configure"
)

//ParametersProcessingReceivedMsgTypeFiltering набор параметров для функции ProcessingReceivedMsgTypeFiltering
type ParametersProcessingReceivedMsgTypeFiltering struct {
	Chans    ChansMsgTypeFiltering
	SMT      *configure.StoringMemoryTask
	Message  *[]byte
	SourceID int
	SourceIP string
}

//ChansMsgTypeFiltering набор каналов
type ChansMsgTypeFiltering struct {
	CwtRes     chan<- configure.MsgWsTransmission
	ChanInCore chan<- *configure.MsgBetweenCoreAndNI
	CwtReq     <-chan configure.MsgWsTransmission
}

//ProcessingReceivedMsgTypeFiltering обработка сообщений связанных с фильтрацией файлов
func ProcessingReceivedMsgTypeFiltering(pprmtf ParametersProcessingReceivedMsgTypeFiltering) error {
	resMsg := configure.MsgTypeFiltration{}

	if err := json.Unmarshal(*pprmtf.Message, &resMsg); err != nil {
		return err
	}

	ftp := configure.FiltrationTaskParameters{
		Status:                          resMsg.Info.TaskStatus,
		NumberFilesMeetFilterParameters: resMsg.Info.NumberFilesMeetFilterParameters,
		NumberProcessedFiles:            resMsg.Info.NumberProcessedFiles,
		NumberFilesFoundResultFiltering: resMsg.Info.NumberFilesFoundResultFiltering,
		NumberDirectoryFiltartion:       resMsg.Info.NumberDirectoryFiltartion,
		NumberErrorProcessedFiles:       resMsg.Info.NumberErrorProcessedFiles,
		SizeFilesMeetFilterParameters:   resMsg.Info.SizeFilesMeetFilterParameters,
		SizeFilesFoundResultFiltering:   resMsg.Info.SizeFilesFoundResultFiltering,
		PathStorageSource:               resMsg.Info.PathStorageSource,
	}

	//обновляем информацию о выполняемой задаче в памяти приложения
	pprmtf.SMT.UpdateTaskFiltrationAllParameters(resMsg.Info.TaskID, &ftp)
	ftp = configure.FiltrationTaskParameters{}

	lfdi := make(map[string]*configure.DetailedFilesInformation, len(resMsg.Info.FoundFilesInformation))
	for n, v := range resMsg.Info.FoundFilesInformation {
		lfdi[n] = &configure.DetailedFilesInformation{
			Size: v.Size,
			Hex:  v.Hex,
		}
	}

	pprmtf.SMT.UpdateListFilesDetailedInformation(resMsg.Info.TaskID, lfdi)

	msgCompleteTask := configure.MsgBetweenCoreAndNI{
		TaskID:  resMsg.Info.TaskID,
		Section: "monitoring task performance",
		Command: "complete task",
	}

	msg := &configure.MsgBetweenCoreAndNI{
		TaskID:   resMsg.Info.TaskID,
		Section:  "filtration control",
		Command:  resMsg.Info.TaskStatus,
		SourceID: pprmtf.SourceID,
		AdvancedOptions: configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus{
			TaskStatus:    resMsg.Info.TaskStatus,
			ListFoundFile: lfdi,
		},
	}

	if resMsg.Info.TaskStatus == "execute" {
		//отправляем в ядро, а от туда в БД и клиенту API
		pprmtf.Chans.ChanInCore <- msg

		return nil
	}

	if resMsg.Info.TaskStatus == "refused" {
		//отправляем в ядро, а от туда в БД и клиенту API
		pprmtf.Chans.ChanInCore <- msg

		//отправляем сообщение о снятии контроля за выполнением задачи
		pprmtf.Chans.ChanInCore <- &msgCompleteTask

		return nil
	}

	//если тип сообщения "stop" или "complete"
	// отправка информации только после получения всех частей
	if resMsg.Info.NumberMessagesParts[0] == resMsg.Info.NumberMessagesParts[1] {
		//отправляем в ядро, а от туда в БД и клиенту API
		pprmtf.Chans.ChanInCore <- msg

		msgJSON, err := json.Marshal(configure.MsgTypeFiltrationControl{
			MsgType: "filtration",
			Info: configure.SettingsFiltrationControl{
				TaskID:  resMsg.Info.TaskID,
				Command: "confirm complete",
			},
		})
		if err != nil {
			return err
		}

		//отправляем источнику сообщение типа 'confirm complete' для того что бы подтвердить останов задачи
		pprmtf.Chans.CwtRes <- configure.MsgWsTransmission{
			DestinationHost: pprmtf.SourceIP,
			Data:            &msgJSON,
		}
	}

	return nil
}
