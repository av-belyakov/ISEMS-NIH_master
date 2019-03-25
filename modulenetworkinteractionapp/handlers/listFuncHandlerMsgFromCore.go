package handlers

import (
	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"errors"
	"sort"
	"strconv"
)

func createStringFromSourceList(l []int) string {
	var strSourceID string

	for i := 0; len(l) > i; i++ {
		es := strconv.Itoa(l[i])

		if i == len(l)-1 {
			strSourceID += es

			continue
		}
		if len(l) > 1 && i == len(l)-2 {
			strSourceID += es + " и "

			continue
		}

		strSourceID += es + ", "
	}

	return strSourceID
}

//createSourceList создает новый список источников на основе полученного из БД
func createSourceList(isl *configure.InformationSourcesList, l []configure.InformationAboutSource) {
	for _, s := range l {
		isl.AddSourceSettings(s.ID, configure.SourceSetting{
			IP:         s.IP,
			Token:      s.Token,
			ClientName: s.NameClientAPI,
			AsServer:   s.AsServer,
			Settings:   s.SourceSetting,
		})
	}
}

func validateUserData(l *[]configure.DetailedListSources, mcpf int8) (*[]configure.SourceSetting, []int) {
	listTrastedSources := make([]configure.SourceSetting, 0, len(*l))
	listInvalidSource := []int{}

	for _, s := range *l {
		ipIsValid, _ := common.CheckStringIP(s.Argument.IP)
		tokenIsValid, _ := common.CheckStringToken(s.Argument.Token)
		foldersIsValid, _ := common.CheckFolders(s.Argument.Settings.StorageFolders)

		if !ipIsValid || !tokenIsValid || !foldersIsValid {
			listInvalidSource = append(listInvalidSource, s.ID)
		}

		if (s.Argument.Settings.MaxCountProcessFiltration > 0) && (s.Argument.Settings.MaxCountProcessFiltration < 10) {
			mcpf = s.Argument.Settings.MaxCountProcessFiltration
		}

		listTrastedSources = append(listTrastedSources, configure.SourceSetting{
			IP:       s.Argument.IP,
			Token:    s.Argument.Token,
			AsServer: s.Argument.Settings.AsServer,
			Settings: configure.InfoServiceSettings{
				EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
				MaxCountProcessFiltration: mcpf,
				StorageFolders:            s.Argument.Settings.StorageFolders,
			},
		})
	}

	return &listTrastedSources, listInvalidSource
}

