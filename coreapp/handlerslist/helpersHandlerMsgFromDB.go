package handlerslist

import (
	"encoding/json"
	"errors"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//getCurrentSourceListForAPI подготавливает список актуальных источников для передаче клиенту API
func getCurrentSourceListForAPI(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	res *configure.MsgBetweenCoreAndDB,
	smt *configure.StoringMemoryTask) error {

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

	st, ok := smt.GetStoringMemoryTask(res.TaskID)
	if !ok {
		return fmt.Errorf("task with %v not found", res.TaskID)
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
	msg.MsgInsturction = "send current source list"
	msg.ClientTaskID = st.ClientTaskID

	msgjson, _ := json.Marshal(&msg)

	if err := senderMsgToAPI(chanToAPI, smt, res.TaskID, st.ClientID, msgjson); err != nil {
		return err
	}

	return nil
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
		ChanToAPI:       chanToAPI,
	}

	//ищем задачу с taskID полученному из БД
	sourceID, tisqt, err := hsm.QTS.SearchTaskForIDQueueTaskStorage(res.TaskID)
	if err != nil {
		emt.MsgHuman = "не найдено задачи по указанному пользователем ID"
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		return fmt.Errorf("not found the tasks specified by the user ID %v%v", res.TaskID, funcName)
	}

	emt.SourceID = sourceID

	//наличие в БД задачи по заданному пользователем ID
	if len(*taskInfoFromDB) == 0 {
		emt.MsgHuman = "Не найдено задачи по указанному пользователем ID, дальнейшее выполнение задачи по выгрузке файлов не возможна"
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//удаляем задачу из очереди
		if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
			return err
		}

		return fmt.Errorf("not found the tasks specified by the user ID %v%v", res.TaskID, funcName)
	}

	//совпадает ли ID источника из задачи с ID источника полученного от пользователя
	if tidb.SourceID != sourceID {
		emt.MsgHuman = "Идентификатор источника указанный пользователем не совпадает с идентификатором полученным из базы данных, дальнейшее выполнение задачи по выгрузке файлов не возможна"
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//удаляем задачу из очереди
		if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
			return err
		}

		return fmt.Errorf("the source ID %v specified by the user does not match the ID %v obtained from the database%v", sourceID, tidb.SourceID, funcName)
	}

	//выполнена ли задача по фильтрации (статус задачи "complite")
	if tidb.DetailedInformationOnFiltering.TaskStatus != "complite" {
		emt.MsgHuman = fmt.Sprintf("Задача с ID %v не имеет статус 'завершена', дальнейшее выполнение задачи по выгрузке файлов не возможна", res.TaskID)
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//удаляем задачу из очереди
		if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
			return err
		}

		return fmt.Errorf("the task with ID %v does not have the status 'completed'%v", res.TaskID, funcName)
	}

	//найденны ли какие либо файлы в результате фильтрации
	if tidb.DetailedInformationOnFiltering.NumberFilesFoundResultFiltering == 0 {
		emt.MsgHuman = "В результате выполненной ранее фильтрации не было найдено ни одного файла, дальнейшее выполнение задачи по выгрузке файлов не возможна"
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//удаляем задачу из очереди
		if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
			return err
		}

		return fmt.Errorf("as a result of the previous filtering, no files were found (task ID %v)%v", res.TaskID, funcName)
	}

	numUserDownloadList := len(tisqt.TaskParameters.DownloadList)

	//совпадают ли файлы переданные пользователем (если передовались) с файлами полученными из БД
	if numUserDownloadList > 0 {
		confirmedListFiles, err := checkFileNameMatches(tidb.ListFilesResultTaskExecution, tisqt.TaskParameters.DownloadList)

		if err != nil {
			emt.MsgHuman = "Внутренняя ошибка, дальнейшее выполнение задачи по выгрузке файлов не возможна"
			if err := ErrorMessage(emt); err != nil {
				return err
			}

			//удаляем задачу из очереди
			if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
				return err
			}

			return err
		}

		if len(confirmedListFiles) == 0 {
			emt.MsgHuman = "Не найдено ни одного совпадения в базе данных для файлов полученных от пользователя"
			if err := ErrorMessage(emt); err != nil {
				return err
			}

			//удаляем задачу из очереди
			if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
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
					MsgType:        "warning",
					MsgDescription: fmt.Sprintf("Не все файлы выбранные для скачивания прошли верификацию. %v и %v файлов передаваться не будут так как отсутствуют на сервере или были переданы ранее.", numFilesInvalid, numUserDownloadList),
				},
				res.TaskIDClientAPI,
				res.IDClientAPI)
		}

		//добавляем информацию о файлах в StoringMemoryQueueTask
		if err := hsm.QTS.AddConfirmedListFiles(sourceID, res.TaskID, confirmedListFiles); err != nil {
			emt.MsgHuman = "Внутренняя ошибка, дальнейшее выполнение задачи по выгрузке файлов не возможна"
			if err := ErrorMessage(emt); err != nil {
				return err
			}

			//удаляем задачу из очереди
			if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
				return err
			}

			return err
		}

	} else {
		nlf := make(map[string]*configure.DownloadFilesInformation, len(tidb.ListFilesResultTaskExecution))

		//формируем новый список не выгружавшихся файлов
		for _, f := range tidb.ListFilesResultTaskExecution {
			//только если файл не загружался
			if !f.FileLoaded {
				nlf[f.FileName] = &configure.DownloadFilesInformation{}
				nlf[f.FileName].Size = f.FileSize
				nlf[f.FileName].Hex = f.FileHex
			}
		}

		//добавляем список подтвержденных и ранее не загружавшихся файлов
		if err := hsm.QTS.AddConfirmedListFiles(sourceID, res.TaskID, nlf); err != nil {
			emt.MsgHuman = "Внутренняя ошибка, дальнейшее выполнение задачи по выгрузке файлов не возможна"
			if err := ErrorMessage(emt); err != nil {
				return err
			}

			//удаляем задачу из очереди
			if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
				return err
			}

			return err
		}
	}

	emt.MsgHuman = "Внутренняя ошибка, дальнейшее выполнение задачи по выгрузке файлов не возможна"

	//добавляем информацию по фильтрации в QueueTaskStorage
	if err := hsm.QTS.AddFiltrationParametersQueueTaskstorage(sourceID, res.TaskID, &tidb.FilteringOption); err != nil {
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//удаляем задачу из очереди
		if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
			return err
		}

		return err
	}

	//добавляем информацию о директории на источнике в которой хранятся отфильтрованные файлы
	if err := hsm.QTS.AddPathDirectoryFilteredFiles(sourceID, res.TaskID, tidb.DetailedInformationOnFiltering.PathDirectoryForFilteredFiles); err != nil {
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//удаляем задачу из очереди
		if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
			return err
		}

		return err
	}

	if err := hsm.QTS.ChangeAvailabilityFilesDownload(sourceID, res.TaskID); err != nil {
		if err := ErrorMessage(emt); err != nil {
			return err
		}

		//удаляем задачу из очереди
		if err := hsm.QTS.DelQueueTaskStorage(sourceID, res.TaskID); err != nil {
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
