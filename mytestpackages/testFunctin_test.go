package mytestpackages_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	//. "ISEMS-NIH_master/mytestpackages"
)

var _ = Describe("Function Test", func() {
	Context("Тест 1: проверкf ip адреса с помощью регулярного вырожения", func() {
		It("Должен вернуть true на валидное IP", func() {
			ok, err := common.CheckStringIP("123.0.56.45")

			Expect(ok).Should(Equal(true))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 2: проверка переданного значения подсети с помощью регулярных вырожений", func() {
		It("Должен вернуть true на валидное значение подсети", func() {
			ok, err := common.CheckStringNetwork("135.36.78.90/31")

			Expect(ok).Should(Equal(true))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	/*Context("Тест 3: проверка пользовательского типа содержащего информацию по задаче фильтрации", func() {
		It("Должен вернуть TRUE если принятые данные валидны", func() {
			fccpf := configure.FiltrationControlCommonParametersFiltration{
				ID: 1234,
				DateTime: configure.DateTimeParameters{
					Start: 1541665800, //1541665800
					End:   1541666700, //1541666700
				},
				Protocol: "tcp",
				Filters: configure.FiltrationControlParametersNetworkFilters{
					IP: configure.FiltrationControlIPorNetorPortParameters{
						Src: []string{"69.32.6.15", "89.63.1.66"},
					},
					Network: configure.FiltrationControlIPorNetorPortParameters{
						Dst: []string{"89.120.36.5/32", "78.40.36.56/23"},
					},
					Port: configure.FiltrationControlIPorNetorPortParameters{
						Any: []string{"80", "8080", "65536"},
					},
				},
			}

			str, isOK := handlerslist.CheckParametersFiltration(&fccpf)

			fmt.Printf("message: %v\n", str)

			Expect(isOK).Should(Equal(true))
			Expect(str).Should(Equal(""))
		})
	})*/
})
