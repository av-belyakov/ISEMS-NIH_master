package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//msgChannelProcessorReceivingFiles параметры канала взаимодействия между 'ControllerReceivingRequestedFiles' и 'processorReceivingFiles'
// MessageType - тип передаваемых данных (1 - text, 2 - binary)
// MsgGenerator - источник сообщения ('Core module', 'NI module')
// Message - информационное сообщение в бинарном виде
type msgChannelProcessorReceivingFiles struct {
	MessageType  int
	MsgGenerator string
	Message      *[]byte
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

type messageFromCore string

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
				_ = saveMessageApp.LogMessage("info", fmt.Sprintf("it is not possible to send a request to download files, the source with ID %v is not connected", msg.SourceID))

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
			//начало выполнения задачи (запрос из Ядра)
			case "give my the files":
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

			//останов выполнения задачи (запрос из Ядра)
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

				c := []byte("stop receiving files")

				hrp.chanToHandler <- msgChannelProcessorReceivingFiles{
					MessageType:  1,
					MsgGenerator: "Core module",
					Message:      &c,
				}

			//ответы приходящие от источника в рамках выполнения конкретной задачи
			case "taken from the source":
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
					MessageType:  msg.MsgType,
					MsgGenerator: "NI module",
					Message:      msg.Message,
				}

			//сообщения о разрыве соединения
			case "":
				/*

					!!! ОБРАБОТАТЬ РАЗРЫВ СОЕДИНЕНИЯ С ИСТОЧНИКОМ !!!

				*/

			}
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

	/*
	   Алгоритм передачи и приема файлов
	   1. Запрос файла 'give me the file' (master -> slave)
	   2.1. Сообщение 'ready for the transfer' - готовность к передачи файла (slave -> master)
	   2.2. Сообщение 'file transfer not possible' - невозможно передать файл (slave -> master)
	   3. Готовность к приему файла 'ready to receive file' (master -> slave)
	   4. ПЕРЕДАЧА БИНАРНОГО ФАЙЛА
	   5. Сообщение о завершении передачи файла 'file transfer complited' (slave -> master)
	   6. Запрос нового файла 'give me the file' (master -> slave) цикл повторяется

	*/

	go func() {
		//начальный запрос на передачу файла
		mtd := configure.MsgTypeDownload{
			MsgType: "download files",
			Info: configure.DetailInfoMsgDownload{
				TaskID:         taskID,
				PathDirStorage: pathDirStorage,
			},
		}

		//читаем список файлов
		for fn, fi := range ti.TaskParameter.DownloadTask.DownloadingFilesInformation {
			//делаем первый запрос на скачивание файла
			mtd.Info.TaskStatus = "give me the file"
			mtd.Info.FileOptions = configure.DownloadFileOptions{
				Name: fn,
				Size: fi.Size,
				Hex:  fi.Hex,
			}

			msgJSON, err := json.Marshal(mtd)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				continue
			}
			cwtRes <- configure.MsgWsTransmission{
				DestinationHost: sourceIP,
				Data:            &msgJSON,
			}

			msg := <-chanOut

			listFileDescriptors := map[string]*os.File{}

			//текстовые данные
			if msg.MessageType == 1 {
				if msg.MsgGenerator == "Core module" {
					command := fmt.Sprint(*msg.Message)

					//остановить скачивание файлов
					if command == "stop receiving files" {
						/*
							- Сообщение о том что задача успешно ОСТАНОВЛЕНА
							- Записать инофрмацию о задаче в БД

							После записи информации в БД УЖЕ В Core modules
							после ответа из БД удалить задачу из StoringeMemoryTask и
							StoringMemoryQueueTask

							- завершить подпрограмму, тем самым остановив цикл
							по запросам файлов у источника
						*/
					}

				} else if msg.MsgGenerator == "NI module" {
					var msgRes configure.MsgTypeDownload
					err := json.Unmarshal(*msg.Message, &msgRes)
					if err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

						continue
					}

					switch msgRes.Info.TaskStatus {
					//готовность к приему файла
					case "ready for the transfer":
						/*
							- Создать линк файла для записи бинарных данных
							из расчета что одновременно могут передаваться
							несколько файлов
							map[<file_hex>]*os.Writer

							- Отправить источнику сообщение о готовности к
							приему данных
						*/

						//отправляем источнику запрос на получение файла
						//msgJSON

						if _, ok := listFileDescriptors[msgRes.Info.FileOptions.Hex]; !ok {
							//создаем дескриптор файла для последующей записи в него
							f, err := os.Create(path.Join(pathDirStorage, msgRes.Info.FileOptions.Name))
							if err != nil {
								_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

								continue
							}

							listFileDescriptors[msgRes.Info.FileOptions.Hex] = f

							msgJSON, err := json.Marshal(configure.MsgTypeDownload{
								MsgType: "download files",
								Info: configure.DetailInfoMsgDownload{
									TaskID:     taskID,
									TaskStatus: "ready to receive file",
								},
							})
							if err != nil {
								_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

								continue
							}
							cwtRes <- configure.MsgWsTransmission{
								DestinationHost: sourceIP,
								Data:            &msgJSON,
							}
						}

					//передача файла успешно завершена
					case "file transfer completed":
						/*

							- Отправить новый запрос на скачивание файла
							такой же как и самый первый 'give me the file'
						*/

						//закрыть дескриптор файла listFileDescription[msg.Message.FileOptions.Hex].Close()

					//сообщение о невозможности передачи файла
					case "file transfer not possible":
						//закрыть дескриптор файла listFileDescription[msg.Message.FileOptions.Hex].Close()

					}
				} else {
					_ = saveMessageApp.LogMessage("error", "unknown generator events")

					continue
				}
			}

			//бинарные данные
			if msg.MessageType == 2 {
				//listFileDescription[msg.Message.FileOptions.Hex]
			}
		}

		/*
			Так как список файлов для скачивания закончился

			- Сообщение о том что задача успешно ЗАВЕРШЕНА

							- Записать инофрмацию о задаче в БД

							После записи информации в БД УЖЕ В Core modules
							после ответа из БД удалить задачу из StoringeMemoryTask и
							StoringMemoryQueueTask

		*/
	}()

	return chanOut, nil
}
