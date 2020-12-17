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
	maxTotalSizeDownloadFiles int64,
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

		case "filtration control":
			//пока заглушка

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
				if err := sendMsgCompliteTaskSearchShortInformationAboutTask(res, hsm.TSSQ, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

			//полная информация по taskID, может быть не полным список файлов (максимальный размер не больше 50)
			case "information by task ID":
				if err := sendMsgCompliteTaskSearchInformationByTaskID(res, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

			//дополнительный список файлов по конкретной задаче
			case "list files by task ID":
				if err := sendMsgCompliteTaskListFilesByTaskID(res, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

			//краткая информация по задаче и параметры для перевода задачи в состояние 'обработана'
			case "mark an task as completed processed":
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

			case "delete all information about a task":
				resMsg := configure.DeleteInformationListTaskCompletedResponse{
					MsgOption: configure.MarkTaskCompletedResponseOption{SuccessStatus: true},
				}
				resMsg.MsgType = "command"
				resMsg.MsgSection = "information search control"
				resMsg.MsgInstruction = "delete all information about a task"
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

			case "get common analytics information about task ID":
				if err := sendMsgCompliteTaskGetCommonAnalyticsInformationAboutTaskID(res, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
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
			if err := getCurrentSourceListForAPI(outCoreChans.OutCoreChanAPI, hsm.ISL, res, taskInfo.ClientID, taskInfo.ClientTaskID); err != nil {
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

			//fmt.Printf("func '%v', from %v,  for %v, section: %v try send API filtration complete \n", funcName, res.MsgGenerator, res.MsgRecipient, res.MsgSection)

			taskInfo, taskIsExist := hsm.SMT.GetStoringMemoryTask(res.TaskID)
			if !taskIsExist {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprintf("task with %v not found%v (section: 'filtration control')", res.TaskID, funcName),
					FuncName:    funcName,
				})

				return
			}

			//fmt.Printf("func '%v', from %v,  for %v, section: %v information about task is exist (restore task) \n", funcName, res.MsgGenerator, res.MsgRecipient, res.MsgSection)

			//клиенту API
			ao, ok := res.AdvancedOptions.(configure.TypeFiltrationMsgFoundFileInformationAndTaskStatus)
			if ok && taskIsExist {

				//fmt.Printf("func '%v', Section: %v, send to client API ---> (restore task) BEFORE\n", funcName, res.MsgSection)

				//упаковываем в JSON и отправляем информацию о ходе фильтрации клиенту API
				// при чем если статус 'execute', то отправляем еще и содержимое поля 'FoundFilesInformation',
				// а если статус фильтрации 'stop' или 'complete' то данное поле не заполняем
				if err := sendInformationFiltrationTask(outCoreChans.OutCoreChanAPI, hsm.SMT, taskInfo, ao.ListFoundFile, taskInfo.TaskParameter.FiltrationTask.ID, res.TaskID); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

				}

				//запуск автоматической передачи файлов
				if err := HandlerAutomaticDownloadFiles(res.TaskID, hsm.SMT, hsm.QTS, maxTotalSizeDownloadFiles, outCoreChans.OutCoreChanAPI); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}
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

			//не отправляем сообщение так как запрос был сформирован автоматически
			if mn.RequestIsGeneratedAutomatically {
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
