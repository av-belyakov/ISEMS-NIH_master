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
	mtsfda int64,
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

			fmt.Println("function 'handlerMsgFromDB' SECTION - 'filtration source'")

			//получаем всю информацию по выполняемой задаче
			taskInfo, ok := hsm.SMT.GetStoringMemoryTask(res.TaskID)
			if !ok {
				_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with ID %v not found", res.TaskID))

				return
			}

			isNotComplete := taskInfo.TaskParameter.FiltrationTask.Status != "complete"
			moreThanMax := taskInfo.TaskParameter.FiltrationTask.SizeFilesFoundResultFiltering > mtsfda
			taskTypeNotFiltr := taskInfo.TaskType != "filtration control"

			fmt.Printf("function 'handlerMsgFromDB' INSTRACTION - %v\n", res.Instruction)
			fmt.Printf("function 'handlerMsgFromDB' STATUS = %v\n", taskInfo.TaskParameter.FiltrationTask.Status)

			//отправляем сообщение пользователю об завершении фильтрации
			if (taskInfo.TaskParameter.FiltrationTask.Status == "complete") && (res.Instruction == "filtration complete") {
				if err := sendMsgCompleteTaskFiltration(res.TaskID, taskInfo, outCoreChans.OutCoreChanAPI); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}
			}

			fmt.Printf("function 'handlerMsgFromDB' STATUS:%v, SIZE:%v, TASK TYPE:%v\n", taskInfo.TaskParameter.FiltrationTask.Status, taskInfo.TaskParameter.FiltrationTask.SizeFilesFoundResultFiltering, taskInfo.TaskType)
			fmt.Printf("function 'handlerMsgFromDB' resipient - API module, section - 'filtration control', isNotComplete - %v, SizeFilesFoundResultFiltering (%v) > mtsfda (%v), taskTypeNotFiltr - %v\n", isNotComplete, taskInfo.TaskParameter.FiltrationTask.SizeFilesFoundResultFiltering, mtsfda, taskTypeNotFiltr)

			if taskTypeNotFiltr || isNotComplete || moreThanMax {
				fmt.Println("function 'handlerMsgFromDB', отмечаем задачу как завершенную в списке очередей")

				//отмечаем задачу как завершенную в списке очередей
				if err := hsm.QTS.ChangeTaskStatusQueueTask(taskInfo.TaskParameter.FiltrationTask.ID, res.TaskID, "complete"); err != nil {
					fmt.Printf("function 'handlerMsgFromDB', ERROR = %v\n", err)

					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}

				return
			}

			fmt.Println("function 'handlerMsgFromDB', add task download after filtration to QueueTaskStorage")

			sourceID := taskInfo.TaskParameter.FiltrationTask.ID

			//получаем параметры фильтрации
			qti, err := hsm.QTS.GetQueueTaskStorage(sourceID, res.TaskID)
			if err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

				return
			}
			fmt.Printf("--- BEFORE ADD QTS: %v\n", qti)

			//добавляем задачу в очередь
			hsm.QTS.AddQueueTaskStorage(res.TaskID, sourceID, configure.CommonTaskInfo{
				IDClientAPI:     res.IDClientAPI,
				TaskIDClientAPI: res.TaskIDClientAPI,
				TaskType:        "download control",
			}, &configure.DescriptionParametersReceivedFromUser{
				FilterationParameters:         qti.TaskParameters.FilterationParameters,
				PathDirectoryForFilteredFiles: taskInfo.TaskParameter.FiltrationTask.PathStorageSource,
				DownloadList:                  []string{},
			})

			one, _ := hsm.QTS.GetQueueTaskStorage(sourceID, res.TaskID)
			fmt.Printf("--- AFTER ADD QTS: %v\n", one)

			//устанавливаем проверочный статус источника для данной задачи как подключен
			if err := hsm.QTS.ChangeAvailabilityConnectionOnConnection(sourceID, res.TaskID); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

			//изменяем статус наличия файлов для скачивания
			if err := hsm.QTS.ChangeAvailabilityFilesDownload(sourceID, res.TaskID); err != nil {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
			}

			two, _ := hsm.QTS.GetQueueTaskStorage(sourceID, res.TaskID)
			fmt.Printf("--- AFTER CHANGE QTS: %v\n", two)

			/*
				   До сего момента вроде все правельно,
				   а вот дальше не понятно и ничего не работает
				   наверное стоит убрать вызов CompleteStoringMemoryTask

				   !!! еще проверить по фильтрации !!!
					1. останов фильтрации (YES)
					2. выполнение нескольких процессов фильтрации
					в том числе обработку при превышении кол-во одновременно
					запущеных процессов фильтрации (несколько процессов выполняются
					успешно, однако выполнение всех процессов прерывается если
				останавливаешь один из них)
					3. обработку сообщений при отключении и подключении клиента
					API при выполнении фильтрации
					4. отключение и подключения самого мастера при выполнении
					фильтрации
			*/

			//устанавливаем статус задачи в "complete" для ее последующего удаления
			//hsm.SMT.CompleteStoringMemoryTask(res.TaskID)

			fmt.Println("function 'handlerMsgFromDB', complete storing memory task --- ")

			/*
				!!!!
				   Не доделал очереди для фильтрации.
				   // - Добавление задачи в очередь и удаление ее из очереди при завершении
				   // фильтрации вроде бы сделал, на до бы еще проверить логику.

				   // !!! ДУМАЮ СТОИТ ПРОДУМАТЬ удаления задачи из очереди
				   // через событие 'monitoring task performance'

				   // - Нужно сделать удаление при отмене фильтрации.

				   // - Продумать действия при подвисании задачи!!!
				    - На основании полученной логики с очередями по фильтрации продумать
				    автоматическую загрузку файлов при завершении фильтрации. Здесь
				    основная сложность это совпадение ID задачи, так как задача по фильтрации
				    еще не удалилась а уже нужно добавлять задачу по скачиванию с таким же ID.
				    По этому не проходит проверка на наличие дубликатов задач.
				!!!!
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

			fmt.Println("function 'HandlerMsgFromDB', section 'filtartion control', recipient - 'NI module' (INDEX NOT FOUNT)")

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

			fmt.Println("function 'HandlerMsgFromDB', section 'filtartion control', recipient - 'NI module' (INDEX FOUND)")

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
