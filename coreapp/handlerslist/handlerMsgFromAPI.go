package handlerslist

import (
	"encoding/json"
	"fmt"
	"time"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromAPI обработчик сообщений приходящих от модуля API
func HandlerMsgFromAPI(
	outCoreChans HandlerOutChans,
	msg *configure.MsgBetweenCoreAndAPI,
	hsm HandlersStoringMemory,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) {

	funcName := "HeaderMsgFromAPI"

	msgc := configure.MsgCommon{}

	nsErrMsg := notifications.NotificationSettingsToClientAPI{
		MsgType: "danger",
		MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
			TaskAction: "задача отклонена",
			Message:    "получен некорректный формат JSON сообщения",
		}),
	}

	msgJSON, ok := msg.MsgJSON.([]byte)
	if !ok {
		notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, "", msg.IDClientAPI)
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "bad cast type JSON messages",
			FuncName:    funcName,
		})

		return
	}

	if err := json.Unmarshal(msgJSON, &msgc); err != nil {
		notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, "", msg.IDClientAPI)
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "bad cast type JSON messages",
			FuncName:    funcName,
		})

		return
	}

	//логируем запросы клиентов
	saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
		TypeMessage: "requests",
		Description: fmt.Sprintf("client name: '%v' (%v), request: type = %v, section = %v, instruction = %v, client task ID = %v", msg.ClientName, msg.ClientIP, msgc.MsgType, msgc.MsgSection, msgc.MsgInstruction, msgc.ClientTaskID),
	})

	if msgc.MsgType == "information" {
		if msgc.MsgSection == "source control" {
			if msgc.MsgInstruction == "send new source list" {
				var scmo configure.SourceControlMsgOptions

				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, msgc.ClientTaskID, msg.IDClientAPI)
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: "bad cast type JSON messages",
						FuncName:    funcName,
					})

					return
				}

				taskID := common.GetUniqIDFormatMD5(msg.IDClientAPI + "_" + msgc.ClientTaskID)

				//добавляем новую задачу
				hsm.SMT.AddStoringMemoryTask(taskID, configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					ClientName:      msg.ClientName,
					Section:         "source control",
					Command:         "load list",
					AdvancedOptions: scmo.MsgOptions,
				}

				return
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, msgc.ClientTaskID, msg.IDClientAPI)
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: "in the json message is not found the right option for 'MsgSection'",
				FuncName:    funcName,
			})

			return
		}

		notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, msgc.ClientTaskID, msg.IDClientAPI)
		saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
			Description: "in the json message is not found the right option for 'MsgSection'",
			FuncName:    funcName,
		})

		return
	}

	if msgc.MsgType == "command" {
		switch msgc.MsgSection {

		// УПРАВЛЕНИЕ ИСТОЧНИКАМИ
		case "source control":
			//получить актуальный список источников
			if msgc.MsgInstruction == "get an updated list of sources" {
				taskID := common.GetUniqIDFormatMD5(msg.IDClientAPI + "_" + msgc.ClientTaskID)

				//добавляем новую задачу
				hsm.SMT.AddStoringMemoryTask(taskID, configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
					MsgGenerator: "API module",
					MsgRecipient: "DB module",
					MsgSection:   "source control",
					Instruction:  "find_all",
					TaskID:       taskID,
				}

				return
			}

			//выполнить какие либо действия над источниками
			if msgc.MsgInstruction == "performing an action" {
				var scmo configure.SourceControlMsgOptions
				if err := json.Unmarshal(msgJSON, &scmo); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, msgc.ClientTaskID, msg.IDClientAPI)
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: "bad cast type JSON messages",
						FuncName:    funcName,
					})

					return
				}

				taskID := common.GetUniqIDFormatMD5(msg.IDClientAPI + "_" + msgc.ClientTaskID)

				//добавляем новую задачу
				hsm.SMT.AddStoringMemoryTask(taskID, configure.TaskDescription{
					ClientID:                        msg.IDClientAPI,
					ClientTaskID:                    msgc.ClientTaskID,
					TaskType:                        msgc.MsgSection,
					ModuleThatSetTask:               "API module",
					ModuleResponsibleImplementation: "NI module",
					TimeUpdate:                      time.Now().Unix(),
				})

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:          taskID,
					ClientName:      msg.ClientName,
					Section:         "source control",
					Command:         "perform actions on sources",
					AdvancedOptions: scmo,
				}

				return
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, msgc.ClientTaskID, msg.IDClientAPI)
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: "in the json message is not found the right option for 'MsgInstruction'",
				FuncName:    funcName,
			})

			return

		// УПРАВЛЕНИЕ ФИЛЬТРАЦИЕЙ
		case "filtration control":
			//обработка команды на запуск фильтрации
			if msgc.MsgInstruction == "to start filtering" {
				var fcts configure.FiltrationControlTypeStart
				if err := json.Unmarshal(msgJSON, &fcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, "", msg.IDClientAPI)
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: "bad cast type JSON messages",
						FuncName:    funcName,
					})

					return
				}

				go handlerFiltrationControlTypeStart(&fcts, hsm, msg.IDClientAPI, saveMessageApp, outCoreChans.OutCoreChanAPI)

				return
			}

			//команда на останов фильтрации
			if msgc.MsgInstruction == "to cancel filtering" {
				//ищем ожидающую в очереди задачу по ее ID
				sourceID, taskID, err := hsm.QTS.SearchTaskForClientIDQueueTaskStorage(msgc.ClientTaskID)
				if err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					notifications.SendNotificationToClientAPI(
						outCoreChans.OutCoreChanAPI,
						notifications.NotificationSettingsToClientAPI{
							MsgType: "danger",
							MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
								SourceID:   sourceID,
								TaskType:   "фильтрация",
								TaskAction: "задача отклонена",
								Message:    fmt.Sprintf("'по переданному идентификатору %v ожидающих или выполняемых задач не обнаружено'", msgc.ClientTaskID),
							}),
						},
						msgc.ClientTaskID,
						msg.IDClientAPI)

					return
				}

				//проверяем наличие задачи в StoringMemoryTask
				isExist := hsm.SMT.CheckIsExistMemoryTask(taskID)
				if !isExist {
					notifications.SendNotificationToClientAPI(
						outCoreChans.OutCoreChanAPI,
						notifications.NotificationSettingsToClientAPI{
							MsgType: "danger",
							MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
								SourceID:   sourceID,
								TaskType:   "фильтрация",
								TaskAction: "задача отклонена",
								Message:    fmt.Sprintf("'по переданному идентификатору %v выполняемых задач не обнаружено'", msgc.ClientTaskID),
							}),
						},
						msgc.ClientTaskID,
						msg.IDClientAPI)

					return
				}

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:     taskID,
					ClientName: msg.ClientName,
					Section:    "filtration control",
					Command:    "stop",
					SourceID:   sourceID,
				}

				return
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, msgc.ClientTaskID, msg.IDClientAPI)
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: "in the json message is not found the right option for 'MsgInstruction'",
				FuncName:    funcName,
			})

			return

		// УПРАВЛЕНИЕ ВЫГРУЗКОЙ ФАЙЛОВ
		case "download control":
			emt := ErrorMessageType{
				IDClientAPI: msg.IDClientAPI,
				Section:     "download control",
				Instruction: "task processing",
				MsgType:     "danger",
				ChanToAPI:   outCoreChans.OutCoreChanAPI,
			}

			if msgc.MsgInstruction == "to start downloading" {
				var dcts configure.DownloadControlTypeStart

				if err := json.Unmarshal(msgJSON, &dcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, "", msg.IDClientAPI)
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: "bad cast type JSON messages",
						FuncName:    funcName,
					})

					return
				}

				emt.TaskID = dcts.MsgOption.TaskIDApp
				emt.TaskIDClientAPI = dcts.ClientTaskID
				emt.Sources = []int{dcts.MsgOption.ID}

				//ищем источник по указанному идентификатору
				sourceInfo, ok := hsm.ISL.GetSourceSetting(dcts.MsgOption.ID)
				if !ok {
					emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID:   dcts.MsgOption.ID,
						TaskType:   "скачивание файлов",
						TaskAction: "задача отклонена",
						Message:    "запись для источника отсутствует в памяти приложения",
					})
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprintf("source ID %v was not found%v", dcts.MsgOption.ID, funcName),
						FuncName:    funcName,
					})

					//сообщение о том что задача была отклонена
					if err := ErrorMessage(emt); err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}

					return
				}

				//проверяем подключение источника
				if !sourceInfo.ConnectionStatus {
					emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID:   dcts.MsgOption.ID,
						TaskType:   "скачивание файлов",
						TaskAction: "задача отклонена",
						Message:    "источник не подключен",
					})
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprintf("source ID %v is not connected%v", dcts.MsgOption.ID, funcName),
						FuncName:    funcName,
					})

					//сообщение о том что задача была отклонена
					if err := ErrorMessage(emt); err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}

					return
				}

				//проверяем наличие в очереди задачи с указанным ID
				_, ti, err := hsm.QTS.SearchTaskForIDQueueTaskStorage(dcts.MsgOption.TaskIDApp)
				if err == nil {
					var errMsg string

					if ti.TaskStatus == "execution" {
						//проверяем наличие выполняемой задачи
						if smti, ok := hsm.SMT.GetStoringMemoryTask(dcts.MsgOption.TaskIDApp); ok {
							//проверяем завершена ли задача
							if smti.TaskStatus {
								errMsg = fmt.Sprintf("Task with ID '%v' for source ID %v rejected. You cannot add a task with the same ID many times in a short period of time.", dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID)
								emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
									SourceID:   dcts.MsgOption.ID,
									TaskType:   "скачивание файлов",
									TaskAction: "задача отклонена",
									Message:    "нельзя добавлять задачу с одним и тем же идентификатором множество раз в течении небольшого периода времени",
								})
							} else {
								errMsg = fmt.Sprintf("You cannot add a task with ID '%v' to a source with ID %v because it is already running", dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID)
								emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
									SourceID:   dcts.MsgOption.ID,
									TaskType:   "скачивание файлов",
									TaskAction: "задача отклонена",
									Message:    "невозможно добавить задачу источнику, задача по скачиванию файлов уже выполняется",
								})
							}
						}
					} else if ti.TaskStatus == "wait" {
						errMsg = fmt.Sprintf("Unable to add task with ID '%v' because it is already pending", dcts.MsgOption.TaskIDApp)
						emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
							SourceID:   dcts.MsgOption.ID,
							TaskType:   "скачивание файлов",
							TaskAction: "задача отклонена",
							Message:    "невозможно добавить задачу источнику, задача по скачиванию файлов уже выполняется",
						})
					} else {
						errMsg = fmt.Sprintf("Unable to add task with ID '%v'. The task has been completed, but has not yet been removed from the pending task list", dcts.MsgOption.ID)
						emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
							SourceID:   dcts.MsgOption.ID,
							TaskType:   "скачивание файлов",
							TaskAction: "задача отклонена",
							Message:    "невозможно добавить задачу так как задача была выполнена, однако из списка задач ожидающих выполнения пока не удалена",
						})
					}

					if len(errMsg) > 0 {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: errMsg,
							FuncName:    funcName,
						})

						//сообщение о том что задача была отклонена
						if err := ErrorMessage(emt); err != nil {
							saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
								Description: fmt.Sprint(err),
								FuncName:    funcName,
							})
						}

						return
					}
				}

				//добавляем задачу в очередь
				hsm.QTS.AddQueueTaskStorage(dcts.MsgOption.TaskIDApp, dcts.MsgOption.ID, configure.CommonTaskInfo{
					IDClientAPI:     msg.IDClientAPI,
					TaskIDClientAPI: dcts.ClientTaskID,
					TaskType:        "download control",
				}, &configure.DescriptionParametersReceivedFromUser{
					DownloadList: dcts.MsgOption.FileList,
				})

				//устанавливаем проверочный статус источника для данной задачи как подключен
				if err := hsm.QTS.ChangeAvailabilityConnectionOnConnection(dcts.MsgOption.ID, dcts.MsgOption.TaskIDApp); err != nil {
					emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID:   dcts.MsgOption.ID,
						TaskType:   "скачивание файлов",
						TaskAction: "задача отклонена",
						Message:    "запись для источника отсутствует в памяти приложения",
					})
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					//сообщение о том что задача была отклонена
					if err := ErrorMessage(emt); err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}

					return
				}

				//запрос в БД о наличии задачи с указанным ID и файлов для скачивания
				outCoreChans.OutCoreChanDB <- &configure.MsgBetweenCoreAndDB{
					MsgGenerator:    "Core module",
					MsgRecipient:    "DB module",
					MsgSection:      "download control",
					Instruction:     "finding information about a task",
					IDClientAPI:     msg.IDClientAPI,
					TaskID:          dcts.MsgOption.TaskIDApp,
					TaskIDClientAPI: dcts.ClientTaskID,
				}

				//информационное сообщение о том что задача добавлена в очередь
				notifications.SendNotificationToClientAPI(
					outCoreChans.OutCoreChanAPI,
					notifications.NotificationSettingsToClientAPI{
						MsgType: "success",
						MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
							SourceID:   dcts.MsgOption.ID,
							TaskType:   "скачивание файлов",
							TaskAction: "добавлена в очередь",
						}),
						Sources: []int{dcts.MsgOption.ID},
					},
					msgc.ClientTaskID,
					msg.IDClientAPI)
			}

			if msgc.MsgInstruction == "to cancel downloading" {
				var dcts configure.DownloadControlTypeStart

				if err := json.Unmarshal(msgJSON, &dcts); err != nil {
					notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, "", msg.IDClientAPI)
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: "bad cast type JSON messages",
						FuncName:    funcName,
					})

					return
				}

				emt.TaskID = dcts.MsgOption.TaskIDApp
				emt.TaskIDClientAPI = dcts.ClientTaskID
				emt.Sources = []int{dcts.MsgOption.ID}
				emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID:   dcts.MsgOption.ID,
					TaskType:   "скачивание файлов",
					TaskAction: "задача отклонена",
					Message:    "невозможен останов задачи, не найдена задача с заданным идентификатором",
				})

				//ищем задачу в очереди задач и в выполняемых задачах
				if _, err := hsm.QTS.GetQueueTaskStorage(dcts.MsgOption.ID, dcts.MsgOption.TaskIDApp); err != nil {
					if err := ErrorMessage(emt); err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}

					return
				}

				if _, ok := hsm.SMT.GetStoringMemoryTask(dcts.MsgOption.TaskIDApp); !ok {
					//если задача есть в очереди но еще не выполнялась ставим
					// ей статус 'complete'
					if err := hsm.QTS.ChangeTaskStatusQueueTask(dcts.MsgOption.ID, dcts.MsgOption.TaskIDApp, "complete"); err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}

					//сообщение об успешном снятии задачи из очереди ожидания
					resMsg := configure.DownloadControlTypeInfo{
						MsgOption: configure.DownloadControlMsgTypeInfo{
							ID:        dcts.MsgOption.ID,
							TaskIDApp: dcts.MsgOption.TaskIDApp,
							Status:    "stop",
						},
					}
					resMsg.MsgType = "information"
					resMsg.MsgSection = "download control"
					resMsg.MsgInstruction = "task processing"
					resMsg.ClientTaskID = dcts.ClientTaskID

					msgJSON, err := json.Marshal(resMsg)
					if err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}

					emt.ChanToAPI <- &configure.MsgBetweenCoreAndAPI{
						MsgGenerator: "Core module",
						MsgRecipient: "API module",
						IDClientAPI:  msg.IDClientAPI,
						MsgJSON:      msgJSON,
					}

					notifications.SendNotificationToClientAPI(
						outCoreChans.OutCoreChanAPI,
						notifications.NotificationSettingsToClientAPI{
							MsgType: "success",
							MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
								SourceID:   dcts.MsgOption.ID,
								TaskType:   "скачивание файлов",
								TaskAction: "удалена из очереди ожидающих задач",
							}),
							Sources: []int{dcts.MsgOption.ID},
						},
						dcts.ClientTaskID,
						msg.IDClientAPI)

					return
				}

				outCoreChans.OutCoreChanNI <- &configure.MsgBetweenCoreAndNI{
					TaskID:     dcts.MsgOption.TaskIDApp,
					ClientName: msg.ClientName,
					Section:    "download control",
					Command:    "stop",
					SourceID:   dcts.MsgOption.ID,
				}
			}

			return

		// УПРАВЛЕНИЕ ПОИСКОМ ИНФОРМАЦИИ В БД ПРИЛОЖЕНИЯ
		case "information search control":
			fmt.Println("func 'HandlerMsgFromAPI' MsgType: 'command', MsgSection: 'information search control'")

			return

		default:
			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, nsErrMsg, msgc.ClientTaskID, msg.IDClientAPI)
			saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
				Description: "in the json message is not found the right option for 'MsgInstruction'",
				FuncName:    funcName,
			})

			return
		}
	}
}
