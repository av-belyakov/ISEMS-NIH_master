package handlers

import (
	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
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

		serverPort := s.Argument.Settings.Port
		if s.Argument.Settings.Port == 0 {
			serverPort = 13113
		}

		listTrastedSources = append(listTrastedSources, configure.SourceSetting{
			IP:       s.Argument.IP,
			Token:    s.Argument.Token,
			AsServer: s.Argument.Settings.AsServer,
			Settings: configure.InfoServiceSettings{
				IfAsServerThenPort:        serverPort,
				EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
				MaxCountProcessFiltration: mcpf,
				StorageFolders:            s.Argument.Settings.StorageFolders,
			},
		})
	}

	return &listTrastedSources, listInvalidSource
}

//updateSourceList обновляет информацию по источникам
func updateSourceList(
	isl *configure.InformationSourcesList,
	qts *configure.QueueTaskStorage,
	l []configure.DetailedListSources,
	clientName string,
	mcpf int8) ([]int, []int) {

	var listTaskExecuted []int

	listTrastedSources, listInvalidSource := validateUserData(&l, mcpf)

	if len(*listTrastedSources) == 0 {
		return listTaskExecuted, listInvalidSource
	}

	typeAreaNetwork := "ip"

	//если список источников в памяти приложения пуст
	if isl.GetCountSources() == 0 {
		for _, s := range l {
			if strings.ToLower(s.Argument.Settings.TypeAreaNetwork) == "pppoe" {
				typeAreaNetwork = "pppoe"
			}

			isl.AddSourceSettings(s.ID, configure.SourceSetting{
				IP:         s.Argument.IP,
				Token:      s.Argument.Token,
				ClientName: clientName,
				AsServer:   s.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					IfAsServerThenPort:        s.Argument.Settings.Port,
					EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: s.Argument.Settings.MaxCountProcessFiltration,
					StorageFolders:            s.Argument.Settings.StorageFolders,
					TypeAreaNetwork:           typeAreaNetwork,
				},
			})
		}

		return listTaskExecuted, listInvalidSource
	}

	var sourcesIsReconnect bool

	_, listDisconnected := isl.GetListsConnectedAndDisconnectedSources()

	for _, source := range l {
		if strings.ToLower(source.Argument.Settings.TypeAreaNetwork) == "pppoe" {
			typeAreaNetwork = "pppoe"
		}

		//если источника нет в списке
		s, isExist := isl.GetSourceSetting(source.ID)
		if !isExist {
			isl.AddSourceSettings(source.ID, configure.SourceSetting{
				IP:         source.Argument.IP,
				Token:      source.Argument.Token,
				ClientName: clientName,
				AsServer:   source.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					IfAsServerThenPort:        source.Argument.Settings.Port,
					EnableTelemetry:           source.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: source.Argument.Settings.MaxCountProcessFiltration,
					StorageFolders:            source.Argument.Settings.StorageFolders,
					TypeAreaNetwork:           typeAreaNetwork,
				},
			})

			continue
		}

		//если на источнике ожидает выполнения или выполняется какая либо задача задача
		if _, ok := qts.GetAllTaskQueueTaskStorage(source.ID); ok {
			listTaskExecuted = append(listTaskExecuted, source.ID)
		}

		//проверяем имеет ли право клиент делать какие либо изменения с информацией по источнику
		if clientName != s.ClientName && clientName != "root token" {
			listInvalidSource = append(listInvalidSource, source.ID)

			continue
		}

		changeToken := (s.Token != source.Argument.Token)
		changeIP := (s.IP != source.Argument.IP)
		changeAsServer := (s.AsServer != source.Argument.Settings.AsServer)
		changeEnTelemetry := (s.Settings.EnableTelemetry != source.Argument.Settings.EnableTelemetry)

		//проверяем параметры подключения
		if changeToken || changeIP || changeAsServer || changeEnTelemetry {
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
					IfAsServerThenPort:        source.Argument.Settings.Port,
					EnableTelemetry:           source.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: source.Argument.Settings.MaxCountProcessFiltration,
					StorageFolders:            source.Argument.Settings.StorageFolders,
					TypeAreaNetwork:           typeAreaNetwork,
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

	typeAreaNetwork := "ip"

	sort.Ints(trastedSoures)
	for _, s := range *l {
		if sort.SearchInts(trastedSoures, s.ID) != -1 {
			if (s.Argument.Settings.MaxCountProcessFiltration > 0) && (s.Argument.Settings.MaxCountProcessFiltration < 10) {
				mcpf = s.Argument.Settings.MaxCountProcessFiltration
			}

			if strings.ToLower(s.Argument.Settings.TypeAreaNetwork) == "pppoe" {
				typeAreaNetwork = "pppoe"
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
					IfAsServerThenPort:        s.Argument.Settings.Port,
					EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: mcpf,
					StorageFolders:            s.Argument.Settings.StorageFolders,
					TypeAreaNetwork:           typeAreaNetwork,
				},
			})
		}
	}

	return &list, nil
}

