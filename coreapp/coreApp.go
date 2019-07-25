package coreapp

/*
* Ядро приложения
*
* Версия 0.11, дата релиза 27.02.2019
* */

import (
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduleapiapp"
	"ISEMS-NIH_master/moduledbinteraction"
	"ISEMS-NIH_master/modulenetworkinteractionapp"
)

//CoreApp запускает все обработчики уровня ядра
func CoreApp(appConf *configure.AppConfig, linkConnection *configure.MongoDBConnect) {
	//инициализация репозитория для учета выполняемых задач
	smt := configure.NewRepositorySMT()

	//инициализация репозитория для хранения очередей задач
	qts := configure.NewRepositoryQTS()

	//инициализация репозитория для хранения информации по источникам
	isl := configure.NewRepositoryISL()

	//инициализация отслеживания выполнения задач
	chanCheckTask := smt.CheckTimeUpdateStoringMemoryTask(55)

	//инициализация отслеживания очередности выполнения задач
	chanMsgInfoQueueTaskStorage := qts.CheckTimeQueueTaskStorage(isl, 3)

	//инициализация модуля для взаимодействия с БД
	chanOutCoreDB, chanInCoreDB := moduledbinteraction.MainDBInteraction(appConf.ConnectionDB.NameDB, linkConnection, smt, qts)

	//инициализация модуля для взаимодействия с API (обработчик внешних запросов)
	chanOutCoreAPI, chanInCoreAPI := moduleapiapp.MainAPIApp(appConf)

	//инициализация модуля сетевого взаимодействия (взаимодействие с сенсорами)
	chanOutCoreNI, chanInCoreNI := modulenetworkinteractionapp.MainNetworkInteraction(appConf, smt, qts, isl)

	chanColl := configure.ChannelCollectionCoreApp{
		OutCoreChanDB:  chanOutCoreDB,  //->БД
		InCoreChanDB:   chanInCoreDB,   //<-БД
		OutCoreChanAPI: chanOutCoreAPI, //->API
		InCoreChanAPI:  chanInCoreAPI,  //<-API
		OutCoreChanNI:  chanOutCoreNI,  //->NI
		InCoreChanNI:   chanInCoreNI,   //<-NI
	}

	//запуск подпрограммы для маршрутизации запросов внутри приложения
<<<<<<< HEAD
	Routing(appConf, &chanColl, smt, qts, chanCheckTask, chanMsgInfoQueueTaskStorage)
=======
	Routing(appConf, &chanColl, smt, qts, isl, chanCheckTask, chanMsgInfoQueueTaskStorage)
>>>>>>> ISEMS-NIH_master 06.08.2019
}
