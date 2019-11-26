package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromDB обработчик сообщений приходящих от модуля взаимодействия с базой данных
func HandlerMsgFromDB(
	outCoreChans HandlerOutChans,
	res *configure.MsgBetweenCoreAndDB,
	hsm HandlersStoringMemory,
	mtsfda int64,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanDropNI <-chan string) {

	funcName := "HandlerMsgFromDB"

	switch res.MsgRecipient {
	case "Core module":
		switch res.MsgSection {
		case "error notification":
			//если сообщение об ошибке только для ядра приложения
			if en, ok := res.AdvancedOptions.(configure.ErrorNotification); ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(en.ErrorBody),
					FuncName:    funcName,
				})

				return
			}

		case "download control":
			switch res.Instruction {
			case "all information about task":
				/* ФАКТИЧЕСКИ ЭТО ЗАПУСК СКАЧИВАНИЯ ФАЙЛОВ */

				//проверяем ряд параметров в задаче для изменения проверочного статуса задачи в QueueStoringMemoryTask
				if err := checkParametersDownloadTask(res, hsm, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}
			}
		}

	case "API module":
		taskInfo, taskIsExist := hsm.SMT.GetStoringMemoryTask(res.TaskID)
		if !taskIsExist {
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprintf("task with %v not found%v", res.TaskID, funcName),
				FuncName:    funcName,
			})

			return
		}

		switch res.MsgSection {
		case "source list":
			if err := getCurrentSourceListForAPI(outCoreChans.OutCoreChanAPI, res, taskInfo.ClientID, taskInfo.ClientTaskID); err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				return
			}

			//устанавливаем статус задачи как выполненую
			hsm.SMT.CompleteStoringMemoryTask(res.TaskID)

		case "source control":
			//пока заглушка

		case "source telemetry":
			//пока заглушка

		case "filtration control":
			isNotComplete := taskInfo.TaskParameter.FiltrationTask.Status != "complete"
			moreThanMax := taskInfo.TaskParameter.FiltrationTask.SizeFilesFoundResultFiltering > mtsfda
			sizeFilesFoundIsZero := taskInfo.TaskParameter.FiltrationTask.SizeFilesFoundResultFiltering == 0
			taskTypeNotFiltr := taskInfo.TaskType != "filtration control"

			if taskTypeNotFiltr || isNotComplete || moreThanMax || sizeFilesFoundIsZero {
				//отмечаем задачу как завершенную в списке очередей
				if err := hsm.QTS.ChangeTaskStatusQueueTask(taskInfo.TaskParameter.FiltrationTask.ID, res.TaskID, "complete"); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

				return
			}

			listFoundFiles := taskInfo.TaskParameter.FiltrationTask.FoundFilesInformation

			//готовим список файлов предназначенный для загрузки
			listDownloadFiles := make(map[string]*configure.DownloadFilesInformation, len(listFoundFiles))
			for fn, v := range listFoundFiles {
				listDownloadFiles[fn] = &configure.DownloadFilesInformation{}
				listDownloadFiles[fn].Size = v.Size
				listDownloadFiles[fn].Hex = v.Hex
			}

			sourceID := taskInfo.TaskParameter.FiltrationTask.ID

			//получаем параметры фильтрации
			qti, err := hsm.QTS.GetQueueTaskStorage(sourceID, res.TaskID)
			if err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				return
			}

			//добавляем задачу в очередь
			hsm.QTS.AddQueueTaskStorage(res.TaskID, sourceID, configure.CommonTaskInfo{
				IDClientAPI:     res.IDClientAPI,
				TaskIDClientAPI: res.TaskIDClientAPI,
				TaskType:        "download control",
			}, &configure.DescriptionParametersReceivedFromUser{
				FilterationParameters:         qti.TaskParameters.FilterationParameters,
				PathDirectoryForFilteredFiles: taskInfo.TaskParameter.FiltrationTask.PathStorageSource,
			})

			//информационное сообщение о том что задача добавлена в очередь
			notifications.SendNotificationToClientAPI(
				outCoreChans.OutCoreChanAPI,
				notifications.NotificationSettingsToClientAPI{
					MsgType: "success",
					MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID:   sourceID,
						TaskType:   "скачивание файлов",
						TaskAction: "задача автоматически добавлена в очередь",
					}),
					Sources: []int{sourceID},
				},
				res.TaskIDClientAPI,
				res.IDClientAPI)

			//добавляем подтвержденный список файлов для скачивания
			if err := hsm.QTS.AddConfirmedListFiles(sourceID, res.TaskID, listDownloadFiles); err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})
			}

			//устанавливаем проверочный статус источника для данной задачи как подключен
			if err := hsm.QTS.ChangeAvailabilityConnectionOnConnection(sourceID, res.TaskID); err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})
			}

			//изменяем статус наличия файлов для скачивания
			if err := hsm.QTS.ChangeAvailabilityFilesDownload(sourceID, res.TaskID); err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})
			}

		case "download control":
			//пока заглушка

		case "information search results":
			//пока заглушка

		case "error notification":
			en, ok := res.AdvancedOptions.(configure.ErrorNotification)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error section type 'error notification'",
					FuncName:    funcName,
				})

				return
			}

			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: fmt.Sprint(en.ErrorBody),
				FuncName:    funcName,
			})

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType: "danger",
				MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
					Message: "ошибка базы данных при обработке запроса",
				}),
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, res.IDClientAPI)

		case "message notification":
			mn, ok := res.AdvancedOptions.(configure.MessageNotification)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error section type 'message notification'",
					FuncName:    funcName,
				})

				return
			}

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType:        mn.CriticalityMessage,
				MsgDescription: mn.HumanDescriptionNotification,
				Sources:        mn.Sources,
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, res.IDClientAPI)
		}
	case "NI module":
		switch res.MsgSection {
		case "source list":
			outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
				Section:         "source control",
				Command:         "create list",
				AdvancedOptions: res.AdvancedOptions,
			}

		case "source control":
			//пока заглушка

		case "filtration control":
			tfmfi, ok := res.AdvancedOptions.(configure.TypeFiltrationMsgFoundIndex)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error section type 'message notification'",
					FuncName:    funcName,
				})

				return
			}

			msg := configure.MsgBetweenCoreAndNI{
				TaskID:   res.TaskID,
				Section:  "filtration control",
				Command:  "start",
				SourceID: tfmfi.FilteringOption.ID,
			}

			mtfc := configure.MsgTypeFiltrationControl{
				MsgType: "filtration",
				Info: configure.SettingsFiltrationControl{
					TaskID:  res.TaskID,
					Command: "start",
					Options: tfmfi.FilteringOption,
				},
			}

			if !tfmfi.IndexIsFound {
				msgJSON, err := json.Marshal(mtfc)
				if err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					return
				}

				//если индексы не найдены
				msg.AdvancedOptions = msgJSON
				outCoreChans.OutCoreChanNI <- &msg

				return
			}

			//размер части сообщения
			const maxChunk = 100
			var numIndexFiles int
			var tmpList map[string]int

			for k, v := range tfmfi.IndexData {
				nf := len(v)
				numIndexFiles += nf

				tmpList[k] = nf
			}

			numChunk := common.GetCountPartsMessage(tmpList, maxChunk)

			//если индексы найдены
			mtfc.Info.IndexIsFound = true
			mtfc.Info.CountIndexFiles = numIndexFiles
			mtfc.Info.NumberMessagesFrom = [2]int{0, numChunk}

			//отправляем первое сообщение (фактически нулевое, так как оно не содержит списка файлов)
			msgJSON, err := json.Marshal(mtfc)
			if err != nil {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprint(err),
					FuncName:    funcName,
				})

				return
			}

			msg.AdvancedOptions = msgJSON
			outCoreChans.OutCoreChanNI <- &msg

			//информация о задаче по заданному ID
			t, ok := hsm.SMT.GetStoringMemoryTask(res.TaskID)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprintf("task with %v not found%v", res.TaskID, funcName),
					FuncName:    funcName,
				})

				return
			}

			//отправляем последующие сообщения содержащие списки файлов,
			// параметры фильтрации данные сообщения уже не содержат
		DONE:
			for i := 1; i < numChunk; i++ {
				select {
				case <-t.ChanStopTransferListFiles:
					break DONE

				default:
					listFiles := common.GetChunkListFiles(i, maxChunk, numChunk, tfmfi.IndexData)

					mtfc.Info.NumberMessagesFrom[0] = i
					mtfc.Info.ListFilesReceivedIndex = listFiles

					msgJSON, err := json.Marshal(mtfc)
					if err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})

						return
					}

					msg.AdvancedOptions = msgJSON
					outCoreChans.OutCoreChanNI <- &msg

				}
			}

		case "download control":
			//пока заглушка

		}
	default:
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "the module receiver is not defined, request processing is impossible",
			FuncName:    funcName,
		})
	}
}
