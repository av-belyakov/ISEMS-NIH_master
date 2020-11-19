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
	SourceID   int
	TaskID     string
	Data       *[]byte
	LFD        *ListFileDescription
	SMT        *configure.StoringMemoryTask
	ChanInCore chan<- *configure.MsgBetweenCoreAndNI
}

type typeProcessingDownloadFile struct {
	sourceID       int
	sourceIP       string
	taskID         string
	taskInfo       *configure.TaskDescription
	smt            *configure.StoringMemoryTask
	saveMessageApp *savemessageapp.PathDirLocationLogFiles
	channels       listChannels
}

type listChannels struct {
	chanInCore chan<- *configure.MsgBetweenCoreAndNI
	chanOut    <-chan MsgChannelProcessorReceivingFiles
	cwtRes     chan<- configure.MsgWsTransmission
	chanDone   chan<- struct{}
}

//ListFileDescription хранит список файловых дескрипторов и канал для доступа к ним
type ListFileDescription struct {
	list    map[string]*os.File
	chanReq chan channelReqSettings
}

type channelReqSettings struct {
	command, fileHex, filePath string
	channelRes                 chan channelResSettings
}

type channelResSettings struct {
	fd  *os.File
	err error
}

//NewListFileDescription создание нового репозитория со списком дескрипторов файлов
func NewListFileDescription() *ListFileDescription {
	lfd := ListFileDescription{}
	lfd.list = map[string]*os.File{}
	lfd.chanReq = make(chan channelReqSettings)

	go func() {
		for msg := range lfd.chanReq {
			switch msg.command {
			case "add":
				crs := channelResSettings{}
				if _, ok := lfd.list[msg.fileHex]; !ok {
					f, err := os.Create(msg.filePath)
					if err != nil {
						crs.err = err
					} else {
						lfd.list[msg.fileHex] = f
					}
				}

				msg.channelRes <- crs

				close(msg.channelRes)

			case "get":
				crs := channelResSettings{}
				fd, ok := lfd.list[msg.fileHex]
				if !ok {
					crs.err = fmt.Errorf("file descriptor with ID '%v' not found", msg.fileHex)
				} else {
					crs.fd = fd
				}

				msg.channelRes <- crs

				close(msg.channelRes)

			case "del":
				if fd, ok := lfd.list[msg.fileHex]; !ok {
					fd.Close()
				}

				delete(lfd.list, msg.fileHex)

				msg.channelRes <- channelResSettings{}

				close(msg.channelRes)
			}
		}
	}()

	return &lfd
}

func (lfd *ListFileDescription) addFileDescription(fh, fp string) error {
	chanRes := make(chan channelResSettings)

	lfd.chanReq <- channelReqSettings{
		command:    "add",
		fileHex:    fh,
		filePath:   fp,
		channelRes: chanRes,
	}

	return (<-chanRes).err
}

func (lfd *ListFileDescription) getFileDescription(fh string) (*os.File, error) {
	chanRes := make(chan channelResSettings)

	lfd.chanReq <- channelReqSettings{
		command:    "get",
		fileHex:    fh,
		channelRes: chanRes,
	}

	res := <-chanRes

	return res.fd, res.err
}

func (lfd *ListFileDescription) delFileDescription(fh string) {
	chanRes := make(chan channelResSettings)

	lfd.chanReq <- channelReqSettings{
		command:    "del",
		fileHex:    fh,
		channelRes: chanRes,
	}

	<-chanRes
}

//processorReceivingFiles управляет приемом файлов в рамках одной задачи
func processorReceivingFiles(
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	sourceIP, taskID string,
	smt *configure.StoringMemoryTask,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	cwtRes chan<- configure.MsgWsTransmission) (chan MsgChannelProcessorReceivingFiles, chan struct{}, error) {

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

	ti, ok := smt.GetStoringMemoryTask(taskID)
	if !ok {
		return nil, nil, fmt.Errorf("task with ID %v not found", taskID)
	}

	sourceID := ti.TaskParameter.DownloadTask.ID

	chanOut := make(chan MsgChannelProcessorReceivingFiles)
	chanDone := make(chan struct{})

	//отправляем информационное сообщение пользователю
	chanInCore <- &configure.MsgBetweenCoreAndNI{
		TaskID:  taskID,
		Section: "message notification",
		Command: "send client API",
		AdvancedOptions: configure.MessageNotification{
			SourceReport:                 "NI module",
			Section:                      "download files",
			TypeActionPerformed:          "start",
			CriticalityMessage:           "info",
			HumanDescriptionNotification: "идет подготовка списка скачиваемых файлов",
			Sources:                      []int{sourceID},
		},
	}

	//проверяем наличие файлов для скачивания
	if len(ti.TaskParameter.ListFilesDetailedInformation) == 0 {
		return chanOut, chanDone, fmt.Errorf("the list of files suitable for downloading from the source is empty")
	}

	go processingDownloadFile(typeProcessingDownloadFile{
		sourceID:       sourceID,
		sourceIP:       sourceIP,
		taskID:         taskID,
		taskInfo:       ti,
		smt:            smt,
		saveMessageApp: saveMessageApp,
		channels: listChannels{
			chanInCore: chanInCore,
			chanOut:    chanOut,
			cwtRes:     cwtRes,
			chanDone:   chanDone,
		},
	})

	return chanOut, chanDone, nil
}

