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
	fmt.Println("START 'Route' module core app")

	for {
		select {
		case data := <-ism.ChannelCollection.ChannelFromModuleAPI:
			fmt.Println("MESSAGE FROM channel 'ChannelFromModuleAPI'")

			fmt.Println(data)

		case data := <-ism.ChannelCollection.ChannelFromMNICommon:
			fmt.Println("MESSAGE FROM channel 'ChannelFromMNICommon'")

			fmt.Println(data)

		case data := <-ism.ChannelCollection.ChannelFromMNIService:
			fmt.Println("MESSAGE FROM channel 'ChannelFromMNIService'")

			fmt.Println(data)

			if data.Type == "get_list" {

				fmt.Println("/***************************************")
				fmt.Println("ДЛЯ СТАРТА HTTPS СЕРВЕРА НЕОБХОДИМ СПИСОК ИСТОЧНИКОВ")
				fmt.Println("***************************************/")

			}
		}
	}
}
