package mytestpackages_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master/mytestpackages"
)

var _ = Describe("QueueTaskStorage", func() {
	qts := configure.NewRepositoryQTS()

	sourceID := 100
	clientID := "mifw77g6f63g"

	taskID := qts.AddQueueTaskStorage(
		sourceID,
		configure.CommonTaskInfo{
			IDClientAPI:     clientID,
			TaskIDClientAPI: "00e0ie0ir0i0r4",
			TaskType:        "download",
		},
		&configure.DescriptionParametersReceivedFromUser{
			DownloadList: []string{"file_name_1.tdp", "file_name_2.tdp"},
		})

	Context("Тест 1: Добавление в очередь новой задачи", func() {
		It("Должна вернутся информация по заданному taskID", func() {
			i, err := qts.GetQueueTaskStorage(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(i.TaskParameters.DownloadList)).Should(Equal(2))

		})
	})

	Context("Тест 2: Изменение статуса задачи", func() {
		It("Должен изменится статус задачи на 'execution'", func() {
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
		It("Должно вернутся значение 'TRUE'", func() {
			var err error

			err = qts.ChangeAvailabilityConnection(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())

			i, err := qts.GetQueueTaskStorage(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(i.CheckingStatusItems.AvailabilityConnection).Should(BeTrue())
		})
	})

	Context("Тест 4: Проверяем наличие файлов для скачивания", func() {
		It("Должно вернутся значение 'TRUE'", func() {
			var err error

			err = qts.ChangeAvailabilityFilesDownload(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())

			i, err := qts.GetQueueTaskStorage(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(i.CheckingStatusItems.AvailabilityFilesDownload).Should(BeTrue())
		})
	})

	Context("Тест 5: Удаление задачи из очередей", func() {
		It("Информация о задаче не может быть удалена из очереди так как задача в процессе выполнения", func() {
			e := qts.DelQueueTaskStorage(sourceID, taskID)
			Expect(e).Should(HaveOccurred())
		})

		It("Вся информация о задаче должна быть удалена из очереди", func() {
			var err error

			err = qts.ChangeTaskStatusQueueTask(sourceID, taskID, "complite")
			Expect(err).ShouldNot(HaveOccurred())

			err = qts.DelQueueTaskStorage(sourceID, taskID)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = qts.GetQueueTaskStorage(sourceID, taskID)
			Expect(err).Should(HaveOccurred())
		})
	})
})
