package configure

/*
* Описание типов коллекций хранящихся в БД
* */

//InformationAboutSource информация об источнике в коллекции 'sources_list'
type InformationAboutSource struct {
	ID            int                 `json:"id" bson:"id"`
	IP            string              `json:"ip" bson:"ip"`
	Token         string              `json:"token" bson:"token"`
	ShortName     string              `json:"short_name" bson:"short_name"`
	Description   string              `json:"description" bson:"description"`
	AsServer      bool                `json:"as_server" bson:"as_server"`
	NameClientAPI string              `json:"name_client_api" bson:"name_client_api"`
	SourceSetting InfoServiceSettings `json:"source_setting" bson:"source_setting"`
}

//InfoServiceSettings содержит настройки источника
type InfoServiceSettings struct {
	EnableTelemetry           bool     `json:"enable_telemetry" bson:"enable_telemetry"`
	MaxCountProcessFiltration int8     `json:"max_count_process_filtration" bson:"max_count_process_filtration"`
	StorageFolders            []string `json:"storage_folders" bson:"storage_folders"`
	TypeAreaNetwork           string   `json:"type_area_network" bson:"type_area_network"`
	IfAsServerThenPort        int      `json:"if_as_server_then_port" bson:"if_as_server_then_port"`
}

//InformationAboutTask подробная информация связанная с задачей по фильтрации
// TaskID - уникальный идентификатор задачи полученный от приложения
// ClientID - уникальный идентификатор клиента
// ClientTaskID - уникальный идентификатор задачи полученный от клиента
// GeneralInformationAboutTask - общая информация о задаче
// SourceID - идентификатор источника на котором выполняется задача
// UserInitiatedFilteringProcess - имя пользователя инициировавшего процесс фильтрации
// UserInitiatedFileDownloadProcess - имя пользователя инициировавшего процесс скачивания файлов
// FilteringOption - параметры фильтрации полученные от клиента
type InformationAboutTask struct {
	TaskID                           string                                 `json:"task_id" bson:"task_id"`
	ClientID                         string                                 `json:"client_id" bson:"client_id"`
	ClientTaskID                     string                                 `json:"client_task_id" bson:"client_task_id"`
	GeneralInformationAboutTask      DescriptionGeneralInformationAboutTask `json:"general_information_about_task" bson:"general_information_about_task"`
	SourceID                         int                                    `json:"source_id" bson:"source_id"`
	UserInitiatedFilteringProcess    string                                 `json:"user_initiated_filtering_process" bson:"user_initiated_filtering_process"`
	UserInitiatedFileDownloadProcess string                                 `json:"user_initiated_file_download_process" bson:"user_initiated_file_download_process"`
	FilteringOption                  FilteringOption                        `json:"filtering_option" bson:"filtering_option"`
	DetailedInformationOnFiltering   DetailedInformationFiltering           `json:"detailed_information_on_filtering" bson:"detailed_information_on_filtering"`
	DetailedInformationOnDownloading DetailedInformationDownloading         `json:"detailed_information_on_downloading" bson:"detailed_information_on_downloading"`
	ListFilesResultTaskExecution     []*FilesInformation                    `json:"list_files_result_task_execution" bson:"list_files_result_task_execution"`
}

//DescriptionGeneralInformationAboutTask описание общей информации о задаче
// TaskProcessed - была ли обработана задача (фильтрация файлов выполнена, какие либо файлы были загружены,
// клиент API пометил задачу как отработанную)
// DateTimeProcessed - дата и время завершения задачи
// ClientID - идентификатор клиента API
// DetailDescription - детальное описание
type DescriptionGeneralInformationAboutTask struct {
	TaskProcessed     bool                                         `json:"task_processed" bson:"task_processed"`
	DateTimeProcessed int64                                        `json:"date_time_processed" bson:"date_time_processed"`
	ClientID          string                                       `json:"client_id" bson:"client_id"`
	DetailDescription DetailDescriptionGeneralInformationAboutTask `json:"detail_description_general_information_about_task" bson:"detail_description_general_information_about_task"`
}

