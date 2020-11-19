package coreapp

/*
* 				Ядро приложения
* Маршрутизация сообщений получаемых через каналы
* */

import (
	"fmt"
	"time"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/coreapp/handlerslist"
	"ISEMS-NIH_master/directorypathshaper"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//TypeRoutingCore тип для функции Routing
type TypeRoutingCore struct {
	AppConf                     *configure.AppConfig
	ChanColl                    *configure.ChannelCollectionCoreApp
	SMT                         *configure.StoringMemoryTask
	QTS                         *configure.QueueTaskStorage
	ISL                         *configure.InformationSourcesList
	TSSQ                        *configure.TemporaryStorageSearchQueries
	SaveMessageApp              *savemessageapp.PathDirLocationLogFiles
	ChanCheckTask               <-chan configure.MsgChanStoringMemoryTask
	ChanMsgInfoQueueTaskStorage <-chan configure.MessageInformationQueueTaskStorage
}

//Routing маршрутизирует данные поступающие в ядро из каналов
func Routing(trc TypeRoutingCore) {
	//при старте приложения запрашиваем список источников в БД
	trc.ChanColl.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
		MsgGenerator: "NI module",
		MsgRecipient: "DB module",
		MsgSection:   "source control",
		Instruction:  "find_all",
	}
	funcName := "Routing"

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
		for msg := range trc.ChanMsgInfoQueueTaskStorage {
			emt := handlerslist.ErrorMessageType{
				SourceID:    msg.SourceID,
				TaskID:      msg.TaskID,
				MsgType:     "danger",
				Instruction: "task processing",
				ChanToAPI:   trc.ChanColl.OutCoreChanAPI,
			}

			qti, err := trc.QTS.GetQueueTaskStorage(msg.SourceID, msg.TaskID)
			if err != nil {
				trc.SaveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				continue
			}

			emt.TaskIDClientAPI = qti.TaskIDClientAPI
			emt.IDClientAPI = qti.IDClientAPI

			si, ok := trc.ISL.GetSourceSetting(msg.SourceID)
			if !ok {
				trc.SaveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprintf("no information found on source ID %v", msg.SourceID),
					FuncName:    funcName,
				})

				//отправляем сообщение пользователю
				emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID: msg.SourceID,
					Message:  "не найдена информация по источнику",
				})

				if err := handlerslist.ErrorMessage(emt); err != nil {
					trc.SaveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

				//изменяем статус задачи в storingMemoryQueueTask
				// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
				// функцией 'CheckTimeQueueTaskStorage')
				if err := trc.QTS.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "complete"); err != nil {
					trc.SaveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

				continue
			}

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType: "info",
				Sources: []int{msg.SourceID},
			}

			if qti.TaskType == "filtration control" {
				emt.Section = "filtration control"

				//добавляем задачу в 'StoringMemoryTask'
				trc.SMT.AddStoringMemoryTask(msg.TaskID, configure.TaskDescription{
					ClientID:                        qti.IDClientAPI,
					ClientTaskID:                    qti.TaskIDClientAPI,
					UserName:                        qti.UserName,
					TaskType:                        "filtration control",
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
					TimeInterval: configure.TimeIntervalTaskExecution{
						Start: time.Now().Unix(),
						End:   time.Now().Unix(),
					},
					TaskParameter: configure.DescriptionTaskParameters{
						FiltrationTask: &configure.FiltrationTaskParameters{
							ID:     msg.SourceID,
							Status: "wait",
						},
						DownloadTask:                 &configure.DownloadTaskParameters{},
						ListFilesDetailedInformation: map[string]*configure.DetailedFilesInformation{},
					},
				})

				//сохраняем параметры задачи в БД
				trc.ChanColl.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
					MsgGenerator:    "Core module",
					MsgRecipient:    "DB module",
					MsgSection:      "filtration control",
					Instruction:     "insert",
					IDClientAPI:     qti.IDClientAPI,
					TaskID:          msg.TaskID,
					TaskIDClientAPI: qti.TaskIDClientAPI,
					AdvancedOptions: msg.SourceID,
				}

				ns.MsgDescription = common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   msg.SourceID,
					TaskType:   "фильтрация",
					TaskAction: "подготовка к выполнению задачи",
				})

				//отправляем информационное сообщение пользователю о начале выполнения задачи
				notifications.SendNotificationToClientAPI(trc.ChanColl.OutCoreChanAPI, ns, qti.TaskIDClientAPI, qti.IDClientAPI)
			}

			if qti.TaskType == "download control" {
				emt.Section = "download control"

				npfp := directorypathshaper.NecessaryParametersFiltrationProblem{
					SourceID:         msg.SourceID,
					SourceShortName:  si.ShortName,
					TaskID:           msg.TaskID,
					PathRoot:         trc.AppConf.DirectoryLongTermStorageDownloadedFiles.Raw,
					FiltrationOption: qti.TaskParameters.FilterationParameters,
				}

				//создаем директорию для хранения файлов и формируем файл README.xml с кратким описание задачи
				pathStorage, err := directorypathshaper.FileStorageDirectiry(&npfp)
				if err != nil {
					//отправляем сообщение пользователю
					emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID:   msg.SourceID,
						TaskType:   "скачивание файлов",
						TaskAction: "задача отклонена",
						Message:    "невозможно создать директорию для хранения файлов или запись скачиваемых файлов в созданную директорию невозможна",
					})

					if err := handlerslist.ErrorMessage(emt); err != nil {
						trc.SaveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}

					//изменяем статус задачи в storingMemoryQueueTask
					// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
					// функцией 'CheckTimeQueueTaskStorage')
					if err := trc.QTS.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "complete"); err != nil {
						trc.SaveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}

					continue
				}

				//добавляем задачу в 'StoringMemoryTask'
				trc.SMT.AddStoringMemoryTask(msg.TaskID, configure.TaskDescription{
					ClientID:                        qti.IDClientAPI,
					ClientTaskID:                    qti.TaskIDClientAPI,
					UserName:                        qti.UserName,
					TaskType:                        "download control",
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
					TimeInterval: configure.TimeIntervalTaskExecution{
						Start: time.Now().Unix(),
						End:   time.Now().Unix(),
					},
					TaskParameter: configure.DescriptionTaskParameters{
						FiltrationTask: &configure.FiltrationTaskParameters{
							PathStorageSource: qti.TaskParameters.PathDirectoryForFilteredFiles,
						},
						DownloadTask: &configure.DownloadTaskParameters{
							ID:                                  msg.SourceID,
							Status:                              "wait",
							NumberFilesTotal:                    len(qti.TaskParameters.ConfirmedListFiles),
							PathDirectoryStorageDownloadedFiles: pathStorage,
						},
					},
				})

				//добавляем список файлов которые необходимо выгрузить
				trc.SMT.UpdateListFilesDetailedInformation(msg.TaskID, qti.TaskParameters.ConfirmedListFiles)

				/*
				   Параметр TaskDescription.TaskParameter.DownloadTask.NumberFilesTotal
				   содержит общее кол-во файлов запрашиваемых пользователем или их
				   общее кол-во когда пользователь список не присылал. Данный параметр
				   может отличатся от аналогичного в таблице БД где он обозначает
				   общее кол-во файлов которые можно скачать, а не запрошенные пользователем
				*/

				ns.MsgDescription = common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   msg.SourceID,
					TaskType:   "скачивание файлов",
					TaskAction: "подготовка к выполнению задачи",
				})

				//отправляем информационное сообщение пользователю о начале выполнения задачи
				notifications.SendNotificationToClientAPI(trc.ChanColl.OutCoreChanAPI, ns, qti.TaskIDClientAPI, qti.IDClientAPI)

				//отправляем в NI module для вызова обработчика задания
				trc.ChanColl.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
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
		SMT:  trc.SMT,
		QTS:  trc.QTS,
		ISL:  trc.ISL,
		TSSQ: trc.TSSQ,
	}

	OutCoreChans := handlerslist.HandlerOutChans{
		OutCoreChanAPI: trc.ChanColl.OutCoreChanAPI,
		OutCoreChanDB:  trc.ChanColl.OutCoreChanDB,
		OutCoreChanNI:  trc.ChanColl.OutCoreChanNI,
	}

	//обработчик запросов от модулей приложения
	for {
		select {
		//CHANNEL FROM API
		case data := <-trc.ChanColl.InCoreChanAPI:
			go handlerslist.HandlerMsgFromAPI(OutCoreChans, data, hsm, trc.SaveMessageApp)

		//CHANNEL FROM DATABASE
		case data := <-trc.ChanColl.InCoreChanDB:
			go handlerslist.HandlerMsgFromDB(OutCoreChans, data, hsm, trc.AppConf.MaximumTotalSizeFilesDownloadedAutomatically, trc.SaveMessageApp, trc.ChanColl.ChanDropNI)

		//CHANNEL FROM NETWORK INTERACTION
		case data := <-trc.ChanColl.InCoreChanNI:
			//go handlerslist.HandlerMsgFromNI(OutCoreChans, data, hsm, saveMessageApp)
			if err := handlerslist.HandlerMsgFromNI(OutCoreChans, data, hsm); err != nil {
				trc.SaveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})
			}

		//сообщение клиенту API о том что задача с указанным ID долго выполняется
		case infoHungTask := <-trc.ChanCheckTask:
			if ti, ok := trc.SMT.GetStoringMemoryTask(infoHungTask.ID); ok {
				nsErrJSON := notifications.NotificationSettingsToClientAPI{
					MsgType:        infoHungTask.Type,
					MsgDescription: infoHungTask.Description,
				}

				notifications.SendNotificationToClientAPI(trc.ChanColl.OutCoreChanAPI, nsErrJSON, ti.ClientTaskID, ti.ClientID)
			}
		}
	}
}
