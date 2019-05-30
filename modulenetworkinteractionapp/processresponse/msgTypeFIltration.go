package processresponse

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//ProcessingReceivedMsgTypeFiltering обработка сообщений связанных с фильтрацией файлов
func ProcessingReceivedMsgTypeFiltering(
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	isl *configure.InformationSourcesList,
	smt *configure.StoringMemoryTask,
	message *[]byte,
	sourceID int,
	cwtReq <-chan configure.MsgWsTransmission) {

	fmt.Println("START function 'ProcessingReceivedMsgTypeFilteringn'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	resMsg := configure.MsgTypeFiltration{}

	if err := json.Unmarshal(*message, &resMsg); err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
	}

	switch resMsg.Info.TaskStatus {
	case "execute":
		/* Просто отправляем в ядро, а от туда в БД иклиенту API */

		chanInCore <- &configure.MsgBetweenCoreAndNI{
			TaskID:          resMsg.Info.TaskID,
			Section:         "filtration control",
			Command:         "execute",
			SourceID:        sourceID,
			AdvancedOptions: resMsg,
		}

	case "stop":

		/*
		   Тоже самое что и при 'execute', но с начала объединяем все списки
		   файлов найденных в результате фильтрации в один и отправляем сообщение
		   'confirm complite'
		*/

	case "complite":

		/*
			Тоже самое что и при 'execute', но с начала объединяем все списки
			файлов найденных в результате фильтрации в один и отправляем сообщение
			'confirm complite'
		*/

	}
}

/*

	!!! ВНИМАНИЕ !!!
	При получении сообщения о завершении фильтрации, 'stop' или 'complite'
	и сборке всего списка файлов полученного в результате фильтрации,
	то есть обработать параметр 'NumberMessagesFrom' [2]int
	ОТПРАВИТЬ сообщение "confirm complite" что бы удалить задачу на
	стороне ISEMS-NIH_slave

		//конвертирование принятой информации из формата JSON

						//запись информации о ходе выполнения задачи в память
						smt.UpdateTaskFiltrationAllParameters(taskID, configure.FiltrationTaskParameters{
							ID:                       190,
							Status:                   "execute",
							NumberFilesToBeFiltered:  231,
							SizeFilesToBeFiltered:    4738959669055,
							CountDirectoryFiltartion: 3,
							NumberFilesProcessed:     12,
							NumberFilesFound:         3,
							SizeFilesFound:           0,
							PathStorageSource:        "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
							FoundFilesInformation: map[string]*configure.FoundFilesInformation{
								"1438535410_2015_08_02____20_10_10_644263.tdp": &configure.FoundFilesInformation{
									Size: 1448375,
									Hex:  "fj9j939j9t88232",
								},
							},
						})

						//формируем из ответа от ISEMS_NIH_slave
					dif := configure.DetailedInformationFiltering{
						TaskStatus                      string
						TimeIntervalTaskExecution       TimeInterval
						WasIndexUsed                    bool
						NumberFilesMeetFilterParameters int
						NumberProcessedFiles            int
						NumberFilesFoundResultFiltering int
						NumberDirectoryFiltartion       int
						SizeFilesMeetFilterParameters   int64
						SizeFilesFoundResultFiltering   int64
						PathDirectoryForFilteredFiles   string
						ListFilesFoundResultFiltering   []*InformationFilesFoundResultFiltering
					}

						//отправка инофрмации на запись в БД
						chanInCore<-&configure.MsgBetweenCoreAndDB{
							Section:         "filtration",
						Command:         "update",
						SourceID:        sourceID,
						TaskID: "получаем из ответа ISEMS_NIH_slave",
						AdvancedOptions: dif,
						}
*/
