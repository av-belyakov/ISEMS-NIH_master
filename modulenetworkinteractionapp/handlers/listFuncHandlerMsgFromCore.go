package handlers

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
)

type validateUserDataSourceSettings struct {
	SourceID  int
	ShortName string
	IP        string
	Token     string
	AsServer  bool
	Settings  configure.InfoServiceSettings
}

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
			ShortName:  s.ShortName,
			IP:         s.IP,
			Token:      s.Token,
			ClientName: s.NameClientAPI,
			AsServer:   s.AsServer,
			Settings:   s.SourceSetting,
		})
	}
}

func validateUserData(l *[]configure.DetailedListSources, mcpf int8) (*[]validateUserDataSourceSettings, []int) {
	listTrastedSources := make([]validateUserDataSourceSettings, 0, len(*l))
	listInvalidSource := []int{}

	for _, s := range *l {
		ipIsValid, _ := common.CheckStringIP(s.Argument.IP)
		tokenIsValid, _ := common.CheckStringToken(s.Argument.Token)
		foldersIsValid, _ := common.CheckFolders(s.Argument.Settings.StorageFolders)

		if !ipIsValid || !tokenIsValid || !foldersIsValid {
			listInvalidSource = append(listInvalidSource, s.ID)

			continue
		}

		if (s.Argument.Settings.MaxCountProcessFiltration > 0) && (s.Argument.Settings.MaxCountProcessFiltration < 10) {
			mcpf = s.Argument.Settings.MaxCountProcessFiltration
		}

		serverPort := s.Argument.Settings.Port
		if s.Argument.Settings.Port == 0 {
			serverPort = 13113
		}

		typeAreaNetwork := "ip"
		if strings.ToLower(s.Argument.Settings.TypeAreaNetwork) == "pppoe" {
			typeAreaNetwork = "pppoe"
		}

		listTrastedSources = append(listTrastedSources, validateUserDataSourceSettings{
			SourceID:  s.ID,
			ShortName: s.Argument.ShortName,
			IP:        s.Argument.IP,
			Token:     s.Argument.Token,
			AsServer:  s.Argument.Settings.AsServer,
			Settings: configure.InfoServiceSettings{
				IfAsServerThenPort:        serverPort,
				EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
				MaxCountProcessFiltration: mcpf,
				StorageFolders:            s.Argument.Settings.StorageFolders,
				TypeAreaNetwork:           typeAreaNetwork,
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

	//если список источников в памяти приложения пуст
	if isl.GetCountSources() == 0 {
		for _, si := range *listTrastedSources {
			isl.AddSourceSettings(si.SourceID, configure.SourceSetting{
				ShortName:  si.ShortName,
				IP:         si.IP,
				Token:      si.Token,
				ClientName: clientName,
				AsServer:   si.AsServer,
				Settings: configure.InfoServiceSettings{
					IfAsServerThenPort:        si.Settings.IfAsServerThenPort,
					EnableTelemetry:           si.Settings.EnableTelemetry,
					MaxCountProcessFiltration: si.Settings.MaxCountProcessFiltration,
					StorageFolders:            si.Settings.StorageFolders,
					TypeAreaNetwork:           si.Settings.TypeAreaNetwork,
				},
			})
		}

		return listTaskExecuted, listInvalidSource
	}

	var sourcesIsReconnect bool

	_, listDisconnected := isl.GetListsConnectedAndDisconnectedSources()

	for _, si := range *listTrastedSources {
		//если источника нет в списке
		s, isExist := isl.GetSourceSetting(si.SourceID)
		if !isExist {
			isl.AddSourceSettings(si.SourceID, configure.SourceSetting{
				ShortName:  si.ShortName,
				IP:         si.IP,
				Token:      si.Token,
				ClientName: clientName,
				AsServer:   si.AsServer,
				Settings: configure.InfoServiceSettings{
					IfAsServerThenPort:        si.Settings.IfAsServerThenPort,
					EnableTelemetry:           si.Settings.EnableTelemetry,
					MaxCountProcessFiltration: si.Settings.MaxCountProcessFiltration,
					StorageFolders:            si.Settings.StorageFolders,
					TypeAreaNetwork:           si.Settings.TypeAreaNetwork,
				},
			})

			continue
		}

		//если на источнике ожидает выполнения или выполняется какая либо задача задача
		if _, ok := qts.GetAllTaskQueueTaskStorage(si.SourceID); ok {
			listTaskExecuted = append(listTaskExecuted, si.SourceID)
		}

		//проверяем имеет ли право клиент делать какие либо изменения с информацией по источнику
		if clientName != s.ClientName && clientName != "root token" {
			listInvalidSource = append(listInvalidSource, si.SourceID)

			continue
		}

		changeToken := (s.Token != si.Token)
		changeIP := (s.IP != si.IP)
		changeAsServer := (s.AsServer != si.AsServer)
		changeEnTelemetry := (s.Settings.EnableTelemetry != si.Settings.EnableTelemetry)

		//проверяем параметры подключения
		if changeToken || changeIP || changeAsServer || changeEnTelemetry {
			sourcesIsReconnect = true
		}

		//полная замена информации об источнике
		if _, ok := listDisconnected[si.SourceID]; ok {
			isl.DelSourceSettings(si.SourceID)

			isl.AddSourceSettings(si.SourceID, configure.SourceSetting{
				ShortName:  si.ShortName,
				IP:         si.IP,
				Token:      si.Token,
				ClientName: clientName,
				AsServer:   si.AsServer,
				Settings: configure.InfoServiceSettings{
					IfAsServerThenPort:        si.Settings.IfAsServerThenPort,
					EnableTelemetry:           si.Settings.EnableTelemetry,
					MaxCountProcessFiltration: si.Settings.MaxCountProcessFiltration,
					StorageFolders:            si.Settings.StorageFolders,
					TypeAreaNetwork:           si.Settings.TypeAreaNetwork,
				},
			})

			continue
		}

		if sourcesIsReconnect {
			//закрываем соединение и удаляем дискриптор
			if cl, isExist := isl.GetLinkWebsocketConnect(si.IP); isExist {
				cl.Link.Close()

				isl.DelLinkWebsocketConnection(si.IP)
			}

			sourcesIsReconnect = false
		}
	}

	return listTaskExecuted, listInvalidSource
}

//getSourceListToStoreDB формирует список источников для последующей их записи в БД
func getSourceListToStoreDB(trustedSoures []int, l *[]configure.DetailedListSources, clientName string, mcpf int8) (*[]configure.InformationAboutSource, error) {
	list := make([]configure.InformationAboutSource, 0, len(*l))

	typeAreaNetwork := "ip"

	sort.Ints(trustedSoures)
	for _, s := range *l {
		for _, trustedSoure := range trustedSoures {
			if trustedSoure == s.ID {
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

		//проверяем максимальное кол-во одновременно запущеных задач фильтрации
		if s.Argument.Settings.MaxCountProcessFiltration > 1 && 10 > s.Argument.Settings.MaxCountProcessFiltration {
			mcpf = s.Argument.Settings.MaxCountProcessFiltration
		}

		//есть ли источник с таким ID
		if sourceInfo, ok = isl.GetSourceSetting(s.ID); !ok {
			if s.ActionType == "add" {
				isl.AddSourceSettings(s.ID, configure.SourceSetting{
					ShortName:  s.Argument.ShortName,
					IP:         s.Argument.IP,
					Token:      s.Argument.Token,
					ClientName: clientName,
					AsServer:   s.Argument.Settings.AsServer,
					Settings: configure.InfoServiceSettings{
						IfAsServerThenPort:        s.Argument.Settings.Port,
						EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
						MaxCountProcessFiltration: mcpf,
						StorageFolders:            s.Argument.Settings.StorageFolders,
						TypeAreaNetwork:           typeAreaNetwork,
					},
				})

				aie.IsSuccess = true
				aie.MessageFailure = common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   s.ID,
					TaskType:   "управление источниками",
					TaskAction: "задача выполнена",
					Message:    "источник был успешно добавлен в базу данных",
				})
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
			aie.MessageFailure = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   s.ID,
				TaskType:   "управление источниками",
				TaskAction: "задача отклонена",
				Message:    "невозможно добавить источник, источник с таким ID уже существует",
			})

			listActionIsExecuted = append(listActionIsExecuted, aie)

			continue
		}

		if s.ActionType == "update" || s.ActionType == "delete" {
			//проверяем имеет ли право клиент делать какие либо изменения с информацией по источнику
			if (clientName != sourceInfo.ClientName) && (clientName != "root token") {
				aie.IsSuccess = false
				aie.MessageFailure = common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   s.ID,
					TaskType:   "управление источниками",
					TaskAction: "задача отклонена",
					Message:    "недостаточно прав для выполнения действий с источником, возможно он был добавлен другим клиентом",
				})

				listActionIsExecuted = append(listActionIsExecuted, aie)

				continue
			}

			//проверяем ожидает или выполняется на источнике какая либо задача
			if listSourceTask, ok := qts.GetAllTaskQueueTaskStorage(s.ID); ok {

				fmt.Printf("func 'listFuncHandlerMsgFromCore', ===== ALL TASKS source ID '%v' (%v)\n", s.ID, listSourceTask)

				if len(listSourceTask) > 0 {
					aie.IsSuccess = false
					aie.MessageFailure = common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID:   s.ID,
						TaskType:   "управление источниками",
						TaskAction: "задача отклонена",
						Message:    "невозможно выполнить действия на источнике, так как в настоящее время на данном сенсоре ожидает выполнения или уже выполняется какая либо задача",
					})

					listActionIsExecuted = append(listActionIsExecuted, aie)

					continue
				}
			}
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
				ShortName:  s.Argument.ShortName,
				IP:         s.Argument.IP,
				Token:      s.Argument.Token,
				ClientName: clientName,
				AsServer:   s.Argument.Settings.AsServer,
				Settings: configure.InfoServiceSettings{
					IfAsServerThenPort:        s.Argument.Settings.Port,
					EnableTelemetry:           s.Argument.Settings.EnableTelemetry,
					MaxCountProcessFiltration: mcpf,
					StorageFolders:            s.Argument.Settings.StorageFolders,
					TypeAreaNetwork:           typeAreaNetwork,
				},
			})

			aie.IsSuccess = true
			aie.MessageFailure = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   s.ID,
				TaskType:   "управление источниками",
				TaskAction: "задача выполнена",
				Message:    "информация по источнику была успешно обновлена",
			})

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
			aie.MessageFailure = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   s.ID,
				TaskType:   "управление источниками",
				TaskAction: "задача выполнена",
				Message:    "информация по источнику была успешно удалена",
			})

			listActionIsExecuted = append(listActionIsExecuted, aie)

			continue
		}

		//при переподключении
		if s.ActionType == "reconnect" {
			if !sourceInfo.ConnectionStatus {
				aie.IsSuccess = false
				aie.MessageFailure = common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   s.ID,
					TaskType:   "управление источниками",
					TaskAction: "задача отклонена",
					Message:    "невозможно выполнить переподключение, источник не подключен",
				})

				listActionIsExecuted = append(listActionIsExecuted, aie)

				continue
			}

			//проверяем не ожидает ли или выполняется скачивание файлов с источника
			if qts.IsExistTaskDownloadQueueTaskStorage(s.ID) {
				aie.IsSuccess = false
				aie.MessageFailure = common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   s.ID,
					TaskType:   "управление источниками",
					TaskAction: "задача отклонена",
					Message:    "невозможно выполнить переподключение, с источника осуществляется скачивание файлов",
				})

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

//getSourceListForWriteToDB получаем ID источников по которым нужно актуализировать информацию
// в БД, к ним относятся источники для которых выполненно действие add, delete, update
func getSourceListsForWriteToDB(
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

			//проверяем максимальное кол-во одновременно запущеных задач фильтрации
			if s.Argument.Settings.MaxCountProcessFiltration > 1 && 10 > s.Argument.Settings.MaxCountProcessFiltration {
				mcpf = s.Argument.Settings.MaxCountProcessFiltration
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
