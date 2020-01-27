package mytestpackages

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
	//	. "ISEMS-NIH_master"
)

var _ = Describe("Mytestpackages/TemporaryStotageSearchQueries", func() {
	clientID := "jf9ej9393hfh48h48h49h99rg94994hg94"

	tssq := configure.NewRepositoryTSSQ()

	Context("Тест №1. Создание временного хранилища задач по поиску", func() {
		It("Должна быть успешно создана новое хранилище задач для хранения результатов поиска", func() {

			fmt.Printf("New TemporaryStotageSearchQueries %v\n", tssq)

			Expect(tssq).ShouldNot(BeNil())
		})
	})

	Context("Тест #2. Проверяем генерацию идентификатора задачи по поиску информации в БД", func() {
		It("Идентификаторы сгенерированые на основе одних и техже параметров должны быть равны", func() {
			oneTaskID := configure.CreateTmpStorageID(clientID, &configure.SearchParameters{})
			twoTaskID := configure.CreateTmpStorageID(clientID, &configure.SearchParameters{})

			fmt.Printf("______ tmpTaskID: %q\n", oneTaskID)

			Expect(oneTaskID).To(Equal(twoTaskID))
		})
	})

	Context("Тест №3. Добавление информации о новой задаче по поиску в БД", func() {
		It("Должна быть успешно добавлена новая задача, ошибки быть не должно", func() {
			tmpStorageID, err := tssq.CreateNewSearchTask(clientID, &configure.SearchParameters{})

			fmt.Printf("temporary storage ID: %q\n", tmpStorageID)

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест №4. Проверка работы функции 'CheckTimeDeleteTemporaryStorageSearchQueries'", func() {
		It("Функция CheckTimeDeleteTemporaryStorageSearchQueries должна вернуть True через 6 секунд", func(done Done) {
			msg := tssq.CheckTimeDeleteTemporaryStorageSearchQueries(1)

			fmt.Println(msg)

			Expect(msg).To(ContainSubstring("DELETE"))
			close(done)
		})
	})
})
