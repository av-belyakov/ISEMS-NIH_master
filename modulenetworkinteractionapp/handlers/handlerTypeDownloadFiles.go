package handlers

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//msgChannelProcessorReceivingFiles параметры канала взаимодействия между 'ControllerReceivingRequestedFiles' и 'processorReceivingFiles'
// TaskStatus - состояние задачи
// MessageType - тип передаваемых данных (1 - text, 2 - binary)
// Message - информационное сообщение в двоичном формате
type msgChannelProcessorReceivingFiles struct {
	TaskStatus  string
	MessageType int
	Message     *[]byte
}

//listHandlerReceivingFile список задач по скачиванию файлов
// ключ - IP источника
type listHandlerReceivingFile map[string]listTaskReceivingFile

//listTaskReceivingFile список задач по приему файлов на данном источнике
// ключ - ID задачи
type listTaskReceivingFile map[string]handlerRecivingParameters

//handlerRecivingParameters описание параметров
type handlerRecivingParameters struct {
	chanToHandler chan msgChannelProcessorReceivingFiles
}

//ControllerReceivingRequestedFiles обработчик приема запрашиваемых файлов
func ControllerReceivingRequestedFiles(
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	isl *configure.InformationSourcesList,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	cwtRes chan<- configure.MsgWsTransmission) chan *configure.MsgChannelReceivingFiles {

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

	chanIn := make(chan *configure.MsgChannelReceivingFiles)
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
				_ = saveMessageApp.LogMessage("info", fmt.Sprintf("It is not possible to send a request to download files, the source with ID %v is not connected", msg.SourceID))

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

			errMsg := fmt.Sprintf("Source with ID %v not found", msg.SourceID)

			switch msg.Command {
			//начало выполнения задачи
			case "give my the file":
				if len(lhrf[si.IP]) == 0 {
					lhrf[si.IP] = listTaskReceivingFile{}
				}

				//запуск обработчика задачи по скачиванию файлов
				channel, err := processorReceivingFiles(chanInCore, msg.SourceID, si.IP, msg.TaskID, smt, saveMessageApp, cwtRes)
				if err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				lhrf[si.IP][msg.TaskID] = handlerRecivingParameters{
					chanToHandler: channel,
				}

			//останов выполнения задачи
			case "stop receiving files":
				if _, ok := lhrf[si.IP]; !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}
				hrp, ok := lhrf[si.IP][msg.TaskID]
				if !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				hrp.chanToHandler <- msgChannelProcessorReceivingFiles{
					TaskStatus: msg.Command,
				}

			//ответы приходящие от источника в рамках выполнения конкретной задачи
			case "taken from the source":
				resMsg := configure.MsgTypeDownload{}

				if err := json.Unmarshal(*msg.Message, &resMsg); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					continue
				}

				if _, ok := lhrf[si.IP]; !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}
				hrp, ok := lhrf[si.IP][msg.TaskID]
				if !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				//ответы приходящие от источника (команды для processorReceivingFiles)
				hrp.chanToHandler <- msgChannelProcessorReceivingFiles{
					TaskStatus:  resMsg.Info.TaskStatus,
					MessageType: msg.MsgType,
					Message:     msg.Message,
				}

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
func processorReceivingFiles(
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	sourceID int,
	sourceIP, taskID string,
	smt *configure.StoringMemoryTask,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	cwtRes chan<- configure.MsgWsTransmission) (chan msgChannelProcessorReceivingFiles, error) {

	ti, ok := smt.GetStoringMemoryTask(taskID)
	if !ok {
		return nil, fmt.Errorf("task with ID %v not found", taskID)
	}

	chanOut := make(chan msgChannelProcessorReceivingFiles)

	pathDirStorage := ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles

	go func() {
		//читаем список файлов
		for fn, fi := range ti.TaskParameter.DownloadTask.DownloadingFilesInformation {
			msg := <-chanOut

			//текстовые данные
			if msg.MessageType == 1 {
				switch msg.TaskStatus {
				case "stop receiving files":
					/*
						- Сообщение о том что задача успешно ОСТАНОВЛЕНА
						- Записать инофрмацию о задаче в БД

						После записи информации в БД УЖЕ В Core modules
						после ответа из БД удалить задачу из StoringeMemoryTask и
						StoringMemoryQueueTask
					*/

				case "ready for the transfer":
					/*
						- Создать линк файла для записи бинарных данных

						- Отправить источнику сообщение о готовности к
						приему данных
					*/

					//отправляем источнику запрос на получение файла
					msgJSON := configure.MsgTypeDownload{
						MsgType: "download files",
						Info:    configure.DetailInfoMsgDownload{},
					}

				case "file transfer completed":
					/*
						- Сообщение о том что задача успешно ЗАВЕРШЕНА

						- Записать инофрмацию о задаче в БД

						После записи информации в БД УЖЕ В Core modules
						после ответа из БД удалить задачу из StoringeMemoryTask и
						StoringMemoryQueueTask
					*/

				}
			}

			//бинарные данные
			if msg.MessageType == 2 {

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