//updateSourceList при получении от клиента API обновляет информацию по источникам
func updateSourceList(isl *configure.InformationSourcesList, l []configure.DetailedListSources, clientName string, mcpf int8) ([]int, []int) {
	var listTaskExecuted []int

	listTrastedSources, listInvalidSource := validateUserData(&l, mcpf)

	/*for _, s := range l {
		ipIsValid, _ := common.CheckStringIP(s.Argument.IP)
		tokenIsValid, _ := common.CheckStringToken(s.Argument.Token)
		foldersIsValid, _ := common.CheckFolders(s.Argument.Settings.StorageFolders)

		if !ipIsValid || !tokenIsValid || !foldersIsValid {
			listInvalidSource = append(listInvalidSource, s.ID)
		}

		if (s.Argument.Settings.MaxCountProcessFiltration > 0) && (s.Argument.Settings.MaxCountProcessFiltration < 10) {
			mcpf = s.Argument.Settings.MaxCountProcessFiltration
		}

		listTrastedSources = append(listTrastedSources, configure.SourceSetting{
			IP:       s.Argument.IP,
			Token:    s.Argument.Token,
			AsServer: s.Argument.Settings.AsServer,
			Settings: configure.InfoServiceSettings{
				EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
				MaxCountProcessFiltration: mcpf,
				StorageFolders:            s.Argument.Settings.StorageFolders,
			},
		})
	}*/

	if len(*listTrastedSources) == 0 {
		return listTaskExecuted, listInvalidSource
	}

	//если список источников в памяти приложения пуст
	if isl.GetCountSources() == 0 {
		for _, s := range l {
			isl.AddSourceSettings(s.ID, configure.SourceSetting{
				IP:         s.Argument.IP,
				Token:      s.Argument.Token,
				ClientName: clientName,
				AsServer:   s.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: s.Argument.Settings.MaxCountProcessFiltration,
					StorageFolders:            s.Argument.Settings.StorageFolders,
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
				IP:         source.Argument.IP,
				Token:      source.Argument.Token,
				ClientName: clientName,
				AsServer:   source.Argument.Settings.AsServer,
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

		//проверяем имеет ли право клиент делать какие либо изменения с информацией по источнику
		if clientName != s.ClientName && clientName != "root token" {
			listInvalidSource = append(listInvalidSource, source.ID)

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
				IP:         source.Argument.IP,
				Token:      source.Argument.Token,
				ClientName: clientName,
				AsServer:   source.Argument.Settings.AsServer,
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

//getSourceListToStoreDB формирует список источников для последующей их записи в БД
func getSourceListToStoreDB(trastedSoures []int, l *[]configure.DetailedListSources, clientName string, mcpf int8) (*[]configure.InformationAboutSource, error) {
	list := make([]configure.InformationAboutSource, 0, len(*l))

	sort.Ints(trastedSoures)
	for _, s := range *l {
		if sort.SearchInts(trastedSoures, s.ID) != -1 {
			if (s.Argument.Settings.MaxCountProcessFiltration > 0) && (s.Argument.Settings.MaxCountProcessFiltration < 10) {
				mcpf = s.Argument.Settings.MaxCountProcessFiltration
			}

			list = append(list, configure.InformationAboutSource{
				ID:            s.ID,
				IP:            s.Argument.IP,
				Token:         s.Argument.Token,
				ShortName:     s.Argument.ShortName,
				Description:   s.Argument.Description,
				AsServer:      s.Argument.Settings.AsServer,
				NameClientAPI: clientName,
				SourceSetting: configure.InfoServiceSettings{
					EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: mcpf,
					StorageFolders:            s.Argument.Settings.StorageFolders,
				},
			})
		}
	}

	return &list, nil
}

//splitSourceList разделяет список источников по выполняемым над ними действиям
/*func splitSourceList(l *[]configure.DetailedListSources) (lau, ldr, listActionUndefined []int) {
	for _, s := range *l {
		if s.ActionType == "add" || s.ActionType == "update" {
			lau = append(lau, s.ID)
		} else if s.ActionType == "delete" || s.ActionType == "reconnect" || s.ActionType == "status request" {
			ldr = append(ldr, s.ID)
		} else {
			listActionUndefined = append(listActionUndefined, s.ID)
		}
	}

	return lau, ldr, listActionUndefined
}*/

//performActionSelectedSources выполняет действия только с заданными источниками
func performActionSelectedSources(isl *configure.InformationSourcesList, l *[]configure.DetailedListSources, clientName string, mcpf int8) (*[]configure.ActionTypeListSources, *[]int, error) {
	listTrastedSources, listInvalidSource := validateUserData(l, mcpf)
	listActionIsExecuted := make([]configure.ActionTypeListSources, 0, len(*l))

	if len(*listTrastedSources) == 0 {
		return &listActionIsExecuted, &listInvalidSource, errors.New("parameters of all sources passed by the user have incorrect values, the action on any source will not be performed")
	}

	for _, s := range *l {
		aie := configure.ActionTypeListSources{
			ID:         s.ID,
			Status:     "disconnect",
			ActionType: s.ActionType,
		}

		strID := strconv.Itoa(s.ID)

		sourceInfo, ok := isl.GetSourceSetting(s.ID)
		//если источник не найден
		if !ok {
			if s.ActionType == "add" {
				isl.AddSourceSettings(s.ID, configure.SourceSetting{
					IP:         s.Argument.IP,
					Token:      s.Argument.Token,
					ClientName: clientName,
					AsServer:   s.Argument.Settings.AsServer,
					Settings: configure.InfoServiceSettings{
						EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
						MaxCountProcessFiltration: s.Argument.Settings.MaxCountProcessFiltration,
						StorageFolders:            s.Argument.Settings.StorageFolders,
					},
				})

				aie.IsSuccess = true
			}

			continue
		}

		if sourceInfo.ConnectionStatus {
			aie.Status = "connect"
		}

		//если источник найден
		if s.ActionType == "add" {
			aie.IsSuccess = false
			aie.MessageFailure = "невозможно добавить сенсор, сенсор с ID " + strID + " уже существует"

			continue
		}

		if s.ActionType == "update" || s.ActionType == "delete" {
			//проверяем имеет ли право клиент делать какие либо изменения с информацией по источнику
			if clientName != sourceInfo.ClientName && clientName != "root token" {
				aie.IsSuccess = false
				aie.MessageFailure = "недостаточно прав для выполнения действий с сенсором ID " + strID + ", возможно данный сенсор был добавлен другим клиентом"

				continue
			}

			//проверяем выполняется ли на источнике задача
			if len(sourceInfo.CurrentTasks) > 0 {
				aie.IsSuccess = false
				aie.MessageFailure = "невозможно выполнить действия на сенсоре с ID " + strID + ", так как в настоящее время на данном сенсоре выполняется задача"

				continue
			}
		}

		if s.ActionType == "reconnect" {
			if !sourceInfo.ConnectionStatus {
				aie.IsSuccess = false
				aie.MessageFailure = "невозможно выполнить переподключение, сенсор с ID " + strID + " не подключен"

				continue
			}

			if len(isl.GetListTasksPerformedSourceByType(s.ID, "download")) > 0 {
				aie.IsSuccess = false
				aie.MessageFailure = "невозможно выполнить переподключение, с сенсора с ID " + strID + " осуществляется загрузка файлов"

				continue
			}
		}

		/*
			ДОДЕЛАТЬ ПРОВЕРКИ, ОБНОВИТЬ ИЛИ УДАЛИТЬ ИСТОЧНИК,
			ВЫПОЛНИТЬ ПЕРЕПОДКЛЮЧЕНИЕ
		*/
	}

	/*
		sl: [ ARRAY
				{
					id: <INT> // ID - уникальный цифровой идентификатор источника
					s: <STRING> // status - 'connect'/'disconnect',
					at: <STRING>, // actionType — тип действия, 'add' / 'delete' / 'update' / 'reconnect' / 'none'
					is: <BOOL>, // isSuccess — успешность (true/flase)
					mf: <STRING>'' //  messageFailure — сообщение об ошибке
				}
			]
	*/

	return &listActionIsExecuted, &listInvalidSource, nil
}
