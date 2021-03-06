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
	list    map[string]ListFileDescriptionOptions
	chanReq chan channelReqSettings
	chanRes chan channelResSettings
}

//ListFileDescriptionOptions хранит опции файловых дескрипторов
// fileIsFullyAccepted - принят ли полностью файл
// fileDescription - файловый дескриптор
type ListFileDescriptionOptions struct {
	fileIsFullyAccepted bool
	fileDescription     *os.File
}

type channelReqSettings struct {
	command, fileHex, filePath string
}

type channelResSettings struct {
	fileStatus bool
	fd         *os.File
	err        error
}

type typeWriteBinaryFileRes struct {
	fileName, fileHex            string
	fileIsLoaded, fileIsSlowDown bool
}

//NewListFileDescription создание нового репозитория со списком дескрипторов файлов
func NewListFileDescription() *ListFileDescription {
	lfd := ListFileDescription{}
	lfd.list = map[string]ListFileDescriptionOptions{}
	lfd.chanReq = make(chan channelReqSettings)
	lfd.chanRes = make(chan channelResSettings)

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
						lfd.list[msg.fileHex] = ListFileDescriptionOptions{fileDescription: f}
					}
				}

				lfd.chanRes <- crs

			case "get":
				crs := channelResSettings{}
				fd, ok := lfd.list[msg.fileHex]
				if !ok {
					crs.err = fmt.Errorf("func 'NewListFileDescription', command: 'get', file descriptor with ID '%v' not found", msg.fileHex)
				} else {
					crs.fd = fd.fileDescription
				}

				lfd.chanRes <- crs

			case "get status accepted":
				crs := channelResSettings{}
				fd, ok := lfd.list[msg.fileHex]
				if !ok {
					crs.err = fmt.Errorf("func 'NewListFileDescription', command: 'get status accepted', file status accepted with ID '%v' not found", msg.fileHex)
				} else {
					crs.fileStatus = fd.fileIsFullyAccepted
				}

				lfd.chanRes <- crs

			case "close":
				if fd, ok := lfd.list[msg.fileHex]; ok {
					fd.fileDescription.Close()
					fd.fileIsFullyAccepted = true
				}

				lfd.chanRes <- channelResSettings{}

			case "del all":
				for fileHex := range lfd.list {
					delete(lfd.list, fileHex)
				}

				lfd.chanRes <- channelResSettings{}
			}
		}
	}()

	return &lfd
}

func (lfd *ListFileDescription) addFileDescription(fh, fp string) error {
	lfd.chanReq <- channelReqSettings{
		command:  "add",
		fileHex:  fh,
		filePath: fp,
	}

	return (<-lfd.chanRes).err
}

func (lfd *ListFileDescription) getFileDescription(fh string) (*os.File, error) {
	lfd.chanReq <- channelReqSettings{
		command: "get",
		fileHex: fh,
	}

	res := <-lfd.chanRes

	return res.fd, res.err
}

func (lfd *ListFileDescription) getStatusFileFullyAccepted(fh string) (bool, error) {
	lfd.chanReq <- channelReqSettings{
		command: "get status accepted",
		fileHex: fh,
	}

	res := <-lfd.chanRes

	return res.fileStatus, res.err
}

func (lfd *ListFileDescription) closeFileDescription(fh string) {
	lfd.chanReq <- channelReqSettings{
		command: "close",
		fileHex: fh,
	}

	<-lfd.chanRes
}

func (lfd *ListFileDescription) delAllFileDescription() {
	lfd.chanReq <- channelReqSettings{
		command: "del all",
	}

	<-lfd.chanRes
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

	clf, err := smt.GetCountListFilesDetailedInformation(taskID)
	if err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    "processorReceivingFiles",
		})

		//fmt.Printf("task with ID '%v' not found the list of files suitable for downloading from the source is empty", taskID)

		return chanOut, chanDone, fmt.Errorf("task with ID '%v' not found the list of files suitable for downloading from the source is empty", taskID)
	}

	if clf == 0 {

		//fmt.Printf("the list of files suitable for downloading from the source is empty, for task with ID '%v'", taskID)

		return chanOut, chanDone, fmt.Errorf("the list of files suitable for downloading from the source is empty, for task with ID '%v'", taskID)
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
	funcName := "processingDownloadFile"

	//задача завершена успешно
	msgToCore := configure.MsgBetweenCoreAndNI{
		TaskID:   tpdf.taskID,
		Section:  "download control",
		Command:  "task completed",
		SourceID: tpdf.sourceID,
	}

	sdf := statusDownloadFile{Status: "success"}
	lfd := NewListFileDescription()

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

	/*fmt.Printf("func '%v', READING LIST FILES\n", funcName)
	for fn, fi := range tpdf.taskInfo.TaskParameter.ListFilesDetailedInformation {
		fmt.Printf("func '%v', files name: %v, size: '%v', hex: '%v'\n", funcName, fn, fi.Size, fi.Hex)
	}*/

	lf, ok := tpdf.smt.GetListFilesDetailedInformation(tpdf.taskID)
	if !ok {
		tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprintf("for the specified task ID '%v', the list of files intended for downloading was not found", tpdf.taskID),
			FuncName:    funcName,
		})

		msgToCore.Command = "task stoped error"
		tpdf.channels.chanDone <- struct{}{}

		return
	}

