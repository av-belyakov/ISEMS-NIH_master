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

			//fmt.Println("func 'HandlerMsgFromDB', MsgRecipient: 'Core module', MsgSection: 'error notification'")

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

		case "information search control":
			switch res.Instruction {
			case "short search result":
				//fmt.Printf("func 'HandlerMsgFromDB', Instruction: 'information search control', Response: '%v'\n", res)

				if err := sendMsgCompliteTaskSearchShortInformationAboutTask(res, hsm.TSSQ, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

			//полная информация по taskID, может быть не полным список файлов (максимальный размер не больше 50)
			case "information by task ID":
				//fmt.Printf("func 'HandlerMsgFromDB', Section: 'information search control', Instruction: 'information by task ID', Response: '%v'\n", res)

				if err := sendMsgCompliteTaskSearchInformationByTaskID(res, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

			//дополнительный список файлов по конкретной задаче
			case "list files by task ID":
				//fmt.Printf("func 'HandlerMsgFromDB', Section: 'information search control', Instruction: 'list files by task ID', Response: '%v'\n", res)

				if err := sendMsgCompliteTaskListFilesByTaskID(res, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

				//краткая информация по задаче и параметры для перевода задачи в состояние 'обработана'
			case "mark an task as completed processed":
				fmt.Printf("func 'HandlerMsgFromDB', Section: 'information search control', Instruction: 'mark an task as completed', Response: '%v'\n", res)

				resMsg := configure.MarkTaskCompletedResponse{
					MsgOption: configure.MarkTaskCompletedResponseOption{
						SuccessStatus: false,
						RequestTaskID: res.TaskID,
					},
				}
				resMsg.MsgType = "command"
				resMsg.MsgSection = "information search control"
				resMsg.MsgInstruction = "mark an task as completed"
				resMsg.ClientTaskID = res.TaskIDClientAPI

				tgitfmtcp, ok := res.AdvancedOptions.(configure.TypeGetInfoTaskFromMarkTaskCompleteProcess)
				if !ok {
					notifications.SendNotificationToClientAPI(
						outCoreChans.OutCoreChanAPI,
						notifications.NotificationSettingsToClientAPI{
							MsgType: "danger",
							MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
								TaskType:   "изменение статуса задачи на 'завершена'",
								TaskAction: "задача отклонена",
								Message:    "внутренняя ошибка приложения",
							}),
						},
						res.TaskIDClientAPI,
						res.IDClientAPI)

					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: "type conversion error section type 'information search control'",
						FuncName:    funcName,
					})

					msgJSON, err := json.Marshal(resMsg)
					if err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})

						return
					}

					outCoreChans.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
						MsgGenerator: "Core module",
						MsgRecipient: "API module",
						IDClientAPI:  res.IDClientAPI,
						MsgJSON:      msgJSON,
					}

					return
				}

				//задача по переданному ID не найдена
				if !tgitfmtcp.TaskIsExist {
					fmt.Println("func 'HandlerMsgFromDB', информационное сообщение о том что, задача по переданному ID не найдена")

					notifications.SendNotificationToClientAPI(
						outCoreChans.OutCoreChanAPI,
						notifications.NotificationSettingsToClientAPI{
							MsgType: "warning",
							MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
								TaskType:   "изменение статуса задачи на 'завершена'",
								TaskAction: "задача отклонена",
								Message:    "невозможно отметить задачу как завершенную, так как задачи по принятому идентификатору в СУБД не найдено",
							}),
						},
						res.TaskIDClientAPI,
						res.IDClientAPI)

					msgJSON, err := json.Marshal(resMsg)
					if err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})

						return
					}

					outCoreChans.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
						MsgGenerator: "Core module",
						MsgRecipient: "API module",
						IDClientAPI:  res.IDClientAPI,
						MsgJSON:      msgJSON,
					}

					return
				}

				//проверяем была ли выполненна задача по фильтрации
				// и загружался ли хотя бы один файл
				if !tgitfmtcp.FiltrationTaskStatus || !tgitfmtcp.FilesDownloaded {

					fmt.Println("func 'HandlerMsgFromDB', информационное сообщение о том что, невозможно отметить задачу как завершенную")

					//информационное сообщение о том что, невозможно отметить задачу как завершенную
					notifications.SendNotificationToClientAPI(
						outCoreChans.OutCoreChanAPI,
						notifications.NotificationSettingsToClientAPI{
							MsgType: "warning",
							MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
								TaskType:   "изменение статуса задачи на 'завершена'",
								TaskAction: "задача отклонена",
								Message:    "невозможно отметить задачу как завершенную, так как данная задача не была выполнена полностью или не один из найденных файлов выгружен не был",
							}),
						},
						res.TaskIDClientAPI,
						res.IDClientAPI)

					msgJSON, err := json.Marshal(resMsg)
					if err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})

						return
					}

					outCoreChans.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
						MsgGenerator: "Core module",
						MsgRecipient: "API module",
						IDClientAPI:  res.IDClientAPI,
						MsgJSON:      msgJSON,
					}

					return
				}

				fmt.Println("func 'HandlerMsgFromDB', отправляем запрос модулю БД для изменения состояния задачи")

				//отправляем запрос модулю БД для изменения состояния задачи
				outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
					MsgGenerator:    "Core module",
					MsgRecipient:    "DB module",
					MsgSection:      "information search control",
					Instruction:     "mark an task as completed",
					TaskID:          res.TaskID,
					IDClientAPI:     res.IDClientAPI,
					TaskIDClientAPI: res.TaskIDClientAPI,
					AdvancedOptions: tgitfmtcp,
				}

			case "mark an task as completed":

				fmt.Printf("func 'HandlerMsgFromDB', Section: 'information search control', Instruction: 'mark an task as completed', Response: '%v'\n", res)

				//информационное сообщение о том что, невозможно отметить задачу как завершенную
				notifications.SendNotificationToClientAPI(
					outCoreChans.OutCoreChanAPI,
					notifications.NotificationSettingsToClientAPI{
						MsgType: "success",
						MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
							TaskType:   "изменение статуса задачи на 'завершена'",
							TaskAction: "задача выполнена",
							Message:    "выбранная задача была успешно отмечена как 'завершена'",
						}),
					},
					res.TaskIDClientAPI,
					res.IDClientAPI)

				resMsg := configure.MarkTaskCompletedResponse{
					MsgOption: configure.MarkTaskCompletedResponseOption{
						SuccessStatus: true,
						RequestTaskID: res.TaskID,
					},
				}
				resMsg.MsgType = "command"
				resMsg.MsgSection = "information search control"
				resMsg.MsgInstruction = "mark an task as completed"
				resMsg.ClientTaskID = res.TaskIDClientAPI
				msgJSON, err := json.Marshal(resMsg)
				if err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					return
				}

				outCoreChans.OutCoreChanAPI <- &configure.MsgBetweenCoreAndAPI{
					MsgGenerator: "Core module",
					MsgRecipient: "API module",
					IDClientAPI:  res.IDClientAPI,
					MsgJSON:      msgJSON,
				}

			}
		}

	case "API module":
		if res.MsgSection == "error notification" {
			en, ok := res.AdvancedOptions.(configure.ErrorNotification)
			if !ok {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: "type conversion error section type 'error notification'",
					FuncName:    funcName,
				})
			} else {
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

				notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, res.TaskIDClientAPI, res.IDClientAPI)
			}
		}

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

				fmt.Printf("ERROR: %v\n", err)

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

		case "information search control":
			//пока заглушка

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
