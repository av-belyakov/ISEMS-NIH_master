package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//getCurrentSourceListForAPI подготавливает список актуальных источников для передаче клиенту API
func getCurrentSourceListForAPI(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	res *configure.MsgBetweenCoreAndDB,
	clientID, clientTaskID string) error {

	funcName := ", function 'getCurrentSourceListForAPI'"

	listSource, ok := res.AdvancedOptions.([]configure.InformationAboutSource)
	if !ok {
		return fmt.Errorf("type conversion error section type 'error notification'%v", funcName)
	}

	list := make([]configure.ShortListSources, 0, len(listSource))

	//формируем ответ клиенту API
	for _, s := range listSource {
		list = append(list, configure.ShortListSources{
			ID:          s.ID,
			IP:          s.IP,
			ShortName:   s.ShortName,
			Description: s.Description,
		})
	}

	msg := configure.SourceControlCurrentListSources{
		MsgOptions: configure.SourceControlCurrentListSourcesList{
			TaskInfo: configure.MsgTaskInfo{
				State: "end",
			},
			SourceList: list,
		},
	}
	msg.MsgType = "information"
	msg.MsgSection = "source control"
	msg.MsgInstruction = "send current source list"
	msg.ClientTaskID = clientTaskID

	msgjson, _ := json.Marshal(&msg)

	//отправляем данные клиенту
	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  clientID,
		MsgJSON:      msgjson,
	}

	return nil
}

//sendMsgCompleteTaskFiltration отправляет информационное сообщение об
// окончании фильтрации но только если восстанавливалась информация о задаче
// в StoringMemoryTask после разрыва соединения и задача находится в
// статусе 'complete'
func sendMsgCompleteTaskFiltration(
	taskID string,
	taskInfo *configure.TaskDescription,
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) error {

	ti := taskInfo.TaskParameter.FiltrationTask
	resMsg := configure.FiltrationControlTypeInfo{
		MsgOption: configure.FiltrationControlMsgTypeInfo{
			ID:                              taskInfo.TaskParameter.FiltrationTask.ID,
			TaskIDApp:                       taskID,
			Status:                          ti.Status,
			NumberFilesMeetFilterParameters: ti.NumberFilesMeetFilterParameters,
			NumberProcessedFiles:            ti.NumberProcessedFiles,
			NumberFilesFoundResultFiltering: ti.NumberFilesFoundResultFiltering,
			NumberErrorProcessedFiles:       ti.NumberErrorProcessedFiles,
			NumberDirectoryFiltartion:       ti.NumberDirectoryFiltartion,
			SizeFilesMeetFilterParameters:   ti.SizeFilesMeetFilterParameters,
			SizeFilesFoundResultFiltering:   ti.SizeFilesFoundResultFiltering,
			PathStorageSource:               ti.PathStorageSource,
		},
	}

	resMsg.MsgType = "information"
	resMsg.MsgSection = "filtration control"
	resMsg.MsgInstruction = "task processing"
	resMsg.ClientTaskID = taskInfo.ClientTaskID

	msgJSON, err := json.Marshal(resMsg)
	if err != nil {
		return err
	}

	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  taskInfo.ClientID,
		MsgJSON:      msgJSON,
	}

	return nil
}

