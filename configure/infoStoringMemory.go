package configure

/*
* Описание типа для хранения в памяти часто используемых параметров
*
* Версия 0.11, дата релиза 21.02.2019
* */

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mongodb/mongo-go-driver/mongo"
)

/*
--- ВЗАИМОДЕЙСТВИЕ С МОДУЛЕМ API ---
*/

//MessageAPI набор параметров для взаимодействия с модулем API
type MessageAPI struct {
	MsgID, MsgType string
	MsgDate        int
}

/*
--- ВЗАИМОДЕЙСТВИЕ С МОДУЛЕМ NetworkInteraction ---
		(СЕРВИСНЫЕ СООБЩЕНИЯ)
*/

//MessageTypeInfoStatusSource статус сенсора (ИНФОРМАЦИОННОЕ)
type MessageTypeInfoStatusSource struct {
	SourceID         string //идентификатор источника в виде текста
	ConnectionStatus string //connect/disconnet
	ConnectionTime   int64  //Unix time
}

//ServiceMessageInfoStatusSource сервисное сообщение о статусе источников
type ServiceMessageInfoStatusSource struct {
	Type       string //get_list/change_list/send_list
	SourceList []MessageTypeInfoStatusSource
}

/*
--- ВЗАИМОДЕЙСТВИЕ С МОДУЛЕМ NetworkInteraction ---
		(СООБЩЕНИЯ ОБЩЕГО НАЗНАЧЕНИЯ)
*/

//Memory содержит информацию об используемой ПО
type telemetryMemory struct {
	Total int
	Used  int
	Free  int
}

//MessageTelemetryData информация о технических параметрах источника
type MessageTelemetryData struct {
	IPAddress          string
	CurrentDateTime    int64
	DiskSpace          []map[string]string
	TimeInterval       map[string]map[string]int
	RandomAccessMemory telemetryMemory
	LoadCPU            float64
	LoadNetwork        map[string]map[string]int
}

//informationFilterProcess информация об обработанных файлах
type informationFilterProcess struct {
	AllCountFilesSearch  int //количество найденных файлов
	AllSizeFilesSearch   int //общий размер найденных файлов
	CountFilesProcessing int //количество обработанных файлов
	CountFilesSearch     int //количество файлов найданных под заданные параметры
	SizeFilesSearch      int //общий размер файлов удовлетворяющих заданным параметрам
}

//MessageTypeInfoFiltering ход выполнения фильтрации (ИНФОРМАЦИОННОЕ)
type MessageTypeInfoFiltering struct {
	TaskIndex     string //ID задачи
	ProcessStatus string //статус процесса ready, start, execute, complete, stop
	Info          informationFilterProcess
}

//fileInformation информация о найденном файле
type fileInformation struct {
	FileHash       string //хеш сумма
	FileSize       int    //размер
	FileCreateTime int    //дата создания в формате Unix
}

//MessageTypeInfoFilteringSearchListFiles список файлов найденных в результате фильтрации (ИНФОРМАЦИОННОЕ)
type MessageTypeInfoFilteringSearchListFiles struct {
	TaskIndex string                     //ID задачи
	ListFiles map[string]fileInformation //ключ имя файла
}

//MessageTypeInfoDownload ход выполнения скачивания файлов (ИНФОРМАЦИОННОЕ)
type MessageTypeInfoDownload struct {
	TaskIndex     string //ID задачи
	ProcessStatus string //статус процесса ready, execute, execute_complited
	InfoFile      fileInformation
}

//MessageTypeErrorMessage информация об ошибках
type MessageTypeErrorMessage struct {
	LocationError string //sensor/module_ni
	MsgError      error
	Description   string //возможно дополнительное описание
}

//параметры фильтрации файлов
type parametersFilterFiles struct {
	DateTimeStart, DateTimeEnd int
	ListIP, ListNetwork        string
}

//MessageTypeCommandFiltering команды связанные с задачами по фильтрации (ИСПОЛНИТЕЛЬНЫЕ)
type MessageTypeCommandFiltering struct {
	TaskIndex     string //ID задачи
	ProcessStatus string //статус процесса start, stop
	Info          parametersFilterFiles
}

//parametersDownloadFiles информация о скачиваемых файлов
type parametersDownloadFiles struct {
	DownloadDirectoryFiles     string   //директория где хранятся файлы на сенсоре
	DownloadSelectedFiles      bool     //скачивать выбранные файлы или все
	CountDownloadSelectedFiles int      //общее число файлов
	NumberMessageParts         [2]int   //число частей сообщения
	ListDownloadSelectedFiles  []string //список файлов выбранных для скачивания
}

//MessageTypeCommandStartDownload сообщение при старте задачи
type MessageTypeCommandStartDownload struct {
	TaskIndex     string //ID задачи
	ProcessStatus string //статус процесса start
	Info          parametersDownloadFiles
}

//MessageTypeCommandDownload команды связанные с задачами по скачиванию файлов (ИСПОЛНИТЕЛЬНЫЕ)
type MessageTypeCommandDownload struct {
	TaskIndex     string //ID задачи
	ProcessStatus string //статус процесса ready, waiting_for_transfer, execute_success'/'execute_failure
}

