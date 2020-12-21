package handlers

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//ChanHandlerMessagesReceivedFilesFiltering канал для обработчика входящей информации о фильтрации
type ChanHandlerMessagesReceivedFilesFiltering struct {
	SourceID int
	SourceIP string
	Message  *[]byte
}

//HandlerMessagesReceivedFilesFiltering обработчик сообщений о фильтрации поступающих с источников
func HandlerMessagesReceivedFilesFiltering(
	smt *configure.StoringMemoryTask,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanToCore chan<- *configure.MsgBetweenCoreAndNI,
	cwtRes chan<- configure.MsgWsTransmission) chan *ChanHandlerMessagesReceivedFilesFiltering {

	funcName := "handlerMessagesReceivedFilesFiltering"

	chmrff := make(chan *ChanHandlerMessagesReceivedFilesFiltering)

	go func() {
		for data := range chmrff {
			var resMsg configure.MsgTypeFiltration

			if err := json.Unmarshal(*data.Message, &resMsg); err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				continue
			}

			//обновляем информацию о выполняемой задаче в памяти приложения
			smt.UpdateTaskFiltrationAllParameters(resMsg.Info.TaskID, &configure.FiltrationTaskParameters{
				Status:                          resMsg.Info.TaskStatus,
				NumberFilesMeetFilterParameters: resMsg.Info.NumberFilesMeetFilterParameters,
				NumberProcessedFiles:            resMsg.Info.NumberProcessedFiles,
				NumberFilesFoundResultFiltering: resMsg.Info.NumberFilesFoundResultFiltering,
				NumberDirectoryFiltartion:       resMsg.Info.NumberDirectoryFiltartion,
				NumberErrorProcessedFiles:       resMsg.Info.NumberErrorProcessedFiles,
				SizeFilesMeetFilterParameters:   resMsg.Info.SizeFilesMeetFilterParameters,
				SizeFilesFoundResultFiltering:   resMsg.Info.SizeFilesFoundResultFiltering,
				PathStorageSource:               resMsg.Info.PathStorageSource,
			})

			lfdi := make(map[string]*configure.DetailedFilesInformation, len(resMsg.Info.FoundFilesInformation))
			for n, v := range resMsg.Info.FoundFilesInformation {
				lfdi[n] = &configure.DetailedFilesInformation{
					Size: v.Size,
					Hex:  v.Hex,
				}
			}

			//fmt.Printf("func '%v', ------------ FILTRATION received list files: '%v'\n", funcName, lfdi)

			smt.UpdateListFilesDetailedInformation(resMsg.Info.TaskID, lfdi)

			msgCompleteTask := configure.MsgBetweenCoreAndNI{
				TaskID:  resMsg.Info.TaskID,
				Section: "monitoring task performance",
				Command: "complete task",
			}

			msg := &configure.MsgBetweenCoreAndNI{
				TaskID:   resMsg.Info.TaskID,
				Section:  "filtration control",
				Command:  resMsg.Info.TaskStatus,
				SourceID: data.SourceID,
				AdvancedOptions: configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus{
					TaskStatus:                      resMsg.Info.TaskStatus,
					ListFoundFile:                   lfdi,
					NumberFilesMeetFilterParameters: resMsg.Info.NumberFilesMeetFilterParameters,
					NumberProcessedFiles:            resMsg.Info.NumberProcessedFiles,
					NumberFilesFoundResultFiltering: resMsg.Info.NumberFilesFoundResultFiltering,
					NumberDirectoryFiltartion:       resMsg.Info.NumberDirectoryFiltartion,
					NumberErrorProcessedFiles:       resMsg.Info.NumberErrorProcessedFiles,
					SizeFilesMeetFilterParameters:   resMsg.Info.SizeFilesMeetFilterParameters,
					SizeFilesFoundResultFiltering:   resMsg.Info.SizeFilesFoundResultFiltering,
					PathStorageSource:               resMsg.Info.PathStorageSource,
				},
			}

			//fmt.Printf("func '%v', task status: '%v' ---------\n", funcName, resMsg.Info.TaskStatus)

			if resMsg.Info.TaskStatus == "execute" {
				//отправляем в ядро, а от туда в БД и клиенту API
				chanToCore <- msg

				continue
			}

			if resMsg.Info.TaskStatus == "refused" {
				//отправляем в ядро, а от туда в БД и клиенту API
				chanToCore <- msg

				//отправляем сообщение о снятии контроля за выполнением задачи
				chanToCore <- &msgCompleteTask

				continue
			}

			//fmt.Printf("func '%v', task status: '%v', send to DB\n", funcName, resMsg.Info.TaskStatus)

			//если тип сообщения "stop" или "complete"
			// отправка информации только после получения всех частей
			if resMsg.Info.NumberMessagesParts[0] == resMsg.Info.NumberMessagesParts[1] {
				//отправляем в ядро, а от туда в БД и клиенту API
				chanToCore <- msg

				msgJSON, err := json.Marshal(configure.MsgTypeFiltrationControl{
					MsgType: "filtration",
					Info: configure.SettingsFiltrationControl{
						TaskID:  resMsg.Info.TaskID,
						Command: "confirm complete",
					},
				})
				if err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					continue
				}

				//fmt.Printf("func '%v', task status: '%v', send to ISEMS-NIH_slave message 'confirm complete'\n", funcName, resMsg.Info.TaskStatus)

				//отправляем источнику сообщение типа 'confirm complete' для того что бы подтвердить останов задачи
				cwtRes <- configure.MsgWsTransmission{
					DestinationHost: data.SourceIP,
					Data:            &msgJSON,
				}
			}
		}
	}()

	return chmrff
}
