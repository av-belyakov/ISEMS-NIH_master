package mytestpackages

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"
)

func createList() []string {
	l := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		l = append(l, fmt.Sprintf("new value %v", i))
	}

	return l
}

func cycleForAndSwith(list []string) error {
	var err error

	for _, v := range list {

		//fmt.Printf("number: '%v'\n", k)

		switch v {
		case "new value 2":
			fmt.Println("case new value 2 (Before)")

			continue

			fmt.Println("case new value 2 (After)")

		case "new value 6":
			fmt.Println("case new value 6 (Before)")

			break

			fmt.Println("case new value 6 (After)")
			err = fmt.Errorf("case new value 6")

		default:

			//fmt.Printf("-= default case: %v =-\n", v)
		}
	}

	return err
}

//MongoDBConnect содержит дискриптор соединения с БД
type MongoDBConnect struct {
	Connect *mongo.Client
	CTX     context.Context
}

//sourcesListSetting настройки источников, ключ ID источника
type sourcesListSetting map[int]sourceSetting

//SourceSetting параметры источника
// ConnectionStatus - статус соединения с источником
// IP - ip адрес источника
// ShortName - краткое название источника
// DateLastConnected - время последнего соединения (в формате unix timestamp)
// Token - токен для авторизации
// ClientName - имя клиента API (нужно для того чтобы контролировать управление определенным источником)
// AccessIsAllowed - разрешен ли доступ, по умолчанию false (при проверке токена ставится true если он верен)
// AsServer - false запуск как клиент, true запуск как сервер
type sourceSetting struct {
	ConnectionStatus  bool
	IP                string
	ShortName         string
	DateLastConnected int64
	Token             string
	ClientName        string
	AccessIsAllowed   bool
	AsServer          bool
	Settings          InfoServiceSettings
}

//InfoServiceSettings содержит настройки источника
type InfoServiceSettings struct {
	EnableTelemetry           bool     `json:"enable_telemetry" bson:"enable_telemetry"`
	MaxCountProcessFiltration int8     `json:"max_count_process_filtration" bson:"max_count_process_filtration"`
	StorageFolders            []string `json:"storage_folders" bson:"storage_folders"`
	TypeAreaNetwork           string   `json:"type_area_network" bson:"type_area_network"`
	IfAsServerThenPort        int      `json:"if_as_server_then_port" bson:"if_as_server_then_port"`
}

//InformationSourcesList информация об источниках
type InformationSourcesList struct {
	sourcesListSetting
	sourcesListConnection
	chanReq chan chanReqSetting
}

type chanReqSetting struct {
	actionType string
	id         int
	setting    sourceSetting
	link       *websocket.Conn
	chanRes    chan chanResSetting
}

type chanResSetting struct {
	err                   error
	id                    int
	setting               *sourceSetting
	additionalInformation interface{}
}

type sourceConnectDisconnectLists struct {
	listConnected, listDisconnected map[int]string
}

//sourcesListConnection дескрипторы соединения с источниками по протоколу websocket
type sourcesListConnection map[string]WssConnection

//WssConnection дескриптор соединения по протоколу websocket
type WssConnection struct {
	Link *websocket.Conn
	//mu   sync.Mutex
}