func processingDownloadFile(tpdf typeProcessingDownloadFile) {
	sdf := statusDownloadFile{Status: "success"}
	lfd := NewListFileDescription()
	funcName := "processingDownloadFile"

	//начальный запрос на передачу файла
	mtd := configure.MsgTypeDownload{
		MsgType: "download files",
		Info: configure.DetailInfoMsgDownload{
			TaskID:         tpdf.taskID,
			PathDirStorage: tpdf.taskInfo.TaskParameter.FiltrationTask.PathStorageSource,
		},
	}
	defer func(mtd configure.MsgTypeDownload) {
		mtd = configure.MsgTypeDownload{}
	}(mtd)

	pathDirStorage := tpdf.taskInfo.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles

DONE:
	//читаем список файлов
	for fn, fi := range tpdf.taskInfo.TaskParameter.ListFilesDetailedInformation {
		//делаем первый запрос на скачивание файла
		mtd.Info.Command = "give me the file"
		mtd.Info.FileOptions = configure.DownloadFileOptions{
			Name: fn,
			Size: fi.Size,
			Hex:  fi.Hex,
		}

		msgJSON, err := json.Marshal(&mtd)
		if err != nil {
			tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(err),
				FuncName:    funcName,
			})

			continue
		}

		tpdf.channels.cwtRes <- configure.MsgWsTransmission{
			DestinationHost: tpdf.sourceIP,
			Data:            &msgJSON,
		}

	NEWFILE:
		for msg := range tpdf.channels.chanOut {
			//обновляем значение таймера (что бы задача не была удалена по таймауту)
			tpdf.smt.TimerUpdateStoringMemoryTask(tpdf.taskID)

			/* текстовый тип сообщения */
			if msg.MessageType == 1 {
				msgReq := configure.MsgTypeDownload{
					MsgType: "download files",
					Info: configure.DetailInfoMsgDownload{
						TaskID: tpdf.taskID,
					},
				}

				if msg.MsgGenerator == "Core module" {
					command := string(*msg.Message)

					//остановить скачивание файлов
					if command == "stop receiving files" {
						//отмечаем задачу как находящуюся в процессе останова
						tpdf.smt.IsSlowDownStoringMemoryTask(tpdf.taskID)

						msgReq.Info.Command = "stop receiving files"

						msgJSON, err := json.Marshal(msgReq)
						if err != nil {
							tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
								Description: fmt.Sprint(err),
								FuncName:    funcName,
							})

							continue
						}

						//отправляем команду останов на slave и ждем подтверждения останова
						tpdf.channels.cwtRes <- configure.MsgWsTransmission{
							DestinationHost: tpdf.sourceIP,
							Data:            &msgJSON,
						}
					}

					//разрыв соединения (остановить загрузку файлов)
					if command == "to stop the task because of a disconnection" {
						//отмечаем задачу как находящуюся в процессе останова
						tpdf.smt.IsSlowDownStoringMemoryTask(tpdf.taskID)

						//закрываем дескриптор файла и удаляем файл
						lfd.delFileDescription(fi.Hex)
						_ = os.Remove(path.Join(pathDirStorage, fn))

						sdf.Status = "task stoped disconnect"

						break DONE
					}

				} else if msg.MsgGenerator == "NI module" {
					var msgRes configure.MsgTypeDownload
					if err := json.Unmarshal(*msg.Message, &msgRes); err != nil {
						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})

						continue
					}

					/* получаем информацию о задаче */
					ti, ok := tpdf.smt.GetStoringMemoryTask(tpdf.taskID)
					if !ok {
						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprintf("task with ID %v not found", tpdf.taskID),
							FuncName:    funcName,
						})

						sdf.Status = "error"
						sdf.ErrMsg = err

						break DONE
					}

					switch msgRes.Info.Command {
					//готовность к передаче файла (slave -> master)
					case "ready for the transfer":
						//если задача находится в стадии останова игнорировать ответ slave
						if ti.IsSlowDown {
							break
						}

						//создаем дескриптор файла для последующей записи в него
						lfd.addFileDescription(msgRes.Info.FileOptions.Hex, path.Join(pathDirStorage, msgRes.Info.FileOptions.Name))

						dtp := configure.DownloadTaskParameters{
							Status:                              "wait",
							NumberFilesTotal:                    ti.TaskParameter.DownloadTask.NumberFilesTotal,
							NumberFilesDownloaded:               ti.TaskParameter.DownloadTask.NumberFilesDownloaded,
							NumberFilesDownloadedError:          ti.TaskParameter.DownloadTask.NumberFilesDownloadedError,
							PathDirectoryStorageDownloadedFiles: ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles,
							FileInformation: configure.DetailedFileInformation{
								Name:         msgRes.Info.FileOptions.Name,
								Hex:          msgRes.Info.FileOptions.Hex,
								FullSizeByte: msgRes.Info.FileOptions.Size,
								NumChunk:     msgRes.Info.FileOptions.NumChunk,
								ChunkSize:    msgRes.Info.FileOptions.ChunkSize,
							},
						}
						//обновляем информацию о задаче
						tpdf.smt.UpdateTaskDownloadAllParameters(tpdf.taskID, &dtp)
						dtp = configure.DownloadTaskParameters{}

						msgReq.Info.Command = "ready to receive file"
						msgJSON, err := json.Marshal(msgReq)
						if err != nil {
							tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
								Description: fmt.Sprint(err),
								FuncName:    funcName,
							})

							sdf.Status = "error"
							sdf.ErrMsg = err

							break DONE
						}

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

					//сообщение об успешном останове передачи файла (slave -> master)
					case "file transfer stopped successfully":
						//закрываем дескриптор файла и удаляем файл
						lfd.delFileDescription(fi.Hex)
						_ = os.Remove(path.Join(pathDirStorage, fn))

						sdf.Status = "task stoped client"

						break DONE

					//невозможно остановить передачу файла (slave -> master)
					case "impossible to stop file transfer":
						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprintf("it is impossible to stop file transfer (source ID: %v, task ID: %v)", tpdf.sourceID, tpdf.taskID),
							FuncName:    funcName,
						})

					default:
						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprintf("received unknown command ('%v')\n", msgRes.Info.Command),
							FuncName:    funcName,
						})
					}
				} else {
					tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: "unknown generator events",
						FuncName:    funcName,
					})

					break NEWFILE
				}
			}

			/* бинарный тип сообщения */
			if msg.MessageType == 2 {
				writeBinaryFileResult := writingBinaryFile(parametersWritingBinaryFile{
					SourceID:   tpdf.sourceID,
					TaskID:     tpdf.taskID,
					Data:       msg.Message,
					LFD:        lfd,
					SMT:        tpdf.smt,
					ChanInCore: tpdf.channels.chanInCore,
				})
				if writeBinaryFileResult.err != nil {
					tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					sdf.Status = "error"
					sdf.ErrMsg = err

					break DONE
				}

				if (writeBinaryFileResult.fileIsLoaded || writeBinaryFileResult.fileLoadedError) && !writeBinaryFileResult.fileIsSlowDown {
					msgRes := configure.MsgTypeDownload{
						MsgType: "download files",
						Info: configure.DetailInfoMsgDownload{
							TaskID:  tpdf.taskID,
							Command: "file successfully accepted",
						},
					}

					//если файл загружен полностью но контрольная сумма не совпадает
					if writeBinaryFileResult.fileLoadedError {
						msgRes.Info.Command = "file received with error"

						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprintf("the checksum value for the downloaded file '%v' is incorrect (task ID %v)", writeBinaryFileResult.fileName, tpdf.taskID),
							FuncName:    funcName,
						})
					}

					msgJSON, err := json.Marshal(&msgRes)
					if err != nil {
						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})

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

	tpdf.channels.chanInCore <- &msgToCore
	tpdf.channels.chanDone <- struct{}{}
}

