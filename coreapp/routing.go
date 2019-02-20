package coreapp

/*
* Ядро приложения
* Маршрутизация сообщений получаемых через каналы
*
* Версия 0.1, дата релиза 20.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//Routing маршрутизирует данные поступающие в ядро из каналов
func Routing(appConf *configure.AppConfig, ism *configure.InformationStoringMemory) {
	select {
	case data := <-ism.ChannelCollection.ChanMessageToAPI:
		fmt.Println("MESSAGE FROM channel 'ChanMessageToAPI'")

		fmt.Println(data)

	case data := <-ism.ChannelCollection.ChanMessageFromAPI:
		fmt.Println("MESSAGE FROM channel 'ChanMessageFromAPI'")

		fmt.Println(data)
	}
}
