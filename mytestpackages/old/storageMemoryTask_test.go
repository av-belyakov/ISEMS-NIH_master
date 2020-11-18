package mytestpackages_test

import (
	"fmt"
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
	taskID := common.GetUniqIDFormatMD5("wdmw99d9jw9j")

	//добавляем новую задачу
	smt.AddStoringMemoryTask(taskID, configure.TaskDescription{
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
		TaskParameter: configure.DescriptionTaskParameters{
			FiltrationTask: &configure.FiltrationTaskParameters{
				ID:     190,
				Status: "wait",
			},
		},
	})

	Context("Тест 1: Добавляем новую задачу и проверяем ее наличие", func() {
		It("Задача должна существовать, а так же должен быть ID источника и статус", func() {
			ti, isFound := smt.GetStoringMemoryTask(taskID)

			Expect(isFound).To(Equal(true))
			Expect(ti.TaskParameter.FiltrationTask.ID).Should(Equal(190))
			Expect(ti.TaskParameter.FiltrationTask.Status).Should(Equal("wait"))
		})
	})

	Context("Тест 2: Делаем ПЕРВЫЙ update параметров и проверяем их наличие", func() {
		It("Должны существовать некоторые параметры заданные пользователем", func() {
			//делаем update параметров фильтрации
			smt.UpdateTaskFiltrationAllParameters(taskID, &configure.FiltrationTaskParameters{
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
			smt.UpdateTaskFiltrationAllParameters(taskID, &configure.FiltrationTaskParameters{
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

			smt.UpdateTaskFiltrationAllParameters(taskID, &configure.FiltrationTaskParameters{
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

			smt.UpdateTaskFiltrationAllParameters(taskID, &configure.FiltrationTaskParameters{
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

			//fmt.Println(task)

			Expect(ok).Should(Equal(true))
			Expect(len(task.TaskParameter.FiltrationTask.FoundFilesInformation)).Should(Equal(3))
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

			smt.UpdateTaskFiltrationAllParameters(taskID, &configure.FiltrationTaskParameters{
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

	cliID := common.GetUniqIDFormatMD5("client_download id")
	cliTaskID := common.GetUniqIDFormatMD5("client_download_task id")
	tasID := common.GetUniqIDFormatMD5("fggg0k40kg04k04jh0j459tj49jt4g0j")
	pathStorage := "/home/ISEMS_NIH_master/ISEMS_NIH_master_OBJECT/313-OBU_ITC_Lipetsk/2019/August/11/11.08.2019T15:45-12.08.2019T07:23_hfeh8e83h38gh88485hg48"

	dfl := map[string]*configure.DownloadFilesInformation{}

	for i := 1; i < 10; i++ {
		fn := fmt.Sprintf("File_name_%v", i)
		dfl[fn] = &configure.DownloadFilesInformation{}

		dfl[fn].Size = int64(38 * i)
		dfl[fn].Hex = "ffw020f29f29f293"
	}

	fileTestName := "File_name_2"

	Context("Тест 5: Проверка добавления информации по задаче скачивания файлов", func() {
		It("Должна быть добавлена новая задача по скачиванию файлов, список файлов должен присутствовать", func() {
			smt.AddStoringMemoryTask(tasID, configure.TaskDescription{
				ClientID:                        cliID,
				ClientTaskID:                    cliTaskID,
				TaskType:                        "download control",
				ModuleThatSetTask:               "API module",
				ModuleResponsibleImplementation: "NI module",
				TimeUpdate:                      time.Now().Unix(),
				TimeInterval: configure.TimeIntervalTaskExecution{
					Start: time.Now().Unix(),
					End:   time.Now().Unix(),
				},
				TaskParameter: configure.DescriptionTaskParameters{
					DownloadTask: &configure.DownloadTaskParameters{
						ID:                                  3031,
						Status:                              "wait",
						NumberFilesTotal:                    len(dfl),
						PathDirectoryStorageDownloadedFiles: pathStorage,
						DownloadingFilesInformation:         dfl,
					},
				},
			})

			i, ok := smt.GetStoringMemoryTask(tasID)

			Expect(ok).Should(BeTrue())
			Expect(i.TaskParameter.DownloadTask.ID).Should(Equal(3031))
			Expect(len(i.TaskParameter.DownloadTask.DownloadingFilesInformation)).Should(Equal(9))
		})
	})

	Context("Тест 6: Проверка изменеия информации по задаче скачивания файлов", func() {
		It("Должна быть изменена информация по скачиванию файлов", func() {
			smt.UpdateTaskDownloadAllParameters(tasID, &configure.DownloadTaskParameters{
				Status:                              "execute",
				NumberFilesTotal:                    len(dfl),
				NumberFilesDownloaded:               1,
				PathDirectoryStorageDownloadedFiles: pathStorage,
				FileInformation: configure.DetailedFileInformation{
					Name:                fileTestName,
					Hex:                 "ffw020f29f29f293",
					FullSizeByte:        int64(38 * 2),
					AcceptedSizeByte:    38,
					AcceptedSizePercent: 50,
				},
			})

			//получаем информацию по задаче
			i, ok := smt.GetStoringMemoryTask(tasID)

			shortFileInfo, ok := i.TaskParameter.DownloadTask.DownloadingFilesInformation[fileTestName]

			Expect(ok).Should(BeTrue())
			Expect(shortFileInfo.IsLoaded).ShouldNot(BeTrue())
			Expect(i.TaskParameter.DownloadTask.FileInformation.Name).Should(Equal(fileTestName))
			Expect(i.TaskParameter.DownloadTask.FileInformation.AcceptedSizeByte).Should(Equal(int64(38)))
		})
	})

	Context("Тест 7: Проверка изменеия информации по задаче скачивания файлов", func() {
		It("Файл должен быть отмечен как скаченный", func() {
			smt.UpdateTaskDownloadAllParameters(tasID, &configure.DownloadTaskParameters{
				Status:                              "execute",
				NumberFilesTotal:                    len(dfl),
				NumberFilesDownloaded:               1,
				PathDirectoryStorageDownloadedFiles: pathStorage,
				FileInformation: configure.DetailedFileInformation{
					Name:                "File_name_2",
					Hex:                 "ffw020f29f29f293",
					FullSizeByte:        int64(38 * 2),
					AcceptedSizeByte:    int64(38 * 2),
					AcceptedSizePercent: 100,
				},
			})

			i, ok := smt.GetStoringMemoryTask(tasID)

			//shortFileInfo, ok := i.TaskParameter.DownloadTask.DownloadingFilesInformation[fileTestName]

			Expect(ok).Should(BeTrue())
			//Expect(shortFileInfo.IsLoaded).Should(BeTrue())
			Expect(i.TaskParameter.DownloadTask.FileInformation.Name).Should(Equal(fileTestName))
			Expect(i.TaskParameter.DownloadTask.FileInformation.AcceptedSizeByte).Should(Equal(int64(76)))
		})
	})

	Context("Тест 8: Проверка изменения статуса IsLoaded для выбранного файла", func() {
		It("Статус файла с заданными именами должны быть изменены на true", func() {
			//"1438535410_2015_08_02____20_10_13_23267.tdp" "1438535410_2015_08_02____20_10_14_724263.tdp"

			smt.UpdateTaskDownloadFileIsLoaded(tasID, &configure.DownloadTaskParameters{
				DownloadingFilesInformation: map[string]*configure.DownloadFilesInformation{
					"File_name_3": &configure.DownloadFilesInformation{},
					"File_name_5": &configure.DownloadFilesInformation{},
				},
			})

			i, ok := smt.GetStoringMemoryTask(tasID)

			for fn, param := range i.TaskParameter.DownloadTask.DownloadingFilesInformation {
				fmt.Printf("file name: %v, param 'IsLoaded': %v\n", fn, param.IsLoaded)

				if fn == "File_name_3" || fn == "File_name_5" {
					Expect(param.IsLoaded).Should(BeTrue())
				}
			}

			Expect(ok).Should(BeTrue())
		})
	})

	Context("Тест 9: Проверка изменения статуса задачи", func() {
		It("Должен изменится статус задачи на 'завершенная'", func() {
			smt.CompleteStoringMemoryTask(tasID)

			i, ok := smt.GetStoringMemoryTask(tasID)

			Expect(ok).Should(BeTrue())
			Expect(i.TaskStatus).Should(BeTrue())
		})
	})
})
