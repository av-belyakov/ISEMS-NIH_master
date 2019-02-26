package coreapp

/*
* Ядро приложения
*
* Версия 0.1, дата релиза 20.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/moduleapiapp"
	"ISEMS-NIH_master/modulenetworkinteractionapp"
)

//CoreApp запускает все обработчики уровня ядра
func CoreApp(appConf *configure.AppConfig, linkConnection *configure.MongoDBConnect, ism *configure.InformationStoringMemory, chanColl *configure.ChannelCollection) {
	fmt.Println("START module 'CoreAppMain'...")

	//запуск подпрограммы для взаимодействия с БД
	go DatabaseInteraction(appConf.ConnectionDB.NameDB, linkConnection, ism)

	//инициализация модуля для взаимодействия с API (обработчик внешних запросов)
	go moduleapiapp.MainAppAPI(chanColl.ChannelFromModuleAPI, appConf, ism, chanColl.ChannelToModuleAPI)

	//инициализация модуля сетевого взаимодействия (взаимодействие с сенсорами)
	go modulenetworkinteractionapp.MainNetworkInteraction(appConf, ism, chanColl)

	//запуск подпрограммы для маршрутизации запросов внутри приложения
	Routing(appConf, ism, chanColl)
}