//performActionSelectedSources выполняет действия только с заданными источниками
func performActionSelectedSources(
	isl *configure.InformationSourcesList,
	qts *configure.QueueTaskStorage,
	l *[]configure.DetailedListSources,
	clientName string,
	mcpf int8) (*[]configure.ActionTypeListSources, *[]int, error) {

	listTrastedSources, listInvalidSource := validateUserData(l, mcpf)
	listActionIsExecuted := make([]configure.ActionTypeListSources, 0, len(*l))

	if len(*listTrastedSources) == 0 {
		return &listActionIsExecuted, &listInvalidSource, errors.New("parameters of all sources passed by the user have incorrect values, the action on any source will not be performed")
	}

	var ok bool
	var sourceInfo *configure.SourceSetting

	typeAreaNetwork := "ip"

	for _, s := range *l {
		aie := configure.ActionTypeListSources{
			ID:         s.ID,
			Status:     "disconnect",
			ActionType: s.ActionType,
		}

		if strings.ToLower(s.Argument.Settings.TypeAreaNetwork) == "pppoe" {
			typeAreaNetwork = "pppoe"
		}

		//есть ли источник с таким ID
		if sourceInfo, ok = isl.GetSourceSetting(s.ID); !ok {
			if s.ActionType == "add" {
				isl.AddSourceSettings(s.ID, configure.SourceSetting{
					IP:         s.Argument.IP,
					Token:      s.Argument.Token,
					ClientName: clientName,
					AsServer:   s.Argument.Settings.AsServer,
					Settings: configure.InfoServiceSettings{
						IfAsServerThenPort:        s.Argument.Settings.Port,
						EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
						MaxCountProcessFiltration: s.Argument.Settings.MaxCountProcessFiltration,
						StorageFolders:            s.Argument.Settings.StorageFolders,
						TypeAreaNetwork:           typeAreaNetwork,
					},
				})

				aie.IsSuccess = true
			}
			listActionIsExecuted = append(listActionIsExecuted, aie)

			continue
		}

		if sourceInfo.ConnectionStatus {
			aie.Status = "connect"
		}

		//обработка запроса статуса соединений
		if s.ActionType == "status request" {
			aie.IsSuccess = true

			listActionIsExecuted = append(listActionIsExecuted, aie)

			continue
		}

		//если источник найден
		if s.ActionType == "add" {
			aie.IsSuccess = false
			aie.MessageFailure = fmt.Sprintf("невозможно добавить сенсор, сенсор с ID %v уже существует", s.ID)

			listActionIsExecuted = append(listActionIsExecuted, aie)

			continue
		}

		if s.ActionType == "update" || s.ActionType == "delete" {
			//проверяем имеет ли право клиент делать какие либо изменения с информацией по источнику
			if (clientName != sourceInfo.ClientName) && (clientName != "root token") {
				aie.IsSuccess = false
				aie.MessageFailure = fmt.Sprintf("недостаточно прав для выполнения действий с сенсором ID %v, возможно данный сенсор был добавлен другим клиентом", s.ID)

				listActionIsExecuted = append(listActionIsExecuted, aie)

				continue
			}

			//проверяем ожидает или выполняется на источнике какая либо задача
			if _, ok := qts.GetAllTaskQueueTaskStorage(s.ID); ok {
				aie.IsSuccess = false
				aie.MessageFailure = fmt.Sprintf("невозможно выполнить действия на сенсоре с ID %v, так как в настоящее время на данном сенсоре ожидает выполнения или уже выполняется какая либо задача", s.ID)

				listActionIsExecuted = append(listActionIsExecuted, aie)

				continue
			}

			/*if len(sourceInfo.CurrentTasks) > 0 {
				aie.IsSuccess = false
				aie.MessageFailure = "невозможно выполнить действия на сенсоре с ID " + strID + ", так как в настоящее время на данном сенсоре выполняется задача"

				listActionIsExecuted = append(listActionIsExecuted, aie)

				continue
			}*/
		}

		if s.ActionType == "update" {
			changeToken := (sourceInfo.Token != s.Argument.Token)
			changeIP := (sourceInfo.IP != s.Argument.IP)
			changeAsServer := (sourceInfo.AsServer != s.Argument.Settings.AsServer)
			changeEnTelemetry := (sourceInfo.Settings.EnableTelemetry != s.Argument.Settings.EnableTelemetry)

			//проверяем параметры подключения
			if changeToken || changeIP || changeAsServer || changeEnTelemetry {
				//закрываем соединение и удаляем дискриптор
				if cl, isExist := isl.GetLinkWebsocketConnect(sourceInfo.IP); isExist {
					cl.Link.Close()

					isl.DelLinkWebsocketConnection(sourceInfo.IP)
				}
			}

			isl.AddSourceSettings(s.ID, configure.SourceSetting{
				IP:         s.Argument.IP,
				Token:      s.Argument.Token,
				ClientName: clientName,
				AsServer:   s.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					IfAsServerThenPort:        s.Argument.Settings.Port,
					EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: s.Argument.Settings.MaxCountProcessFiltration,
					StorageFolders:            s.Argument.Settings.StorageFolders,
					TypeAreaNetwork:           typeAreaNetwork,
				},
			})

			aie.IsSuccess = true

			listActionIsExecuted = append(listActionIsExecuted, aie)

			continue
		}

		if s.ActionType == "delete" {
			//закрываем соединение и удаляем дискриптор
			if cl, isExist := isl.GetLinkWebsocketConnect(sourceInfo.IP); isExist {
				cl.Link.Close()

				isl.DelLinkWebsocketConnection(sourceInfo.IP)
			}

			//удаление всей информации об источнике
			isl.DelSourceSettings(s.ID)

			aie.IsSuccess = true

			listActionIsExecuted = append(listActionIsExecuted, aie)

			continue
		}

		//при переподключении
		if s.ActionType == "reconnect" {
			if !sourceInfo.ConnectionStatus {
				aie.IsSuccess = false
				aie.MessageFailure = fmt.Sprintf("невозможно выполнить переподключение, сенсор с ID %v не подключен", s.ID)

				listActionIsExecuted = append(listActionIsExecuted, aie)

				continue
			}

			//проверяем не ожидает ли или выполняется скачивание файлов с источника
			if qts.IsExistTaskDownloadQueueTaskStorage(s.ID) {
				aie.IsSuccess = false
				aie.MessageFailure = fmt.Sprintf("невозможно выполнить переподключение, с сенсора с ID %v осуществляется загрузка файлов", s.ID)

				listActionIsExecuted = append(listActionIsExecuted, aie)

				continue
			}

			//закрываем соединение и удаляем дискриптор
			if cl, isExist := isl.GetLinkWebsocketConnect(sourceInfo.IP); isExist {
				cl.Link.Close()

				isl.DelLinkWebsocketConnection(sourceInfo.IP)
			}

			aie.IsSuccess = true

			listActionIsExecuted = append(listActionIsExecuted, aie)
		}
	}

	return &listActionIsExecuted, &listInvalidSource, nil
}

