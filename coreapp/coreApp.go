package coreapp

/*
* Ядро приложения
*
* Версия 0.1, дата релиза 20.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//CoreApp запускает все обработчики уровня ядра
func CoreApp(appConf *configure.AppConfig, ism *configure.InformationStoringMemory, linkConnection *configure.MongoDBConnect) {
	fmt.Println("START module 'CoreAppMain'...")

	//запуск подпрограммы для маршрутизации запросов внутри приложения
	go Routing(appConf, ism)

	//запуск подпрограммы для взаимодействия с БД
	go DatabaseInteraction(appConf.ConnectionDB.NameDB, ism, linkConnection)
}
