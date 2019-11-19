package handlers

import (
	"fmt"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//MsgChannelProcessorReceivingFiles параметры канала взаимодействия между 'ControllerReceivingRequestedFiles' и 'processorReceivingFiles'
// MessageType - тип передаваемых данных (1 - text, 2 - binary)
// MsgGenerator - источник сообщения ('Core module', 'NI module')
// Message - информационное сообщение в бинарном виде
type MsgChannelProcessorReceivingFiles struct {
	MessageType  int
	MsgGenerator string
	Message      *[]byte
}

//TypeHandlerReceivingFile репозитория для хранения каналов взаимодействия с обработчиками записи файлов
type TypeHandlerReceivingFile struct {
	ListHandler             listHandlerReceivingFile
	ChannelCommunicationReq chan typeChannelCommunication
}

type typeChannelCommunication struct {
	handlerIP              string
	handlerID              string
	actionType             string
	msgForChunnelProcessor MsgChannelProcessorReceivingFiles
	channelErrMsg          chan error
	//	channel                chan typeChannelError
	channelCommunication chan MsgChannelProcessorReceivingFiles
}

//typeChannelError передает ошибку
type typeChannelError error

//listHandlerReceivingFile список задач по скачиванию файлов
// ключ - IP источника
type listHandlerReceivingFile map[string]listTaskReceivingFile

//listTaskReceivingFile список задач по приему файлов на данном источнике
// ключ - ID задачи
type listTaskReceivingFile map[string]handlerRecivingParameters

//handlerRecivingParameters описание параметров
type handlerRecivingParameters struct {
	chanToHandler chan MsgChannelProcessorReceivingFiles
}

//NewListHandlerReceivingFile создание нового репозитория для хранения каналов взаимодействия с обработчиками записи файлов
func NewListHandlerReceivingFile() *TypeHandlerReceivingFile {
	thrf := TypeHandlerReceivingFile{}
	thrf.ListHandler = listHandlerReceivingFile{}
	thrf.ChannelCommunicationReq = make(chan typeChannelCommunication)

	go func() {
		for msg := range thrf.ChannelCommunicationReq {
			switch msg.actionType {
			case "set":
				if _, ok := thrf.ListHandler[msg.handlerIP]; !ok {
					thrf.ListHandler[msg.handlerIP] = listTaskReceivingFile{}
				}

				thrf.ListHandler[msg.handlerIP][msg.handlerID] = handlerRecivingParameters{
					chanToHandler: msg.channelCommunication,
				}

				msg.channelErrMsg <- nil

			case "send data":
				if _, ok := thrf.ListHandler[msg.handlerIP]; !ok {

					fmt.Println("_=-+_+=-+_+ func 'NewListHandlerReceivingFile', client IP NOT found")

					msg.channelErrMsg <- fmt.Errorf("client IP not found")

					break
				}
				hrp, ok := thrf.ListHandler[msg.handlerIP][msg.handlerID]
				if !ok {

					fmt.Println("_=-+_+=-+_+ func 'NewListHandlerReceivingFile', task ID NOT found")

					msg.channelErrMsg <- fmt.Errorf("task ID not found")

					break
				}

				hrp.chanToHandler <- msg.msgForChunnelProcessor

				msg.channelErrMsg <- nil

			//fmt.Println("_=-+_+=-+_+ func 'NewListHandlerReceivingFile', CHANNEL TASK ID FOUND")

			case "del":
				if _, ok := thrf.ListHandler[msg.handlerIP]; !ok {
					msg.channelErrMsg <- fmt.Errorf("client IP not found")

					break
				}
				/*hrp*/ _, ok := thrf.ListHandler[msg.handlerIP][msg.handlerID]
				if !ok {
					msg.channelErrMsg <- fmt.Errorf("task ID not found")

					break
				}

				//close(hrp.chanToHandler)

				delete(thrf.ListHandler[msg.handlerIP], msg.handlerID)
				//				thrf.ListHandler[msg.handlerIP][msg.handlerID] = handlerRecivingParameters{}

				fmt.Println("_=-+_+=-+_+ func 'NewListHandlerReceivingFile', DELETE CHANNEL ----")

				msg.channelErrMsg <- nil
			}
		}
	}()

	return &thrf
}

//SetHendlerReceivingFile добавляет новый канал взаимодействия
func (thrf *TypeHandlerReceivingFile) SetHendlerReceivingFile(ip, id string, channel chan MsgChannelProcessorReceivingFiles) error {
	//	chanRes := make(chan handlerRecivingParameters)
	chanResErr := make(chan error)

	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType:           "set",
		handlerIP:            ip,
		handlerID:            id,
		channelErrMsg:        chanResErr,
		channelCommunication: channel,
	}

	return <-chanResErr
}

