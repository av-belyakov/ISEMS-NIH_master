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
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	res *configure.MsgBetweenCoreAndDB,
	smt *configure.StoringMemoryTask,
	chanDropNI <-chan string,
	chanToNI chan<- *configure.MsgBetweenCoreAndNI) {

	fmt.Println("START function 'HandlerMsgFromDB' module coreapp")
	//	fmt.Printf("%v", res)

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'HandlerMsgFromDB'"

	taskInfo, taskIDIsExist := smt.GetStoringMemoryTask(res.TaskID)

	if res.MsgRecipient == "API module" {
		if !taskIDIsExist {
			_ = saveMessageApp.LogMessage("error", fmt.Sprintf("task with %v not found%v", res.TaskID, funcName))
		}

		switch res.MsgSection {
		case "source list":
			getCurrentSourceListForAPI(chanToAPI, res, smt)

		case "source control":
			//пока заглушка

		case "source telemetry":
			//пока заглушка

		case "filtration control":
			//если фильтрация завершилась, то есть статус задачи "stop" или "complite"
			if res.Instruction == "filtration complite" {
				ns := notifications.NotificationSettingsToClientAPI{
					MsgType:        "success",
					MsgDescription: "Задача по фильтрации сетевого трафика завершена",
					Sources:        []int{taskInfo.TaskParameter.FiltrationTask.ID},
				}

				//отправляем информационное сообщение пользователю
				notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, res.IDClientAPI)

				//устанавливаем статус задачи в "complite" для ее последующего удаления
				smt.CompleteStoringMemoryTask(res.TaskID)
			}

			//если фильтрация была отклонена
			if res.Instruction == "filtration rejected" {
				//устанавливаем статус задачи в "complite" для ее последующего удаления
				smt.CompleteStoringMemoryTask(res.TaskID)
			}

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

			notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, res.IDClientAPI)

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

			notifications.SendNotificationToClientAPI(chanToAPI, ns, taskInfo.ClientTaskID, res.IDClientAPI)
		}
	} else if res.MsgRecipient == "NI module" {
		switch res.MsgSection {
		case "source list":
			chanToNI <- &configure.MsgBetweenCoreAndNI{
				Section:         "source control",
				Command:         "create list",
				AdvancedOptions: res.AdvancedOptions,
			}

		case "source control":
			//пока заглушка

		case "filtration control":

			fmt.Println(" ***** CORE MODULE (handlerMsgFromDB), Resived MSG 'filtration' ****")
			fmt.Printf("%v\n", res)

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
				chanToNI <- &msg

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
			chanToNI <- &msg

			//информация о задаче по заданному ID
			t, ok := smt.GetStoringMemoryTask(res.TaskID)
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
					chanToNI <- &msg

				}
			}

		case "download control":
			//пока заглушка

		}
	} else if res.MsgRecipient == "Core module" {
		fmt.Printf("RESIPENT MSG FOR CORE %v", res)

		if res.MsgSection == "error notification" {
			//если сообщение об ошибке только для ядра приложения
			if en, ok := res.AdvancedOptions.(configure.ErrorNotification); ok {
				_ = saveMessageApp.LogMessage("error", fmt.Sprint(en.ErrorBody))

				return
			}
		}
	} else {
		_ = saveMessageApp.LogMessage("error", "the module receiver is not defined, request processing is impossible"+funcName)
	}
}
