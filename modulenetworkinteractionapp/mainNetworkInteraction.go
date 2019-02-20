package modulenetworkinteractionapp

/*
* Модуль сетевого взаимодействия
* Выполняет следующие функции:
* - осуществляет взаимодействие с ядром приложения
* - осуществляет обмен данными и агригирацию данных получаемых от модулей wssServerNI и wssClientNI
* - выгружает файлы сет. трафика и объекты в долговременное хранилище
*
* Версия 0.1, дата релиза 20.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//MainNetworkInteraction осуществляет общее управление
func MainNetworkInteraction(appConf *configure.AppConfig, ism *configure.InformationStoringMemory) {
	fmt.Println("START module 'MainNetworkInteraction'...")

	//запуск модуля wssServerNI
	go WssServerNetworkInteraction(appConf, ism)

	//запуск модуля wssClientNI
	go WssClientNetworkInteraction(appConf.TimeReconnectClient, ism)

	//обработка данных получаемых через каналы
	/*	select {

		}*/
}
