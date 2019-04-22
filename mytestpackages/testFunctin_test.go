package mytestpackages_test

import (
	"fmt"

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

	Context("Тест 3: проверяем количество частей сообщения", func() {
		It("Должен вернуть определенное количество частей", func() {
			chankSize := 10
			list := map[string]int{
				"d_1": 45,
				"d_2": 48,
				"d_3": 47,
				"d_4": 45,
				"d_5": 44,
				"d_6": 51,
			}

			Expect(common.GetCountPartsMessage(list, chankSize)).Should(Equal(6))
		})
	})

	Context("Тест 4: делим список с файлами на части", func() {
		It("Функция должна вернуть определенное количество частей со списками файлов", func() {
			lf := map[string][]string{
				"D_1": []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9"},
				"D_2": []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11"},
				"D_3": []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8"},
				"D_4": []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10"},
				"D_5": []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12", "f13"},
			}

			getListFiles := func(numPart, sizeChunk, countParts int, listFilesFilter map[string][]string) map[string][]string {
				lff := map[string][]string{}

				fmt.Printf("Parts: %v\n", numPart)

				for disk, files := range listFilesFilter {
					if numPart == 1 {
						if len(files) < sizeChunk {
							lff[disk] = files[:]
						} else {
							lff[disk] = files[:sizeChunk]
						}

						continue
					}

					num := sizeChunk * (numPart - 1)
					numEnd := num + sizeChunk

					if numPart == countParts {
						if num < len(files) {
							lff[disk] = files[num:]

							continue
						}

						lff[disk] = []string{}
					}

					if numPart < countParts {
						if num > len(files) {
							lff[disk] = []string{}

							continue
						}

						if numEnd < len(files) {
							lff[disk] = files[num:numEnd]

							continue
						}

						lff[disk] = files[num:]
					}

				}
				return lff
			}

			sizeChunk := 4

			lft := map[string]int{}
			for d, f := range lf {
				lft[d] = len(f)
			}

			countParts := common.GetCountPartsMessage(lft, sizeChunk)

			for i := 1; i <= countParts; i++ {
				list := getListFiles(i, sizeChunk, countParts, lf)

				fmt.Printf("List: %v\n", list)

			}

			Expect(true).Should(Equal(true))

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
