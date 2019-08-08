package coreapp

/*
* Ядро приложения
* Маршрутизация сообщений получаемых через каналы
*
* Версия 0.4, дата релиза 01.08.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/coreapp/handlerslist"
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
		//инициализируем функцию конструктор для записи лог-файлов
		saveMessageApp := savemessageapp.New()

		for msg := range chanMsgInfoQueueTaskStorage {
			qti, err := qts.GetQueueTaskStorage(msg.SourceID, msg.TaskID)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				continue
			}

			if qti.TaskType == "filteration" {
				/*

				   ФИльтрацию переделаем позже, после выполнения части
				   раздела по выгрузки файлов

				*/
			}

			if qti.TaskType == "download" {
				/*
					   После предыдущих манипуляций в QueueTaskSorage
					   есть вся информация о задаче на выполнение

					   теперь выполнить следующие действия

					    - создать директорию для хранения обработанных файлов (и обработать ошибки)
						информацию по фильтрации взять из QUeueTaskStatrage

						- добавить задачу в StoringMemoryTask
					    - запустить обработчик для выполнений действий по скачиванию файлов

				*/

				//создание директорий куда будут сохранятся скачанные файлы
				/*pathStorageDirectory, err := directorypathshaper.CreatePathDirectory()
				if err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					//отправить сообщение пользователю
					nsErrJSON := notifications.NotificationSettingsToClientAPI{
						MsgType:        "danger",
						MsgDescription: "Внутренняя ошибка, невозможно создать директории для сохранения скаченных файлов",
					}

					notifications.SendNotificationToClientAPI(cc.OutCoreChanAPI, nsErrJSON, qti.TaskIDClientAPI, qti.IDClientAPI)

					//удалить всю информацию о задаче из очереди
					if e := qts.DelQueueTaskStorage(msg.SourceID, msg.TaskID); e != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(e))
					}

					continue
				}

				/*
				   Сделать и протестировать модуль создания вложеных директорий для
				   хранения скаченных файлов
				*/

				//добавление новой задачи в StoringMemoryTask

				//изменение значений в таблице БД (статуса задачи и пути сохранения файлов)

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
			go handlerslist.HandlerMsgFromDB(OutCoreChans, data, hsm, cc.ChanDropNI)

		//CHANNEL FROM API
		case data := <-cc.InCoreChanAPI:
			go handlerslist.HandlerMsgFromAPI(OutCoreChans, data, hsm)

		//CHANNEL FROM NETWORK INTERACTION
		case data := <-cc.InCoreChanNI:
			go handlerslist.HandlerMsgFromNI(OutCoreChans, data, hsm)

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