//sendMsgCompliteTaskSearchShortInformationAboutTask отправляет краткую информацию полученную
// при поиске в БД ранее выполняемых задач по фильтрации и скачиванию
func sendMsgCompliteTaskSearchShortInformationAboutTask(
	res *configure.MsgBetweenCoreAndDB,
	tssq *configure.TemporaryStorageSearchQueries,
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) error {

	const chunkSize = 100

	//получаем информацию о задаче
	info, err := tssq.GetInformationAboutSearchTask(res.TaskID)
	if err != nil {
		return err
	}

	numTaskFound := len(info.ListFoundInformation.List)

	//	fmt.Printf("func 'sendMsgCompliteTaskSearchShortInformationAboutTask', Info: '%v'\n", info)

	resMsg := configure.SearchInformationResponseCommanInfo{
		MsgOption: configure.SearchInformationResponseOptionCommanInfo{
			TaskIDApp:             res.TaskID,
			Status:                "complete",
			TotalNumberTasksFound: numTaskFound,
			PaginationOptions: configure.PaginationOption{
				ChunkSize:          chunkSize,
				ChunkNumber:        1,
				ChunkCurrentNumber: 1,
			},
		},
	}

	resMsg.MsgType = "information"
	resMsg.MsgSection = "information search control"
	resMsg.MsgInstruction = "processing information search task"
	resMsg.ClientTaskID = res.TaskIDClientAPI

	//отправляем всю информацию целиком
	if numTaskFound < chunkSize {

		fmt.Println("func 'sendMsgCompliteTaskSearchShortInformationAboutTask', SEND FULL")

		resMsg.MsgOption.ShortListFoundTasks = info.ListFoundInformation.List

		msgJSON, err := json.Marshal(resMsg)
		if err != nil {
			return err
		}

		chanToAPI <- &configure.MsgBetweenCoreAndAPI{
			MsgGenerator: "Core module",
			MsgRecipient: "API module",
			IDClientAPI:  res.IDClientAPI,
			MsgJSON:      msgJSON,
		}

		return nil
	}

	numChunk := getNumParts(numTaskFound, chunkSize)
	resMsg.MsgOption.PaginationOptions.ChunkNumber = numChunk

	//сегментируем найденый список
	for i := 0; i < numChunk; i++ {
		//		fmt.Printf("Send chunks num: '%v' from '%v' (full task found: '%v')\n", i, numChunk, numTaskFound)

		num := i * chunkSize
		if i == numChunk-1 {
			resMsg.MsgOption.ShortListFoundTasks = info.ListFoundInformation.List[num:]
		} else {
			resMsg.MsgOption.ShortListFoundTasks = info.ListFoundInformation.List[num:(num + chunkSize)]
		}

		resMsg.MsgOption.PaginationOptions.ChunkCurrentNumber = i + 1
		msgJSON, err := json.Marshal(resMsg)
		if err != nil {
			return err
		}

		chanToAPI <- &configure.MsgBetweenCoreAndAPI{
			MsgGenerator: "Core module",
			MsgRecipient: "API module",
			IDClientAPI:  res.IDClientAPI,
			MsgJSON:      msgJSON,
		}
	}

	return nil
}

func getNumParts(fullSize, sizeChunk int) int {
	return int(math.Round(float64(fullSize) / float64(sizeChunk)))
}

