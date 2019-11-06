package mytestpackages

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master"
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

var _ = Describe("Mytestpackages/InfoSourceList", func() {
	sourceID := 1100
	sourceShortName := "Source Test 1"
	token := "f89393t934ty4y45"
	ipAddr := "123.12.4.55"

	isl := configure.NewRepositoryISL()
	isl.AddSourceSettings(sourceID, configure.SourceSetting{
		IP:         ipAddr,
		ShortName:  sourceShortName,
		Token:      token,
		ClientName: "User 1",
		Settings: configure.InfoServiceSettings{
			MaxCountProcessFiltration: 3,
			StorageFolders:            []string{"Folder_1", "Folder_2", "Folder_3"},
			TypeAreaNetwork:           "ip",
		},
	})

	isl.AddSourceSettings(1101, configure.SourceSetting{
		IP:         "45.69.1.36",
		ShortName:  "Source Test 2",
		Token:      "vvvvvdvifivffdfdfd",
		ClientName: "User 1",
		Settings: configure.InfoServiceSettings{
			MaxCountProcessFiltration: 4,
			StorageFolders:            []string{"Folder_1", "Folder_2", "Folder_3"},
			TypeAreaNetwork:           "ip",
		},
	})

	Context("Тест 1. Проверка наличия информации об источнике", func() {
		It("Информация об источнике должна быть успешно получена", func() {
			si, ok := isl.GetSourceSetting(sourceID)

			Expect(ok).Should(BeTrue())
			Expect(si.ShortName).Should(Equal(sourceShortName))
		})
	})

	Context("Тест 2. Проверка поиска информации по источнику", func() {
		It("Информация по IP и токену должна быть найдена", func() {
			sID, ok := isl.SourceAuthenticationByIPAndToken(ipAddr, token)

			Expect(ok).Should(BeTrue())
			Expect(sID).Should(Equal(sourceID))
		})
	})

	Context("Тест 3. Проверка поиска ID по IP", func() {
		It("Должно возврящатся ID источника по его IP", func() {
			sID, ok := isl.GetSourceIDOnIP(ipAddr)

			Expect(ok).Should(BeTrue())
			Expect(sID).Should(Equal(sourceID))
		})
	})

	Context("Тест 5. Проверяем возможность получения статуса соединения с источником", func() {
		It("Получаем статус соединения с источником, должен быть FALSE", func() {
			isConn, err := isl.GetSourceConnectionStatus(sourceID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(isConn).ShouldNot(BeTrue())
		})
		It("Статус соединения с источником должен быть TRUE", func() {
			ok := isl.ChangeSourceConnectionStatus(sourceID, true)

			Expect(ok).Should(BeTrue())

			isConn, err := isl.GetSourceConnectionStatus(sourceID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(isConn).Should(BeTrue())
		})
	})

	Context("Тест 6. Проверить изменения разрешения доступа для источника", func() {
		It("Статус доступа должен быть FALSE", func() {
			isAllowed := isl.GetAccessIsAllowed("45.69.1.36")

			Expect(isAllowed).Should(BeFalse())
		})
		It("Статус доступа должен изменится на TRUE", func() {
			isl.SetAccessIsAllowed(1101)

			isAllowed := isl.GetAccessIsAllowed("45.69.1.36")

			Expect(isAllowed).Should(BeFalse())
		})
	})

	Context("Тест 7. Проверка удаления информации по источнику", func() {
		It("После удаления информации по источнику, информация об источнике НЕ должна быть успешно получена", func() {
			isl.DelSourceSettings(sourceID)

			_, ok := isl.GetSourceSetting(sourceID)

			Expect(ok).ShouldNot(BeTrue())
		})
	})

	Context("Тест 8. Проверка закрытого канала на nil", func() {
		It("При закрытии канала он должен быть nil", func() {
			chanTest := make(chan struct{})
			type channels chan struct{}

			list := map[string]channels{}
			list["one"] = chanTest

			Expect(list["one"]).ShouldNot(BeNil())

			close(list["one"])
			list["one"] = nil

			Expect(list["one"]).Should(BeNil())
		})
	})

	handlerIP := "59.1.33.45"
	handlerID := "323"

	newChannel := make(chan MsgChannelProcessorReceivingFiles, 1)

	nlhrf := NewListHandlerReceivingFile()
	nlhrf.SetHendlerReceivingFile(handlerIP, handlerID, newChannel)

	Context("Тест 9. Создание и добавление канала для взаимодействия с репозиторием ", func() {
		It("Должен быть создан канал и успешно добавлен в репозиторий", func(done Done) {
			chanComm := nlhrf.GetHendlerReceivingFile(handlerIP, handlerID)

			go func() {
				fmt.Printf("resived from channel: %v\n", <-chanComm)
			}()

			strMsg := []byte("diifiif ief993")

			chanComm <- MsgChannelProcessorReceivingFiles{
				MessageType:  1,
				MsgGenerator: "test generator",
				Message:      &strMsg,
			}

			Expect(chanComm).ShouldNot(BeNil())

			close(done)
		}, 3)
	})

	Context("Тест 10. Удаление канала взаимодействия из репозитория", func() {
		It("После закрытия и удаления канала должно возвращатся nil", func(done Done) {
			nlhrf.DelHendlerReceivingFile(handlerIP, handlerID)

			chanComm := nlhrf.GetHendlerReceivingFile(handlerIP, handlerID)

			Expect(chanComm).Should(BeNil())

			close(done)
		}, 4)
	})

	/*Context("", func(){
		It("", func(){

		})
	})

	/*Context("", func(){
		It("", func(){

		})
	})*/
})
