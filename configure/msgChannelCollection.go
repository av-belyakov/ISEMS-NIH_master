package configure

//ChannelCollectionCoreApp коллекция каналов для coreAppRoute
type ChannelCollectionCoreApp struct {
	OutCoreChanDB, InCoreChanDB   chan *MsgBetweenCoreAndDB
	OutCoreChanAPI, InCoreChanAPI chan *MsgBetweenCoreAndAPI
	OutCoreChanNI, InCoreChanNI   chan *MsgBetweenCoreAndNI
}

//MsgWsTransmission содержит информацию для передачи подключенному источнику
type MsgWsTransmission struct {
	DestinationHost string
	Data            []byte
}

// MsgBetweenCoreAndNI используется для взаимодействия между ядром приложения и модулем сет. взаимодействия
// TaskID - ID задачи
// ClientName - имя клиента, используется для управления источниками
// Section:
//  - 'source control'
//  - 'filtration control'
//  - 'download control'
//  - 'error notification'
//  - 'message notification'
// Command:
//  - для source_control:
//		* 'load list'
// 		* 'update list' (для каждого источника свое возможно действие например,
//			один источник нужно добавить 'add', другой удалить 'del', а третий
// 			перезапустить 'reconnect')
// 		* 'keep list sources in database' (сохранить список источников в БД)
// 		* 'send list sources to client api'
//  - для filtration_control:
//		* 'start'
//		* 'stop'
//		* ''
//  - для download_control:
//		* 'start'
//		* 'stop'
//  - 'error notification'
//      * 'send client API'
//      * 'no send client API'
//  - 'message notification'
//      * 'send client API'
//      * 'no send client API'
type MsgBetweenCoreAndNI struct {
	TaskID          string
	ClientName      string
	Section         string
	Command         string
	AdvancedOptions interface{}
}

//MsgBetweenCoreAndAPI используется для взаимодействия между ядром приложения и модулем API приложения
//по каналам с данной структурой передаются следующие виды сообщений:
// MsgGenerator - API module/Core module
// MsgType:
//  - information
//  - command
// MsgSection:
//  - 'source control'
//  - 'filtration control'
//  - 'download control'
//  - 'information search control'
//  - 'source telemetry info'
//  - 'error notification'
//  - 'message notification'
// IDClientAPI - уникальный идентификатор клиента API
type MsgBetweenCoreAndAPI struct {
	MsgGenerator string
	MsgRecipient string
	IDClientAPI  string
	ClientName   string
	ClientIP     string
	MsgJSON      interface{}
}

//MsgBetweenCoreAndDB используется для взаимодействия между ядром и модулем взаимодействия с БД
//по каналам с данной структурой передаются следующие виды сообщений:
// MsgGenerator - DB module/Core module (источник сообщения)
// MsgRecipient - API module/NI module/DB module/Core module (получатель сообщения)
// MsgSection:
//  - 'source control'
//  - 'source telemetry'
//  - 'filtration'
//  - 'download'
//  - 'information search results'
//  - 'error notification'
//  - 'message notification'
//  - 'information search'
// Insturction:
//  - insert
//  - find
//  - find_all
//  - update
//  - delete
// IDClientAPI - уникальный идентификатор клиента API
// TaskID - уникальный идентификатор задачи присвоенный ядром приложения
type MsgBetweenCoreAndDB struct {
	MsgGenerator    string
	MsgRecipient    string
	MsgSection      string
	Instruction     string
	IDClientAPI     string
	TaskID          string
	AdvancedOptions interface{}
}

// --- AdvancedOptions взависимости от типов сообщений ---

//ErrorNotification содержит информацию об ошибке
// SourceReport - DB module/NI module/API module
// ErrorBody - тело ошибки (stack trace)
// HumanDescriptionError - сообщение для пользователя
// Sources - срез ID источников связанных с данным сообщением
type ErrorNotification struct {
	SourceReport, HumanDescriptionError string
	ErrorBody                           error
	Sources                             []int
}

//MessageNotification содержит информационное сообщение о выполненном действии
// SourceReport - DB module/NI module/API module
// Section - раздел к которому относится действие
// TypeActionPerformed - тип выполненного действия
// CriticalityMessage - критичность сообщения ('info', 'success', 'warning', 'danger')
// HumanDescriptionError - сообщение для пользователя
// Sources - срез ID источников связанных с данным сообщением
type MessageNotification struct {
	SourceReport                 string
	Section                      string
	TypeActionPerformed          string
	CriticalityMessage           string
	HumanDescriptionNotification string
	Sources                      []int
}

/* РАССМОТРЕТЬ ПОЗЖЕ */

//--- ВЗАИМОДЕЙСТВИЕ С МОДУЛЕМ NetworkInteraction ---
//		(СЕРВИСНЫЕ СООБЩЕНИЯ)
/*

//MessageTypeInfoStatusSource статус сенсора (ИНФОРМАЦИОННОЕ)
type MessageTypeInfoStatusSource struct {
	IP               string //ip адрес источника
	ConnectionStatus string //connect/disconnet
	ConnectionTime   int64  //Unix time
}

//ServiceMessageInfoStatusSource сервисное сообщение о статусе источников
type ServiceMessageInfoStatusSource struct {
	Type       string //change_sources
	SourceList []MessageTypeInfoStatusSource
}


//--- ВЗАИМОДЕЙСТВИЕ С МОДУЛЕМ NetworkInteraction ---
//		(СООБЩЕНИЯ ОБЩЕГО НАЗНАЧЕНИЯ)


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

//MessageTypeInfofiltration ход выполнения фильтрации (ИНФОРМАЦИОННОЕ)
type MessageTypeInfofiltration struct {
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

//MessageTypeInfofiltrationSearchListFiles список файлов найденных в результате фильтрации (ИНФОРМАЦИОННОЕ)
type MessageTypeInfofiltrationSearchListFiles struct {
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

//MessageTypeCommandfiltration команды связанные с задачами по фильтрации (ИСПОЛНИТЕЛЬНЫЕ)
type MessageTypeCommandfiltration struct {
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
	SubType string //подтип сообщения, если команда filtration/download,
	/*
	   если информационное
	   - change_sensor_status
	   - info_filtration
	   - info_filtration_search_list_files
	   - info_download
	   - error_message
*/
/*Message interface{} //сообщение в любом виде, используется контролируемое приведение типа
}
*/
