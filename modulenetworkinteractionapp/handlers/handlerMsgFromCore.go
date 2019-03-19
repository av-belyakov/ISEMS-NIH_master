package handlers

/*
* Обработчик запросов от ядра приложения
*
* Версия 0.1, дата релиза 18.03.2019
* */

import (
	"ISEMS-NIH_master/common"
	"fmt"
	"strconv"

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
				clientNotify := configure.MsgBetweenCoreAndNI{
					TaskID:  msg.TaskID,
					Section: "message notification",
					Command: "send client API",
				}

				//проверяем прислал ли пользователь данные по источникам
				if len(ado.MsgOptions.SourceList) == 0 {
					clientNotify.AdvancedOptions = configure.MessageNotification{
						SourceReport:                 "NI module",
						Section:                      "source control",
						TypeActionPerformed:          "load list",
						CriticalityMessage:           "warning",
						HumanDescriptionNotification: "Получен пустой список сенсоров",
					}

					chanInCore <- clientNotify

					return
				}

				notAddSourceList, listInvalidSource := updateSourceList(isl, ado.MsgOptions.SourceList)
				if len(listInvalidSource) != 0 {
					strSourceID := createStringFromSourceList(listInvalidSource)

					clientNotify.AdvancedOptions = configure.MessageNotification{
						SourceReport:                 "NI module",
						Section:                      "source control",
						TypeActionPerformed:          "load list",
						CriticalityMessage:           "warning",
						Sources:                      listInvalidSource,
						HumanDescriptionNotification: "Обновление списка сенсоров выполнено не полностью, параметры сенсоров " + strSourceID + " содержат некорректные значения",
					}

					chanInCore <- clientNotify
				} else {
					hdn := "Обновление настроек сенсоров выполнено успешно"
					cm := "success"
					if len(notAddSourceList) > 0 {
						strSourceID := createStringFromSourceList(notAddSourceList)
						hdn = "На источнике (-ах) " + strSourceID + " выполняются задачи, изменение настроек не доступно"
						cm = "info"
					}

					clientNotify.AdvancedOptions = configure.MessageNotification{
						SourceReport:                 "NI module",
						Section:                      "source control",
						TypeActionPerformed:          "load list",
						CriticalityMessage:           cm,
						Sources:                      notAddSourceList,
						HumanDescriptionNotification: hdn,
					}

					chanInCore <- clientNotify
				}

				msgToCore := configure.MsgBetweenCoreAndNI{
					TaskID:          msg.TaskID,
					Section:         "source control",
					Command:         "keep list sources in database",
					AdvancedOptions: isl.GetSourceList(),
				}

				//новый список источников для сохранения в БД
				chanInCore <- msgToCore

				//новый список источников для клиента API
				msgToCore.Command = "send list sources to client api"
				chanInCore <- msgToCore
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

func createStringFromSourceList(l []int) string {
	var strSourceID string

	for i := 0; len(l) > i; i++ {
		es := strconv.Itoa(l[i])

		if i == len(l)-1 {
			strSourceID += es

			continue
		}
		strSourceID += es + ", "
	}

	return strSourceID
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
func updateSourceList(isl *configure.InformationSourcesList, l []configure.SourceListFromAPI) ([]int, []int) {
	fmt.Printf("\n function 'updateSourceList' list sources from client API \n%v\n", l)

	var listTaskExecuted, listInvalidSource []int
	listTrastedSources := []configure.SourceSetting{}

	for _, s := range l {
		ipIsValid, _ := common.CheckStringIP(s.Argument.IP)

		tokenIsValid, _ := common.CheckStringToken(s.Argument.Token)

		foldersIsValid, _ := common.CheckFolders(s.Argument.Settings.StorageFolders)

		fmt.Println("++++ ipIsValid:", ipIsValid, "tokenIsValid:", tokenIsValid, "foldersIsValide:", foldersIsValid, "+++++")

		if !ipIsValid || !tokenIsValid || !foldersIsValid {
			listInvalidSource = append(listInvalidSource, s.ID)
		}

		listTrastedSources = append(listTrastedSources, configure.SourceSetting{
			IP:       s.Argument.IP,
			Token:    s.Argument.Token,
			AsServer: s.Argument.Settings.AsServer,
			Settings: configure.InfoServiceSettings{
				EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
				MaxCountProcessFiltration: s.Argument.Settings.MaxCountProcessFiltration,
				StorageFolders:            s.Argument.Settings.StorageFolders,
			},
		})
	}

	if len(listTrastedSources) == 0 {
		return listTaskExecuted, listInvalidSource
	}

	//если список источников в памяти приложения пуст
	if isl.GetCountSources() == 0 {
		for _, source := range l {
			isl.AddSourceSettings(source.ID, configure.SourceSetting{
				IP:       source.Argument.IP,
				Token:    source.Argument.Token,
				AsServer: source.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					EnableTelemetry:           source.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: source.Argument.Settings.MaxCountProcessFiltration,
					StorageFolders:            source.Argument.Settings.StorageFolders,
				},
			})
		}

		return listTaskExecuted, listInvalidSource
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
					MaxCountProcessFiltration: source.Argument.Settings.MaxCountProcessFiltration,
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
					MaxCountProcessFiltration: source.Argument.Settings.MaxCountProcessFiltration,
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

	return listTaskExecuted, listInvalidSource
}

//performActionSelectedSources выполняет действия только с заданными источниками
func performActionSelectedSources() {

}
