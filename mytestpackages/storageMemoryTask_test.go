package mytestpackages_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master/mytestpackages"
)

var _ = Describe("StorageMemoryTask", func() {
	smt := configure.NewRepositorySMT()

	//генерируем хеш для clientID
	clientID := common.GetUniqIDFormatMD5("client id")
	clientTaskID := common.GetUniqIDFormatMD5("client_task id")

	//добавляем новую задачу
	taskID := smt.AddStoringMemoryTask(configure.TaskDescription{
		ClientID:                        clientID,
		ClientTaskID:                    clientTaskID,
		TaskType:                        "filtration control",
		ModuleThatSetTask:               "API module",
		ModuleResponsibleImplementation: "NI module",
		TimeUpdate:                      time.Now().Unix(),
		TimeInterval: configure.TimeIntervalTaskExecution{
			Start: time.Now().Unix(),
			End:   time.Now().Unix(),
		},
	})

	Context("Тест 1: Добавляем новую задачу и проверяем ее наличие", func() {
		It("Задача должна существовать", func() {
			_, isFound := smt.GetStoringMemoryTask(taskID)

			Expect(isFound).To(Equal(true))
		})
	})

	Context("Тест 2: Делаем ПЕРВЫЙ update параметров и проверяем их наличие", func() {
		It("Должны существовать некоторые параметры заданные пользователем", func() {
			//делаем update параметров фильтрации
			smt.UpdateTaskFiltrationAllParameters(taskID, configure.FiltrationTaskParameters{
				ID:                              190,
				Status:                          "execute",
				NumberFilesMeetFilterParameters: 231,
				NumberProcessedFiles:            1,
				NumberFilesFoundResultFiltering: 0,
				NumberDirectoryFiltartion:       3,
				NumberErrorProcessedFiles:       0,
				SizeFilesMeetFilterParameters:   4738959669055,
				SizeFilesFoundResultFiltering:   0,
				PathStorageSource:               "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
			})

			//проверяем заданные параметры фильтрации
			task, ok := smt.GetStoringMemoryTask(taskID)
			taskInfo := task.TaskParameter.FiltrationTask

			Expect(ok).Should(Equal(true))

			Expect(taskInfo.ID).Should(Equal(190))
			Expect(taskInfo.Status).Should(Equal("execute"))
			Expect(taskInfo.NumberProcessedFiles).Should(Equal(1))
		})
	})

	Context("Тест 3: Формируем список файлов найденных по результатам фильтрации и проверяем их наличие", func() {
		It("Должнен быть сформирован список файлов найденных по результатам фильтрации", func() {
			//делаем update параметров фильтрации
			smt.UpdateTaskFiltrationAllParameters(taskID, configure.FiltrationTaskParameters{
				ID:                              190,
				Status:                          "execute",
				NumberFilesMeetFilterParameters: 231,
				NumberProcessedFiles:            2,
				NumberFilesFoundResultFiltering: 1,
				NumberDirectoryFiltartion:       3,
				NumberErrorProcessedFiles:       0,
				SizeFilesMeetFilterParameters:   4738959669055,
				SizeFilesFoundResultFiltering:   0,
				PathStorageSource:               "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
				FoundFilesInformation: map[string]*configure.FoundFilesInformation{
					"1438537240_2015_08_02____20_40_40_104870.tdp": &configure.FoundFilesInformation{
						Size: 3484834452,
						Hex:  "fj9j939j9t88232",
					},
				},
			})

			smt.UpdateTaskFiltrationAllParameters(taskID, configure.FiltrationTaskParameters{
				ID:                              190,
				Status:                          "execute",
				NumberFilesMeetFilterParameters: 231,
				NumberProcessedFiles:            10,
				NumberFilesFoundResultFiltering: 2,
				NumberDirectoryFiltartion:       3,
				NumberErrorProcessedFiles:       0,
				SizeFilesMeetFilterParameters:   4738959669055,
				SizeFilesFoundResultFiltering:   0,
				PathStorageSource:               "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
				FoundFilesInformation: map[string]*configure.FoundFilesInformation{
					"1438536146_2015_08_02____20_22_26_974623.tdp": &configure.FoundFilesInformation{
						Size: 748956954,
						Hex:  "fj9j939j9t88232",
					},
				},
			})

			smt.UpdateTaskFiltrationAllParameters(taskID, configure.FiltrationTaskParameters{
				ID:                              190,
				Status:                          "execute",
				NumberFilesMeetFilterParameters: 231,
				NumberProcessedFiles:            12,
				NumberFilesFoundResultFiltering: 3,
				NumberDirectoryFiltartion:       3,
				NumberErrorProcessedFiles:       0,
				SizeFilesMeetFilterParameters:   4738959669055,
				SizeFilesFoundResultFiltering:   0,
				PathStorageSource:               "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
				FoundFilesInformation: map[string]*configure.FoundFilesInformation{
					"1438535410_2015_08_02____20_10_10_644263.tdp": &configure.FoundFilesInformation{
						Size: 1448375,
						Hex:  "fj9j939j9t88232",
					},
				},
			})

			//проверяем заданные параметры фильтрации
			task, ok := smt.GetStoringMemoryTask(taskID)
			taskInfo := task.TaskParameter.FiltrationTask

			Expect(ok).Should(Equal(true))
			Expect(len(taskInfo.FoundFilesInformation)).Should(Equal(3))
		})
	})

	Context("Тест 4: Проверка функции добавления списка найденных файлов", func() {
		It("По результатам должен быть получен список файлов кол-во которых соответствует заданным условиям", func() {
			newFilesList := map[string]*configure.FoundFilesInformation{
				"1438535410_2015_08_02____20_10_12_556677.tdp": &configure.FoundFilesInformation{
					Size: 545868,
					Hex:  "fefee888f88e7f7e7e",
				},
				"1438535410_2015_08_02____20_10_13_23267.tdp": &configure.FoundFilesInformation{
					Size: 34454666,
					Hex:  "fere0r0r30r33999994ffd",
				},
				"1438535410_2015_08_02____20_10_14_724263.tdp": &configure.FoundFilesInformation{
					Size: 34400005,
					Hex:  "iewe8828e2e82888484848df8s",
				},
			}

			smt.UpdateTaskFiltrationAllParameters(taskID, configure.FiltrationTaskParameters{
				ID:                              190,
				Status:                          "execute",
				NumberFilesMeetFilterParameters: 231,
				NumberProcessedFiles:            231,
				NumberFilesFoundResultFiltering: 6,
				NumberDirectoryFiltartion:       3,
				NumberErrorProcessedFiles:       0,
				SizeFilesMeetFilterParameters:   4738959669055,
				SizeFilesFoundResultFiltering:   0,
				PathStorageSource:               "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
				FoundFilesInformation:           newFilesList,
			})

			//smt.UpdateTaskFiltrationFilesList(taskID, newFilesList)

			//проверяем заданные параметры фильтрации
			task, ok := smt.GetStoringMemoryTask(taskID)
			taskInfo := task.TaskParameter.FiltrationTask

			Expect(ok).Should(Equal(true))
			Expect(len(taskInfo.FoundFilesInformation)).Should(Equal(6))
		})
	})
})