//SendChunkReceivingData отправляет через канал яасти принятого файла или информации
func (thrf *TypeHandlerReceivingFile) SendChunkReceivingData(ip, id string, msgSend MsgChannelProcessorReceivingFiles) error {
	chanResErr := make(chan error)

	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType:             "send data",
		handlerIP:              ip,
		handlerID:              id,
		msgForChunnelProcessor: msgSend,
		channelErrMsg:          chanResErr,
	}

	return <-chanResErr
}

//DelHendlerReceivingFile закрывает и удаляет канал
func (thrf *TypeHandlerReceivingFile) DelHendlerReceivingFile(ip, id string) error {
	//	chanRes := make(chan handlerRecivingParameters)
	chanResErr := make(chan error)

	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType:    "del",
		handlerIP:     ip,
		handlerID:     id,
		channelErrMsg: chanResErr,
	}

	return <-chanResErr
}

type messageFromCore string

type statusDownloadFile struct {
	Status string
	ErrMsg error
}

//ControllerReceivingRequestedFiles обработчик приема запрашиваемых файлов
func ControllerReceivingRequestedFiles(
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	isl *configure.InformationSourcesList,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI,
	cwtRes chan<- configure.MsgWsTransmission) chan *configure.MsgChannelReceivingFiles {

	clientNotify := configure.MsgBetweenCoreAndNI{
		Section: "message notification",
		Command: "send client API",
	}

	//обработка ошибки
	handlerTaskWarning := func(taskID string, msg configure.MsgBetweenCoreAndNI) {
		chanInCore <- &msg

		smt.UpdateTaskDownloadAllParameters(taskID, configure.DownloadTaskParameters{Status: "refused"})

		fmt.Println("func 'handlerTaskWarning', снимаем отслеживание выполнения задачи")

		//снимаем отслеживание выполнения задачи
		chanInCore <- &configure.MsgBetweenCoreAndNI{
			TaskID:  taskID,
			Section: "monitoring task performance",
			Command: "complete task",
		}
	}

	chanIn := make(chan *configure.MsgChannelReceivingFiles)
	lhrf := NewListHandlerReceivingFile()

	go func(
		lhrf *TypeHandlerReceivingFile,
		clientNotify configure.MsgBetweenCoreAndNI,
		chanIn <-chan *configure.MsgChannelReceivingFiles) {

		defer fmt.Println("====== ATTEMPTED!!! go func in 'ControllerReceivingRequestedFiles' BE STOPED =============")

		ao := configure.MessageNotification{
			SourceReport:        "NI module",
			Section:             "download control",
			TypeActionPerformed: "task processing",
			CriticalityMessage:  "warning",
		}

		for msg := range chanIn {
			clientNotify.TaskID = msg.TaskID
			ao.Sources = []int{msg.SourceID}

			if msg.Command != "taken from the source" {
				fmt.Printf("\tfunc 'ControllerReceivingRequestedFiles' resived new msg DOWNLOAD TASK for task ID %v, MSG %v\n", msg.TaskID, msg)
			}

			//получаем IP адрес и параметры источника
			si, ok := isl.GetSourceSetting(msg.SourceID)

			//fmt.Printf("\tfunc 'ControllerReceivingRequestedFiles' SOURCE INFO: %v, OK %v\n", si, ok)

			if !ok || !si.ConnectionStatus {
				_ = saveMessageApp.LogMessage("info", fmt.Sprintf("it is not possible to send a request to download files, the source with ID %v is not connected", msg.SourceID))

				fmt.Println("\tfunc 'ControllerReceivingRequestedFiles' ERROR 0000")

				humanNotify := common.PatternUserMessage(&common.TypePatternUserMessage{
					SourceID: msg.SourceID,
					TaskType: "скачивание файлов",
					Message:  "не возможно отправить запрос на скачивание файлов, источник не подключен",
				})
				if !ok {
					humanNotify = common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID: msg.SourceID,
						TaskType: "скачивание файлов",
						Message:  "источник не найден",
					})

					fmt.Println("\tfunc 'ControllerReceivingRequestedFiles' ERROR 1111")

					//изменяем статус задачи в storingMemoryQueueTask
					// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
					// функцией 'CheckTimeQueueTaskStorage')
					if err := qts.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "complete"); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}
				}

				ao.HumanDescriptionNotification = humanNotify
				clientNotify.AdvancedOptions = ao

				handlerTaskWarning(msg.TaskID, clientNotify)

				continue
			}

			//fmt.Printf("\tfunc 'ControllerReceivingRequestedFiles' RESIVED SOURCE PARAMETERS: %v\n", si)

			ao.HumanDescriptionNotification = common.PatternUserMessage(&common.TypePatternUserMessage{
				SourceID: msg.SourceID,
				TaskType: "скачивание файлов",
				Message:  "источник не найден",
			})
			clientNotify.AdvancedOptions = ao

			//errMsg := fmt.Sprintf("Source with ID %v not found", msg.SourceID)

			//fmt.Println("\tfunc 'ControllerReceivingRequestedFiles' 222222222")

			switch msg.Command {
			//начало выполнения задачи (запрос из Ядра)
			case "give my the files":
				fmt.Println("\tfunc 'ControllerReceivingRequestedFiles' запуск обработчика задачи по скачиванию файлов")

				//запуск обработчика задачи по скачиванию файлов
				channel, chanHandlerStoped, err := processorReceivingFiles(chanInCore, si.IP, msg.TaskID, smt, saveMessageApp, cwtRes)
				if err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					fmt.Printf("func 'handlerTypeDownloadFiles', ERROR (processorReceivingFiles):%v\n", err)

					ao.HumanDescriptionNotification = common.PatternUserMessage(&common.TypePatternUserMessage{
						SourceID:   msg.SourceID,
						TaskType:   "скачивание файлов",
						TaskAction: "задача отклонена",
						Message:    "не найдены файлы для скачивания",
					})
					clientNotify.AdvancedOptions = ao

					handlerTaskWarning(msg.TaskID, clientNotify)

					break
				}

				lhrf.SetHendlerReceivingFile(si.IP, msg.TaskID, channel)

				go func() {
					<-chanHandlerStoped

					fmt.Println("\tfunc 'ControllerReceivingRequestedFiles' принято сообщение от обработчика о его останове")

					//удаляем канал для взаимодействия с обработчиком так как
					// обработчик к этому времени завершил свою работу
					if err := lhrf.DelHendlerReceivingFile(si.IP, msg.TaskID); err != nil {
						_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
					}
				}()

			//останов выполнения задачи (запрос из Ядра)
			case "stop receiving files":

				fmt.Println("func 'ControllerReceivingRequestedFiles', COMMAND: 'stop receiving files'")
				fmt.Println("func 'ControllerReceivingRequestedFiles', SEND MSG 'stop receiving files' TO HANDLER (=-BEFORE-=) --->>>>")

				c := []byte("stop receiving files")
				if err := lhrf.SendChunkReceivingData(
					si.IP,
					msg.TaskID,
					MsgChannelProcessorReceivingFiles{
						MessageType:  1,
						MsgGenerator: "Core module",
						Message:      &c,
					}); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}

				fmt.Println("func 'ControllerReceivingRequestedFiles', SEND MSG 'stop receiving files' TO HANDLER (=-AFTER-=) --->>>>")

			//останов выполнения задачи из-за разрыва соединения (запрос из Ядра)
			case "to stop the task because of a disconnection":
				c := []byte("to stop the task because of a disconnection")
				if err := lhrf.SendChunkReceivingData(
					si.IP,
					msg.TaskID,
					MsgChannelProcessorReceivingFiles{
						MessageType:  1,
						MsgGenerator: "Core module",
						Message:      &c,
					}); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}

			//ответы приходящие от источника в рамках выполнения конкретной задачи
			case "taken from the source":

				//fmt.Printf("func ' ControllerReceivingRequestedFiles', RESIVED MSG 'taken from the source': '%v'\n", msg)
				//fmt.Println("func ' ControllerReceivingRequestedFiles', send ---> to handler func 'processorReceivingFiles'")

				if err := lhrf.SendChunkReceivingData(
					si.IP,
					msg.TaskID,
					MsgChannelProcessorReceivingFiles{
						MessageType:  msg.MsgType,
						MsgGenerator: "NI module",
						Message:      msg.Message,
					}); err != nil {

					fmt.Printf("func 'handlerTypeDownloadFiles', SECTION:'taken from the source' ERROR: '%v'\n", err)

					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}

				/*chanToHandler, err := lhrf.GetHendlerReceivingFile(si.IP, msg.TaskID)
				if err != nil {

					fmt.Printf("============= func 'handlerTypeDownloadFiles', MSG 'taken from the source' ERROR: chan eqval 'nil' (%v)\n", string(*msg.Message)[:100])

					_ = saveMessageApp.LogMessage("error", errMsg)

					break
				}

				//ответы приходящие от источника (команды для processorReceivingFiles)
				chanToHandler <- MsgChannelProcessorReceivingFiles{
					MessageType:  msg.MsgType,
					MsgGenerator: "NI module",
					Message:      msg.Message,
				}*/

			}
		}
	}(lhrf, clientNotify, chanIn)

	return chanIn
}
