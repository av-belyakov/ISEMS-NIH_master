package moduleapiapp

/*
* Модуль API, маршрутизация запросов
*
* Версия 0.1, дата релиза 28.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//RouteCoreMessage маршрутизатор сообщений от ядра приложения
func RouteCoreMessage(chanToCore chan<- configure.MsgBetweenCoreAndAPI, chanFromCore <-chan configure.MsgBetweenCoreAndAPI) {
	fmt.Println("START module 'RouteCoreMessage' (API)...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	fmt.Printf("%v\n", chanFromCore)

	for msg := range chanFromCore {
		fmt.Println("++++ ROUTE APIApp resived MSG from CoreApp")
		fmt.Printf("%T %v\n", msg, msg)

		if msg.MsgGenerator == "Core module" {
			if msg.MsgType == "command" {
				switch msg.DataType {
				case "source_control":
					message, ok := msg.AdvancedOptions.(configure.MsgCommandSourceControl)
					if !ok {
						_ = saveMessageApp.LogMessage("error", "an incorrect command is accepted from CoreApp (module API route)")
						break
					}

					if message.ListRequest {
						/*
							Запрашиваем у клиента API полный список источников
						*/

						fmt.Println("--- ModuleAPIApp, resived request to SOURCE LIST", "send...")

						// --- ТЕСТОВЫЙ ОТВЕТ ---
						chanToCore <- configure.MsgBetweenCoreAndAPI{
							MsgGenerator: "API module",
							MsgType:      "information",
							DataType:     "change_status_source",
							IDClientAPI:  "du68whfh733hjf9393",
							AdvancedOptions: configure.MsgInfoChangeStatusSource{
								SourceListIsExist: true,
								SourceList: []configure.SourceCharacteristicForConnection{
									{9, "127.0.0.1", "fmdif3o444fdf344k0fiif"},
									{10, "192.168.0.10", "fmdif3o444fdf344k0fiif"},
									{11, "192.168.0.11", "ttrr9gr9r9e9f9fadx94"},
									{12, "192.168.0.12", "2n3n3iixcxcc3444xfg0222"},
									{13, "192.168.0.13", "osdsoc9c933cc9cn939f9f33"},
									{14, "192.168.0.14", "hgdfffffff9333ffodffodofff0"},
								},
							},
						}
						//------------------------
					}
				}

			}

			if msg.MsgType == "information" {
				/* пока заглушка */
			}
		}

	}
}
