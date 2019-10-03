package handlerslist

import (
	"ISEMS-NIH_master/common"
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

//HandlerMsgFromDB обработчик сообщений приходящих от модуля взаимодействия с базой данных
func HandlerMsgFromDB(
	outCoreChans HandlerOutChans,
	res *configure.MsgBetweenCoreAndDB,
	hsm HandlersStoringMemory,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanDropNI <-chan string) {

	funcName := ", function 'HandlerMsgFromDB'"

	taskInfo, taskIDIsExist := hsm.SMT.GetStoringMemoryTask(res.TaskID)

	switch res.MsgRecipient {
	case "Core module":
		switch res.MsgSection {
		case "error notification":
			//если сообщение об ошибке только для ядра приложения
			if en, ok := res.AdvancedOptions.(configure.ErrorNotification); ok {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(en.ErrorBody))

				return
			}

		case "all information about task":
			/* ФАКТИЧЕСКИ ЭТО ЗАПУСК СКАЧИВАНИЯ ФАЙЛОВ */

			//проверяем ряд параметров в задаче для изменения проверочного статуса задачи в QueueStoringMemoryTask
			if err := checkParametersDownloadTask(res, hsm, outCoreChans.OutCoreChanAPI); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}
		}

	case "API module":
		if !taskIDIsExist {
			_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with %v not found%v", res.TaskID, funcName))

			return
		}

		switch res.MsgSection {
		case "source list":
			if err := getCurrentSourceListForAPI(outCoreChans.OutCoreChanAPI, res, hsm.SMT); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

		case "source control":
			//пока заглушка

		case "source telemetry":
			//пока заглушка

		case "filtration control":
			//получаем всю информацию по выполняемой задаче
			taskInfo, ok := hsm.SMT.GetStoringMemoryTask(res.TaskID)
			if !ok {
				_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with ID %v not found", res.TaskID))

				return
			}

			//отмечаем задачу как завершенную в списке очередей
			if err := hsm.QTS.ChangeTaskStatusQueueTask(taskInfo.TaskParameter.FiltrationTask.ID, res.TaskID, "complete"); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

			//устанавливаем статус задачи в "complete" для ее последующего удаления
			hsm.SMT.CompleteStoringMemoryTask(res.TaskID)

			/*
			   Не доделал очереди для фильтрации.
			    - Добавление задачи в очередь и удаление ее из очереди при завершении
			   фильтрации вроде бы сделал, на до бы еще проверить логику.
			    - Нужно сделать удаление при отмене фильтрации.

			    - Продумать действия при подвисании задачи!!!
			    - На основании полученной логики с очередями по фильтрации продумать
			    автоматическую загрузку файлов при завершении фильтрации. Здесь
			    основная сложность это совпадение ID задачи, так как задача по фильтрации
			    еще не удалилась а уже нужно добавлять задачу по скачиванию с таким же ID.
			    По этому не проходит проверка на наличие дубликатов задач.
			*/

		case "download control":
			//пока заглушка

		case "information search results":
			//пока заглушка

		case "error notification":
			en, ok := res.AdvancedOptions.(configure.ErrorNotification)
			if !ok {
				_ = saveMessageApp.LogMessage("error", "type conversion error section type 'error notification'"+funcName)

				return
			}

			_ = saveMessageApp.LogMessage("error", fmt.Sprint(en.ErrorBody))

			ns := notifications.NotificationSettingsToClientAPI{
				MsgType:        "danger",
				MsgDescription: "Ошибка базы данных при обработке запроса",
			}

			notifications.SendNotificationToClientAPI(outCoreChans.OutCoreChanAPI, ns, taskInfo.ClientTaskID, res.IDClientAPI)

		case "message notification":
			mn, ok := res.AdvancedOptions.(configure.MessageNotification)
			if !ok {
				_ = saveMessageApp.LogMessage("error", "type conversion error section type 'message notification'"+funcName)

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
				_ = saveMessageApp.LogMessage("error", "type conversion error section type 'message notification'"+funcName)

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
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

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
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}

			msg.AdvancedOptions = msgJSON

			outCoreChans.OutCoreChanNI <- &msg

			//информация о задаче по заданному ID
			t, ok := hsm.SMT.GetStoringMemoryTask(res.TaskID)
			if !ok {
				_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with %v not found%v", res.TaskID, funcName))

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
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

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
		_ = saveMessageApp.LogMessage("error", "the module receiver is not defined, request processing is impossible"+funcName)
	}
}
