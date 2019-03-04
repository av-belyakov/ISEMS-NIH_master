package configure

//ChannelCollectionCoreApp коллекция каналов для coreAppRoute
type ChannelCollectionCoreApp struct {
	OutCoreChanDB, InCoreChanDB   chan MsgBetweenCoreAndDB
	OutCoreChanAPI, InCoreChanAPI chan MsgBetweenCoreAndAPI
	OutCoreChanNI, InCoreChanNI   chan MsgBetweenCoreAndNI
}

//MsgWsTransmission содержит информацию для передачи подключенному источнику
type MsgWsTransmission struct {
	DestinationHost string
	Data            []byte
}

// MsgBetweenCoreAndNI используется для взаимодействия между ядром приложения и модулем сет. взаимодействия
// по каналам с данной структурой передаются следующие виды сообщений:
// MsgType:
//  - information
//  - command
// DataType:
//  - change_status_source (info)
//  - source_telemetry (info)
//  - filtration (info)
//  - download (info)
//  - information_search_results (info)
//  - error_notification (info)
//  - source_control (command)
//  - filtration (command)
//  - download (command)
//  - information_search (command)
// IDClientAPI - уникальный идентификатор клиента API
//
//1. информационные (information):
//  - изменение статуса источника TypeRequiredAction (change_status_source)
//  - получение данных телеметрии об источнике TypeRequiredAction (source_telemetry)
//  - ход выполнения ФИЛЬТРАЦИИ TypeRequiredAction (filtration)
//  - ход выполнения СКАЧИВАНИЯ файлов TypeRequiredAction (download)
//  - информация об ошибках соединения TypeRequiredAction (error_message)
// 2. команда (instraction):
//  - управление источником TypeRequiredAction(add_source, delete_source, change_setting_source)
//  - управление фильтрацией сет. трафика TypeRequiredAction (filtration_start, filtration_stop)
//  - управление скачиванием файлов TypeRequiredAction (download_start, download_stop, download_resume)
type MsgBetweenCoreAndNI struct {
	SourceID        string
	MsgType         string
	DataType        string
	IDClientAPI     string
	AdvancedOptions interface{}

	TypeRequiredAction string
	Data               []byte
}

//MsgBetweenCoreAndAPI используется для взаимодействия между ядром приложения и модулем API приложения
//по каналам с данной структурой передаются следующие виды сообщений:
// MsgGenerator - API module/Core module
// MsgType:
//  - information
//  - command
// DataType:
//  - change_status_source (info)
//  - source_telemetry (info)
//  - filtration (info)
//  - download (info)
//  - information_search_results (info)
//  - error_notification (info)
//  - source_control (command)
//  - filtration (command)
//  - download (command)
//  - information_search (command)
// IDClientAPI - уникальный идентификатор клиента API
type MsgBetweenCoreAndAPI struct {
	MsgGenerator    string
	MsgType         string
	DataType        string
	IDClientAPI     string
	AdvancedOptions interface{}
}

//MsgBetweenCoreAndDB используется для взаимодействия между ядром и модулем взаимодействия с БД
//по каналам с данной структурой передаются следующие виды сообщений:
// MsgGenerator - DB module/Core module (источник сообщения)
// MsgRecipient - API module/NI module/DB module/Core module (получатель сообщения)
// MsgDirection - request/response (направление, запрос или ответ)
// DataType:
//  - sources_list
//  - change_status_source
//  - source_telemetry
//  - filtration
//  - download
//  - information_search_results
//  - error_notification
//  - source_control
//  - information_search
// IDClientAPI - уникальный идентификатор клиента API
type MsgBetweenCoreAndDB struct {
	MsgGenerator    string
	MsgRecipient    string
	MsgDirection    string
	DataType        string
	IDClientAPI     string
	AdvancedOptions interface{}
}

//AdvancedOptions взависимости от типов сообщений

//ErrorNotification содержит информацию об ошибке
// SourceReport - DB module/NI module/API module
// ErrorBody - тело ошибки (stack trace)
// HumanDescriptionError - сообщение для пользователя
type ErrorNotification struct {
	SourceReport, HumanDescriptionError string
	ErrorBody                           error
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