//DetailDescriptionGeneralInformationAboutTask необязательное детальное описание получаемое от клиента
// UserNameProcessed - имя пользователя который закрыл задачу (внутри клиента API)
// DescriptionProcessingResults - описание задачи или причины закрытия
type DetailDescriptionGeneralInformationAboutTask struct {
	UserNameProcessed            string `json:"user_name_processed" bson:"user_name_processed"`
	DescriptionProcessingResults string `json:"description_processing_results" bson:"description_processing_results"`
}

//FilteringOption опции фильтрации
type FilteringOption struct {
	DateTime TimeInterval         `json:"date_time_interval" bson:"date_time_interval"`
	Protocol string               `json:"protocol" bson:"protocol"`
	Filters  FilteringExpressions `json:"filters" bson:"filters"`
}

//TimeInterval временной интервал
type TimeInterval struct {
	Start int64 `json:"start" bson:"start"`
	End   int64 `json:"end" bson:"end"`
}

//DetailedInformationFiltering детальная информация о ходе фильтрации
// TaskStatus - состояние задачи 'not executed'/'wait'/'refused'/'execute'/'not fully completed'/'complete'
// TimeIntervalTaskExecution - временной интервал начало, окончание выполнения задачи
// WasIndexUsed - использовался ли данные по индексам для поиска файлов удовлетворяющих параметрам фильтрации
// NumberFilesMeetFilterParameters - кол-во файлов удовлетворяющих параметрам фильтрации
// NumberProcessedFiles - кол-во обработанных файлов
// NumberFilesFoundResultFiltering - кол-во найденных, в результате фильтрации, файлов
// NumberDirectoryFiltartion - кол-во директорий по которым выполняется фильтрация
// NumberErrorProcessedFiles - кол-во не обработанных файлов или файлов обработанных с ошибками
// SizeFilesMeetFilterParameters - общий размер файлов (в байтах) удовлетворяющих параметрам фильтрации
// SizeFilesFoundResultFiltering - общий размер найденных, в результате фильтрации, файлов (в байтах)
// PathDirectoryForFilteredFiles - путь к директории в которой хранятся отфильтрованные файлы
type DetailedInformationFiltering struct {
	TaskStatus                      string       `json:"task_status" bson:"task_status"`
	TimeIntervalTaskExecution       TimeInterval `json:"time_interval_task_execution" bson:"time_interval_task_execution"`
	WasIndexUsed                    bool         `json:"was_index_used" bson:"was_index_used"`
	NumberFilesMeetFilterParameters int          `json:"number_files_meet_filter_parameters" bson:"number_files_meet_filter_parameters"`
	NumberProcessedFiles            int          `json:"number_processed_files" bson:"number_processed_files"`
	NumberFilesFoundResultFiltering int          `json:"number_files_found_result_filtering" bson:"number_files_found_result_filtering"`
	NumberDirectoryFiltartion       int          `json:"number_directory_filtartion" bson:"number_directory_filtartion"`
	NumberErrorProcessedFiles       int          `json:"number_error_processed_files" bson:"number_error_processed_files"`
	SizeFilesMeetFilterParameters   int64        `json:"size_files_meet_filter_parameters" bson:"size_files_meet_filter_parameters"`
	SizeFilesFoundResultFiltering   int64        `json:"size_files_found_result_filtering" bson:"size_files_found_result_filtering"`
	PathDirectoryForFilteredFiles   string       `json:"path_directory_for_filtered_files" bson:"path_directory_for_filtered_files"`
}