//MessageNetworkInteraction набор параметров для взаимодействия с модулем Network Interaction
type MessageNetworkInteraction struct {
	MsgType string //тип сообщения (команда/информационное)
	SubType string //подтип сообщения, если команда filtering/download,
	/*
	   если информационное
	   - change_sensor_status
	   - info_filtering
	   - info_filtering_search_list_files
	   - info_download
	   - error_message
	*/
	Message interface{} //сообщение в любом виде, используется контролируемое приведение типа
}

/*
--- НАБОР КАНАЛОВ ---
*/

//channelCollection набор каналов
type channelCollection struct {
	ChannelToModuleAPI    chan MessageAPI
	ChannelFromModuleAPI  chan MessageAPI
	ChannelToMNICommon    chan MessageNetworkInteraction
	ChannelFromMNICommon  chan MessageNetworkInteraction
	ChannelToMNIService   chan ServiceMessageInfoStatusSource
	ChannelFromMNIService chan ServiceMessageInfoStatusSource
}

//ChanReguestDatabase содержит запросы для модуля обеспечивающего доступ к БД
type ChanReguestDatabase struct {
}

//ChanResponseDatabase содержит ответы от модуля обеспечивающего доступ к БД
type ChanResponseDatabase struct {
}

/*
--- ДОЛГОВРЕМЕННОЕ ХРАНЕНИЕ ВРЕМЕННЫХ ФАЙЛОВ ---
*/

//ParametersSource описание состояния источника
type ParametersSource struct {
	ConnectionStatus  bool //true/false
	ID                string
	DateLastConnected int64 //Unix time
	Token             string
	AccessIsAllowed   bool              //разрешен ли доступ, по умолчанию false (при проверке токена ставится true если он верен)
	ToServer          bool              //false - как клиент, true - как сервер
	CurrentTasks      map[string]string // задачи для данного источника,
	//key - ID задачи, value - ее тип 'in queuq' или 'in process'
	LinkWsConnection *websocket.Conn
}

//SourcesList список источников
type SourcesList map[string]ParametersSource

//InformationStoringMemory часто используемые параметры
type InformationStoringMemory struct {
	SourcesList       //key - ip источника в виде строки
	ChannelCollection channelCollection
}

//SearchSourceToken поиск id источника по его токену и ip
func (ism *InformationStoringMemory) SearchSourceToken(host, token string) (string, bool) {
	for sourceIP, settings := range ism.SourcesList {
		if sourceIP == host && settings.Token == token {
			//разрешаем соединение с данным источником
			settings := ism.SourcesList[host]
			settings.AccessIsAllowed = true
			ism.SourcesList[host] = settings

			return settings.ID, true
		}
	}

	return "", false
}

//GetSourceSetting получить все настройки источника по его ip
func (ism *InformationStoringMemory) GetSourceSetting(host string) (ParametersSource, bool) {
	for ip, settings := range ism.SourcesList {
		if ip == host {
			return settings, true
		}
	}

	return ParametersSource{}, false
}

//ChangeSourceConnectionStatus изменить состояние источника
func (ism *InformationStoringMemory) ChangeSourceConnectionStatus(host string) bool {
	if _, isExist := ism.SourcesList[host]; isExist {
		sourceSetting := ism.SourcesList[host]
		sourceSetting.ConnectionStatus = !sourceSetting.ConnectionStatus
		if !sourceSetting.ConnectionStatus {
			//статус источника = false (тоесть disconnect) удаляем линк соединения по websocket
			sourceSetting.LinkWsConnection = nil
			sourceSetting.AccessIsAllowed = false
		} else {
			//статус источника = true, добавить время последнего соединения
			sourceSetting.DateLastConnected = time.Now().Unix()
		}

		ism.SourcesList[host] = sourceSetting

		return true
	}

	return false
}

//AddLinkWebsocketConnect добавить линк соединения по websocket
func (ism *InformationStoringMemory) AddLinkWebsocketConnect(host string, lwsc *websocket.Conn) {
	if _, isExist := ism.SourcesList[host]; isExist {
		sourceSetting := ism.SourcesList[host]
		sourceSetting.LinkWsConnection = lwsc
		ism.SourcesList[host] = sourceSetting
	}
}

//GetLinkWebsocketConnect получить линк соединения по websocket
func (ism *InformationStoringMemory) GetLinkWebsocketConnect(host string) (*websocket.Conn, bool) {
	if _, isExist := ism.SourcesList[host]; isExist {
		return ism.SourcesList[host].LinkWsConnection, true
	}

	return nil, false
}

//GetAccessIsAllowed возвращает значение подтверждающее или откланяющее права доступа источника
func (ism *InformationStoringMemory) GetAccessIsAllowed(host string) bool {
	if _, isExist := ism.SourcesList[host]; isExist {
		return ism.SourcesList[host].AccessIsAllowed
	}

	return false
}

//MongoDBConnect содержит дискриптор соединения с БД
type MongoDBConnect struct {
	Connect *mongo.Client
	CTX     context.Context
}
