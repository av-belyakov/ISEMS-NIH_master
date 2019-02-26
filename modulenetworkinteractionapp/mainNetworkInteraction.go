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
	"ISEMS-NIH_master/savemessageapp"
)

/*
cOut chan<- configure.MessageNetworkInteraction
cIn <-chan configure.MessageNetworkInteraction
ccss chan struct{}
*/

//MainNetworkInteraction осуществляет общее управление
func MainNetworkInteraction(appConf *configure.AppConfig, ism *configure.InformationStoringMemory, chanColl *configure.ChannelCollection) {
	fmt.Println("START module 'MainNetworkInteraction'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//изменение статуса источника
	changeStatusSourceToWssModule := make(chan struct{})
	changeStatusSourceFromWssModule := make(chan map[string]string)
	//запуск маршрутизатора запросов получаемых с источников
	go Routing(changeStatusSourceToWssModule, chanColl, ism, changeStatusSourceFromWssModule)

	//запуск модуля wssServerNI
	go WssServerNetworkInteraction(changeStatusSourceFromWssModule, appConf, ism)

	//запуск модуля wssClientNI
	go WssClientNetworkInteraction(changeStatusSourceFromWssModule, appConf.TimeReconnectClient, ism, changeStatusSourceToWssModule)

	go func() {
		for msg := range chanColl.Cwt {
			sourceIP, data := msg.DestinationHost, msg.Data
			if conn, ok := ism.GetLinkWebsocketConnect(sourceIP); ok {
				if err := conn.SendWsMessage(1, data); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}
			}
		}
	}()
}
