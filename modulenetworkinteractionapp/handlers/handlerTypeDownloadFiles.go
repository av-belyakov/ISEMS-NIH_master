package handlers

import (
	"fmt"

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
	ChannelErrorResponse    chan error
}

type typeChannelCommunication struct {
	handlerIP              string
	handlerID              string
	actionType             string
	msgForChunnelProcessor MsgChannelProcessorReceivingFiles
	channelCommunication   chan MsgChannelProcessorReceivingFiles
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
	thrf := TypeHandlerReceivingFile{
		ListHandler:             listHandlerReceivingFile{},
		ChannelCommunicationReq: make(chan typeChannelCommunication),
		ChannelErrorResponse:    make(chan error),
	}

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

				thrf.ChannelErrorResponse <- nil

			case "send data":
				if _, ok := thrf.ListHandler[msg.handlerIP]; !ok {
					thrf.ChannelErrorResponse <- fmt.Errorf("not action 'send data', client IP '%v' not found", msg.handlerIP)

					continue
				}
				hrp, ok := thrf.ListHandler[msg.handlerIP][msg.handlerID]
				if !ok {
					thrf.ChannelErrorResponse <- fmt.Errorf("not action 'send data', task ID '%v' not found", msg.handlerID)

					continue
				}

				hrp.chanToHandler <- msg.msgForChunnelProcessor
				thrf.ChannelErrorResponse <- nil

			case "del":
				if _, ok := thrf.ListHandler[msg.handlerIP]; !ok {
					thrf.ChannelErrorResponse <- fmt.Errorf("not action 'delete', client IP '%v' not found", msg.handlerIP)

					continue
				}
				_, ok := thrf.ListHandler[msg.handlerIP][msg.handlerID]
				if !ok {
					thrf.ChannelErrorResponse <- fmt.Errorf("not action 'delete', task ID '%v' not found", msg.handlerID)

					continue
				}

				delete(thrf.ListHandler[msg.handlerIP], msg.handlerID)

				thrf.ChannelErrorResponse <- nil
			}
		}
	}()

	return &thrf
}

//SetHendlerReceivingFile добавляет новый канал взаимодействия
func (thrf *TypeHandlerReceivingFile) SetHendlerReceivingFile(ip, id string, channel chan MsgChannelProcessorReceivingFiles) error {
	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType:           "set",
		handlerIP:            ip,
		handlerID:            id,
		channelCommunication: channel,
	}

	return <-thrf.ChannelErrorResponse
}

//SendChunkReceivingData отправляет через канал части принятого файла или информации
func (thrf *TypeHandlerReceivingFile) SendChunkReceivingData(ip, id string, msgSend MsgChannelProcessorReceivingFiles) error {
	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType:             "send data",
		handlerIP:              ip,
		handlerID:              id,
		msgForChunnelProcessor: msgSend,
	}

	return <-thrf.ChannelErrorResponse
}

//DelHendlerReceivingFile закрывает и удаляет канал по ID задачи с ний связанной
func (thrf *TypeHandlerReceivingFile) DelHendlerReceivingFile(ip, id string) error {
	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType: "del",
		handlerIP:  ip,
		handlerID:  id,
	}

	return <-thrf.ChannelErrorResponse
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

	funcName := "ControllerReceivingRequestedFiles"

	clientNotify := configure.MsgBetweenCoreAndNI{
		Section: "message notification",
		Command: "send client API",
	}

	//обработка ошибки
	handlerTaskWarning := func(taskID string, msg configure.MsgBetweenCoreAndNI) {
		chanInCore <- &msg

		smt.UpdateTaskDownloadAllParameters(taskID, &configure.DownloadTaskParameters{Status: "refused"})

		//снимаем отслеживание выполнения задачи
		chanInCore <- &configure.MsgBetweenCoreAndNI{
			TaskID:  taskID,
			Section: "monitoring task performance",
			Command: "complete task",
		}
	}

	chanIn := make(chan *configure.MsgChannelReceivingFiles)
	lhrf := NewListHandlerReceivingFile()

	go func(lhrf *TypeHandlerReceivingFile,
		clientNotify configure.MsgBetweenCoreAndNI,
		chanIn <-chan *configure.MsgChannelReceivingFiles) {

		ao := configure.MessageNotification{
			SourceReport:        "NI module",
			Section:             "download files",
			TypeActionPerformed: "stop",
			CriticalityMessage:  "warning",
		}

		for msg := range chanIn {
			clientNotify.TaskID = msg.TaskID
			ao.Sources = []int{msg.SourceID}

			//получаем IP адрес и параметры источника
			si, ok := isl.GetSourceSetting(msg.SourceID)
			if !ok || !si.ConnectionStatus {
				saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
					Description: fmt.Sprintf("it is not possible to send a request to download files, the source with ID %v is not connected", msg.SourceID),
					FuncName:    funcName,
				})

				humanNotify := "не возможно отправить запрос на скачивание файлов, источник не подключен"

				if !ok {
					humanNotify = "источник не найден"

					//изменяем статус задачи в storingMemoryQueueTask
					// на 'complete' (ПОСЛЕ ЭТОГО ОНА БУДЕТ АВТОМАТИЧЕСКИ УДАЛЕНА
					// функцией 'CheckTimeQueueTaskStorage')
					if err := qts.ChangeTaskStatusQueueTask(msg.SourceID, msg.TaskID, "complete"); err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}
				}

				ao.HumanDescriptionNotification = humanNotify
				clientNotify.AdvancedOptions = ao

				handlerTaskWarning(msg.TaskID, clientNotify)

				continue
			}

			ao.HumanDescriptionNotification = "источник не найден"
			clientNotify.AdvancedOptions = ao

			switch msg.Command {
			//начало выполнения задачи (запрос из Ядра)
			case "give my the files":

				fmt.Printf("func '%v', received new request to dofnload file\n", funcName)

				//запуск обработчика задачи по скачиванию файлов
				channel, chanHandlerStoped, err := processorReceivingFiles(chanInCore, si.IP, msg.TaskID, smt, saveMessageApp, cwtRes)
				if err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})

					ao.HumanDescriptionNotification = "не найдены файлы для скачивания"
					clientNotify.AdvancedOptions = ao

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				lhrf.SetHendlerReceivingFile(si.IP, msg.TaskID, channel)

				go func() {
					<-chanHandlerStoped

					//удаляем канал для взаимодействия с обработчиком так как
					// обработчик к этому времени завершил свою работу
					if err := lhrf.DelHendlerReceivingFile(si.IP, msg.TaskID); err != nil {
						saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
							Description: fmt.Sprint(err),
							FuncName:    funcName,
						})
					}
				}()

			//останов выполнения задачи (запрос из Ядра)
			case "stop receiving files":
				c := []byte("stop receiving files")
				if err := lhrf.SendChunkReceivingData(
					si.IP,
					msg.TaskID,
					MsgChannelProcessorReceivingFiles{
						MessageType:  1,
						MsgGenerator: "Core module",
						Message:      &c,
					}); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

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
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}

			//ответы приходящие от источника в рамках выполнения конкретной задачи
			case "taken from the source":
				if err := lhrf.SendChunkReceivingData(
					si.IP,
					msg.TaskID,
					MsgChannelProcessorReceivingFiles{
						MessageType:  msg.MsgType,
						MsgGenerator: "NI module",
						Message:      msg.Message,
					}); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    funcName,
					})
				}
			}
		}
	}(lhrf, clientNotify, chanIn)

	return chanIn
}
