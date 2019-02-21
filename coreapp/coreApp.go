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
func CoreApp(appConf *configure.AppConfig, ism *configure.InformationStoringMemory, linkConnection *configure.MongoDBConnect) {
	fmt.Println("START module 'CoreAppMain'...")

	//запуск подпрограммы для взаимодействия с БД
	go DatabaseInteraction(appConf.ConnectionDB.NameDB, ism, linkConnection)

	//инициализация модуля для взаимодействия с API (обработчик внешних запросов)
	go moduleapiapp.MainAppAPI(appConf, ism)

	//инициализация модуля сетевого взаимодействия (взаимодействие с сенсорами)
	go modulenetworkinteractionapp.MainNetworkInteraction(appConf, ism)

	//запуск подпрограммы для маршрутизации запросов внутри приложения
	Routing(appConf, ism)
}
