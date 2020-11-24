package mytestpackages

import (
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
	//	. "ISEMS-NIH_master"
)

type listTask []listTaskInfo

type listTaskInfo struct {
	time int64
	id   string
}

func (lt listTask) Len() int {
	return len(lt)
}

func (lt listTask) Swap(i, j int) {
	lt[i], lt[j] = lt[j], lt[i]
}

func (lt listTask) Less(i, j int) bool {
	return lt[i].time < lt[j].time
}

type searchTaskDescription struct {
	addedDataTask int64
}

var _ = Describe("Mytestpackages/TemporaryStotageSearchQueries", func() {
	clientID := "jf9ej9393hfh48h48h49h99rg94994hg94"
	sp := configure.SearchParameters{
		TaskProcessed: false,
		ID:            1010,
		InstalledFilteringOption: configure.SearchFilteringOptions{
			DateTime: configure.DateTimeParameters{
				Start: 1576713600,
				End:   1576886400,
			},
			Protocol: "any",
			NetworkFilters: configure.FiltrationControlParametersNetworkFilters{
				IP: configure.FiltrationControlIPorNetorPortParameters{
					Any: []string{"104.238.175.16", "115.171.23.128"},
					Src: []string{"72.105.58.23"},
				},
				Port: configure.FiltrationControlIPorNetorPortParameters{
					Dst: []string{"8080"},
				},
			},
		},
	}

	tssq := configure.NewRepositoryTSSQ(configure.TypeRepositoryTSSQ{
		TickerSec:      3,
		TimeExpiration: 4,
		MaxCacheSize:   10,
	})
	oneTaskID, _, _ := tssq.CreateNewSearchTask(clientID, &sp)
	//oneTaskID := configure.CreateTmpStorageID(clientID, &sp)

	//fmt.Printf("TASK ID MAJOR: %q\n", oneTaskID)

	Context("Тест №1. Создание временного хранилища задач по поиску", func() {
		It("Должна быть успешно создана новое хранилище задач для хранения результатов поиска", func() {

			//fmt.Printf("New TemporaryStotageSearchQueries %v\n", tssq)

			Expect(tssq).ShouldNot(BeNil())
		})
	})

	Context("Тест #2. Проверяем генерацию идентификатора задачи по поиску информации в БД", func() {
		It("Идентификаторы сгенерированые на основе одних и техже параметров должны быть равны", func() {
			twoTaskID := configure.CreateTmpStorageID(clientID, &sp)

			//fmt.Printf("______ oneTaskID: %q\n", oneTaskID)
			//fmt.Printf("______ twoTaskID: %q\n", twoTaskID)

			Expect(oneTaskID).Should(Equal(twoTaskID))
		})
	})

	Context("Тест №3. Добавление информации о новой задаче по поиску в БД, если она уже есть", func() {
		It("Так как уже задача с ID уже существует (clientID и параметры поиска одинаковы) возвращается вся информация о задаче", func() {
			tmpStorageID, info, err := tssq.CreateNewSearchTask(clientID, &sp)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(tmpStorageID).Should(Equal(oneTaskID))
			Expect(info).ShouldNot(BeNil())
		})
	})

	Context("Тест №4. Добавление информации о новой задаче по поиску в БД, если ее еще нет в хранилище", func() {
		It("Должен вернуться новый ID задачи и nil вместо информации о задаче", func() {
			tmpStorageID, info, err := tssq.CreateNewSearchTask("vn9h83h33f4f84g8", &configure.SearchParameters{ID: 1021})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(tmpStorageID).ShouldNot(Equal(oneTaskID))
			Expect(info).Should(BeNil())
		})
	})

	Context("Тест №5. Проверка работы функции 'GetInformationAboutSearchTask'", func() {
		It("Функция GetInformationAboutSearchTask должна вернуть ошибку при поиске информации по ID задачи", func() {
			info, err := tssq.GetInformationAboutSearchTask("e9ve990")

			Expect(info).Should(BeNil())
			Expect(err).Should(HaveOccurred())
		})

		It("Функция GetInformationAboutSearchTask должна вернуть информацию по существующему ID задачи", func() {
			info, err := tssq.GetInformationAboutSearchTask(oneTaskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(info).ShouldNot(BeNil())
			Expect(info.SearchParameters.ID).Should(Equal(1010))
		})
	})

	Context("Тест №6. Проверить сортировку отображения с информацией о задаче", func() {
		It("Отображение должно правильно быть отсортировано", func() {
			getOldestRecord := func(list map[string]*searchTaskDescription) string {
				ls := make(listTask, 0, len(list))

				for key, value := range list {
					ls = append(ls, listTaskInfo{time: value.addedDataTask, id: key})
				}

				sort.Sort(ls)

				return ls[0].id
			}

			data := map[string]*searchTaskDescription{
				"dg77gd7dddddddg8": &searchTaskDescription{addedDataTask: 123456789},
				"vkndide38fh8hffd": &searchTaskDescription{addedDataTask: 124456789},
				"92hd8h8fe8f83h39": &searchTaskDescription{addedDataTask: 121456789},
				"c90wj99c9939939f": &searchTaskDescription{addedDataTask: 123452789},
				"cc89wdh999393r9r": &searchTaskDescription{addedDataTask: 323456789},
				"cncwh8hw8hd8ef83": &searchTaskDescription{addedDataTask: 23456789},
			}

			oldestRecord := getOldestRecord(data)

			//fmt.Printf("The oldest record: %q\n", oldestRecord)

			Expect(oldestRecord).Should(Equal("cncwh8hw8hd8ef83"))
		})
	})

	Context("Тест №7. Проверяем статус задачи который меняется при передачи информации клиенту API", func() {
		It("Изменяем статус задачи на 'информация передается клиенту'. Результат должен быть TRUE.", func() {
			err := tssq.ChangeStatusTransmissionTask(oneTaskID, true)

			Expect(err).ShouldNot(HaveOccurred())

			taskInfo, err := tssq.GetInformationAboutSearchTask(oneTaskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect((*taskInfo).TransmissionStatus).Should(BeTrue())
		})
		It("Изменяем статус задачи на 'информация передается клиенту'. Результат должен быть FALSE.", func() {
			err := tssq.ChangeStatusTransmissionTask(oneTaskID, false)

			Expect(err).ShouldNot(HaveOccurred())

			taskInfo, err := tssq.GetInformationAboutSearchTask(oneTaskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(taskInfo.TransmissionStatus).Should(BeFalse())
		})
	})

	Context("Тест №8.1. Проверяем актуальность информации о задаче (не было ли изменения в БД)", func() {
		It("Информация о задаче должна быть актуальна, так как по данной задачи ЕЩЕ идет поиск. Параметр должен быть FALSE", func() {
			tssq.ChangingStatusInformationRelevance([]string{oneTaskID})
			taskInfo, err := tssq.GetInformationAboutSearchTask(oneTaskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(taskInfo.NotRelevance).Should(BeFalse())
		})
	})

	Context("Тест 9. Проверяем добавление найденной информации и изменяем задачу на неактуальную так как с момента поиска в БД произошли изменение", func() {
		It("Должен быть добавлен список информации полученный из БД и должен быть успешно изменен параметор NotRelevance на TRUE", func() {
			err := tssq.AddInformationFoundSearchResult(oneTaskID, []*configure.BriefTaskInformation{
				&configure.BriefTaskInformation{
					TaskID:              "ficn939f994994hf94h9gh",
					ClientTaskID:        "nc9939h3f9h39h9fh949",
					SourceID:            1000,
					FilteringTaskStatus: "execute",
				},
				&configure.BriefTaskInformation{
					TaskID:              "cmeomof39fhn9h44h9h94",
					ClientTaskID:        "nc9939h3f9h39h9fh949",
					SourceID:            1001,
					FilteringTaskStatus: "execute",
				},
				&configure.BriefTaskInformation{
					TaskID:              "cm9f39fh9h9h49h94hg9h4g",
					ClientTaskID:        "nc9939h3f9h39h9fh949",
					SourceID:            1000,
					FilteringTaskStatus: "complete",
				},
			})

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест №8.2. Проверяем актуальность информации о задаче (не было ли изменения в БД)", func() {
		It("Информация о задаче должна быть не актуальна, то есть проверяемый параметр должен быть TRUE", func() {
			tssq.ChangingStatusInformationRelevance([]string{oneTaskID})
			taskInfo, err := tssq.GetInformationAboutSearchTask(oneTaskID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(taskInfo.NotRelevance).Should(BeTrue())
			Expect(len(taskInfo.ListFoundInformation.List)).Should(Equal(3))
		})
	})

	/*
				Context("", func(){
		It("", func(){

		})
			})
	*/

	/*

	   !!! Дальше писать методы доступа к temporaryStorageSearchQueries и тестировать их !!!

	*/

	/*Context("Тест №7. Проверка работы функции 'CheckTimeDeleteTemporaryStorageSearchQueries'", func() {
		It("Функция CheckTimeDeleteTemporaryStorageSearchQueries должна вернуть True через 6 секунд", func(done Done) {
			msg := tssq.CheckTimeDeleteTemporaryStorageSearchQueries(1)

			fmt.Println(msg)

			Expect(msg).To(ContainSubstring("DELETE"))
			close(done)
		})
	})*/
})
