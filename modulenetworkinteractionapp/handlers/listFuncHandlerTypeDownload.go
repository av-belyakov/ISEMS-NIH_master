package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

type parametersWritingBinaryFile struct {
	SourceID            int
	TaskID              string
	Data                *[]byte
	ListFileDescriptors map[string]*os.File
	SMT                 *configure.StoringMemoryTask
	ChanInCore          chan<- *configure.MsgBetweenCoreAndNI
}

type typeProcessingDownloadFile struct {
	sourceID       int
	sourceIP       string
	taskID         string
	taskInfo       configure.TaskDescription
	smt            *configure.StoringMemoryTask
	saveMessageApp *savemessageapp.PathDirLocationLogFiles
	channels       listChannels
}

type listChannels struct {
	chanInCore  chan<- *configure.MsgBetweenCoreAndNI
	chanOutCore <-chan msgChannelProcessorReceivingFiles
	cwtRes      chan<- configure.MsgWsTransmission
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

	fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', '%v'\n", ti.TaskParameter.DownloadTask.DownloadingFilesInformation)

	//отправляем информационное сообщение пользователю
	chanInCore <- &configure.MsgBetweenCoreAndNI{
		TaskID:  taskID,
		Section: "message notification",
		Command: "send client API",
		AdvancedOptions: configure.MessageNotification{
			SourceReport:                 "NI module",
			Section:                      "download control",
			TypeActionPerformed:          "task processing",
			CriticalityMessage:           "info",
			HumanDescriptionNotification: fmt.Sprintf("Инициализирована задача по скачиванию файлов с источника %v, идет подготовка списка загружаемых файлов", sourceID),
			Sources:                      []int{ti.TaskParameter.DownloadTask.ID},
		},
	}

	//проверяем наличие файлов для скачивания
	if len(ti.TaskParameter.DownloadTask.DownloadingFilesInformation) == 0 {
		return chanOut, fmt.Errorf("the list of files suitable for downloading from the source is empty")
	}

	go processingDownloadFile(typeProcessingDownloadFile{
		sourceID:       sourceID,
		sourceIP:       sourceIP,
		taskID:         taskID,
		taskInfo:       ti,
		smt:            smt,
		saveMessageApp: saveMessageApp,
		channels: listChannels{
			chanInCore:  chanInCore,
			chanOutCore: chanOut,
			cwtRes:      cwtRes,
		},
	})

	return chanOut, nil
}

