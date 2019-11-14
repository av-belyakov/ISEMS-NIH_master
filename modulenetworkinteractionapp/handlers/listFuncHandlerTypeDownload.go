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
	taskInfo       configure.TaskDescription
	smt            *configure.StoringMemoryTask
	saveMessageApp *savemessageapp.PathDirLocationLogFiles
	channels       listChannels
}

type listChannels struct {
	chanInCore  chan<- *configure.MsgBetweenCoreAndNI
	chanOutCore <-chan MsgChannelProcessorReceivingFiles
	cwtRes      chan<- configure.MsgWsTransmission
	chanDone    chan<- struct{}
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

					fmt.Printf("NewListFileDescription (get) count list = %v\n", len(lfd.list))

					crs.err = fmt.Errorf("file descriptor with ID '%v' not found", msg.fileHex)
				} else {
					crs.fd = fd
				}

				msg.channelRes <- crs

				close(msg.channelRes)

			case "del":
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
	sourceID int,
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

	fmt.Println("\tDOWNLOAD: func 'processorReceivingFiles', START...")

	ti, ok := smt.GetStoringMemoryTask(taskID)
	if !ok {

		fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', task with ID %v not found\n", taskID)

		return nil, nil, fmt.Errorf("task with ID %v not found", taskID)
	}

	chanOut := make(chan MsgChannelProcessorReceivingFiles)
	chanDone := make(chan struct{})

	//fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', '%v'\n", ti)

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
			chanInCore:  chanInCore,
			chanOutCore: chanOut,
			cwtRes:      cwtRes,
			chanDone:    chanDone,
		},
	})

	return chanOut, chanDone, nil
}

