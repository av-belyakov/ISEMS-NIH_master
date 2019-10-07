package processresponse

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//ParametersProcessingReceivedMsgTypeFiltering набор параметров для функции ProcessingReceivedMsgTypeFiltering
type ParametersProcessingReceivedMsgTypeFiltering struct {
	CwtRes         chan<- configure.MsgWsTransmission
	ChanInCore     chan<- *configure.MsgBetweenCoreAndNI
	CwtReq         <-chan configure.MsgWsTransmission
	Isl            *configure.InformationSourcesList
	SMT            *configure.StoringMemoryTask
	Message        *[]byte
	SourceID       int
	SourceIP       string
	SaveMessageApp *savemessageapp.PathDirLocationLogFiles
}

//ProcessingReceivedMsgTypeFiltering обработка сообщений связанных с фильтрацией файлов
func ProcessingReceivedMsgTypeFiltering(pprmtf ParametersProcessingReceivedMsgTypeFiltering) {
	resMsg := configure.MsgTypeFiltration{}

	if err := json.Unmarshal(*pprmtf.Message, &resMsg); err != nil {
		_ = pprmtf.SaveMessageApp.LogMessage("error", fmt.Sprint(err))
	}

	ffi := make(map[string]*configure.FoundFilesInformation, len(resMsg.Info.FoundFilesInformation))
	for n, v := range resMsg.Info.FoundFilesInformation {
		ffi[n] = &configure.FoundFilesInformation{
			Size: v.Size,
			Hex:  v.Hex,
		}
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
		FoundFilesInformation:           ffi,
	}

	//обновляем информацию о выполняемой задаче в памяти приложения
	pprmtf.SMT.UpdateTaskFiltrationAllParameters(resMsg.Info.TaskID, ftp)

	msgCompliteTask := configure.MsgBetweenCoreAndNI{
		TaskID:  resMsg.Info.TaskID,
		Section: "monitoring task performance",
		Command: "complete task",
	}

	msg := &configure.MsgBetweenCoreAndNI{
		TaskID:          resMsg.Info.TaskID,
		Section:         "filtration control",
		Command:         resMsg.Info.TaskStatus,
		SourceID:        pprmtf.SourceID,
		AdvancedOptions: ffi,
	}

	fmt.Printf("\tпринята информация о задаче с ID '%v', статус задачи - %v\n", resMsg.Info.TaskID, resMsg.Info.TaskStatus)

	if resMsg.Info.TaskStatus == "execute" {
		//отправляем в ядро, а от туда в БД и клиенту API
		pprmtf.ChanInCore <- msg

		return
	}

	if resMsg.Info.TaskStatus == "refused" {
		//отправляем в ядро, а от туда в БД и клиенту API
		pprmtf.ChanInCore <- msg

		//отправляем сообщение о снятии контроля за выполнением задачи
		pprmtf.ChanInCore <- &msgCompliteTask

		return
	}

	//если тип сообщения "stop" или "complete"

	//отправка информации только после получения всех частей
	if resMsg.Info.NumberMessagesParts[0] == resMsg.Info.NumberMessagesParts[1] {
		//отправляем в ядро, а от туда в БД и клиенту API
		pprmtf.ChanInCore <- msg

		//отправляем сообщение о снятии контроля за выполнением задачи
		pprmtf.ChanInCore <- &msgCompliteTask

		resConfirmComplite := configure.MsgTypeFiltrationControl{
			MsgType: "filtration",
			Info: configure.SettingsFiltrationControl{
				TaskID:  resMsg.Info.TaskID,
				Command: "confirm complete",
			},
		}

		msgJSON, err := json.Marshal(resConfirmComplite)
		if err != nil {
			_ = pprmtf.SaveMessageApp.LogMessage("error", fmt.Sprint(err))

			return
		}

		//отправляем источнику сообщение типа 'confirm complete' для того что бы подтвердить останов задачи
		pprmtf.CwtRes <- configure.MsgWsTransmission{
			DestinationHost: pprmtf.SourceIP,
			Data:            &msgJSON,
		}
	}
}
