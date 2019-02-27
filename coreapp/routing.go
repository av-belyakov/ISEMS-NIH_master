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
func Routing(appConf *configure.AppConfig, ism *configure.InformationStoringMemory, cc *configure.ChannelCollectionCoreApp) {
	fmt.Println("START 'Route' module core app")

	for {
		select {
		case data := <-cc.InCoreChanDB:
			fmt.Println("MESSAGE FROM module DBInteraction")
			//использовать канал cc.OutCoreChanDB для ответа
			fmt.Println(data)

		case data := <-cc.InCoreChanAPI:
			fmt.Println("MESSAGE FROM module API")
			//использовать канал cc.OutCoreChanAPI для ответа
			fmt.Println(data)

		case data := <-cc.InCoreChanNI:
			fmt.Println("MESSAGE FROM module NetworkInteraction")
			//использовать канал cc.OutCoreChanNI для ответа
			fmt.Println(data)
		}
	}
}