func processingDownloadFile(tpdf typeProcessingDownloadFile) {

	sdf := statusDownloadFile{Status: "success"}
	listFileDescriptors := map[string]*os.File{}

	//начальный запрос на передачу файла
	mtd := configure.MsgTypeDownload{
		MsgType: "download files",
		Info: configure.DetailInfoMsgDownload{
			TaskID:         tpdf.taskID,
			PathDirStorage: tpdf.taskInfo.TaskParameter.FiltrationTask.PathStorageSource,
		},
	}

	pathDirStorage := tpdf.taskInfo.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles

	fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', готовим начальный запрос '%v', путь до списка файлов на источнике: '%v'\n", mtd, tpdf.taskInfo.TaskParameter.FiltrationTask.PathStorageSource)

DONE:
	//читаем список файлов
	for fn, fi := range tpdf.taskInfo.TaskParameter.DownloadTask.DownloadingFilesInformation {
		//делаем первый запрос на скачивание файла
		mtd.Info.Command = "give me the file"
		mtd.Info.FileOptions = configure.DownloadFileOptions{
			Name: fn,
			Size: fi.Size,
			Hex:  fi.Hex,
		}

		msgJSON, err := json.Marshal(mtd)
		if err != nil {
			_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

			continue
		}

		//fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', make ONE request to download file, %v\n", mtd)

		tpdf.channels.cwtRes <- configure.MsgWsTransmission{
			DestinationHost: tpdf.sourceIP,
			Data:            &msgJSON,
		}

	NEWFILE:
		for msg := range tpdf.channels.chanOutCore {
			//обновляем значение таймера (что бы задача не была удалена по таймауту)
			tpdf.smt.TimerUpdateStoringMemoryTask(tpdf.taskID)

			//fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', RESIVED MSG, %v\n", msg)

			/* текстовый тип сообщения */
			if msg.MessageType == 1 {
				msgReq := configure.MsgTypeDownload{
					MsgType: "download files",
					Info: configure.DetailInfoMsgDownload{
						TaskID: tpdf.taskID,
					},
				}

				if msg.MsgGenerator == "Core module" {
					command := fmt.Sprint(*msg.Message)

					//остановить скачивание файлов
					if command == "stop receiving files" {
						msgReq.Info.Command = "stop receiving files"

						msgJSON, err := json.Marshal(msgReq)
						if err != nil {
							_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

							continue
						}
						tpdf.channels.cwtRes <- configure.MsgWsTransmission{
							DestinationHost: tpdf.sourceIP,
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
					if err := json.Unmarshal(*msg.Message, &msgRes); err != nil {
						_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

						continue
					}

					//fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', GENERATOR = 'NI module' RESIVED msg '%v'\n", msgRes)

					//fmt.Println("\tDOWNLOAD: func 'processorReceivingFiles', получаем информацию о задаче")

					/* получаем информацию о задаче */
					ti, ok := tpdf.smt.GetStoringMemoryTask(tpdf.taskID)
					if !ok {
						_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprintf("task with ID %v not found", tpdf.taskID))

						sdf.Status = "error"
						sdf.ErrMsg = err

						break DONE
					}

					//fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', TASK INFO '%v'\n", ti)

					fi := ti.TaskParameter.DownloadTask.FileInformation

					switch msgRes.Info.Command {
					//готовность к приему файла (slave -> master)
					case "ready for the transfer":

						//fmt.Println("\tDOWNLOAD: func 'processorReceivingFiles', command 'ready for the transfer'")

						if _, ok := listFileDescriptors[msgRes.Info.FileOptions.Hex]; ok {
							continue
						}

						//создаем дескриптор файла для последующей записи в него
						f, err := os.Create(path.Join(pathDirStorage, msgRes.Info.FileOptions.Name))
						if err != nil {
							_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

							sdf.Status = "error"
							sdf.ErrMsg = err

							break DONE
						}

						listFileDescriptors[msgRes.Info.FileOptions.Hex] = f

						//обновляем информацию о задаче
						tpdf.smt.UpdateTaskDownloadAllParameters(tpdf.taskID, configure.DownloadTaskParameters{
							Status:                              "wait",
							NumberFilesTotal:                    ti.TaskParameter.DownloadTask.NumberFilesTotal,
							NumberFilesDownloaded:               ti.TaskParameter.DownloadTask.NumberFilesDownloaded,
							PathDirectoryStorageDownloadedFiles: ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles,
							FileInformation: configure.DetailedFileInformation{
								Name:         msgRes.Info.FileOptions.Name,
								Hex:          msgRes.Info.FileOptions.Hex,
								FullSizeByte: msgRes.Info.FileOptions.Size,
								NumChunk:     msgRes.Info.FileOptions.NumChunk,
								ChunkSize:    msgRes.Info.FileOptions.ChunkSize,
							},
						})

						msgReq.Info.Command = "ready to receive file"
						msgJSON, err := json.Marshal(msgReq)
						if err != nil {
							_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

							sdf.Status = "error"
							sdf.ErrMsg = err

							break DONE
						}

						//fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', SEND MSG '%v'\n", msgReq)

						tpdf.channels.cwtRes <- configure.MsgWsTransmission{
							DestinationHost: tpdf.sourceIP,
							Data:            &msgJSON,
						}

					//сообщение о невозможности передачи файла (slave -> master)
					case "file transfer not possible":
						dtp := ti.TaskParameter.DownloadTask
						dtp.NumberFilesDownloadedError = dtp.NumberFilesDownloadedError + 1

						//добавляем информацию о не принятом файле
						tpdf.smt.UpdateTaskDownloadAllParameters(tpdf.taskID, dtp)

						//отправляем информацию в Ядро
						tpdf.channels.chanInCore <- &configure.MsgBetweenCoreAndNI{
							TaskID:   tpdf.taskID,
							Section:  "download control",
							Command:  "file download process",
							SourceID: tpdf.sourceID,
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
						_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprintf("it is impossible to stop file transfer (source ID: %v, task ID: %v)", tpdf.sourceID, tpdf.taskID))

					}
				} else {
					_ = tpdf.saveMessageApp.LogMessage("error", "unknown generator events")

					break NEWFILE
				}
			}

			/* бинарный тип сообщения */
			if msg.MessageType == 2 {
				fileIsLoaded, err := writingBinaryFile(parametersWritingBinaryFile{
					SourceID:            tpdf.sourceID,
					TaskID:              tpdf.taskID,
					Data:                msg.Message,
					ListFileDescriptors: listFileDescriptors,
					SMT:                 tpdf.smt,
					ChanInCore:          tpdf.channels.chanInCore,
				})
				if err != nil {
					_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

					fmt.Printf("------ ERROR: %v\n", fmt.Sprint(err))

					sdf.Status = "error"
					sdf.ErrMsg = err

					break NEWFILE
				}

				//если файл полностью загружен запрашиваем следующий файл
				if fileIsLoaded {

					//fmt.Println("\t File success uploaded, REQUEST NEW FILE UPLOAD")

					//отправляем сообщение источнику подтверждающее успешный прием файла
					msgJSON, err := json.Marshal(configure.MsgTypeDownload{
						MsgType: "download files",
						Info: configure.DetailInfoMsgDownload{
							TaskID:  tpdf.taskID,
							Command: "file successfully accepted",
						},
					})
					if err != nil {
						_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

						break NEWFILE
					}

					tpdf.channels.cwtRes <- configure.MsgWsTransmission{
						DestinationHost: tpdf.sourceIP,
						Data:            &msgJSON,
					}

					break NEWFILE
				}
			}
		}
	}

	//задача завершена успешно
	msgToCore := configure.MsgBetweenCoreAndNI{
		TaskID:   tpdf.taskID,
		Section:  "download control",
		Command:  "task completed",
		SourceID: tpdf.sourceID,
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

	fmt.Printf("DOWNLOAD: func 'processorReceivingFiles', TASK COMPLITE MSG:%v\n", msgToCore)

	tpdf.channels.chanInCore <- &msgToCore
}

//writingBinaryFile осуществляет запись бинарного файла
func writingBinaryFile(pwbf parametersWritingBinaryFile) (bool, error) {
	//получаем хеш принимаемого файла
	fileHex := string((*pwbf.Data)[35:67])

	w, ok := pwbf.ListFileDescriptors[fileHex]
	if !ok {
		return false, fmt.Errorf("no file descriptor found for the specified hash %v (task ID %v, source ID %v)", fileHex, pwbf.TaskID, pwbf.SourceID)
	}

	//запись принятых байт
	numWriteByte, err := w.Write((*pwbf.Data)[67:])
	if err != nil {
		return false, err
	}

	ti, ok := pwbf.SMT.GetStoringMemoryTask(pwbf.TaskID)
	if !ok {
		return false, fmt.Errorf("task with ID %v not found", pwbf.TaskID)
	}

	fi := ti.TaskParameter.DownloadTask.FileInformation

	writeByte := fi.AcceptedSizeByte + int64(numWriteByte)

	//fmt.Printf("func 'writingBinaryFile', AcceptedSizeByte = %v, int64(numWriteByte) = %v, writeByte = %v\n", fi.AcceptedSizeByte, int64(numWriteByte), writeByte)
	//fmt.Printf("fi.FullSizeByte: %v\n", fi.FullSizeByte)

	writePercent := int(writeByte / (fi.FullSizeByte / 100))
	numAcceptedChunk := fi.NumAcceptedChunk + 1

	//обновляем информацию о принимаемом файле
	pwbf.SMT.UpdateTaskDownloadAllParameters(pwbf.TaskID, configure.DownloadTaskParameters{
		Status:                              "execute",
		NumberFilesTotal:                    ti.TaskParameter.DownloadTask.NumberFilesTotal,
		NumberFilesDownloaded:               ti.TaskParameter.DownloadTask.NumberFilesDownloaded,
		PathDirectoryStorageDownloadedFiles: ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles,
		FileInformation: configure.DetailedFileInformation{
			Name:                fi.Name,
			Hex:                 fi.Hex,
			FullSizeByte:        fi.FullSizeByte,
			AcceptedSizeByte:    writeByte,
			AcceptedSizePercent: writePercent,
			NumChunk:            fi.NumChunk,
			NumAcceptedChunk:    numAcceptedChunk,
		},
	})

	msgToCore := configure.MsgBetweenCoreAndNI{
		TaskID:   pwbf.TaskID,
		Section:  "download control",
		Command:  "file download process",
		SourceID: pwbf.SourceID,
	}

	//отправляем сообщение Ядру приложения, только если
	// процент увеличился на 1
	if (writePercent > fi.AcceptedSizePercent) && (writePercent != 100) {
		pwbf.ChanInCore <- &msgToCore
	}

	//если все кусочки были переданы (то есть файл считается полностью загруженым)
	if fi.NumChunk == numAcceptedChunk {
		//закрываем дескриптор файла
		w.Close()

		//удаляем дескриптор файла
		delete(pwbf.ListFileDescriptors, fi.Hex)

		filePath := path.Join(ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles, fi.Name)

		//проверяем хеш-сумму файла
		ok := checkDownloadedFile(filePath, fi.Hex, fi.FullSizeByte)
		if !ok {
			pwbf.SMT.IncrementNumberFilesDownloadedError(pwbf.TaskID)

			return false, fmt.Errorf("invalid checksum for file %v (task ID %v)", fi.Name, pwbf.TaskID)
		}

		msgToCore.Command = "file download complete"

		newFileInfo := configure.DownloadFilesInformation{
			IsLoaded:     true,
			TimeDownload: time.Now().Unix(),
		}
		newFileInfo.Size = fi.FullSizeByte
		newFileInfo.Hex = fi.Hex

		//отмечаем файл как успешно принятый
		pwbf.SMT.UpdateTaskDownloadFileIsLoaded(pwbf.TaskID, configure.DownloadTaskParameters{
			DownloadingFilesInformation: map[string]*configure.DownloadFilesInformation{
				fi.Name: &newFileInfo,
			},
		})

		//увеличиваем количество принятых файлов на 1
		pwbf.SMT.IncrementNumberFilesDownloaded(pwbf.TaskID)

		//fmt.Printf("func 'writingBinaryFile', file name:%v, success uploaded\n", fi.Name)

		pwbf.ChanInCore <- &msgToCore

		return true, nil
	}

	return false, nil
}

func checkDownloadedFile(pathFile, fileHex string, fileSize int64) bool {
	fs, fh, err := common.GetFileParameters(pathFile)
	if err != nil {
		return false
	}

	if (fs != fileSize) || (fh != fileHex) {
		return false
	}

	return true
}