//getSourceListForWriteToBD получаем ID источников по которым нужно актуализировать информацию
// в БД, к ним относятся источники для которых выполненно действие
// add, delete, update
func getSourceListsForWriteToBD(
	ml *[]configure.DetailedListSources,
	l *[]configure.ActionTypeListSources,
	clientName string,
	mcpf int8) (*[]configure.InformationAboutSource, *[]configure.InformationAboutSource, *[]int) {

	typeAreaNetwork := "ip"

	var listAdd, listUpdate []configure.InformationAboutSource
	listDel := []int{}

	for _, source := range *l {
		if !source.IsSuccess {
			continue
		}

		if source.ActionType == "delete" {
			listDel = append(listDel, source.ID)

			continue
		}

		for _, s := range *ml {
			if strings.ToLower(s.Argument.Settings.TypeAreaNetwork) == "pppoe" {
				typeAreaNetwork = "pppoe"
			}

			if s.ID == source.ID {
				si := configure.InformationAboutSource{
					ID:            source.ID,
					IP:            s.Argument.IP,
					Token:         s.Argument.Token,
					ShortName:     s.Argument.ShortName,
					Description:   s.Argument.Description,
					AsServer:      s.Argument.Settings.AsServer,
					NameClientAPI: clientName,
					SourceSetting: configure.InfoServiceSettings{
						IfAsServerThenPort:        s.Argument.Settings.Port,
						EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
						MaxCountProcessFiltration: mcpf,
						StorageFolders:            s.Argument.Settings.StorageFolders,
						TypeAreaNetwork:           typeAreaNetwork,
					},
				}

				if source.ActionType == "add" {
					listAdd = append(listAdd, si)

					continue
				}

				if source.ActionType == "update" {
					listUpdate = append(listUpdate, si)

					continue
				}

				break
			}
		}
	}

	return &listAdd, &listUpdate, &listDel
}
