package handlers

/*
* Обработчик запросов от ядра приложения
*
* Версия 0.1, дата релиза 13.03.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//HandlerMsgFromCore обработчик сообщений от ядра приложения
func HandlerMsgFromCore(cwt chan<- configure.MsgWsTransmission, isl *configure.InformationSourcesList, msg configure.MsgBetweenCoreAndNI, chanInCore chan<- configure.MsgBetweenCoreAndNI) {
	fmt.Println("START func HandlerMsgFromCore...")

	//инициализируем функцию конструктор для записи лог-файлов
	//saveMessageApp := savemessageapp.New()

	switch msg.Section {
	case "source control":
		if msg.Command == "create list" {

			fmt.Println("CREATE LIST")

			if sl, ok := msg.AdvancedOptions.([]configure.InformationAboutSource); ok {

				createSourceList(isl, sl)

				fmt.Println("create source list for memory success (OUT DATABASE)")
			}
		}

		if msg.Command == "load list" {
			if ado, ok := msg.AdvancedOptions.(configure.SourceControlMsgOptions); ok {
				fmt.Println("interface{} -> user type is ", ok)

				updateSourceList(chanInCore, isl, ado.MsgOptions.SourceList)
			}

			/*			for k, v := range msg.AdvancedOptions.(configure.SourceControlMsgTypeFromAPI) {
						fmt.Println(k)
						fmt.Printf("%T", v)
					}*/

			//			fmt.Println(msg.AdvancedOptions.([]configure.SourceListFromAPI))
			/*			if sl, ok := msg.AdvancedOptions.(configure.SourceControlMsgTypeFromAPI); ok {

						updateSourceList(chanInCore, isl, sl)

						fmt.Println("create source list for memory success (OUT API MODULE)")
						fmt.Printf("\n%T%v\n", isl, isl)
					}*/
		}

		if msg.Command == "update list" {

		}
		/*if msg.Command == "add" {

		}

		if msg.Command == "del" {

		}

		if msg.Command == "update" {

		}

		if msg.Command == "reconnect" {

		}*/

	case "filtration control":
		if msg.Command == "start" {

		}

		if msg.Command == "stop" {

		}

	case "download control":
		if msg.Command == "start" {

		}

		if msg.Command == "stop" {

		}

	}
}

//createSourceList создает новый список источников на основе полученного из БД
func createSourceList(isl *configure.InformationSourcesList, list []configure.InformationAboutSource) {
	for _, source := range list {
		isl.AddSourceSettings(source.IP, configure.SourceSetting{
			ID:       source.ID,
			Token:    source.Token,
			AsServer: source.AsServer,
			Settings: source.SourceSetting,
		})
	}
}

//updateSourceList при получении от клиента API обновляет информацию по источникам
func updateSourceList(chanInCore chan<- configure.MsgBetweenCoreAndNI, isl *configure.InformationSourcesList, l []configure.SourceListFromAPI) {
	fmt.Printf("\n function 'updateSourceList' list sources from client API \n%v\n", l)
	fmt.Println("Дальше нужно делать после тестов")
}

//performActionSelectedSources выполняет действия только с заданными источниками
func performActionSelectedSources() {

}
