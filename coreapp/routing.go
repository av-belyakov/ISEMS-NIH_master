package coreapp

/*
* Ядро приложения
* Маршрутизация сообщений получаемых через каналы
*
* Версия 0.3, дата релиза 13.03.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/common/notifications"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/coreapp/handlerslist"
	"ISEMS-NIH_master/savemessageapp"
)

//Routing маршрутизирует данные поступающие в ядро из каналов
func Routing(appConf *configure.AppConfig, cc *configure.ChannelCollectionCoreApp, smt *configure.StoringMemoryTask) {
	fmt.Println("START ROUTE module 'CoreApp'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//при старте приложения запрашиваем у БД список источников
	cc.OutCoreChanDB <- configure.MsgBetweenCoreAndDB{
		MsgGenerator: "NI module",
		MsgRecipient: "DB module",
		MsgDirection: "request",
		MsgSection:   "sources_control",
		Instruction:  "find_all",
	}

	for {
		select {
		//CHANNEL FROM DATABASE
		case data := <-cc.InCoreChanDB:
			fmt.Println("MESSAGE FROM module DBInteraction")

			handlerslist.HandlerMsgFromDB(cc.OutCoreChanAPI, &data, cc.OutCoreChanNI)

		//CHANNEL FROM API
		case data := <-cc.InCoreChanAPI:

			fmt.Println("MESSAGE FROM module API")

			if taskID, ns, err := handlerslist.HandlerMsgFromAPI(cc.OutCoreChanNI, &data, smt, cc.OutCoreChanDB); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				//это сообщение для клиента только когда есть какие либо ошибки в функции HandlerMsgFromAPI
				notifications.SendNotificationToClientAPI(cc.OutCoreChanAPI, ns, taskID, data.IDClientAPI)
			}

		//CHANNEL FROM NETWORK INTERACTION
		case data := <-cc.InCoreChanNI:

			fmt.Println("MESSAGE FROM module NetworkInteraction")

			if err := handlerslist.HandlerMsgFromNI(cc.OutCoreChanAPI, &data, smt, cc.OutCoreChanDB); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}
		}
	}
}
