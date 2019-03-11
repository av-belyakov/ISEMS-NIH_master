package coreapp

/*
* Ядро приложения
* Маршрутизация сообщений получаемых через каналы
*
* Версия 0.2, дата релиза 04.03.2019
* */

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/coreapp/handlerslist"
	"ISEMS-NIH_master/savemessageapp"
)

//Routing маршрутизирует данные поступающие в ядро из каналов
func Routing(appConf *configure.AppConfig, ism *configure.InformationStoringMemory, cc *configure.ChannelCollectionCoreApp) {
	fmt.Println("START ROUTE module 'CoreApp'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//при старте приложения запрашиваем у API новый список источников
	//временно оставляем, но новый список запрашивается в модуле API
	//при успешном подключении клиента к wss серверу
	/*cc.OutCoreChanAPI <- configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgType:      "command",
		DataType:     "source_control",
		IDClientAPI:  "",
		AdvancedOptions: configure.MsgCommandSourceControl{
			ListRequest: true,
		},
	}*/

	fmt.Println("ROUTE CoreApp sending data to channel")

	//обработчики для инфрмационных сообщений от модуля API
	handlersInfoMsgFromAPI := map[string]func(chan<- configure.MsgBetweenCoreAndDB, string, interface{}) error{
		"source control": handlerslist.HandlerStatusSourceFromAPI,
	}

	//обработчики для команд от модуля API
	handlersCommandMsgFromAPI := map[string]func(chan<- configure.MsgBetweenCoreAndDB, string, interface{}) error{
		"source control":     handlerslist.HandlerSourceControlFromAPI,
		"filtration":         handlerslist.HandlerFiltrationFromAPI,
		"download":           handlerslist.HandlerDownloadFromAPI,
		"information search": handlerslist.HandlerInformationSearchFromAPI,
	}

	//обработчики для информационных сообщений от модуля взаимодействия с БД
	handlersInfoMsgFromDB := map[string]func(chan<- configure.MsgBetweenCoreAndAPI, chan<- configure.MsgBetweenCoreAndNI, configure.MsgBetweenCoreAndDB) error{
		"sources control":            handlerslist.HandlerSourcesControlFromDB,
		"source telemetry":           handlerslist.HandlerSourceTelemetryFromDB,
		"filtration":                 handlerslist.HandlerFiltrationFromDB,
		"download":                   handlerslist.HandlerDownloadFromDB,
		"information search results": handlerslist.HandlerMsgInfoSearchResultsFromDB,
		"error notification":         handlerslist.HandlerErrorNotificationFromDB,
	}

	//обработчики для информационных сообщений от модуля сетевого взаимодействия Network Interaction
	handlersInfoMsgIN := map[string]func(string, interface{}) error{
		"change status source": handlerslist.HandlerChangeStatusSourceFromNI,
		"source telemetry":     handlerslist.HandlerSourceTelemetryFromNI,
		"filtration":           handlerslist.HandlerFiltrationFromNI,
		"download":             handlerslist.HandlerDownloadFromNI,
		"error notification":   handlerslist.HandlerErrorNotificationFromNI,
	}

	for {
		select {
		//CHANNEL FROM DATABASE
		case data := <-cc.InCoreChanDB:
			fmt.Println("MESSAGE FROM module DBInteraction")
			//использовать канал cc.OutCoreChanDB для ответа
			fmt.Println(data)

			handler, ok := handlersInfoMsgFromDB[data.DataType]
			if !ok {
				_ = saveMessageApp.LogMessage("error", "from the API passed an invalid data type (module Core route)")
			}

			if err := handler(cc.OutCoreChanAPI, cc.OutCoreChanNI, data); err != nil {
				_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
			}

			//
			// получаем список источников для подключения и
			// записываем в память
			//

			//CHANNEL FROM API
		case data := <-cc.InCoreChanAPI:

			fmt.Println("resived message from API module ", data)

			if data.MsgGenerator == "API module" && data.MsgRecipient == "Core module" {
				/*
					сделать учет запросов по ID клиенту?

					развернуть JSON
				*/

				var msgAPI configure.MsgType

				if err = json.Unmarshal(message, &msgAPI); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}
			}

			if data.MsgType == "information" {
				fmt.Println("resived message type 'information' from API module")

				handler, ok := handlersInfoMsgFromAPI[data.DataType]
				if !ok {
					_ = saveMessageApp.LogMessage("error", "from the API passed an invalid data type (module Core route)")
				}

				if err := handler(cc.OutCoreChanDB, data.IDClientAPI, data.AdvancedOptions); err != nil {
					_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
				}
			}

			if data.MsgType == "command" {
				fmt.Println("resived message type COMMAND from API module")

				handler, ok := handlersCommandMsgFromAPI[data.DataType]
				if !ok {
					_ = saveMessageApp.LogMessage("error", "from the API passed an invalid data type (module Core route)")
				}

				if err := handler(cc.OutCoreChanDB, data.IDClientAPI, data.AdvancedOptions); err != nil {
					_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
				}

			}

			fmt.Println("ДАЛЕЕ НУЖНО ОБРАБОТАТЬ И ПЕРЕДАТь через канал модулю БД")

		case data := <-cc.InCoreChanNI:
			//CHANNEL FROM NETWORK INTERACTION
			if data.MsgType == "command" {
				fmt.Println("resived message type 'command' from module NI")
			}
			if data.MsgType == "information" {
				fmt.Println("resived message type 'information' from module NI")

				handler, ok := handlersInfoMsgIN[data.DataType]
				if !ok {
					_ = saveMessageApp.LogMessage("error", "from the API passed an invalid data type (module Core route)")
				}

				if err := handler(data.IDClientAPI, data.AdvancedOptions); err != nil {
					_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
				}
			}
			fmt.Println("MESSAGE FROM module NetworkInteraction")
			//использовать канал cc.OutCoreChanNI для ответа
			fmt.Println(data)
		}
	}
}
