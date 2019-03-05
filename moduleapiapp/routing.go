package moduleapiapp

/*
* Модуль API, маршрутизация запросов
*
* Версия 0.1, дата релиза 28.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//RouteCoreMessage маршрутизатор сообщений от ядра приложения
func RouteCoreMessage(chanToCore chan<- configure.MsgBetweenCoreAndAPI, chanFromCore <-chan configure.MsgBetweenCoreAndAPI) {
	fmt.Println("START module 'RouteCoreMessage' (API)...")

	//инициализируем функцию конструктор для записи лог-файлов
	//saveMessageApp := savemessageapp.New()

	fmt.Printf("%v\n", chanFromCore)

	for msg := range chanFromCore {
		fmt.Println("++++ ROUTE APIApp resived MSG from CoreApp")
		fmt.Printf("%T %v\n", msg, msg)

		if msg.MsgGenerator == "Core module" {
			if msg.MsgType == "command" {
				/* пока заглушка */

				/*switch msg.DataType {
								case "source_control":
									message, ok := msg.AdvancedOptions.(configure.MsgCommandSourceControl)
									if !ok {
										_ = saveMessageApp.LogMessage("error", "an incorrect command is accepted from CoreApp (module API route)")
										break
									}

									if message.ListRequest {

				//							Запрашиваем у клиента API полный список источников


										fmt.Println("--- ModuleAPIApp, resived request to SOURCE LIST", "send...")


									}
								}*/

			}

			if msg.MsgType == "information" {
				/* пока заглушка */
			}
		}

	}
}