//DetailedInformationDownloading детальная информация о ходе скачивания файлов
// TaskStatus - состояние задачи
// TimeIntervalTaskExecution - временной интервал начало, окончание выполнения задачи
// NumberFilesTotal - общее количество файлов подлежащих скачиванию
// NumberFilesDownloaded - количество уже загруженных файлов
// NumberFilesDownloadedError - количество файлов загруженных с ошибкой
// PathDirectoryStorageDownloadedFiles - путь до директории долговременного хранения скаченных файлов
type DetailedInformationDownloading struct {
	TaskStatus                          string       `json:"task_status" bson:"task_status"`
	TimeIntervalTaskExecution           TimeInterval `json:"time_interval_task_execution" bson:"time_interval_task_execution"`
	NumberFilesTotal                    int          `json:"number_files_total" bson:"number_files_total"`
	NumberFilesDownloaded               int          `json:"number_files_downloaded" bson:"number_files_downloaded"`
	NumberFilesDownloadedError          int          `json:"number_files_downloaded_error" bson:"number_files_downloaded_error"`
	PathDirectoryStorageDownloadedFiles string       `json:"path_directory_storage_downloaded_files" bson:"path_directory_storage_downloaded_files"`
}

//FilteringExpressions выражения используемые для фильтрации
type FilteringExpressions struct {
	IP      FilteringNetworkParameters `json:"ip" bson:"ip"`
	Port    FilteringNetworkParameters `json:"port" bson:"port"`
	Network FilteringNetworkParameters `json:"network" bson:"network"`
}

//FilteringNetworkParameters сетевые параметры фильтрации
type FilteringNetworkParameters struct {
	Any []string `json:"any" bson:"any"`
	Src []string `json:"src" bson:"src"`
	Dst []string `json:"dst" bson:"dst"`
}

//FilesInformation информация по файлам найденным в результате фильтрации
// FileName - имя файла
// FileSize - размер файла
// FileHex - хеш сумма файла
// FileLoaded - загружен ли файл
type FilesInformation struct {
	FileName   string `json:"file_name" bson:"file_name"`
	FileSize   int64  `json:"file_size" bson:"file_size"`
	FileHex    string `json:"file_hex" bson:"file_hex"`
	FileLoaded bool   `json:"file_loaded" bson:"file_loaded"`
}

/* описание метаданных получаемых от Joy */

//GeneralDescriptionNetworkPacket информация по сетевым пакетам
// SourceID - идентификатор источника
// TaskID - идентификатор задачи (если есть)
// FileName - название файла
// FileCreationTime - время создания файла
// FileProcessingTime - время обработки файла
// NetworkParameters - общие сетевые параметры
type GeneralDescriptionNetworkPacket struct {
	SourceID           int
	TaskID             string
	FileName           string
	FileCreationTime   int64
	FileProcessingTime int64
	NetworkParameters  CommonNetworkParameters
}

//CommonNetworkParameters общие сетевые параметры
// SrcAddr - ip адрес источника (src address)
// DstAddr - ip адрес назначения (dsc address)
// Proto - тип протокола транспортного уровня
// SrcPort - сетевой порт источника (src port)
// DstPort - сетевой порт назначения (dsc port)
// BytesOut - кол-во отправленных байт
// NumPktsOut - кол-во отправленных пакетов
// BytesIn - кол-во принятых байт
// NumPktsIn - кол-во принятых пакетов
// TimeStart - начальное время пакета
// TimeEnd - конечное время пакета
// Packets - список с подробным описанием пакетов
// PacketsIP - детальное описание ip пакета
// PacketsHTTP - детальное описание HTTP пакета
// PacketsDNS - детальное описание DNS пакета
// PacketsSSH - детальное описание SSH пакета
// PacketsTLS - детальное описание TLS пакета
type CommonNetworkParameters struct {
	SrcAddr     string                       `json:"sa" bson:"sa"`
	DstAddr     string                       `json:"da" bson:"da"`
	Proto       int                          `json:"pr" bson:"pr"`
	SrcPort     int                          `json:"sp" bson:"sp"`
	DstPort     int                          `json:"dp" bson:"dp"`
	BytesOut    int                          `json:"bytes_out" bson:"bytes_out"`
	NumPktsOut  int                          `json:"num_pkts_out" bson:"num_pkts_out"`
	BytesIn     int                          `json:"bytes_in" bson:"bytes_in"`
	NumPktsIn   int                          `json:"num_pkts_in" bson:"num_pkts_in"`
	TimeStart   float32                      `json:"time_start" bson:"time_start"`
	TimeEnd     float32                      `json:"time_end" bson:"time_end"`
	Packets     []DetailedDescriptionPackets `json:"packets" bson:"packets"`
	PacketsIP   DetailedPacketsIP            `json:"ip" bson:"ip"`
	PacketsHTTP DetailedPacketsHTTP          `json:"http" bson:"http"`
	PacketsDNS  DetailedPacketsDNS           `json:"dns" bson:"dns"`
	PacketsSSH  DetailedPacketsSSH           `json:"ssh" bson:"ssh"`
	PacketsTLS  DetailedPacketsTLS           `json:"tls" bson:"tls"`
	//	PacketsDHCP  DetailedPacketsDHCP           `json:"" bson:""`
}