type typeWriteBinaryFileRes struct {
	fileName                                      string
	fileIsLoaded, fileLoadedError, fileIsSlowDown bool
	err                                           error
}

//writingBinaryFile осуществляет запись бинарного файла
func writingBinaryFile(pwbf parametersWritingBinaryFile) *typeWriteBinaryFileRes {
	twbfr := typeWriteBinaryFileRes{}

	/*  очищаем для отладки  */
	var fileHex string
	var fi configure.DetailedFileInformation

	func(fileHex *string, fi *configure.DetailedFileInformation) {
		(*fileHex) = ""
		(*fi) = configure.DetailedFileInformation{}
	}(&fileHex, &fi)

	//получаем хеш принимаемого файла
	fileHex = string((*pwbf.Data)[35:67])

	w, err := pwbf.LFD.getFileDescription(fileHex)
	if err != nil {
		twbfr.err = err

		return &twbfr
	}

	//запись принятых байт
	numWriteByte, err := w.Write((*pwbf.Data)[67:])
	if err != nil {
		twbfr.err = err

		return &twbfr
	}

	ti, ok := pwbf.SMT.GetStoringMemoryTask(pwbf.TaskID)
	if !ok {
		twbfr.err = fmt.Errorf("task with ID %v not found", pwbf.TaskID)
		return &twbfr
	}

	twbfr.fileName = ti.TaskParameter.DownloadTask.FileInformation.Name
	twbfr.fileIsSlowDown = ti.IsSlowDown

	fi = ti.TaskParameter.DownloadTask.FileInformation

	writeByte := fi.AcceptedSizeByte + int64(numWriteByte)

	wp := float64(writeByte) / (float64(fi.FullSizeByte) / 100)
	writePercent := int(wp)
	numAcceptedChunk := fi.NumAcceptedChunk + 1

	msgToCore := configure.MsgBetweenCoreAndNI{
		TaskID:   pwbf.TaskID,
		Section:  "download control",
		Command:  "file download process",
		SourceID: pwbf.SourceID,
		AdvancedOptions: configure.MoreFileInformation{
			Hex:                 fi.Hex,
			AcceptedSizeByte:    writeByte,
			AcceptedSizePercent: writePercent,
		},
	}

	//отправляем сообщение Ядру приложения, только если
	// процент увеличился на 1
	if (writePercent > fi.AcceptedSizePercent) && (writePercent != 100) {
		pwbf.ChanInCore <- &msgToCore
	}

	dtp := configure.DownloadTaskParameters{
		Status:                              "execute",
		NumberFilesTotal:                    ti.TaskParameter.DownloadTask.NumberFilesTotal,
		NumberFilesDownloaded:               ti.TaskParameter.DownloadTask.NumberFilesDownloaded,
		NumberFilesDownloadedError:          ti.TaskParameter.DownloadTask.NumberFilesDownloadedError,
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
	}
	//обновляем информацию о принимаемом файле
	pwbf.SMT.UpdateTaskDownloadAllParameters(pwbf.TaskID, &dtp)
	dtp = configure.DownloadTaskParameters{}

	//если все кусочки были переданы (то есть файл считается полностью загруженым)
	if fi.NumChunk == numAcceptedChunk {
		//закрываем и удаляем дескриптор файла
		pwbf.LFD.delFileDescription(fi.Hex)

		//увеличиваем количество принятых файлов на 1
		pwbf.SMT.IncrementNumberFilesDownloaded(pwbf.TaskID)

		filePath := path.Join(ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles, fi.Name)

		msgToCore.Command = "file download complete"

		//проверяем хеш-сумму файла
		ok := checkDownloadedFile(filePath, fi.Hex, fi.FullSizeByte)
		if !ok {
			pwbf.SMT.IncrementNumberFilesDownloadedError(pwbf.TaskID)

			pwbf.ChanInCore <- &msgToCore

			twbfr.fileLoadedError = true

			return &twbfr
		}

		ndfi := configure.DetailedFilesInformation{
			Hex:          fi.Hex,
			Size:         fi.FullSizeByte,
			IsLoaded:     true,
			TimeDownload: time.Now().Unix(),
		}

		//отмечаем файл как успешно принятый
		pwbf.SMT.UpdateListFilesDetailedInformationFileIsLoaded(pwbf.TaskID, map[string]*configure.DetailedFilesInformation{
			fi.Name: &ndfi,
		})

		pwbf.ChanInCore <- &msgToCore

		twbfr.fileIsLoaded = true
		return &twbfr
	}

	return &twbfr
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
