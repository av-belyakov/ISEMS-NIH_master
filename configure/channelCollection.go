package configure

//ChannelCollection коллекция внутренних каналов передачи данных
type ChannelCollection struct {
	ChannelToModuleAPI    chan MessageAPI
	ChannelFromModuleAPI  chan MessageAPI
	ChannelToMNICommon    chan MessageNetworkInteraction
	ChannelFromMNICommon  chan MessageNetworkInteraction
	ChannelToMNIService   chan ServiceMessageInfoStatusSource
	ChannelFromMNIService chan ServiceMessageInfoStatusSource
	Cwt                   chan MsgWsTransmission

	//ChannelChangeSourceList chan struct{} //информирует об изменении списка источников
}

//MsgWsTransmission содержит информацию для передачи подключенному источнику
type MsgWsTransmission struct {
	DestinationHost string
	Data            []byte
}

//ChanReguestDatabase содержит запросы для модуля обеспечивающего доступ к БД
type ChanReguestDatabase struct {
}

//ChanResponseDatabase содержит ответы от модуля обеспечивающего доступ к БД
type ChanResponseDatabase struct {
}

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
	IP               string //ip адрес источника
	ConnectionStatus string //connect/disconnet
	ConnectionTime   int64  //Unix time
}

//ServiceMessageInfoStatusSource сервисное сообщение о статусе источников
type ServiceMessageInfoStatusSource struct {
	Type       string //change_sources
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
