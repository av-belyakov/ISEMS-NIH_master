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
// 1. информационные (information):
//  - изменение статуса источника TypeRequiredAction (change_status_source)
//  - получение данных телеметрии об источнике TypeRequiredAction (source_telemetry)
//  - ход выполнения ФИЛЬТРАЦИИ TypeRequiredAction (filtering)
//  - ход выполнения СКАЧИВАНИЯ файлов TypeRequiredAction (download)
//  - информация об ошибках соединения TypeRequiredAction (error_message)
// 2. команда (instraction):
//  - управление источником TypeRequiredAction(add_source, delete_source, change_setting_source)
//  - управление фильтрацией сет. трафика TypeRequiredAction (filtering_start, filtering_stop)
//  - управление скачиванием файлов TypeRequiredAction (download_start, download_stop, download_resume)
type MsgBetweenCoreAndNI struct {
	SourceID           string //ID источника с котором связанно действие
	MsgType            string //команда/информационное
	TypeRequiredAction string
	Data               []byte
}

//MsgBetweenCoreAndAPI используется для взаимодействия между ядром приложения и модулем API приложения
//по каналам с данной структурой передаются следующие виды сообщений:
// MsgGenerator - from API/from Core
// MsgType:
//  - information
//  - command
// DataType:
//  - change_status_source (info)
//  - source_telemetry (info)
//  - filtering (info)
//  - download (info)
//  - information_search_results (info)
//  - error_notification (info)
//  - source_control (command)
//  - filtering (command)
//  - download (command)
//  - information_search (command)
type MsgBetweenCoreAndAPI struct {
	MsgGenerator    string
	MsgType         string
	DataType        string
	IDClientAPI     string
	AdvancedOptions interface{}
}

//MsgBetweenCoreAndDB используется для взаимодействия между ядром и модулем взаимодействия с БД
//по каналам с данной структурой передаются следующие виды сообщений:
//
type MsgBetweenCoreAndDB struct {
	MsgID, CollectionName string
	Date                  interface{}
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
/*Message interface{} //сообщение в любом виде, используется контролируемое приведение типа
}
*/
