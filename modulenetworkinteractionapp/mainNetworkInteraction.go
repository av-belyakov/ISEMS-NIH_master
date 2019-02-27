package modulenetworkinteractionapp

/*
* Модуль сетевого взаимодействия
* Выполняет следующие функции:
* - осуществляет взаимодействие с ядром приложения
* - осуществляет обмен данными и агригирацию данных получаемых от модулей wssServerNI и wssClientNI
* - выгружает файлы сет. трафика и объекты в долговременное хранилище
*
* Версия 0.11, дата релиза 27.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//MainNetworkInteraction осуществляет общее управление
func MainNetworkInteraction(appConf *configure.AppConfig, ism *configure.InformationStoringMemory) (chanOutCore, chanInCore chan configure.MsgBetweenCoreAndNI) {
	fmt.Println("START module 'MainNetworkInteraction'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//инициализируем канал для передачи данных через websocket соединение
	cwt := make(chan configure.MsgWsTransmission)

	//инициализируем каналы для передачи данных между ядром приложения и текущем модулем
	chanOutCore = make(chan configure.MsgBetweenCoreAndNI)
	chanInCore = make(chan configure.MsgBetweenCoreAndNI)

	//инициализация каналов управления и состояния источников
	chansStatSource := map[string]chan [2]string{
		"inWssModuleClient":  make(chan [2]string),
		"outWssModuleServer": make(chan [2]string),
		"outWssModuleClient": make(chan [2]string),
	}

	//маршрутизат запросов получаемых от CoreApp
	go RouteCoreRequest(cwt, ism, chanOutCore, chansStatSource)
	//маршрутизат запросов получаемых Wss
	go RouteWssConnectionResponse(cwt, chanInCore, ism)

	//запуск модуля wssServerNI
	go WssServerNetworkInteraction(chansStatSource["outWssModuleServer"], appConf, ism)

	//запуск модуля wssClientNI
	go WssClientNetworkInteraction(chansStatSource["outWssModuleClient"], appConf.TimeReconnectClient, ism, chansStatSource["inWssModule"])

	go func() {
		for msg := range cwt {
			sourceIP, data := msg.DestinationHost, msg.Data
			if conn, ok := ism.GetLinkWebsocketConnect(sourceIP); ok {
				if err := conn.SendWsMessage(1, data); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}
			}
		}
	}()

	return chanOutCore, chanInCore
}
