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
				fmt.Printf("function 'routing' Core module - ERROR %v", err)

				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				continue
			}

			fmt.Printf("function 'routing' Core module - sent new task type %v\n", qti.TaskType)
			fmt.Println(qti)

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

				//изменяем статус задачи в storingMemoryQueueTask
				// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
				// функцией 'CheckTimeQueueTaskStorage')
				if err := qts.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "complete"); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}

				continue
			}

			if qti.TaskType == "filtration control" {
				emt.Section = "filtration control"

				//добавляем задачу в 'StoringMemoryTask'
				smt.AddStoringMemoryTask(msg.TaskID, configure.TaskDescription{
					ClientID:                        qti.IDClientAPI,
					ClientTaskID:                    qti.TaskIDClientAPI,
					TaskType:                        "filtration control",
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
					TimeInterval: configure.TimeIntervalTaskExecution{
						Start: time.Now().Unix(),
						End:   time.Now().Unix(),
					},
					TaskParameter: configure.DescriptionTaskParameters{
						FiltrationTask: configure.FiltrationTaskParameters{
							ID:     msg.SourceID,
							Status: "wait",
						},
					},
				})

				//сохраняем параметры задачи в БД
				cc.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
					MsgGenerator:    "Core module",
					MsgRecipient:    "DB module",
					MsgSection:      "filtration control",
					Instruction:     "insert",
					IDClientAPI:     qti.IDClientAPI,
					TaskID:          msg.TaskID,
					TaskIDClientAPI: qti.TaskIDClientAPI,
					AdvancedOptions: msg.SourceID,
				}

				fmt.Println("function 'routing' Core module - add task FILTRATION in StoringMemoryTask and send insert DB module")
			}

			if qti.TaskType == "download control" {
				emt.Section = "download control"

				npfp := directorypathshaper.NecessaryParametersFiltrationProblem{
					SourceID:         msg.SourceID,
					SourceShortName:  si.ShortName,
					TaskID:           msg.TaskID,
					PathRoot:         appConf.DirectoryLongTermStorageDownloadedFiles.Raw,
					FiltrationOption: qti.TaskParameters.FilterationParameters,
				}

				//создаем директорию для хранения файлов и формируем файл README.xml с кратким описание задачи
				pathStorage, err := directorypathshaper.FileStorageDirectiry(&npfp)
				if err != nil {
					//отправляем сообщение пользователю
					emt.MsgHuman = "Невозможно создать директорию для хранения файлов или запись скачиваемых файлов в созданную директорию невозможен"
					if err := handlerslist.ErrorMessage(emt); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					//изменяем статус задачи в storingMemoryQueueTask
					// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
					// функцией 'CheckTimeQueueTaskStorage')
					if err := qts.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "complete"); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}

					continue
				}

				fmt.Printf("function 'routing' Core module - Создали директорию '%v' для хранения файлов при скачивании (task ID %v)\n", pathStorage, msg.TaskID)

				//изменяем статус задачи в StoringMemoryQueueTask
				/*
					if err := qts.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "execution"); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}
				*/

				//добавляем задачу в 'StoringMemoryTask'
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
						FiltrationTask: configure.FiltrationTaskParameters{
							PathStorageSource: qti.TaskParameters.PathDirectoryForFilteredFiles,
						},
						DownloadTask: configure.DownloadTaskParameters{
							ID:                                  msg.SourceID,
							Status:                              "wait",
							NumberFilesTotal:                    len(qti.TaskParameters.ConfirmedListFiles),
							PathDirectoryStorageDownloadedFiles: pathStorage,
							DownloadingFilesInformation:         qti.TaskParameters.ConfirmedListFiles,
						},
					},
				})
				/*
				   Параметр TaskDescription.TaskParameter.DownloadTask.NumberFilesTotal
				   содержит общее кол-во файлов запрашиваемых пользователем или их
				   общее кол-во когда пользователь список не присылал. Данный параметр
				   может отличатся от аналогичного в таблице БД где он обозночает
				   общее кол-во файлов которые можно скачать, а не запрошенные пользователем
				*/

				nit, _ := smt.GetStoringMemoryTask(msg.TaskID)

				fmt.Printf("function 'routing' Core module - добавили задачу по скачиванию (task ID %v) в StoringMemoryTask: '%v'\n", msg.TaskID, nit)

				//отправляем в NI module для вызова обработчика задания
				cc.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:     msg.TaskID,
					ClientName: si.ClientName,
					Section:    "download control",
					Command:    "start",
					SourceID:   msg.SourceID,
				}

				fmt.Println("function 'routing' Core module - function 'routing' Core module - add task DOWNLOAD in StoringMemoryTask and send NI module")
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
			go handlerslist.HandlerMsgFromDB(OutCoreChans, data, hsm, appConf.MaximumTotalSizeFilesDownloadedAutomatically, saveMessageApp, cc.ChanDropNI)

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
