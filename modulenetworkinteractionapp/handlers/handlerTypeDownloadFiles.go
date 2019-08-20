package handlers

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//MsgChannelReceivingFiles взаимодействие при приеме запрошенных файлов
// SourceID - ID источника
// SourceIP - IP источника
// TaskID - ID задачи
// Command - команда взаимодействия
//  - 'give my the file'
//  - 'ready to receive file'
//  - 'stop receiving files'
// Message - сообщения принимаемые от источников
type MsgChannelReceivingFiles struct {
	SourceID int
	SourceIP string
	TaskID   string
	Command  string
	Message  *[]byte
}

//listHandlerReceivingFile список задач по скачиванию файлов
// ключ - IP источника
type listHandlerReceivingFile map[string]listTaskReceivingFile

//listTaskReceivingFile список задач по приему файлов на данном источнике
// ключ - ID задачи
type listTaskReceivingFile map[string]handlerRecivingParameters

//handlerRecivingParameters описание параметров
type handlerRecivingParameters struct {
	chanToHandler chan string
}

//ControllerReceivingRequestedFiles обработчик приема запрашиваемых файлов
func ControllerReceivingRequestedFiles(
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	isl *configure.InformationSourcesList,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI) chan *MsgChannelReceivingFiles {

	clientNotify := configure.MsgBetweenCoreAndNI{
		Section: "message notification",
		Command: "send client API",
	}

	//сообщение об ошибке и сопутствующие действия
	handlerTaskWarning := func(taskID string, msg configure.MsgBetweenCoreAndNI) {
		chanInCore <- &clientNotify

		smt.UpdateTaskDownloadAllParameters(taskID, configure.DownloadTaskParameters{Status: "refused"})

		//снимаем отслеживание выполнения задачи
		chanInCore <- &configure.MsgBetweenCoreAndNI{
			TaskID:  taskID,
			Section: "monitoring task performance",
			Command: "complete task",
		}
	}

	chanIn := make(chan *MsgChannelReceivingFiles)
	lhrf := listHandlerReceivingFile{}

	go func() {
		for msg := range chanIn {
			clientNotify.TaskID = msg.TaskID
			ao := configure.MessageNotification{
				SourceReport:        "NI module",
				Section:             "download control",
				TypeActionPerformed: "task processing",
				CriticalityMessage:  "warning",
			}

			//получаем IP адрес и параметры источника
			si, ok := isl.GetSourceSetting(msg.SourceID)
			if !ok || !si.ConnectionStatus {
				humanNotify := fmt.Sprintf("Не возможно отправить запрос на скачивание файлов, источник с ID %v не подключен", msg.SourceID)
				if !ok {
					humanNotify = fmt.Sprintf("Источник с ID %v не найден", msg.SourceID)

					//удаляем из списка задач ожидающих выполнение
					if err := qts.DelQueueTaskStorage(msg.SourceID, msg.TaskID); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}
				}

				ao.HumanDescriptionNotification = humanNotify
				clientNotify.AdvancedOptions = ao

				handlerTaskWarning(msg.TaskID, clientNotify)

				continue
			}

			ao.HumanDescriptionNotification = fmt.Sprintf("Источник с ID %v не найден", msg.SourceID)
			clientNotify.AdvancedOptions = ao

			switch msg.Command {
			case "give my the file":
				if len(lhrf[si.IP]) == 0 {
					lhrf[si.IP] = listTaskReceivingFile{}
				}

				chanel, err := processorReceivingFiles(msg.SourceID, si.IP, msg.TaskID, smt)
				if err != nil {
					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				lhrf[si.IP][msg.TaskID] = handlerRecivingParameters{
					chanToHandler: chanel,
				}

			case "stop receiving files":
				if _, ok := lhrf[si.IP]; !ok {
					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}
				hrp, ok := lhrf[si.IP][msg.TaskID]
				if !ok {
					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				hrp.chanToHandler <- msg.Command

			case "taken from the source":
				resMsg := configure.MsgTypeDownload{}

				if err := json.Unmarshal(*msg.Message, &resMsg); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					continue
				}

				if _, ok := lhrf[si.IP]; !ok {
					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}
				hrp, ok := lhrf[si.IP][msg.TaskID]
				if !ok {
					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				//ответы приходящие от источника (команды для processorReceivingFiles)
				hrp.chanToHandler <- resMsg.Info.TaskStatus

			case "":
				/*

					!!! ОБРАБОТАТЬ РАЗРЫВ СОЕДИНЕНИЯ С ИСТОЧНИКОМ !!!

				*/

			}
			/*
				//выполняем запуск процесса по приему файлов (ОТ ЯДРА)
				if msg.Command == "give my the file" {
					if len(lhrf[si.IP]) == 0 {
						lhrf[si.IP] = listTaskReceivingFile{}
					}

					lhrf[si.IP][msg.TaskID] = handlerRecivingParameters{
						chanToHandler: processorReceivingFiles(msg.SourceID, si.IP, msg.TaskID),
					}
				}

				ao.HumanDescriptionNotification = fmt.Sprintf("Источник с ID %v не найден", msg.SourceID)
				clientNotify.AdvancedOptions = ao

				//выполняем останов процесса по приему файлов (ОТ ЯДРА)
				if msg.Command == "stop receiving files" {
					if _, ok := lhrf[si.IP]; !ok {
						handlerTaskWarning(msg.TaskID, clientNotify)

						continue
					}
					hrp, ok := lhrf[si.IP][msg.TaskID]
					if !ok {
						handlerTaskWarning(msg.TaskID, clientNotify)

						continue
					}

					hrp.chanToHandler <- msg.Command
				}

				//обработка всего того что приходит от источника
				if msg.Command == "taken from the source" {
					resMsg := configure.MsgTypeDownload{}

					if err := json.Unmarshal(*msg.Message, &resMsg); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

						continue
					}

					if _, ok := lhrf[si.IP]; !ok {
						handlerTaskWarning(msg.TaskID, clientNotify)

						continue
					}
					hrp, ok := lhrf[si.IP][msg.TaskID]
					if !ok {
						handlerTaskWarning(msg.TaskID, clientNotify)

						continue
					}

					//ответы приходящие от источника (команды для processorReceivingFiles)
					hrp.chanToHandler <- resMsg.Info.TaskStatus

				}*/
		}
	}()

	return chanIn
}

//processorReceivingFiles управляет приемом файлов в рамках одной задачи
func processorReceivingFiles(sourceID int, sourceIP, taskID string, smt *configure.StoringMemoryTask) (chan string, error) {
	chanOut := make(chan string)

	ti, ok := smt.GetStoringMemoryTask(taskID)
	if !ok {
		return chanOut, fmt.Errorf("task with ID %v not found", taskID)
	}

	go func() {
		//читаем список файлов
		for fn, fi := range ti.TaskParameter.DownloadTask.DownloadingFilesInformation {
			//отправляем источнику запрос на получение файла
			msgJSON := configure.MsgTypeDownload{
				MsgType: "download files",
				Info:    configure.DetailInfoMsgDownload{},
			}

			command := <-chanOut
			switch command {
			case "stop receiving files":

			case "ready for the transfer":

			case "file transfer completed":

			}
		}
	}()

	return chanOut, nil
}

//FileDownloadProcessing обработчик выполняющий процесс по скачиванию файлов
func FileDownloadProcessing(
	cwt chan<- configure.MsgWsTransmission,
	isl *configure.InformationSourcesList,
	msg *configure.MsgBetweenCoreAndNI,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI) {

	//msg.TaskID
	//msg.ClientName
	//msg.SourceID

	/*
	   Непосредственно выполняет скачивание файлов с источника
	   отправляя источнику задачи на скачивания по очередно,
	   в каждой задаче свой файл


	*/
}