//NewRepositoryISL инициализация хранилища
func NewRepositoryISL() *InformationSourcesList {
	isl := InformationSourcesList{
		sourcesListSetting:    sourcesListSetting{},
		sourcesListConnection: sourcesListConnection{},
		chanReq:               make(chan chanReqSetting),
	}

	go func() {
		for msg := range isl.chanReq {
			switch msg.actionType {
			case "add source settings":
				isl.sourcesListSetting[msg.id] = msg.setting

				msg.chanRes <- chanResSetting{}
				//close(msg.chanRes)

			case "del info about source":
				delete(isl.sourcesListSetting, msg.id)

				msg.chanRes <- chanResSetting{}
				//close(msg.chanRes)

			case "get source list":
				sl := make(map[int]sourceSetting, len(isl.sourcesListSetting))

				for id, ss := range isl.sourcesListSetting {
					sl[id] = ss
				}

				msg.chanRes <- chanResSetting{additionalInformation: sl}
				//close(msg.chanRes)

			case "get source setting by id":
				si, ok := isl.sourcesListSetting[msg.id]
				if ok {
					msg.chanRes <- chanResSetting{
						setting: &si,
					}

					break
				}

				msg.chanRes <- chanResSetting{
					err: fmt.Errorf("source with ID %v not found", msg.id),
				}

				//close(msg.chanRes)

			case "get source connection status":
				s, ok := isl.sourcesListSetting[msg.id]
				if !ok {
					msg.chanRes <- chanResSetting{
						err: fmt.Errorf("source with ID %v not found", msg.id),
					}

					//close(msg.chanRes)

					break
				}

				msg.chanRes <- chanResSetting{setting: &sourceSetting{ConnectionStatus: s.ConnectionStatus}}
				//close(msg.chanRes)

			case "change source connection status":
				s, ok := isl.sourcesListSetting[msg.id]
				if !ok {
					msg.chanRes <- chanResSetting{
						err: fmt.Errorf("source with ID %v not found", msg.id),
					}

					//close(msg.chanRes)

					break
				}
				s.ConnectionStatus = msg.setting.ConnectionStatus

				if msg.setting.ConnectionStatus {
					s.DateLastConnected = time.Now().Unix()
				} else {
					s.AccessIsAllowed = false
				}

				isl.sourcesListSetting[msg.id] = s

				msg.chanRes <- chanResSetting{}
				//close(msg.chanRes)
			}
		}
	}()

	return &isl
}

//AddSourceSettings добавляет настройки источника
func (isl *InformationSourcesList) AddSourceSettings(id int, settings sourceSetting) {

	fmt.Println("func 'AddSourceSettings', START...")

	chanRes := make(chan chanResSetting)
	defer close(chanRes)

	isl.chanReq <- chanReqSetting{
		actionType: "add source settings",
		id:         id,
		setting:    settings,
		chanRes:    chanRes,
	}

	<-chanRes

	fmt.Println("func 'AddSourceSettings', received from chan")
}

//GetSourceList возвращает список источников
func (isl *InformationSourcesList) GetSourceList() *map[int]sourceSetting {

	fmt.Println("func 'GetSourceList', START...")

	chanRes := make(chan chanResSetting)
	defer close(chanRes)

	isl.chanReq <- chanReqSetting{
		actionType: "get source list",
		chanRes:    chanRes,
	}

	if sl, ok := (<-chanRes).additionalInformation.(map[int]sourceSetting); ok {

		fmt.Println("func 'GetSourceList', received from channel 111")

		return &sl
	}

	fmt.Println("func 'GetSourceList', received from channel 222")

	return &map[int]sourceSetting{}
}

//ChangeSourceConnectionStatus изменяет состояние источника
func (isl *InformationSourcesList) ChangeSourceConnectionStatus(id int, status bool) bool {

	fmt.Printf("func 'ChangeSourceConnectionStatus', received from channel, STATUS: '%v'\n", status)

	chanRes := make(chan chanResSetting)
	defer close(chanRes)

	isl.chanReq <- chanReqSetting{
		actionType: "change source connection status",
		id:         id,
		setting:    sourceSetting{ConnectionStatus: status},
		chanRes:    chanRes,
	}

	if (<-chanRes).err != nil {

		fmt.Println("func 'ChangeSourceConnectionStatus', received from channel 111")

		return false
	}

	fmt.Println("func 'ChangeSourceConnectionStatus', received from channel 222")

	return true
}

//GetSourceConnectionStatus возвращает состояние соединения с источником
func (isl *InformationSourcesList) GetSourceConnectionStatus(id int) (bool, error) {

	fmt.Println("func 'GetSourceConnectionStatus', START...")

	chanRes := make(chan chanResSetting)
	defer close(chanRes)

	isl.chanReq <- chanReqSetting{
		actionType: "get source connection status",
		id:         id,
		chanRes:    chanRes,
	}

	resMsg := <-chanRes

	fmt.Println("func 'GetSourceConnectionStatus', received from channel")

	return resMsg.setting.ConnectionStatus, resMsg.err
}