func processingDownloadFile(tpdf typeProcessingDownloadFile) {
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

		fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', make ONE request to download file, %v\n", mtd)

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
					command := string(*msg.Message)

					fmt.Printf("func 'listFuncHandlerTypeDownload', RESIVED COMMAND: '%v'\n", command)

					//остановить скачивание файлов
					if command == "stop receiving files" {

						//отмечаем задачу как находящуюся в процессе останова
						tpdf.smt.IsSlowDownStoringMemoryTask(tpdf.taskID)

						taskDesc, _ := tpdf.smt.GetStoringMemoryTask(tpdf.taskID)

						fmt.Printf("func 'listFuncHandlerTypeDownload', SECTION 'stop receiving files' IsSlowDown ---> New IsSlowDown '%v'\n", taskDesc.IsSlowDown)

						msgReq.Info.Command = "stop receiving files"

						msgJSON, err := json.Marshal(msgReq)
						if err != nil {
							_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

							continue
						}

						fmt.Printf("func 'listFuncHandlerTypeDownload', SEND ---> SLAVE MSG:%v\n", msgReq)

						//отправляем команду останов на slave и ждем подтверждения останова
						tpdf.channels.cwtRes <- configure.MsgWsTransmission{
							DestinationHost: tpdf.sourceIP,
							Data:            &msgJSON,
						}

						//закрываем дескриптор файла и удаляем файл
						/*lfd.delFileDescription(fi.Hex)
						_ = os.Remove(path.Join(pathDirStorage, fn))

						sdf.Status = "task stoped client"

						break DONE*/
					}

					//разрыв соединения (остановить загрузку файлов)
					if command == "to stop the task because of a disconnection" {
						//закрываем дескриптор файла и удаляем файл
						lfd.delFileDescription(fi.Hex)
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

					switch msgRes.Info.Command {
					//готовность к передаче файла (slave -> master)
					case "ready for the transfer":

						fmt.Printf("------ func 'processingDownloadFile' RESEIVING MSG: 'ready for the transfer' MSG:'%v'\n", msgRes)

						//если задача находится в стадии останова игнорировать ответ slave
						if ti.IsSlowDown {
							break
						}

						//создаем дескриптор файла для последующей записи в него
						lfd.addFileDescription(msgRes.Info.FileOptions.Hex, path.Join(pathDirStorage, msgRes.Info.FileOptions.Name))

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

						fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', SECTION:'ready for the transfer' SEND MSG '%v'\n", msgReq)

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

						fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', SECTION:'file transfer stopped successfully' SEND MSG '%v'\n", msgReq)

						//закрываем дескриптор файла и удаляем файл
						lfd.delFileDescription(fi.Hex)
						_ = os.Remove(path.Join(pathDirStorage, fn))

						sdf.Status = "task stoped client"

						break DONE

					//невозможно остановить передачу файла (slave -> master)
					case "impossible to stop file transfer":
						_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprintf("it is impossible to stop file transfer (source ID: %v, task ID: %v)", tpdf.sourceID, tpdf.taskID))

					default:

						fmt.Printf("\tDOWNLOAD: func 'processorReceivingFiles', RESEIVED COMMAND ------+++++ %v -----+++\n", msgRes.Info.Command)

						_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprintf("received unknown command ('%v')\n", msgRes.Info.Command))

					}
				} else {
					_ = tpdf.saveMessageApp.LogMessage("error", "unknown generator events")

					break NEWFILE
				}
			}

			/* бинарный тип сообщения */
			if msg.MessageType == 2 {
				fileIsLoaded, err := writingBinaryFile(parametersWritingBinaryFile{
					SourceID:   tpdf.sourceID,
					TaskID:     tpdf.taskID,
					Data:       msg.Message,
					LFD:        lfd,
					SMT:        tpdf.smt,
					ChanInCore: tpdf.channels.chanInCore,
				})
				if err != nil {
					_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprint(err))

					fmt.Printf("------ ERROR: %v\n", fmt.Sprint(err))
					/*

					   !!!! Непонятно !!!
					   Почему приходит эта ошибка ------ ERROR: file descriptor with ID '5714cb3ecf81013a0ab15160e9d9a17a' not found
					   Часто возникает после останова и возобновления задачи по скачиванию
					   файлов

					   Посмотреть что отправляет slave с ответом 'ready for the transfer'
					*/
					sdf.Status = "error"
					sdf.ErrMsg = err

					break DONE
				}

				//обновляем информацию о задаче (что она находится в стадии останова)
				ti, ok := tpdf.smt.GetStoringMemoryTask(tpdf.taskID)
				if !ok {
					_ = tpdf.saveMessageApp.LogMessage("error", fmt.Sprintf("task with ID %v not found", tpdf.taskID))

					sdf.Status = "error"
					sdf.ErrMsg = err

					break DONE
				}

				//если файл полностью загружен и задача не находится в
				// стадии выполнения запрашиваем следующий файл
				if fileIsLoaded && !ti.IsSlowDown {

					fmt.Printf("\t\tDOWNLOAD: func 'processorReceivingFiles', File success uploaded, REQUEST NEW FILE UPLOAD, TASK STATUS:'%v'\n", ti.TaskParameter.DownloadTask.Status)

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
	tpdf.channels.chanDone <- struct{}{}
}

//writingBinaryFile осуществляет запись бинарного файла
func writingBinaryFile(pwbf parametersWritingBinaryFile) (bool, error) {
	//получаем хеш принимаемого файла
	fileHex := string((*pwbf.Data)[35:67])

	w, err := pwbf.LFD.getFileDescription(fileHex)
	if err != nil {

		fmt.Printf("+-+-+-++-+ func 'writingBinaryFile', FILE HEX '%v' NOT FOUND\n", fileHex)
		fmt.Printf("__________ Hex string: '%v' ____________\n", string((*pwbf.Data)[:100]))

		return false, err
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

	wp := writeByte / (fi.FullSizeByte / 100)
	writePercent := int(wp)
	numAcceptedChunk := fi.NumAcceptedChunk + 1

	if numAcceptedChunk == 1 || numAcceptedChunk == 2 {
		fmt.Printf("\t---*** Full file size: '%v', write byte: '%v', sum write byte: '%v', all count chunks: '%v', accepted chunk: '%v' PERCENT: '%v'\n", fi.FullSizeByte, numWriteByte, writeByte, fi.NumChunk, numAcceptedChunk, wp)
	}
	//Full file size: '277', write byte: '277', sum write byte: '277', all count chunks: '1', accepted chunk: '1' PERCENT: '138'
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
		pwbf.LFD.delFileDescription(fi.Hex)

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
