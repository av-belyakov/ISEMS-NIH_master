package coreapp

/*
* Ядро приложения
*
* Версия 0.11, дата релиза 27.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduleapiapp"
	"ISEMS-NIH_master/moduledbinteraction"
	"ISEMS-NIH_master/modulenetworkinteractionapp"
)

//CoreApp запускает все обработчики уровня ядра
func CoreApp(appConf *configure.AppConfig, linkConnection *configure.MongoDBConnect) {
	fmt.Println("START module 'CoreAppMain'...")

	//инициализация хранилища для учета выполняемых задач
	smt := configure.NewRepositorySMT()

	//инициализируем отслеживания выполнения задач
	chanCheckTask := smt.CheckTimeUpdateStoringMemoryTask(55)

	//запуск подпрограммы для взаимодействия с БД
	chanOutCoreDB, chanInCoreDB := moduledbinteraction.MainDBInteraction(appConf.ConnectionDB.NameDB, linkConnection, smt)

	//инициализация модуля для взаимодействия с API (обработчик внешних запросов)
	chanOutCoreAPI, chanInCoreAPI := moduleapiapp.MainAPIApp(appConf)

	//инициализация модуля сетевого взаимодействия (взаимодействие с сенсорами)
	chanOutCoreNI, chanInCoreNI := modulenetworkinteractionapp.MainNetworkInteraction(appConf, smt)

	chanColl := configure.ChannelCollectionCoreApp{
		OutCoreChanDB:  chanOutCoreDB,  //->БД
		InCoreChanDB:   chanInCoreDB,   //<-БД
		OutCoreChanAPI: chanOutCoreAPI, //->API
		InCoreChanAPI:  chanInCoreAPI,  //<-API
		OutCoreChanNI:  chanOutCoreNI,  //->NI
		InCoreChanNI:   chanInCoreNI,   //<-NI
	}

	//запуск подпрограммы для маршрутизации запросов внутри приложения
	Routing(appConf, &chanColl, smt, chanCheckTask)
}
