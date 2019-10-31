package mytestpackages

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master"
)

var _ = Describe("Mytestpackages/InfoSourceList", func() {
	sourceID := 1100
	sourceShortName := "Source Test 1"
	token := "f89393t934ty4y45"
	ipAddr := "123.12.4.55"

	isl := configure.NewRepositoryISL()
	isl.AddSourceSettings(sourceID, configure.SourceSetting{
		IP:         ipAddr,
		ShortName:  sourceShortName,
		Token:      token,
		ClientName: "User 1",
		Settings: configure.InfoServiceSettings{
			MaxCountProcessFiltration: 3,
			StorageFolders:            []string{"Folder_1", "Folder_2", "Folder_3"},
			TypeAreaNetwork:           "ip",
		},
	})

	isl.AddSourceSettings(1101, configure.SourceSetting{
		IP:         "45.69.1.36",
		ShortName:  "Source Test 2",
		Token:      "vvvvvdvifivffdfdfd",
		ClientName: "User 1",
		Settings: configure.InfoServiceSettings{
			MaxCountProcessFiltration: 4,
			StorageFolders:            []string{"Folder_1", "Folder_2", "Folder_3"},
			TypeAreaNetwork:           "ip",
		},
	})

	Context("Тест 1. Проверка наличия информации об источнике", func() {
		It("Информация об источнике должна быть успешно получена", func() {
			si, ok := isl.GetSourceSetting(sourceID)

			Expect(ok).Should(BeTrue())
			Expect(si.ShortName).Should(Equal(sourceShortName))
		})
	})

	Context("Тест 2. Проверка поиска информации по источнику", func() {
		It("Информация по IP и токену должна быть найдена", func() {
			sID, ok := isl.SourceAuthenticationByIPAndToken(ipAddr, token)

			Expect(ok).Should(BeTrue())
			Expect(sID).Should(Equal(sourceID))
		})
	})

	Context("Тест 3. Проверка поиска ID по IP", func() {
		It("Должно возврящатся ID источника по его IP", func() {
			sID, ok := isl.GetSourceIDOnIP(ipAddr)

			Expect(ok).Should(BeTrue())
			Expect(sID).Should(Equal(sourceID))
		})
	})

	Context("Тест 5. Проверяем возможность получения статуса соединения с источником", func() {
		It("Получаем статус соединения с источником, должен быть FALSE", func() {
			isConn, err := isl.GetSourceConnectionStatus(sourceID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(isConn).ShouldNot(BeTrue())
		})
		It("Статус соединения с источником должен быть TRUE", func() {
			ok := isl.ChangeSourceConnectionStatus(sourceID, true)

			Expect(ok).Should(BeTrue())

			isConn, err := isl.GetSourceConnectionStatus(sourceID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(isConn).Should(BeTrue())
		})
	})

	Context("Тест 6. Проверить изменения разрешения доступа для источника", func() {
		It("Статус доступа должен быть FALSE", func() {
			isAllowed := isl.GetAccessIsAllowed("45.69.1.36")

			Expect(isAllowed).Should(BeFalse())
		})
		It("Статус доступа должен изменится на TRUE", func() {
			isl.SetAccessIsAllowed(1101)

			isAllowed := isl.GetAccessIsAllowed("45.69.1.36")

			Expect(isAllowed).Should(BeFalse())
		})
	})

	Context("Тест 7. Проверка удаления информации по источнику", func() {
		It("После удаления информации по источнику, информация об источнике НЕ должна быть успешно получена", func() {
			isl.DelSourceSettings(sourceID)

			_, ok := isl.GetSourceSetting(sourceID)

			Expect(ok).ShouldNot(BeTrue())
		})
	})

	/*Context("", func(){
		It("", func(){

		})
	})

	/*Context("", func(){
		It("", func(){

		})
	})*/
})
