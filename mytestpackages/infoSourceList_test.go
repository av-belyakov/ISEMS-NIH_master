package mytestpackages

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master"
)

var _ = Describe("Mytestpackages/InfoSourceList", func() {
	sourceShortName := "Source Test 1"

	isl := configure.NewRepositoryISL()
	isl.AddSourceSettings(1100, configure.SourceSetting{
		IP:         "123.12.4.55",
		ShortName:  sourceShortName,
		Token:      "f89393t934ty4y45",
		ClientName: "User 1",
		Settings: configure.InfoServiceSettings{
			MaxCountProcessFiltration: 3,
			StorageFolders:            []string{"Folder_1", "Folder_2", "Folder_3"},
			TypeAreaNetwork:           "ip",
		},
	})

	Context("Тест 1: Добавляем инфромацию об источнике", func() {
		It("Информация об источнике должна быть успешно добавлена", func() {
			si, ok := isl.GetSourceSetting(1100)

			Expect(ok).Should(BeTrue())
			Expect(si.ShortName).Should(Equal(sourceShortName))
		})
	})

	/*Context("", func(){
		It("", func(){

		})
	})*/
})