//DetailedDescriptionPackets подробное описание пакета
// Byte - размер в байтах
// Direction - направление передачи пакета
// InterPacketTimes - время прибытия между пакетами
type DetailedDescriptionPackets struct {
	Byte             int    `json:"b" bson:"b"`
	Direction        string `json:"dir" bson:"dir"`
	InterPacketTimes int    `json:"ipt" bson:"ipt"`
}

//DetailedPacketsIP делаьное описание ip пакета
// OutputPackets - описание исходящих ip пакетов
// InputPackets - описание входящих ip пакетов
type DetailedPacketsIP struct {
	OutputPackets DetailedOutputInputPackets `json:"out" bson:"ip_out"`
	InputPackets  DetailedOutputInputPackets `json:"in" bson:"ip_in"`
}

//DetailedOutputInputPackets детальное описание входящих и исходящих пакетов
// TTL - время жизни пакета
// ID - id пакета
type DetailedOutputInputPackets struct {
	TTL int   `json:"ttl" bson:"ttl"`
	ID  []int `json:"id" bson:"id"`
}

//DetailedPacketsHTTP детальное описание HTTP пакета
// OutputPacketsHTTP - описание исходящих HTTP пакетов
// InputPacketsHTTP - описание входящих HTTP пакетов
type DetailedPacketsHTTP struct {
	OutputPacketsHTTP DetailedOutputPacketsHTTP `json:"out" bson:"http_out"`
	InputPacketsHTTP  DetailedInputPacketsHTTP  `json:"in" bson:"http_in"`
}

//DetailedPacketsDNS детальное описание DNS пакета
// QueryName - запрашиваемое доменное имя (в запросе)
// ResponsedName - запрашиваемое доменное имя (в ответе)
// ResponseCode - код ответа
// ResourceRecord - ресурсная запись
type DetailedPacketsDNS struct {
	QueryName      string                         `json:"qn" bson:"qname"`
	ResponsedName  string                         `json:"rn" bson:"rname"`
	ResponseCode   int                            `json:"rc" bson:"rcode"`
	ResourceRecord []ResourceRecordDescriptionDNS `json:"rr" bson:"resrec"`
}

//ResourceRecordDescriptionDNS детальное описание ресурсной записи DNS протокола
// Address - ip address
// CanonicalName - каноническое имя
// Type - тип ресурсной записи
// Class - класс ресурсной записи
// RdLength - длинна
// TTL - время жизни
type ResourceRecordDescriptionDNS struct {
	Address       string `json:"a" bson:"a"`
	CanonicalName string `json:"cname" bson:"cname"`
	Type          int    `json:"type" bson:"type"`
	Class         string `json:"class" bson:"class"`
	RdLength      int    `json:"rdlength" bson:"rdlength"`
	TTL           int    `json:"ttl" bson:"ttl"`
}

