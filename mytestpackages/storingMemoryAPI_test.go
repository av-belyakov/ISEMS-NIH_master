package mytestpackages

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
)

var _ = Describe("Mytestpackages/StoringMemoryAPI", func() {
	//инициализируем новый репозиторий
	smapi := configure.NewRepositorySMAPI()

	userIP := "77.21.36.13"

	//добавляем несколько пользователей
	clientID1 := smapi.AddNewClient(userIP, "user name 1")
	clientID2 := smapi.AddNewClient("89.26.1.46", "user name 2")
	_ = smapi.AddNewClient("96.23.1.33", "user name 3")

	Context("Тест 1. Получаем список клиентов", func() {
		It("Должен быть получен список клиентов состоящий из 3", func() {
			Expect(len(smapi.GetClientList())).Should(Equal(3))
		})
	})
	Context("Тест 2. Получить всю информацию о клиенте по его ID", func() {
		It("Информация о клиенте 'clientID1' должна быть получена, ошибки быть не должно", func() {
			cs, err := smapi.GetClientSettings(clientID1)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(cs.IP).Should(Equal(userIP))
		})
	})
	Context("Тест 3. Найти информацию о клиенте по его IP", func() {
		It("Должен вернуть информацию о клиенте по его IP", func() {
			cID, _, isExist := smapi.SearchClientForIP(userIP)

			Expect(isExist).Should(BeTrue())
			Expect(cID).Should(Equal(clientID1))
		})
	})
	Context("Тест 31. Не найти информацию о клиенте по его IP", func() {
		It("Не должен вернуть информацию о клиенте по его IP", func() {
			cID, _, isExist := smapi.SearchClientForIP("12.13.10.3")

			Expect(isExist).Should(BeFalse())
			Expect(cID).Should(Equal(""))
		})
	})
	Context("Тест 4. Удалить клиента", func() {
		It("Клиент должен быть удален", func() {
			smapi.DelClientAPI(clientID1)
		})
		It("Клиент должен быть удален, кол-во клиентов должно быть 2", func() {
			Expect(len(smapi.GetClientList())).Should(Equal(2))
		})
		It("Клиент должен быть удален, поиск по ID клиента вернет ошибку", func() {
			_, err := smapi.GetClientSettings(clientID1)

			Expect(err).Should(HaveOccurred())
		})
	})
	Context("Тест 5. Получить линк соединения у клиента которого нет", func() {
		It("Должна вернутся ошибка", func() {
			_, err := smapi.GetWssClientConnection(clientID1)

			Expect(err).Should(HaveOccurred())
		})
	})
	Context("Тест 6. Получить всю информацию о клиенте по его ID", func() {
		It("Информация о клиенте 'clientID2' должна быть получена, ошибки быть не должно", func() {
			cs, err := smapi.GetClientSettings(clientID2)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(cs.IP).Should(Equal("89.26.1.46"))
		})
	})
})