//checkParametersDownloadTask проверяет ряд параметров в информации о задаче полученной из БД
func checkParametersDownloadTask(
	res *configure.MsgBetweenCoreAndDB,
	hsm HandlersStoringMemory,
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) error {

	funcName := ", function 'checkParametersDownloadTask'"

	taskInfoFromDB, ok := res.AdvancedOptions.(*[]configure.InformationAboutTask)
	if !ok {
		return fmt.Errorf("type conversion error section type 'error notification'%v", funcName)
	}

	tidb := (*taskInfoFromDB)[0]

	emt := ErrorMessageType{
		TaskID:          res.TaskID,
		TaskIDClientAPI: res.TaskIDClientAPI,
		IDClientAPI:     res.IDClientAPI,
		Section:         res.MsgSection,
		Instruction:     res.Instruction,
		MsgType:         "danger",
		Sources:         []int{tidb.SourceID},
		ChanToAPI:       chanToAPI,
	}

	//ищем задачу с taskID полученному из БД
	sourceID, tisqt, err := hsm.QTS.SearchTaskForIDQueueTaskStorage(res.TaskID)
	if err != nil {
		emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID: sourceID,
			TaskType: "скачивание файлов",
			Message:  "задача была аварийно завершена, не найдено задачи по указанному пользователем ID",
		})

		if err := ErrorMessage(emt); err != nil {
			return err
		}

		return fmt.Errorf("not found the tasks specified by the user ID %v%v", res.TaskID, funcName)
	}

	emt.SourceID = sourceID

	//наличие в БД задачи по заданному пользователем ID
	if len(*taskInfoFromDB) == 0 {
		emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID:   sourceID,
			TaskType:   "скачивание файлов",
			TaskAction: "задача отклонена",
			Message:    "получен не верный ID задачи",
		})

		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//изменяем статус задачи в storingMemoryQueueTask
		// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
		// функцией 'CheckTimeQueueTaskStorage')
		if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
			return err
		}

		return fmt.Errorf("not found the tasks specified by the user ID %v%v", res.TaskID, funcName)
	}

	//совпадает ли ID источника задачи из БД с ID источника полученного от пользователя
	if tidb.SourceID != sourceID {
		emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID:   sourceID,
			TaskType:   "скачивание файлов",
			TaskAction: "задача отклонена",
			Message:    "идентификатор источника указанный пользователем не совпадает с идентификатором полученным из базы данных",
		})

		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//изменяем статус задачи в storingMemoryQueueTask
		// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
		// функцией 'CheckTimeQueueTaskStorage')
		if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
			return err
		}

		return fmt.Errorf("the source ID %v specified by the user does not match the ID %v obtained from the database%v", sourceID, tidb.SourceID, funcName)
	}

	//выполнена ли задача по фильтрации (статус задачи "complete")
	if tidb.DetailedInformationOnFiltering.TaskStatus != "complete" {
		emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID: sourceID,
			TaskType: "скачивание файлов",
			Message:  fmt.Sprintf("задача с ID %v не имеет статус 'завершена', дальнейшее выполнение задачи не возможно", res.TaskID),
		})

		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//изменяем статус задачи в storingMemoryQueueTask
		// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
		// функцией 'CheckTimeQueueTaskStorage')
		if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
			return err
		}

		return fmt.Errorf("the task with ID %v does not have the status 'completed'%v", res.TaskID, funcName)
	}

	//найденны ли какие либо файлы в результате фильтрации
	if tidb.DetailedInformationOnFiltering.NumberFilesFoundResultFiltering == 0 {
		emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
			SourceID:   sourceID,
			TaskType:   "скачивание файлов",
			TaskAction: "задача отклонена",
			Message:    "в результате выполненной ранее фильтрации не было найдено ни одного файла, дальнейшее выполнение задачи не возможно",
		})

		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//изменяем статус задачи в storingMemoryQueueTask
		// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
		// функцией 'CheckTimeQueueTaskStorage')
		if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
			return err
		}

		return fmt.Errorf("as a result of the previous filtering, no files were found (task ID %v)%v", res.TaskID, funcName)
	}

	var confirmedListFiles map[string]*configure.DownloadFilesInformation

	numUserDownloadList := len(tisqt.TaskParameters.DownloadList)

	//совпадают ли файлы переданные пользователем (если передовались) с файлами полученными из БД
	if numUserDownloadList == 0 {
		confirmedListFiles = make(map[string]*configure.DownloadFilesInformation, len(tidb.ListFilesResultTaskExecution))

		//формируем новый список не выгружавшихся файлов
		for _, f := range tidb.ListFilesResultTaskExecution {
			//только если файл не загружался
			if !f.FileLoaded {
				confirmedListFiles[f.FileName] = &configure.DownloadFilesInformation{}
				confirmedListFiles[f.FileName].Size = f.FileSize
				confirmedListFiles[f.FileName].Hex = f.FileHex
			}
		}

	} else {
		clf, err := checkFileNameMatches(tidb.ListFilesResultTaskExecution, tisqt.TaskParameters.DownloadList)
		if err != nil {
			emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID: sourceID,
				TaskType: "скачивание файлов",
				Message:  "внутренняя ошибка, дальнейшее выполнение задачи не возможно",
			})

			if err := ErrorMessage(emt); err != nil {
				return err
			}

			//изменяем статус задачи в storingMemoryQueueTask
			// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
			// функцией 'CheckTimeQueueTaskStorage')
			if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
				return err
			}

			return err
		}

		confirmedListFiles = clf
		if len(confirmedListFiles) == 0 {
			emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID:   sourceID,
				TaskType:   "скачивание файлов",
				TaskAction: "задача отклонена",
				Message:    "не найдено ни одного совпадения для скачиваемых с источника файлов, возможно все файлы были скачены ранее или отсутствуют на источнике",
			})

			if err := ErrorMessage(emt); err != nil {
				return err
			}

			//изменяем статус задачи в storingMemoryQueueTask
			// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
			// функцией 'CheckTimeQueueTaskStorage')
			if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
				return err
			}

			return errors.New("no matches found in the database for files received from the user")
		}

		numFilesInvalid := numUserDownloadList - len(confirmedListFiles)

		if numFilesInvalid > 0 {
			//отправляем информационное сообщение
			notifications.SendNotificationToClientAPI(
				chanToAPI,
				notifications.NotificationSettingsToClientAPI{
					MsgType: "warning",
					MsgDescription: common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID:   sourceID,
						TaskType:   "скачивание файлов",
						TaskAction: "задача принята",
						Message:    fmt.Sprintf("не все файлы выбранные для скачивания прошли верификацию. %v из %v файлов передаваться не будут, так как отсутствуют на источнике", numFilesInvalid, numUserDownloadList),
					}),
				},
				res.TaskIDClientAPI,
				res.IDClientAPI)
		}
	}

	emt.MsgHuman = common.PatternUserMessage(&common.TypePatternUserMessage{
		SourceID: sourceID,
		TaskType: "скачивание файлов",
		Message:  "внутренняя ошибка, дальнейшее выполнение задачи не возможно",
	})

	//добавляем список подтвержденных и ранее не загружавшихся файлов
	if err := hsm.QTS.AddConfirmedListFiles(sourceID, res.TaskID, confirmedListFiles); err != nil {
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//изменяем статус задачи в storingMemoryQueueTask
		// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
		// функцией 'CheckTimeQueueTaskStorage')
		if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
			return err
		}

		return err
	}

	//добавляем информацию по фильтрации в QueueTaskStorage
	if err := hsm.QTS.AddFiltrationParametersQueueTaskStorage(sourceID, res.TaskID, &tidb.FilteringOption); err != nil {
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//изменяем статус задачи в storingMemoryQueueTask
		// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
		// функцией 'CheckTimeQueueTaskStorage')
		if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
			return err
		}

		return err
	}

	//добавляем информацию о директории на источнике, в которой хранятся отфильтрованные файлы
	if err := hsm.QTS.AddPathDirectoryFilteredFiles(sourceID, res.TaskID, tidb.DetailedInformationOnFiltering.PathDirectoryForFilteredFiles); err != nil {
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//изменяем статус задачи в storingMemoryQueueTask
		// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
		// функцией 'CheckTimeQueueTaskStorage')
		if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
			return err
		}

		return err
	}

	//изменяем статус наличия файлов для скачивания
	if err := hsm.QTS.ChangeAvailabilityFilesDownload(sourceID, res.TaskID); err != nil {
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//изменяем статус задачи в storingMemoryQueueTask
		// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
		// функцией 'CheckTimeQueueTaskStorage')
		if err := hsm.QTS.ChangeTaskStatusQueueTask(sourceID, res.TaskID, "complete"); err != nil {
			return err
		}

		return err
	}

	return nil
}

//checkFileNameMatches проверяет на совпадение файлов переданных пользователем с файлами полученными из БД
func checkFileNameMatches(lfdb []*configure.FilesInformation, lfqst []string) (map[string]*configure.DownloadFilesInformation, error) {
	type fileInfo struct {
		hex      string
		size     int64
		isLoaded bool
	}

	nlf := make(map[string]*configure.DownloadFilesInformation, len(lfqst))

	if len(lfdb) == 0 {
		return nlf, errors.New("an empty list with files was obtained from the database")
	}

	if len(lfqst) == 0 {
		return nlf, errors.New("an empty list with files was received from the API client")
	}

	tmpList := make(map[string]fileInfo, len(lfdb))

	for _, i := range lfdb {
		tmpList[i.FileName] = fileInfo{i.FileHex, i.FileSize, i.FileLoaded}
	}

	for _, f := range lfqst {
		if info, ok := tmpList[f]; ok {
			//только если файл не загружался
			if !info.isLoaded {
				nlf[f] = &configure.DownloadFilesInformation{}
				nlf[f].Size = info.size
				nlf[f].Hex = info.hex
			}
		}
	}

	return nlf, nil
}
