package processresponse

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//ProcessingReceivedMsgTypeFiltering обработка сообщений связанных с фильтрацией файлов
func ProcessingReceivedMsgTypeFiltering(
	cwtRes chan<- configure.MsgWsTransmission,
	isl *configure.InformationSourcesList,
	smt *configure.StoringMemoryTask,
	message *[]byte,
	sourceID int,
	sourceIP string,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	cwtReq <-chan configure.MsgWsTransmission) {

	fmt.Println("START function 'ProcessingReceivedMsgTypeFilteringn'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	resMsg := configure.MsgTypeFiltration{}

	if err := json.Unmarshal(*message, &resMsg); err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
	}

	ffi := make(map[string]*configure.FoundFilesInformation, len(resMsg.Info.FoundFilesInformation))
	for n, v := range resMsg.Info.FoundFilesInformation {
		ffi[n] = &configure.FoundFilesInformation{
			Size: v.Size,
			Hex:  v.Hex,
		}
	}

	ftp := configure.FiltrationTaskParameters{
		ID:                              sourceID,
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
	smt.UpdateTaskFiltrationAllParameters(resMsg.Info.TaskID, ftp)

	msg := &configure.MsgBetweenCoreAndNI{
		TaskID:   resMsg.Info.TaskID,
		Section:  "filtration control",
		Command:  resMsg.Info.TaskStatus,
		SourceID: sourceID,
	}

	//обычный ответ по задаче фильтрации
	if (resMsg.Info.TaskStatus == "execute") || (resMsg.Info.TaskStatus == "rejected") {
		//отправляем в ядро, а от туда в БД и клиенту API
		chanInCore <- msg

		return
	}

	//если тип сообщения "stop" или "complite"

	//отправка информации только после получения всех частей
	if resMsg.Info.NumberMessagesParts[0] == resMsg.Info.NumberMessagesParts[1] {

		//отправляем в ядро, а от туда в БД и клиенту API
		chanInCore <- msg

		resConfirmComplite := configure.MsgTypeFiltrationControl{
			MsgType: "filtration",
			Info: configure.SettingsFiltrationControl{
				TaskID:  resMsg.Info.TaskID,
				Command: "confirm complite",
			},
		}

		msgJSON, err := json.Marshal(resConfirmComplite)
		if err != nil {
			_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

			return
		}

		//отправляем источнику сообщение типа 'confirm complite' для того что бы подтвердить останов задачи
		cwtRes <- configure.MsgWsTransmission{
			DestinationHost: sourceIP,
			Data:            &msgJSON,
		}
	}
}