DONE:
	//читаем список файлов
	for fn, fi := range lf {

		//fmt.Printf("func '%v', first request to downlod file name: '%v'\n", funcName, fn)

		//делаем первый запрос на скачивание файла
		mtd.Info.Command = "give me the file"
		mtd.Info.FileOptions = configure.DownloadFileOptions{
			Name: fn,
			Size: fi.Size,
			Hex:  fi.Hex,
		}

		//fmt.Printf("func '%v', reguest to slave -> : %v\n", funcName, mtd.Info)

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

			//fmt.Printf("func '%v', TimerUpdateStoringMemoryTask to downlod file name: '%v'\n", funcName, fn)

			/*
				Такое впечатление что задача удаляется по таймауту
					   из лога

					   2020-12-17 09:31:09.03540284 [+0300 MSK] - file descriptor with ID 'cc446c8fc7cb597d4a729b421392bde8' not found (function 'processingDownloadFile')
					   2020-12-17 09:31:09.04517288 [+0300 MSK] - not action 'send data', task ID '2e1c7c351bafa585ce92f28fdc22e1ea' not found (function 'ControllerReceivingRequestedFiles')

			*/

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

						//fmt.Printf("func '%v', received messgae 'stop receiving files' file name: '%v'\n", funcName, fn)

						//отмечаем задачу как находящуюся в процессе останова
						tpdf.smt.IsSlowDownStoringMemoryTask(tpdf.taskID)

						msgReq.Info.Command = "stop receiving files"

						msgJSON, err := json.Marshal(&msgReq)
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

						//fmt.Printf("func '%v', received messgae 'to stop the task because of a disconnection' file name: '%v'\n", funcName, fn)

						//отмечаем задачу как находящуюся в процессе останова
						tpdf.smt.IsSlowDownStoringMemoryTask(tpdf.taskID)

						//закрываем дескриптор файла и удаляем файл
						lfd.closeFileDescription(fi.Hex)
						//lfd.delFileDescription(fi.Hex)
						_ = os.Remove(path.Join(pathDirStorage, fn))

						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprintf("the task with ID '%v' was stopped, because the network connection was broken", tpdf.taskID),
							FuncName:    funcName,
						})

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
							Description: fmt.Sprintf("task with ID '%v' not found", tpdf.taskID),
							FuncName:    funcName,
						})

						sdf.Status = "error"
						sdf.ErrMsg = err

						break DONE
					}

					switch msgRes.Info.Command {
					//готовность к передаче файла (slave -> master)
					case "ready for the transfer":

						//fmt.Printf("func '%v', received response 'ready for the transfer' for file name: '%v'\n", funcName, fn)

						//если задача находится в стадии останова игнорировать ответ slave
						if ti.IsSlowDown {
							continue
						}

						//fmt.Printf("func '%v', create file descriptor for file name: '%v'\n", funcName, fn)

						//создаем дескриптор файла для последующей записи в него
						if err := lfd.addFileDescription(msgRes.Info.FileOptions.Hex, path.Join(pathDirStorage, msgRes.Info.FileOptions.Name)); err != nil {
							tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
								Description: fmt.Sprint(err),
								FuncName:    funcName,
							})

							break NEWFILE
						}

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
						msgJSON, err := json.Marshal(&msgReq)
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

						//fmt.Printf("func '%v', received response 'file transfer not possible' for file name: '%v'\n", funcName, fn)
						//fmt.Printf("func '%v', received response '%v'\n", funcName, msgRes.Info)

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

						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprintf("message received '%v' from the source %v of the task ID '%v', file name:'%v', file hex: '%v'", msgRes.Info.Command, tpdf.sourceID, tpdf.taskID, fn, fi.Hex),
							FuncName:    funcName,
						})

						break NEWFILE

					//сообщение об успешном останове передачи файла (slave -> master)
					case "file transfer stopped successfully":
						//закрываем дескриптор файла и удаляем файл
						lfd.closeFileDescription(fi.Hex)
						//lfd.delFileDescription(fi.Hex)
						_ = os.Remove(path.Join(pathDirStorage, fn))

						sdf.Status = "task stoped client"

						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprintf("message received '%v' from the source %v of the task ID '%v'", msgRes.Info.Command, tpdf.sourceID, tpdf.taskID),
							FuncName:    funcName,
						})

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

					//fmt.Printf("func '%v', unknown generator events!!! for file name: '%v'\n", funcName, fn)

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
				}, tpdf.saveMessageApp)

				if writeBinaryFileResult.fileIsLoaded && !writeBinaryFileResult.fileIsSlowDown {
					//lfd.delFileDescription(fi.Hex)

					//fmt.Printf("func '%v', writeBinaryFileResult.fileIsLoaded: '%v' || writeBinaryFileResult.fileLoadedError: '%v', !writeBinaryFileResult.fileIsSlowDown: '%v'\n", funcName, writeBinaryFileResult.fileIsLoaded, writeBinaryFileResult.fileLoadedError, !writeBinaryFileResult.fileIsSlowDown)

					msgJSON, err := json.Marshal(&configure.MsgTypeDownload{
						MsgType: "download files",
						Info: configure.DetailInfoMsgDownload{
							TaskID:  tpdf.taskID,
							Command: "file successfully accepted",
						},
					})
					if err != nil {
						tpdf.saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					} else {
						tpdf.channels.cwtRes <- configure.MsgWsTransmission{
							DestinationHost: tpdf.sourceIP,
							Data:            &msgJSON,
						}
					}

					//закрываем дескриптор файла
					lfd.closeFileDescription(fi.Hex)

					break NEWFILE
				}
			}
		}
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

	//очищаем всю информацию о принятых файлах
	lfd.delAllFileDescription()

	tpdf.channels.chanInCore <- &msgToCore
	tpdf.channels.chanDone <- struct{}{}
}

