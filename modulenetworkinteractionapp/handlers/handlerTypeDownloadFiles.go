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

type statusDownloadFile struct {
	Status string
	ErrMsg error
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

	go func(
		lhrf listHandlerReceivingFile,
		clientNotify configure.MsgBetweenCoreAndNI,
		chanIn <-chan *configure.MsgChannelReceivingFiles) {

		for msg := range chanIn {
			clientNotify.TaskID = msg.TaskID
			ao := configure.MessageNotification{
				SourceReport:        "NI module",
				Section:             "download control",
				TypeActionPerformed: "task processing",
				CriticalityMessage:  "warning",
			}

			fmt.Printf("\tfunc 'ControllerReceivingRequestedFiles' resived new msg DOWNLOAD TASK for task ID %v\n", msg.TaskID)

			//получаем IP адрес и параметры источника
			si, ok := isl.GetSourceSetting(msg.SourceID)
			if !ok || !si.ConnectionStatus {
				_ = saveMessageApp.LogMessage("info", fmt.Sprintf("it is not possible to send a request to download files, the source with ID %v is not connected", msg.SourceID))

				humanNotify := fmt.Sprintf("Не возможно отправить запрос на скачивание файлов, источник с ID %v не подключен", msg.SourceID)
				if !ok {
					humanNotify = fmt.Sprintf("Источник с ID %v не найден", msg.SourceID)

					//изменяем статус задачи в storingMemoryQueueTask
					// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
					// функцией 'CheckTimeQueueTaskStorage')
					if err := qts.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "complete"); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}
				}

				ao.HumanDescriptionNotification = humanNotify
				clientNotify.AdvancedOptions = ao

				handlerTaskWarning(msg.TaskID, clientNotify)

				continue
			}

			fmt.Printf("\tfunc 'ControllerReceivingRequestedFiles' RESIVED SOURCE PARAMETERS: %v\n", si)

			ao.HumanDescriptionNotification = fmt.Sprintf("Источник с ID %v не найден", msg.SourceID)
			clientNotify.AdvancedOptions = ao

			errMsg := fmt.Sprintf("Source with ID %v not found", msg.SourceID)

			switch msg.Command {
			//начало выполнения задачи (запрос из Ядра)
			case "give my the files":
				if len(lhrf[si.IP]) == 0 {
					lhrf[si.IP] = listTaskReceivingFile{}
				}

				fmt.Println("\tfunc 'ControllerReceivingRequestedFiles' запуск обработчика задачи по скачиванию файлов")

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

			//останов выполнения задачи из-за разрыва соединения (запрос из Ядра)
			case "to stop the task because of a disconnection":
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

				c := []byte("to stop the task because of a disconnection")

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

			}
		}
	}(lhrf, clientNotify, chanIn)

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

	/*
	   Алгоритм передачи и приема файлов
	   1. Запрос файла 'give me the file' (master -> slave)
	   2.1. Сообщение 'ready for the transfer' - готовность к передачи файла (slave -> master)
	   2.2. Сообщение 'file transfer not possible' - невозможно передать файл (slave -> master)
	   3. Готовность к приему файла 'ready to receive file' (master -> slave)
	   4. ПЕРЕДАЧА БИНАРНОГО ФАЙЛА
	   5. завершением приема файла считается прием последнего части файла
	   6. Запрос нового файла 'give me the file' (master -> slave) цикл повторяется
	*/

	fmt.Println("\tDOWNLOAD: func 'processorReceivingFiles', START...")

	ti, ok := smt.GetStoringMemoryTask(taskID)
	if !ok {

		fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', task with ID %v not found\n", taskID)

		return nil, fmt.Errorf("task with ID %v not found", taskID)
	}

	chanOut := make(chan msgChannelProcessorReceivingFiles)
	pathDirStorage := ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles

	go func() {
		sdf := statusDownloadFile{Status: "success"}
		listFileDescriptors := map[string]*os.File{}

		//начальный запрос на передачу файла
		mtd := configure.MsgTypeDownload{
			MsgType: "download files",
			Info: configure.DetailInfoMsgDownload{
				TaskID:         taskID,
				PathDirStorage: pathDirStorage,
			},
		}

		fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', готовим начальный запрос %v not found, читаем список файлов %v\n", mtd, ti.TaskParameter.FiltrationTask.PathStorageSource)

	DONE:
		//читаем список файлов
		for fn, fi := range ti.TaskParameter.DownloadTask.DownloadingFilesInformation {
			//делаем первый запрос на скачивание файла
			mtd.Info.Command = "give me the file"
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

			fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', make ONE request to download file, %v\n", mtd)

			cwtRes <- configure.MsgWsTransmission{
				DestinationHost: sourceIP,
				Data:            &msgJSON,
			}

		NEWFILE:
			for msg := range chanOut {
				//обновляем значение таймера (что бы задача не была удалена по таймауту)
				smt.TimerUpdateStoringMemoryTask(taskID)

				//текстовые данные
				if msg.MessageType == 1 {
					msgReq := configure.MsgTypeDownload{
						MsgType: "download files",
						Info: configure.DetailInfoMsgDownload{
							TaskID: taskID,
						},
					}

					if msg.MsgGenerator == "Core module" {
						command := fmt.Sprint(*msg.Message)

						//остановить скачивание файлов
						if command == "stop receiving files" {
							msgReq.Info.Command = "stop receiving files"

							msgJSON, err := json.Marshal(msgReq)
							if err != nil {
								_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

								continue
							}
							cwtRes <- configure.MsgWsTransmission{
								DestinationHost: sourceIP,
								Data:            &msgJSON,
							}
						}

						//разрыв соединения (остановить загрузку файлов)
						if command == "to stop the task because of a disconnection" {
							//закрываем дескриптор файла
							if w, ok := listFileDescriptors[fi.Hex]; ok {
								w.Close()

								//удаляем дескриптор файла
								delete(listFileDescriptors, fi.Hex)
							}

							//удаляем файл
							_ = os.Remove(path.Join(pathDirStorage, fn))

							sdf.Status = "task stoped disconnect"

							break DONE
						}

					} else if msg.MsgGenerator == "NI module" {
						var msgRes configure.MsgTypeDownload
						err := json.Unmarshal(*msg.Message, &msgRes)
						if err != nil {
							_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

							continue
						}

						/* получаем информацию о задаче */
						ti, ok := smt.GetStoringMemoryTask(taskID)
						if !ok {
							_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with ID %v not found", taskID))

							sdf.Status = "error"
							sdf.ErrMsg = err

							break DONE
						}

						fi := ti.TaskParameter.DownloadTask.FileInformation

						switch msgRes.Info.Command {
						//готовность к приему файла (slave -> master)
						case "ready for the transfer":
							if _, ok := listFileDescriptors[msgRes.Info.FileOptions.Hex]; ok {
								continue
							}

							//создаем дескриптор файла для последующей записи в него
							f, err := os.Create(path.Join(pathDirStorage, msgRes.Info.FileOptions.Name))
							if err != nil {
								_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

								sdf.Status = "error"
								sdf.ErrMsg = err

								break DONE
							}

							listFileDescriptors[msgRes.Info.FileOptions.Hex] = f

							//обновляем информацию о задаче
							smt.UpdateTaskDownloadAllParameters(taskID, configure.DownloadTaskParameters{
								Status:                              "wait",
								NumberFilesTotal:                    ti.TaskParameter.DownloadTask.NumberFilesTotal,
								NumberFilesDownloaded:               ti.TaskParameter.DownloadTask.NumberFilesDownloaded,
								PathDirectoryStorageDownloadedFiles: ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles,
								FileInformation: configure.DetailedFileInformation{
									Name:         fn,
									Hex:          fi.Hex,
									FullSizeByte: fi.FullSizeByte,
									NumChunk:     msgRes.Info.FileOptions.NumChunk,
									ChunkSize:    msgRes.Info.FileOptions.ChunkSize,
								},
							})

							msgReq.Info.Command = "ready to receive file"
							msgJSON, err := json.Marshal(msgReq)
							if err != nil {
								_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

								sdf.Status = "error"
								sdf.ErrMsg = err

								break DONE
							}

							cwtRes <- configure.MsgWsTransmission{
								DestinationHost: sourceIP,
								Data:            &msgJSON,
							}

						//сообщение о невозможности передачи файла (slave -> master)
						case "file transfer not possible":
							dtp := ti.TaskParameter.DownloadTask
							dtp.NumberFilesDownloadedError = dtp.NumberFilesDownloadedError + 1

							//добавляем информацию о не принятом файле
							smt.UpdateTaskDownloadAllParameters(taskID, dtp)

							//отправляем информацию в Ядро
							chanInCore <- &configure.MsgBetweenCoreAndNI{
								TaskID:   taskID,
								Section:  "download control",
								Command:  "file download process",
								SourceID: sourceID,
							}

							break NEWFILE

						//передача файла успешно остановлена (slave -> master)
						case "file transfer stopped":
							//закрываем дескриптор файла
							if w, ok := listFileDescriptors[fi.Hex]; ok {
								w.Close()

								//удаляем дескриптор файла
								delete(listFileDescriptors, fi.Hex)
							}

							//удаляем файл
							_ = os.Remove(path.Join(pathDirStorage, fn))

							sdf.Status = "task stoped client"

							break DONE

						//невозможно остановить передачу файла
						case "impossible to stop file transfer":
							_ = saveMessageApp.LogMessage("error", fmt.Sprintf("it is impossible to stop file transfer (source ID: %v, task ID: %v)", sourceID, taskID))

						}
					} else {
						_ = saveMessageApp.LogMessage("error", "unknown generator events")

						break NEWFILE
					}
				}

				//бинарные данные
				if msg.MessageType == 2 {
					fileIsLoaded, err := writingBinaryFile(parametersWritingBinaryFile{
						SourceID:            sourceID,
						TaskID:              taskID,
						Data:                msg.Message,
						ListFileDescriptors: listFileDescriptors,
						SMT:                 smt,
						ChanInCore:          chanInCore,
					})
					if err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

						sdf.Status = "error"
						sdf.ErrMsg = err

						break DONE
					}

					//если файл полностью загружен запрашиваем следующий файл
					if fileIsLoaded {
						break NEWFILE
					}
				}
			}
		}

		dtp := ti.TaskParameter.DownloadTask
		dtp.Status = "complete"

		/*
			изменяем состояние задачи по которому данная задача будет
			удалена через определенный промежуток времени
		*/
		smt.UpdateTaskDownloadAllParameters(taskID, dtp)

		//задача завершена успешно
		msgToCore := configure.MsgBetweenCoreAndNI{
			TaskID:   taskID,
			Section:  "download control",
			Command:  "task completed",
			SourceID: sourceID,
		}

		switch sdf.Status {
		//задача остановлена пользователем
		case "task stoped client":
			msgToCore.Command = "file transfer stopped"

		//задача остановлена в связи с разрывом соединения с источником
		case "task stoped disconnect":
			msgToCore.Command = "task stoped disconnect"

		//задача остановлена из-за внутренней ошибки приложения
		case "error":
			msgToCore.Command = "task stoped error"

		}

		chanInCore <- &msgToCore
	}()

	return chanOut, nil
}
