package coreapp

/*
* Ядро приложения
* */

import (
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduleapiapp"
	"ISEMS-NIH_master/moduledbinteraction"
	"ISEMS-NIH_master/modulenetworkinteractionapp"
	"ISEMS-NIH_master/savemessageapp"
)

//CoreApp запускает все обработчики уровня ядра
func CoreApp(appConf *configure.AppConfig, linkConnection *configure.MongoDBConnect, saveMessageApp *savemessageapp.PathDirLocationLogFiles) {

	//------------- ДЕЛАЕМ ДАМП ПАМЯТИ ---------------------
	/*ticker := time.NewTicker(time.Duration(3) * time.Second)

	num := 0
	go func() {
		for range ticker.C {
			s := strconv.Itoa(num)
			logFileName := fmt.Sprintf("memdumpfile_%v.memdump", s)

			//fmt.Printf("Write memdump to file %v\n", logFileName)

			fl, err := os.Create(logFileName)
			if err != nil {
				fmt.Printf("Create file %v, error: %v\n", logFileName, fmt.Sprint(err))
			}

			pprof.Lookup("heap").WriteTo(fl, 0)

			fl.Close()
		}
	}()*/
	//------------------------------------------------------

	//инициализация репозитория для учета выполняемых задач
	smt := configure.NewRepositorySMT()

	//инициализация репозитория для хранения очередей задач
	qts := configure.NewRepositoryQTS(saveMessageApp)

	//инициализация репозитория для хранения информации по источникам
	isl := configure.NewRepositoryISL()

	//инициализация репозитория для кэширования информации по поиску задач в БД
	// TickerSec - интервал проверки информации
	// TimeExpiration - время устаревания кэша в сек.
	// MaxCacheSize - кол-во записей в кэше
	tssq := configure.NewRepositoryTSSQ(configure.TypeRepositoryTSSQ{
		TickerSec:      3,
		TimeExpiration: 15,
	})

	//инициализация отслеживания выполнения задач
	chanCheckTask := smt.CheckTimeUpdateStoringMemoryTask(55)

	//инициализация отслеживания очередности выполнения задач
	chanMsgInfoQueueTaskStorage := qts.CheckTimeQueueTaskStorage(isl, 1, saveMessageApp)

	//инициализация модуля для взаимодействия с БД
	chanOutCoreDB, chanInCoreDB := moduledbinteraction.MainDBInteraction(appConf.ConnectionDB.NameDB, linkConnection, smt, qts, tssq, saveMessageApp)

	//инициализация модуля для взаимодействия с API (обработчик внешних запросов)
	chanOutCoreAPI, chanInCoreAPI := moduleapiapp.MainAPIApp(appConf, saveMessageApp)

	//инициализация модуля сетевого взаимодействия (взаимодействие с сенсорами)
	chanOutCoreNI, chanInCoreNI := modulenetworkinteractionapp.MainNetworkInteraction(appConf, smt, qts, isl, saveMessageApp)

	chanColl := configure.ChannelCollectionCoreApp{
		OutCoreChanDB:  chanOutCoreDB,  //->БД
		InCoreChanDB:   chanInCoreDB,   //<-БД
		OutCoreChanAPI: chanOutCoreAPI, //->API
		InCoreChanAPI:  chanInCoreAPI,  //<-API
		OutCoreChanNI:  chanOutCoreNI,  //->NI
		InCoreChanNI:   chanInCoreNI,   //<-NI
	}

	//запуск подпрограммы для маршрутизации запросов внутри приложения
	Routing(TypeRoutingCore{
		AppConf:                     appConf,
		ChanColl:                    &chanColl,
		SMT:                         smt,
		QTS:                         qts,
		ISL:                         isl,
		TSSQ:                        tssq,
		SaveMessageApp:              saveMessageApp,
		ChanCheckTask:               chanCheckTask,
		ChanMsgInfoQueueTaskStorage: chanMsgInfoQueueTaskStorage,
	})
}