/*
{"rn": {"type": "string",
                                            "title": "response name",
                                            "description": "DNS name"
                                           },

                                     "rc": {"type": "integer",
                                            "title": "return code.",
                                            "description": "The status code returned by the DNS server (0=no problem, other=error)."
                                           },

                                     "rr": {"type": "array",
                                            "title": "response record.",
                                            "description": "A DNS record.",
                                            "items": {"type": "object",
                                                      "title": "DNS record.",
                                                      "description": "A record returned by the DNS server.",
                                                      "properties": {"a": {"type": "string",
                                                                           "title": "address.",
                                                                           "description": "The IP address corresponding to the DNS name."
                                                                          },

                                                                     "ttl": {"type": "integer",
                                                                             "title": "TTL",
                                                                             "description": "Time To Live (TTL); the number of seconds that the address/name correspondence is to be considered valid."
                                                                            }
		                                                    }
                                                     }
                                           },

                                     "required": []
                                    }
*/

//DetailedPacketsSSH детальное описание SSH пакета
type DetailedPacketsSSH struct {
}

//DetailedPacketsTLS детальное описание TLS пакета
type DetailedPacketsTLS struct {
}

//DetailedOutputPacketsHTTP детальное описание изходящих HTTP сообщений
// Method - метод
// URI - уникальный идентификатор ресурса
// Version - версия HTTP протокола
// Host - имя хоста
// UserAgent
// Accept
// AcceptLanguage
// AcceptEncoding
// CacheControl
// Pragma
// Connection
// Body - первые 32 байта данных
type DetailedOutputPacketsHTTP struct {
	Method         string `json:"method" bson:"method"`
	URI            string `json:"uri" bson:"uri"`
	Version        string `json:"version" bson:"version"`
	Host           string `json:"Host" bson:"Host"`
	UserAgent      string `json:"User-Agent" bson:"User-Agent"`
	Accept         string `json:"Accept" bson:"Accept"`
	AcceptLanguage string `json:"Accept-Language" bson:"Accept-Language"`
	AcceptEncoding string `json:"Accept-Encoding" bson:"Accept-Encoding"`
	CacheControl   string `json:"Cache-Control" bson:"Cache-Control"`
	Pragma         string `json:"" bson:""`
	Connection     string `json:"Connection" bson:"Connection"`
	Body           byte   `json:"body" bson:"body"`
}

//DetailedInputPacketsHTTP детальное описание изходящих HTTP сообщений
// Version - версия HTTP протокола
// Code - код ответа сервера
// Reason - описание причины
// ContentType - тип передаваемого контента
// ContentLength - длинна контента
// LastModified - время модификации
// ETag
// Accept-Ranges
// Server
// XAmzCfID
// CacheControl
// Date
// Connection
// Body - первые 32 байта данных
type DetailedInputPacketsHTTP struct {
	Version       string `json:"version" bson:"version"`
	Code          int    `json:"code" bson:"code"`
	Reason        string `json:"reason" bson:"reason"`
	ContentType   string `json:"Content-Type" bson:"Content-Type"`
	ContentLength int    `json:"Content-Length" bson:"Content-Length"`
	LastModified  string `json:"Last-Modified" bson:"Last-Modified"`
	ETag          string `json:"ETag" bson:"ETag"`
	AcceptRanges  string `json:"Accept-Ranges" bson:"Accept-Ranges"`
	Server        string `json:"Server" bson:"Server"`
	XAmzCfID      string `json:"X-Amz-Cf-Id" bson:"X-Amz-Cf-Id"`
	CacheControl  string `json:"Cache-Control" bson:"Cache-Control"`
	Date          string `json:"Date" bson:"Date"`
	Connection    string `json:"Connection" bson:"Connection"`
	Body          byte   `json:"body" bson:"body"`
}
