package coreapp

/*
* Ядро приложения
* Маршрутизация сообщений получаемых через каналы
*
* Версия 0.5, дата релиза 13.08.2019
* */

import (
	"fmt"
	"time"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/coreapp/handlerslist"
	"ISEMS-NIH_master/directorypathshaper"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//Routing маршрутизирует данные поступающие в ядро из каналов
func Routing(
	appConf *configure.AppConfig,
	cc *configure.ChannelCollectionCoreApp,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	isl *configure.InformationSourcesList,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanCheckTask <-chan configure.MsgChanStoringMemoryTask,
	chanMsgInfoQueueTaskStorage <-chan configure.MessageInformationQueueTaskStorage) {

	//при старте приложения запрашиваем список источников в БД
	cc.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
		MsgGenerator: "NI module",
		MsgRecipient: "DB module",
		MsgSection:   "source control",
		Instruction:  "find_all",
	}

	/*
		const logFileName = "memdumpfile"

		fl, err := os.Create(logFileName)
		if err != nil {
			fmt.Printf("Create file %v, error: %v\n", logFileName, fmt.Sprint(err))
		}
		defer fl.Close()

		pprof.Lookup("heap").WriteTo(fl, 0)
	*/

	//обработчик модуля очереди ожидающих задач QueueTaskStorage
	go func() {
		for msg := range chanMsgInfoQueueTaskStorage {
			emt := handlerslist.ErrorMessageType{
				SourceID:    msg.SourceID,
				TaskID:      msg.TaskID,
				MsgType:     "danger",
				Instruction: "task processing",
				ChanToAPI:   cc.OutCoreChanAPI,
			}

			qti, err := qts.GetQueueTaskStorage(msg.SourceID, msg.TaskID)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				continue
			}

			emt.TaskIDClientAPI = qti.TaskIDClientAPI
			emt.IDClientAPI = qti.IDClientAPI

			si, ok := isl.GetSourceSetting(msg.SourceID)
			if !ok {
				_ = saveMessageApp.LogMessage("error", fmt.Sprintf("no information found on source ID %v", msg.SourceID))

				//отправляем сообщение пользователю
				emt.MsgHuman = fmt.Sprintf("Не найдена информация по источнику с ID %v", msg.SourceID)
				if err := handlerslist.ErrorMessage(emt); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}

				//удаляем задачу из очереди
				if err := qts.DelQueueTaskStorage(msg.SourceID, msg.TaskID); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}

				continue
			}

			if qti.TaskType == "filteration" {
				emt.Section = "filtration control"
				/*

				   ФИльтрацию переделаем позже, после выполнения части
				   раздела по выгрузки файлов

				*/
			}

			if qti.TaskType == "download" {
				emt.Section = "download control"

				npfp := directorypathshaper.NecessaryParametersFiltrationProblem{
					SourceID:         msg.SourceID,
					SourceShortName:  si.ShortName,
					TaskID:           msg.TaskID,
					PathRoot:         appConf.DirectoryLongTermStorageDownloadedFiles.Raw,
					FiltrationOption: qti.TaskParameters.FilterationParameters,
				}

				pathStorage, err := directorypathshaper.FileStorageDirectiry(&npfp)
				if err != nil {
					//отправляем сообщение пользователю
					emt.MsgHuman = "Невозможно создать директорию для хранения файлов или запись скачиваемых файлов в созданную директорию невозможен"
					if err := handlerslist.ErrorMessage(emt); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					//удаляем задачу из очереди
					if err := qts.DelQueueTaskStorage(msg.SourceID, msg.TaskID); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					continue
				}

				//добавляем задачу в StorageMemoryTask
				smt.AddStoringMemoryTask(msg.TaskID, configure.TaskDescription{
					ClientID:                        qti.IDClientAPI,
					ClientTaskID:                    qti.TaskIDClientAPI,
					TaskType:                        "download control",
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
					TimeInterval: configure.TimeIntervalTaskExecution{
						Start: time.Now().Unix(),
						End:   time.Now().Unix(),
					},
					TaskParameter: configure.DescriptionTaskParameters{
						DownloadTask: configure.DownloadTaskParameters{
							ID:                                  msg.SourceID,
							Status:                              "wait",
							PathDirectoryStorageDownloadedFiles: pathStorage,
							DownloadingFilesInformation:         qti.TaskParameters.ConfirmedListFiles,
						},
					},
				})

				//изменяем статус задачи в StoringMemoryQueueTask
				if err := qts.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "execution"); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}
				/*
				   Удалять нельзя, понадобится при обработке разрыва соединения

				   				//удаляем из StoringMemoryQueueTask списки файлов
				   if err := qts.ClearAllListFiles(msg.SourceID, msg.TaskID); err != nil {
				   					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				   				}
				*/

				//отправляем в NI module для вызова обработчика задания
				cc.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:     msg.TaskID,
					ClientName: si.ClientName,
					Section:    "download control",
					Command:    "start",
					SourceID:   msg.SourceID,
				}
			}
		}
	}()

	hsm := handlerslist.HandlersStoringMemory{
		SMT: smt,
		QTS: qts,
		ISL: isl,
	}

	OutCoreChans := handlerslist.HandlerOutChans{
		OutCoreChanAPI: cc.OutCoreChanAPI,
		OutCoreChanDB:  cc.OutCoreChanDB,
		OutCoreChanNI:  cc.OutCoreChanNI,
	}

	//обработчик запросов от модулей приложения
	for {
		select {
		//CHANNEL FROM DATABASE
		case data := <-cc.InCoreChanDB:
			go handlerslist.HandlerMsgFromDB(OutCoreChans, data, hsm, saveMessageApp, cc.ChanDropNI)

		//CHANNEL FROM API
		case data := <-cc.InCoreChanAPI:
			go handlerslist.HandlerMsgFromAPI(OutCoreChans, data, hsm, saveMessageApp)

		//CHANNEL FROM NETWORK INTERACTION
		case data := <-cc.InCoreChanNI:
			go handlerslist.HandlerMsgFromNI(OutCoreChans, data, hsm, saveMessageApp)

		//сообщение клиенту API о том что задача с указанным ID долго выполняется
		case infoHungTask := <-chanCheckTask:
			if ti, ok := smt.GetStoringMemoryTask(infoHungTask.ID); ok {
				nsErrJSON := notifications.NotificationSettingsToClientAPI{
					MsgType:        infoHungTask.Type,
					MsgDescription: infoHungTask.Description,
				}

				notifications.SendNotificationToClientAPI(cc.OutCoreChanAPI, nsErrJSON, ti.ClientTaskID, ti.ClientID)
			}
		}
	}
}
