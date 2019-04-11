package coreapp

/*
* Ядро приложения
* Маршрутизация сообщений получаемых через каналы
*
* Версия 0.3, дата релиза 13.03.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/coreapp/handlerslist"
	"ISEMS-NIH_master/notifications"
)

//Routing маршрутизирует данные поступающие в ядро из каналов
func Routing(
	appConf *configure.AppConfig,
	cc *configure.ChannelCollectionCoreApp,
	smt *configure.StoringMemoryTask,
	chanCheckTask <-chan configure.MsgChanStoringMemoryTask) {

	fmt.Println("START ROUTE module 'CoreApp'...")

	//при старте приложения запрашиваем список источников
	//отправляем запрос в БД
	cc.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
		MsgGenerator: "NI module",
		MsgRecipient: "DB module",
		MsgSection:   "source control",
		Instruction:  "find_all",
	}

	//обработчик запросов от модулей приложения
	for {
		select {
		//CHANNEL FROM DATABASE
		case data := <-cc.InCoreChanDB:
			fmt.Println("MESSAGE FROM module DBInteraction (route Core module)")

			go handlerslist.HandlerMsgFromDB(cc.OutCoreChanAPI, data, smt, cc.OutCoreChanNI)

		//CHANNEL FROM API
		case data := <-cc.InCoreChanAPI:
			fmt.Println("MESSAGE FROM module API (route Core module)")

			go handlerslist.HandlerMsgFromAPI(cc.OutCoreChanNI, data, smt, cc.OutCoreChanDB, cc.OutCoreChanAPI)

		//CHANNEL FROM NETWORK INTERACTION
		case data := <-cc.InCoreChanNI:
			fmt.Println("MESSAGE FROM module NetworkInteraction (route Core module)")

			go handlerslist.HandlerMsgFromNI(cc.OutCoreChanAPI, data, smt, cc.OutCoreChanDB)

		//сообщение клиенту API о том что задача с указанным ID долго выполняется
		case infoHungTask := <-chanCheckTask:
			fmt.Println("сообщение клиенту о том что данная задача долго выполняется")

			if ti, ok := smt.GetStoringMemoryTask(infoHungTask.ID); ok {
				nsErrJSON := notifications.NotificationSettingsToClientAPI{
					MsgType:        infoHungTask.Type,
					MsgDescription: infoHungTask.Description,
				}

				notifications.SendNotificationToClientAPI(cc.OutCoreChanAPI, nsErrJSON, ti.ClientTaskID, ti.ClientID)
			}
		}
	}
}
