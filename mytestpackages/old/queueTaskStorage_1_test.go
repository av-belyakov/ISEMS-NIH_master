package mytestpackages

import (
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master/mytestpackages"
)

func checkFileNameMatches(lfdb []*configure.FilesInformation, lfqst []string) (map[string]configure.DetailedFilesInformation, error) {
	type fileInfo struct {
		hex      string
		size     int64
		isLoaded bool
	}

	nlf := make(map[string]configure.DetailedFilesInformation, len(lfqst))

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
				nlf[f] = configure.DetailedFilesInformation{
					Size: info.size,
					Hex:  info.hex,
				}
			}
		}
	}

	return nlf, nil
}

/*func checkFileNameMatches(lfdb []*configure.FilesInformation, lfqst []string) ([]*configure.DetailedFileInformation, error) {
	type fileInfo struct {
		hex      string
		size     int64
		isLoaded bool
	}

	nlf := make([]*configure.DetailedFileInformation, 0, len(lfqst))

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
				nlf = append(nlf, &configure.DetailedFileInformation{
					Name:         f,
					Hex:          info.hex,
					FullSizeByte: info.size,
				})
			}
		}
	}

	return nlf, nil
}*/

var _ = Describe("QueueTaskStorage", func() {
	qts := configure.NewRepositoryQTS()

	sourceID := 100
	clientID := "mifw77g6f63g"

	taskID := common.GetUniqIDFormatMD5("idwi99d92")

	listFilesUser := make([]string, 0, 6)
	for i := 5; i < 10; i++ {
		listFilesUser = append(listFilesUser, fmt.Sprintf("file_000%v", i))
	}
	listFilesUser = append(listFilesUser, "file_0022")

	testPathDirForFilterFiles := "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_June_17_0_0_b072340dc4e69d6df229a26c14edbfc8"

	qts.AddQueueTaskStorage(
		taskID,
		sourceID,
		configure.CommonTaskInfo{
			IDClientAPI:     clientID,
			TaskIDClientAPI: "00e0ie0ir0i0r4",
			TaskType:        "download control",
		},
		&configure.DescriptionParametersReceivedFromUser{
			DownloadList:                  listFilesUser,
			PathDirectoryForFilteredFiles: testPathDirForFilterFiles,
		})

	listFilesDB := make([]*configure.FilesInformation, 0, 10)

	for num := 0; num < 10; num++ {
		listFilesDB = append(listFilesDB, &configure.FilesInformation{
			FileName: fmt.Sprintf("file_000%v", num),
			FileHex:  "nifeh3883hf8f98gf49494hg94",
			FileSize: time.Now().Unix() + int64(num),
		})
	}

	listFilesDB = append(listFilesDB, &configure.FilesInformation{
		FileName:   "file_0022",
		FileHex:    "dif9wj9fj9932f33hg94",
		FileSize:   59549549,
		FileLoaded: true,
	})

	Context("Тест 1: Добавление в очередь новой задачи", func() {
		It("Должна вернуться информация по заданным sourceID и taskID", func() {
			i, err := qts.GetQueueTaskStorage(sourceID, taskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(i.TaskParameters.DownloadList)).Should(Equal(6))
			Expect(i.TaskParameters.PathDirectoryForFilteredFiles).Should(Equal(testPathDirForFilterFiles))
		})
	})

	Context("Тест 2: Изменение статуса задачи", func() {
		It("Должен измениться статус задачи на 'execution'", func() {
			var err error

			err = qts.ChangeTaskStatusQueueTask(sourceID, taskID, "execution")
			Expect(err).ShouldNot(HaveOccurred())

			i, err := qts.GetQueueTaskStorage(sourceID, taskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(i.TaskStatus).Should(Equal("execution"))
			Expect(i.IDClientAPI).Should(Equal(clientID))
		})
	})

	Context("Тест 3: Проверяем наличие соединения с источником", func() {
		It("Должно вернуться значение 'TRUE'", func() {
			var err error

			err = qts.ChangeAvailabilityConnectionOnConnection(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())

			i, err := qts.GetQueueTaskStorage(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(i.CheckingStatusItems.AvailabilityConnection).Should(BeTrue())
		})
	})

	Context("Тест 4: Проверяем наличие файлов для скачивания", func() {
		It("Должно вернуться значение 'TRUE'", func() {
			var err error

			err = qts.ChangeAvailabilityFilesDownload(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())

			i, err := qts.GetQueueTaskStorage(sourceID, taskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(i.CheckingStatusItems.AvailabilityFilesDownload).Should(BeTrue())
		})
	})

	Context("Тест 5: Проверяем поиск информации о задаче только по ID задачи", func() {
		_ = qts.ChangeAvailabilityFilesDownload(sourceID, taskID)

		sID, i, err := qts.SearchTaskForIDQueueTaskStorage(taskID)

		It("Должна вернуться информация о задаче", func() {
			Expect(err).ShouldNot(HaveOccurred())
			Expect(i.CheckingStatusItems.AvailabilityFilesDownload).Should(BeTrue())
			Expect(sID).Should(Equal(100))
		})

		It("Должна вернутся ошибка так как задачи с указанным ID не найдено", func() {
			_, _, err := qts.SearchTaskForIDQueueTaskStorage("fff9993j9f3")
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("Тест 6: Проверить функцию отвечающую за сравнение имен файлов из двух разных списков", func() {
		/*ait, _ := qts.GetAllTaskQueueTaskStorage(sourceID)

		fmt.Println("++++++++++++++++++++++++++")
		fmt.Println(ait)
		fmt.Println("++++++++++++++++++++++++++")*/

		It("Функция должна вернуть список совпадющих имен файлов", func() {
			ti, err := qts.GetQueueTaskStorage(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())

			finalyList, err := checkFileNameMatches(listFilesDB, ti.TaskParameters.DownloadList)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(finalyList)).Should(Equal(5))
		})
	})

	Context("Тест 7: Проверка добавления нового списка в storingMemoryQueryTask", func() {
		ti, _ := qts.GetQueueTaskStorage(sourceID, taskID)
		finalyList, _ := checkFileNameMatches(listFilesDB, ti.TaskParameters.DownloadList)

		It("При добавлении нового списка файлов не должно быть ошибок", func() {
			err := qts.AddConfirmedListFiles(sourceID, taskID, finalyList)
			Expect(err).ShouldNot(HaveOccurred())
			/*
				infoTask, err := qts.GetQueueTaskStorage(sourceID, taskID)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(infoTask.TaskParameters.DownloadList)).Should(Equal(0))

				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(infoTask.TaskParameters.ConfirmedListFiles)).Should(Equal(5))*/
		})

	})

	Context("Тест 8: Проверка старого и нового списка файлов", func() {
		It("Старый, пользовательский список файлов должен быть удален", func() {

			infoTask, err := qts.GetQueueTaskStorage(sourceID, taskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(infoTask.TaskParameters.DownloadList)).Should(Equal(0))
		})

		It("Новый список файлов должен быть добавлен в StoringMemoryQueueTask", func() {

			infoTask, err := qts.GetQueueTaskStorage(sourceID, taskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(infoTask.TaskParameters.ConfirmedListFiles)).Should(Equal(5))
		})
	})

	/*Context("Тест 8: Удаление задачи из очередей", func() {
		It("Информация о задаче не может быть удалена из очереди так как задача в процессе выполнения", func() {
			e := qts.delQueueTaskStorage(sourceID, taskID)
			Expect(e).Should(HaveOccurred())
		})

		It("Вся информация о задаче должна быть удалена из очереди", func() {
			var err error

			err = qts.ChangeTaskStatusQueueTask(sourceID, taskID, "complete")
			Expect(err).ShouldNot(HaveOccurred())

			err = qts.delQueueTaskStorage(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = qts.GetQueueTaskStorage(sourceID, taskID)
			Expect(err).Should(HaveOccurred())
		})
	})*/

	Context("Тест 9: Проверка поиска задачи после удаления всех задач из очереди", func() {
		It("Должна вернутся ошибка так как в очереди нет задач", func() {
			_, _, err := qts.SearchTaskForIDQueueTaskStorage(taskID)
			Expect(err).Should(HaveOccurred())
		})
	})
})
