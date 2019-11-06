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
}

type typeChannelCommunication struct {
	handlerIP            string
	handlerID            string
	actionType           string
	channel              chan handlerRecivingParameters
	channelCommunication chan MsgChannelProcessorReceivingFiles
}

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

				msg.channel <- handlerRecivingParameters{}

			case "get":
				if _, ok := thrf.ListHandler[msg.handlerIP]; !ok {
					msg.channel <- handlerRecivingParameters{
						chanToHandler: nil,
					}
				}
				hrp, ok := thrf.ListHandler[msg.handlerIP][msg.handlerID]
				if !ok {
					msg.channel <- handlerRecivingParameters{
						chanToHandler: nil,
					}
				}

				msg.channel <- handlerRecivingParameters{
					chanToHandler: hrp.chanToHandler,
				}

			case "del":
				if _, ok := thrf.ListHandler[msg.handlerIP]; !ok {
					msg.channel <- handlerRecivingParameters{}

					break
				}
				hrp, ok := thrf.ListHandler[msg.handlerIP][msg.handlerID]
				if !ok {
					msg.channel <- handlerRecivingParameters{}

					break
				}

				close(hrp.chanToHandler)

				thrf.ListHandler[msg.handlerIP][msg.handlerID] = handlerRecivingParameters{
					chanToHandler: nil,
				}

				msg.channel <- handlerRecivingParameters{}
			}
		}
	}()

	return &thrf
}

//SetHendlerReceivingFile добавляет новый канал взаимодействия
func (thrf *TypeHandlerReceivingFile) SetHendlerReceivingFile(ip, id string, channel chan MsgChannelProcessorReceivingFiles) {
	chanRes := make(chan handlerRecivingParameters)

	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType:           "set",
		handlerIP:            ip,
		handlerID:            id,
		channel:              chanRes,
		channelCommunication: channel,
	}

	<-chanRes

	return
}

//GetHendlerReceivingFile возврашает канал для доступа к обработчику файлов
func (thrf *TypeHandlerReceivingFile) GetHendlerReceivingFile(ip, id string) chan MsgChannelProcessorReceivingFiles {
	chanRes := make(chan handlerRecivingParameters)

	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType: "get",
		handlerIP:  ip,
		handlerID:  id,
		channel:    chanRes,
	}

	return (<-chanRes).chanToHandler
}

//DelHendlerReceivingFile закрывает и удаляет канал
func (thrf *TypeHandlerReceivingFile) DelHendlerReceivingFile(ip, id string) {
	chanRes := make(chan handlerRecivingParameters)

	thrf.ChannelCommunicationReq <- typeChannelCommunication{
		actionType: "del",
		handlerIP:  ip,
		handlerID:  id,
		channel:    chanRes,
	}

	<-chanRes
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

	//сообщение об ошибке и сопутствующие действия
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
	lhrf := listHandlerReceivingFile{}

	go func(
		lhrf listHandlerReceivingFile,
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

				humanNotify := fmt.Sprintf("Не возможно отправить запрос на скачивание файлов, источник с ID %v не подключен", msg.SourceID)
				if !ok {
					humanNotify = fmt.Sprintf("Источник с ID %v не найден", msg.SourceID)

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

			ao.HumanDescriptionNotification = fmt.Sprintf("Источник с ID %v не найден", msg.SourceID)
			clientNotify.AdvancedOptions = ao

			errMsg := fmt.Sprintf("Source with ID %v not found", msg.SourceID)

			//fmt.Println("\tfunc 'ControllerReceivingRequestedFiles' 222222222")

			switch msg.Command {
			//начало выполнения задачи (запрос из Ядра)
			case "give my the files":
				if len(lhrf[si.IP]) == 0 {
					lhrf[si.IP] = listTaskReceivingFile{}
				}

				fmt.Println("\tfunc 'ControllerReceivingRequestedFiles' запуск обработчика задачи по скачиванию файлов")

				//запуск обработчика задачи по скачиванию файлов
				channel, err := processorReceivingFiles(chanInCore, msg.SourceID, si.IP, msg.TaskID, smt, saveMessageApp, cwtRes)
				if err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

					fmt.Printf("func 'handlerTypeDownloadFiles', ERROR (processorReceivingFiles):%v\n", err)

					ao.HumanDescriptionNotification = fmt.Sprintf("Отмена выполнения задачи по скачиванию файлов с источника ID %v, не найдены файлы для скачивания", msg.SourceID)
					clientNotify.AdvancedOptions = ao

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				lhrf[si.IP][msg.TaskID] = handlerRecivingParameters{
					chanToHandler: channel,
				}

			//останов выполнения задачи (запрос из Ядра)
			case "stop receiving files":

				fmt.Println("func 'ControllerReceivingRequestedFiles', COMMAND: 'stop receiving files'")

				if _, ok := lhrf[si.IP]; !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					//handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}
				hrp, ok := lhrf[si.IP][msg.TaskID]
				if !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					//handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				c := []byte("stop receiving files")

				fmt.Println("func 'ControllerReceivingRequestedFiles', SEND MSG 'stop receiving files' TO HANDLER (=-BEFORE-=) --->>>>")

				hrp.chanToHandler <- MsgChannelProcessorReceivingFiles{
					MessageType:  1,
					MsgGenerator: "Core module",
					Message:      &c,
				}

				close(hrp.chanToHandler)
				hrp.chanToHandler = nil

				fmt.Println("func 'ControllerReceivingRequestedFiles', SEND MSG 'stop receiving files' TO HANDLER (=-AFTER-=) --->>>>")

			//останов выполнения задачи из-за разрыва соединения (запрос из Ядра)
			case "to stop the task because of a disconnection":
				if _, ok := lhrf[si.IP]; !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}
				hrp, ok := lhrf[si.IP][msg.TaskID]
				if !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				c := []byte("to stop the task because of a disconnection")

				hrp.chanToHandler <- MsgChannelProcessorReceivingFiles{
					MessageType:  1,
					MsgGenerator: "Core module",
					Message:      &c,
				}

			//ответы приходящие от источника в рамках выполнения конкретной задачи
			case "taken from the source":

				//fmt.Printf("func ' ControllerReceivingRequestedFiles', RESIVED MSG 'taken from the source': '%v'\n", msg)

				if _, ok := lhrf[si.IP]; !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}
				hrp, ok := lhrf[si.IP][msg.TaskID]
				if !ok {
					_ = saveMessageApp.LogMessage("error", errMsg)

					handlerTaskWarning(msg.TaskID, clientNotify)

					continue
				}

				//fmt.Println("func ' ControllerReceivingRequestedFiles', send ---> to handler func 'processorReceivingFiles'")

				if hrp.chanToHandler == nil {
					fmt.Println("func 'ControllerReceivingRequestedFiles', Command: 'taken from the source', CHANNEL BE CLOSED")

					continue
				}

				//ответы приходящие от источника (команды для processorReceivingFiles)
				hrp.chanToHandler <- MsgChannelProcessorReceivingFiles{
					MessageType:  msg.MsgType,
					MsgGenerator: "NI module",
					Message:      msg.Message,
				}

			}
		}
	}(lhrf, clientNotify, chanIn)

	return chanIn
}