//GetSourceSetting возвращает все настройки источника по его ID
func (isl *InformationSourcesList) GetSourceSetting(id int) (*sourceSetting, bool) {
	chanRes := make(chan chanResSetting)
	defer close(chanRes)

	isl.chanReq <- chanReqSetting{
		actionType: "get source setting by id",
		id:         id,
		chanRes:    chanRes,
	}

	resMsg := <-chanRes

	if resMsg.err != nil {
		return nil, false
	}

	return resMsg.setting, true
}

func testNextTickAndCycle() bool {

	fmt.Println("func 'testNextTickAndCycle', START...")

	list := make([]int, 15, 15)
	isCycleProcessing := "non-blocking"

	var count, countGoroutine int
	chanDone := make(chan struct{})

	handlerRequest := func(chanDone chan<- struct{}, key int) {
		fmt.Printf("func 'handlerRequest', START for key '%v'...\n", key)

		chanDone <- struct{}{}
	}

	cycleProcessing := func(icp *string, l *[]int) {
		fmt.Println("func 'isCycleProcessing', START")

		if len(*l) == 0 {

			fmt.Println("func 'isCycleProcessing', STOP, list = 0")

			return
		}

		isCycleProcessing = "blocking"

		//time.Sleep(2000 * time.Millisecond)

		var i int

		for k := range *l {
			if k < 3 {
				fmt.Printf("func 'isCycleProcessing', key: '%v'\n", k)
			}

			i = i * k

			go handlerRequest(chanDone, k)

			countGoroutine++
		}

		for {
			<-chanDone

			count++

			if count == countGoroutine {
				break
			}
		}

		fmt.Println("func 'isCycleProcessing', STOP")

		isCycleProcessing = "non-blocking"
	}

	num := 0

	ticker := time.NewTicker(time.Duration(1) * time.Second)
	for range ticker.C {
		if isCycleProcessing == "blocking" {

			fmt.Printf("ticker STOP, ---- num: %v\n", num)

			continue
		}

		if num == 3 {
			break
		}

		go cycleProcessing(&isCycleProcessing, &list)

		num++
	}

	return true
}

var _ = Describe("Testcycle", func() {

	nrISL := NewRepositoryISL()
	nrISL.AddSourceSettings(111, sourceSetting{
		IP:        "45.10.23.6",
		ShortName: "source test name",
		Token:     "djn3h8fh8hh84gt",
	})

	Context("Тест 1: Проверка размера списка полученного в результате работы цикла", func() {
		It("Количество должно быть равно 10", func() {
			list := createList()

			Expect(len(list)).Should(Equal(10))
		})
	})

	Context("Тест 2: Проверка выхода из цикла", func() {
		It("Не должно быть ошибки", func() {
			result := cycleForAndSwith(createList())

			Expect(result).To(BeNil())
			//			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 3: Попытка закрыть канал", func() {
		It("Изменяем статус соединения на ПОДКЛЮЧЕН, ошибки быть не должно", func() {
			ok := nrISL.ChangeSourceConnectionStatus(111, true)

			isConn, err := nrISL.GetSourceConnectionStatus(111)

			Expect(ok).Should(BeTrue())
			Expect(isConn).Should(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Изменяем статус соединения на НЕ ПОДКЛЧЕН, ошибки быть не должно", func() {
			ok := nrISL.ChangeSourceConnectionStatus(111, false)

			isConn, err := nrISL.GetSourceConnectionStatus(111)

			Expect(ok).Should(BeTrue())
			Expect(isConn).ShouldNot(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Получаем информацию об источнике, должно быть полученно заданное значение", func() {
			ss, ok := nrISL.GetSourceSetting(111)

			Expect(ok).Should(BeTrue())
			Expect(ss.IP).Should(Equal("45.10.23.6"))
		})
	})

	Context("Тест 4: проверяем цикл с блокировкой", func() {
		It("have to all OK", func() {
			Expect(testNextTickAndCycle()).Should(BeTrue())
		})
	})
})
