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
	fmt.Println("START ROUTE module CoreApp")

	fmt.Printf("%T %v\n", cc, cc)

	fmt.Println("----------")
	//запрашиваем у API новый список источников
	cc.OutCoreChanAPI <- configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "from API",
		MsgType:      "command",
		DataType:     "source_control",
		IDClientAPI:  "",
		AdvancedOptions: configure.MsgCommandSourceControl{
			ListRequest: true,
		},
	}

	fmt.Println("ROUTE CoreApp sending data to channel")

	for {
		select {
		case data := <-cc.InCoreChanDB:
			fmt.Println("MESSAGE FROM module DBInteraction")
			//использовать канал cc.OutCoreChanDB для ответа
			fmt.Println(data)

		case data := <-cc.InCoreChanAPI:
			fmt.Println("MESSAGE FROM module API")
			//использовать канал cc.OutCoreChanAPI для ответа
			fmt.Printf("%T %v\n", data, data)
			fmt.Println("ДАЛЕЕ НУЖНО ОБРАБОТАТЬ И ПЕРЕДАТь через канал модулю БД")

		case data := <-cc.InCoreChanNI:
			fmt.Println("MESSAGE FROM module NetworkInteraction")
			//использовать канал cc.OutCoreChanNI для ответа
			fmt.Println(data)
		}
	}
}
