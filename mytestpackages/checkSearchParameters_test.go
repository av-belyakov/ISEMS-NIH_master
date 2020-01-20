package mytestpackages

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/coreapp/handlerslist"
)

/*
//SearchInformationAboutTasksRequestOption дополнительные опции для поиска информации по задаче
// TaskProcessed - была ли задача отмечена клиентом API как завершенная
// ID - уникальный цифровой идентификатор источника
// NumberTasksReturnedPart - количество задач в возвращаемой части (не обязательный параметр)
// FilesDownloaded - опции выгрузки файлов
// InformationAboutFiltering - поиск по информации о результатах фильтрации
// InstalledFilteringOption - установленные опции фильтрации

//FilesDownloadedOptions опции выгрузки файлов
// FilesIsDownloaded - выполнялась ли выгрузка файлов
// AllFilesIsDownloaded - были ли выгружены все файлы

//InformationAboutFilteringOptions опции для поиска по информации о результатах фильтрации
// FilesIsFound - были ли найдены файлы
// CountAllFilesMin - минимальное общее количество всех найденных файлов
// CountAllFilesMax - максимальное общее количество всех найденных файлов
// SizeAllFilesMin - общий минимальный размер всех найденных файлов
// SizeAllFilesMax - общий максимальный размер всех найденных файлов

//SearchFilteringOptions искомые опции фильтрации
// DateTime - временной диапазон по которому осуществлялась фильтрация
// Protocol - тип транспортного протокола
// NetworkFilters - сетевые фильтры
*/
var _ = Describe("CheckSearchParameters", func() {
	Context("Тест 1. Тестируем общую проверку типа используемого при поиске информации", func() {
		It("При проверке типа должно возвращаться True", func() {
			searchOptions := &configure.SearchInformationAboutTasksRequestOption{
				ID: 1002,
			}

			msg, ok := handlerslist.CheckParametersSearchCommonInformation(searchOptions)

			fmt.Printf("RESEIVED MESSAGE: %q\n", msg)
			fmt.Println(searchOptions)

			Expect(ok).Should(BeTrue())
		})

		It("При проверке типа должно возвращаться False", func() {
			searchOptions := &configure.SearchInformationAboutTasksRequestOption{
				ID:                      1002,
				NumberTasksReturnedPart: 20,
				InstalledFilteringOption: configure.SearchFilteringOptions{
					DateTime: configure.DateTimeParameters{
						Start: 16643312331,
						End:   156164646133,
					},
					Protocol: "udp",
					NetworkFilters: configure.FiltrationControlParametersNetworkFilters{
						IP: configure.FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							//Src: []string{"96.4.66.3", "9.464.4.3"},
							Dst: []string{},
						},
						Port: configure.FiltrationControlIPorNetorPortParameters{
							//Any: []string{"899998"},
							Src: []string{},
							Dst: []string{},
						},
						Network: configure.FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{"101.63.6.66/35"},
							Dst: []string{},
						},
					},
				},
			}

			msg, ok := handlerslist.CheckParametersSearchCommonInformation(searchOptions)

			fmt.Printf("RESEIVED MESSAGE: %q\n", msg)
			fmt.Println(searchOptions)

			Expect(ok).Should(BeFalse())
		})
	})

	Context("Тест 2. Тестируем функцию проверяющую опции фильтрации на валидность входных параметров", func() {
		It("Тест должен пройти успешно, ошибки быть не должно", func() {
			err := handlerslist.LoopHandler(handlerslist.LoopHandlerParameters{
				handlerslist.OptionsCheckFilterParameters{
					SourceID: 1003,
					TaskType: "поиск информации",
				},
				map[string]map[string]*[]string{
					"IP": map[string]*[]string{
						"Any": &[]string{"12.36.55.1", "59.100.33.6"},
						"Src": &[]string{"95.102.36.5"},
						"Dst": &[]string{},
					},
					"Port": map[string]*[]string{
						"Any": &[]string{},
						"Src": &[]string{},
						"Dst": &[]string{"23", "445"},
					},
					"Network": map[string]*[]string{
						"Any": &[]string{},
						"Src": &[]string{"89.1.36.8/25"},
						"Dst": &[]string{},
					},
				},
				handlerslist.CheckIPPortNetwork,
			})

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должна быть ошибка по невалидному ip адресу", func() {
			err := handlerslist.LoopHandler(handlerslist.LoopHandlerParameters{
				handlerslist.OptionsCheckFilterParameters{
					SourceID: 1003,
					TaskType: "поиск информации",
				},
				map[string]map[string]*[]string{
					"IP": map[string]*[]string{
						"Any": &[]string{"12.36.55.1", "59.100.33.6"},
						"Src": &[]string{"95.102.36.5"},
						"Dst": &[]string{"132.566.11.3"},
					},
					"Port": map[string]*[]string{
						"Any": &[]string{},
						"Src": &[]string{},
						"Dst": &[]string{"23", "445"},
					},
					"Network": map[string]*[]string{
						"Any": &[]string{},
						"Src": &[]string{"89.1.36.8/25"},
						"Dst": &[]string{},
					},
				},
				handlerslist.CheckIPPortNetwork,
			})

			//fmt.Printf("ERROR (IP invalid): %q\n", err)

			Expect(err).Should(HaveOccurred())
		})

		It("Должна быть ошибка по невалидному сетевому порту", func() {
			err := handlerslist.LoopHandler(handlerslist.LoopHandlerParameters{
				handlerslist.OptionsCheckFilterParameters{
					SourceID: 1003,
					TaskType: "поиск информации",
				},
				map[string]map[string]*[]string{
					"IP": map[string]*[]string{
						"Any": &[]string{"12.36.55.1", "59.100.33.6"},
						"Src": &[]string{"95.102.36.5"},
						"Dst": &[]string{},
					},
					"Port": map[string]*[]string{
						"Any": &[]string{"0", "8080"},
						"Src": &[]string{},
						"Dst": &[]string{"23", "445"},
					},
					"Network": map[string]*[]string{
						"Any": &[]string{},
						"Src": &[]string{"89.1.36.8/25"},
						"Dst": &[]string{},
					},
				},
				handlerslist.CheckIPPortNetwork,
			})

			//fmt.Printf("ERROR (PORT invalid): %q\n", err)

			Expect(err).Should(HaveOccurred())

		})

		It("Должна быть ошибка по невалидному значению подсети", func() {
			err := handlerslist.LoopHandler(handlerslist.LoopHandlerParameters{
				handlerslist.OptionsCheckFilterParameters{
					SourceID: 1003,
					TaskType: "поиск информации",
				},
				map[string]map[string]*[]string{
					"IP": map[string]*[]string{
						"Any": &[]string{"12.36.55.1", "59.100.33.6"},
						"Src": &[]string{"95.102.36.5"},
						"Dst": &[]string{},
					},
					"Port": map[string]*[]string{
						"Any": &[]string{},
						"Src": &[]string{},
						"Dst": &[]string{"23", "445"},
					},
					"Network": map[string]*[]string{
						"Any": &[]string{"120.1336.46.3/25"},
						"Src": &[]string{"89.1.36.8/25"},
						"Dst": &[]string{},
					},
				},
				handlerslist.CheckIPPortNetwork,
			})

			//fmt.Printf("ERROR (NETWORK invalid): %q\n", err)

			Expect(err).Should(HaveOccurred())
		})
	})

	/*Context("Тест 3. Тестируем функцию проверяющую опции фильтрации на наличие параметров фильтрации", func() {
		It("Должно вернутся false, так как переданное значение сетевых параметров пустое", func() {
			isNotEmpty := handlerslist.CheckNetworkParametersIsNotEmpty(map[string]map[string]*[]string{
				"IP": map[string]*[]string{
					"Any": &[]string{},
					"Src": &[]string{},
					"Dst": &[]string{},
				},
				"Port": map[string]*[]string{
					"Any": &[]string{},
					"Src": &[]string{},
					"Dst": &[]string{},
				},
				"Network": map[string]*[]string{
					"Any": &[]string{},
					"Src": &[]string{},
					"Dst": &[]string{},
				},
			})

			Expect(isNotEmpty).Should(BeFalse())
		})

		It("Должно вернутся true, так как переданное значение сетевых параметров заполненно", func() {
			isNotEmpty := handlerslist.CheckNetworkParametersIsNotEmpty(map[string]map[string]*[]string{
				"IP": map[string]*[]string{
					"Any": &[]string{"12.36.55.1", "59.100.33.6"},
					"Src": &[]string{"95.102.36.5"},
					"Dst": &[]string{},
				},
				"Port": map[string]*[]string{
					"Any": &[]string{},
					"Src": &[]string{},
					"Dst": &[]string{"23", "445"},
				},
				"Network": map[string]*[]string{
					"Any": &[]string{"120.1336.46.3/25"},
					"Src": &[]string{"89.1.36.8/25"},
					"Dst": &[]string{},
				},
			})

			Expect(isNotEmpty).Should(BeTrue())
		})
	})*/
})
