package handlers

import (
	"fmt"
	"os"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
)

type parametersWritingBinaryFile struct {
	SourceID            int
	TaskID              string
	Data                *[]byte
	ListFileDescriptors map[string]*os.File
	SMT                 *configure.StoringMemoryTask
	ChanInCore          chan<- *configure.MsgBetweenCoreAndNI
}

//writingBinaryFile осуществляет запись бинарного файла
func writingBinaryFile(pwbf parametersWritingBinaryFile) (bool, error) {
	//получаем хеш принимаемого файла
	fileHex := string((*pwbf.Data)[35:67])
	w, ok := pwbf.ListFileDescriptors[fileHex]
	if ok {
		return false, fmt.Errorf("no file descriptor found for the specified hash %v (task ID %v, source ID %v)", fileHex, pwbf.TaskID, pwbf.SourceID)
	}

	//запись принятых байт
	numWriteByte, err := w.Write((*pwbf.Data)[67:])
	if err != nil {
		return false, err
	}

	ti, ok := pwbf.SMT.GetStoringMemoryTask(pwbf.TaskID)
	if !ok {
		return false, fmt.Errorf("task with ID %v not found", pwbf.TaskID)
	}

	fi := ti.TaskParameter.DownloadTask.FileInformation

	writeByte := fi.AcceptedSizeByte + int64(numWriteByte)
	writePercent := int(writeByte / (fi.FullSizeByte / 100))
	numAcceptedChunk := fi.NumAcceptedChunk + 1

	//обновляем информацию о принимаемом файле
	pwbf.SMT.UpdateTaskDownloadAllParameters(pwbf.TaskID, configure.DownloadTaskParameters{
		Status:                              "execute",
		NumberFilesTotal:                    ti.TaskParameter.DownloadTask.NumberFilesTotal,
		NumberFilesDownloaded:               ti.TaskParameter.DownloadTask.NumberFilesDownloaded,
		PathDirectoryStorageDownloadedFiles: ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles,
		FileInformation: configure.DetailedFileInformation{
			Name:                fi.Name,
			Hex:                 fi.Hex,
			FullSizeByte:        fi.FullSizeByte,
			AcceptedSizeByte:    writeByte,
			AcceptedSizePercent: writePercent,
			NumChunk:            fi.NumChunk,
			NumAcceptedChunk:    numAcceptedChunk,
		},
	})

	msgToCore := configure.MsgBetweenCoreAndNI{
		TaskID:   pwbf.TaskID,
		Section:  "download control",
		Command:  "file download process",
		SourceID: pwbf.SourceID,
	}

	//отправляем сообщение Ядру приложения только если
	// процент увеличился на 1
	if writePercent > fi.AcceptedSizePercent {
		pwbf.ChanInCore <- &msgToCore
	}

	//если все кусочки были переданы (то есть файл считается полностью скаченным)
	if fi.NumChunk == numAcceptedChunk {
		//закрываем дескриптор файла
		w.Close()

		//проверяем хеш-сумму файла
		ok := checkDownloadedFile(ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles, fi.Hex, fi.FullSizeByte)
		if !ok {
			return false, fmt.Errorf("invalid checksum for file %v (task ID %v)", fi.Name, pwbf.TaskID)
		}

		msgToCore.Command = "file download complete"

		//отмечаем файл как успешно принятый
		pwbf.SMT.UpdateTaskDownloadFileIsLoaded(pwbf.TaskID, configure.DownloadTaskParameters{
			DownloadingFilesInformation: map[string]*configure.DownloadFilesInformation{
				fi.Name: &configure.DownloadFilesInformation{},
			},
		})

		pwbf.ChanInCore <- &msgToCore

		return true, nil
	}

	return false, nil
}

func checkDownloadedFile(pathFile, fileHex string, fileSize int64) bool {
	fs, fh, err := common.GetFileParameters(pathFile)
	if err != nil {
		return false
	}

	if (fs != fileSize) || (fh != fileHex) {
		return false
	}

	return true
}