//writingBinaryFile осуществляет запись бинарного файла
func writingBinaryFile(pwbf parametersWritingBinaryFile, saveMessageApp *savemessageapp.PathDirLocationLogFiles) *typeWriteBinaryFileRes {
	funcName := "writingBinaryFile"

	var fileHex string
	var fi configure.DetailedFileInformation
	var twbfr typeWriteBinaryFileRes

	/*defer func(fileHex *string, fi *configure.DetailedFileInformation, twbfr *typeWriteBinaryFileRes) {
		(*fileHex) = ""
		(*fi) = configure.DetailedFileInformation{}
		(*twbfr) = typeWriteBinaryFileRes{}
	}(&fileHex, &fi, &twbfr)*/

	//получаем хеш принимаемого файла
	fileHex = string((*pwbf.Data)[35:67])
	twbfr.fileHex = fileHex

	sf, err := pwbf.LFD.getStatusFileFullyAccepted(fileHex)
	if err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})

		return &twbfr
	}

	if sf {
		twbfr.fileIsLoaded = true

		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprintf("file already loaded, fileIsLoaded: '%v'", sf),
			FuncName:    funcName,
		})

		return &twbfr
	}

	w, err := pwbf.LFD.getFileDescription(fileHex)
	if err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})

		return &twbfr
	}

	//запись принятых байт
	numWriteByte, err := w.Write((*pwbf.Data)[67:])
	if err != nil {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprint(err),
			FuncName:    funcName,
		})

		return &twbfr
	}

	ti, ok := pwbf.SMT.GetStoringMemoryTask(pwbf.TaskID)
	if !ok {
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: fmt.Sprintf("task with ID '%v' not found", pwbf.TaskID),
			FuncName:    funcName,
		})

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
		//увеличиваем количество принятых файлов на 1
		pwbf.SMT.IncrementNumberFilesDownloaded(pwbf.TaskID)

		filePath := path.Join(ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles, fi.Name)

		msgToCore.Command = "file download complete"

		//проверяем хеш-сумму файла
		ok := checkDownloadedFile(filePath, fi.Hex, fi.FullSizeByte)
		if !ok {
			pwbf.SMT.IncrementNumberFilesDownloadedError(pwbf.TaskID)

			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprintf("the checksum value for the downloaded file '%v' is incorrect (task ID '%v')", fi.Name, pwbf.TaskID),
				FuncName:    funcName,
			})

			pwbf.ChanInCore <- &msgToCore
			twbfr.fileIsLoaded = true

			return &twbfr
		}

		ndfi := &configure.DetailedFilesInformation{
			Hex:          fi.Hex,
			Size:         fi.FullSizeByte,
			IsLoaded:     true,
			TimeDownload: time.Now().Unix(),
		}

		//отмечаем файл как успешно принятый
		pwbf.SMT.UpdateListFilesDetailedInformationFileIsLoaded(pwbf.TaskID, map[string]*configure.DetailedFilesInformation{fi.Name: ndfi})

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
