package handlers

/*
* Обработчик запросов от ядра приложения
*
* Версия 0.1, дата релиза 18.03.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//HandlerMsgFromCore обработчик сообщений от ядра приложения
func HandlerMsgFromCore(cwt chan<- configure.MsgWsTransmission, isl *configure.InformationSourcesList, msg configure.MsgBetweenCoreAndNI, chanInCore chan<- configure.MsgBetweenCoreAndNI) {
	fmt.Println("START func HandlerMsgFromCore...")

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
				notAddSourceList := updateSourceList(isl, ado.MsgOptions.SourceList)

				fmt.Println("NOT ADD SOURCE LIST", notAddSourceList)

				//информационное сообщение пользователю
				chanInCore <- configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "message notification",
					Command: "send client API",
					AdvancedOptions: configure.MessageNotification{
						SourceReport:                 "NI module",
						Section:                      "source control",
						TypeActionPerformed:          "load list",
						Sources:                      notAddSourceList,
						HumanDescriptionNotification: fmt.Sprintf("На источнике (-ах) %q выполняются задачи, изменение настроек не доступно", notAddSourceList),
					},
				}

				/*											*
				*											*
				*					!!!!					*
				*	сообщение для обновления списка в БД	*
				*					!!!!					*
				*											*
				 */
			}
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
func createSourceList(isl *configure.InformationSourcesList, l []configure.InformationAboutSource) {
	for _, source := range l {
		isl.AddSourceSettings(source.ID, configure.SourceSetting{
			IP:       source.IP,
			Token:    source.Token,
			AsServer: source.AsServer,
			Settings: source.SourceSetting,
		})
	}
}

//updateSourceList при получении от клиента API обновляет информацию по источникам
func updateSourceList(isl *configure.InformationSourcesList, l []configure.SourceListFromAPI) []int {
	fmt.Printf("\n function 'updateSourceList' list sources from client API \n%v\n", l)
	fmt.Println("Дальше нужно делать после тестов")

	var listTaskExecuted []int

	//если список источников в памяти приложения пуст
	if isl.GetCountSources() == 0 {
		for _, source := range l {
			isl.AddSourceSettings(source.ID, configure.SourceSetting{
				IP:       source.Argument.IP,
				Token:    source.Argument.Token,
				AsServer: source.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					EnableTelemetry:           source.Argument.Settings.EnableTelemetry,
					MaxCountProcessfiltration: source.Argument.Settings.MaxCountProcessFiltering,
					StorageFolders:            source.Argument.Settings.StorageFolders,
				},
			})
		}

		return listTaskExecuted
	}

	var sourcesIsReconnect bool

	_, listDisconnected := isl.GetListsConnectedAndDisconnectedSources()
	sourceListTaskExecuted := isl.GetListSourcesWhichTaskExecuted()

	for _, source := range l {
		//если источника нет в списке
		s, isExist := isl.GetSourceSetting(source.ID)
		if !isExist {
			isl.AddSourceSettings(source.ID, configure.SourceSetting{
				IP:       source.Argument.IP,
				Token:    source.Argument.Token,
				AsServer: source.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					EnableTelemetry:           source.Argument.Settings.EnableTelemetry,
					MaxCountProcessfiltration: source.Argument.Settings.MaxCountProcessFiltering,
					StorageFolders:            source.Argument.Settings.StorageFolders,
				},
			})

			continue
		}

		//если на источнике выполняется задача
		if _, ok := sourceListTaskExecuted[source.ID]; ok {
			listTaskExecuted = append(listTaskExecuted, source.ID)

			continue
		}

		//проверяем параметры подключения
		if (s.Token != source.Argument.Token) || (s.AsServer != source.Argument.Settings.AsServer) {
			sourcesIsReconnect = true
		}

		//полная замена информации об источнике
		if _, ok := listDisconnected[source.ID]; ok {
			isl.DelSourceSettings(source.ID)

			isl.AddSourceSettings(source.ID, configure.SourceSetting{
				IP:       source.Argument.IP,
				Token:    source.Argument.Token,
				AsServer: source.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					EnableTelemetry:           source.Argument.Settings.EnableTelemetry,
					MaxCountProcessfiltration: source.Argument.Settings.MaxCountProcessFiltering,
					StorageFolders:            source.Argument.Settings.StorageFolders,
				},
			})

			continue
		}

		if sourcesIsReconnect {
			//закрываем соединение и удаляем дискриптор
			if cl, isExist := isl.GetLinkWebsocketConnect(source.Argument.IP); isExist {
				cl.Link.Close()

				isl.DelLinkWebsocketConnection(source.Argument.IP)
			}

			sourcesIsReconnect = false
		}
	}

	return listTaskExecuted
}

//performActionSelectedSources выполняет действия только с заданными источниками
func performActionSelectedSources() {

}
