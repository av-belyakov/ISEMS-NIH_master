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

//InformationAboutTaskFiltration подробная информация связанная с задачей по фильтрации
// TaskID - уникальный идентификатор задачи полученный от приложения
// ClientID - уникальный идентификатор клиента
// ClientTaskID - уникальный идентификатор задачи полученный от клиента
// SourceID - идентификатор источника на котором выполняется задача
// FilteringOption - параметры фильтрации полученные от клиента
type InformationAboutTaskFiltration struct {
	TaskID                           string                         `json:"task_id" bson:"task_id"`
	ClientID                         string                         `json:"client_id" bson:"client_id"`
	ClientTaskID                     string                         `json:"client_task_id" bson:"client_task_id"`
	SourceID                         int                            `json:"source_id" bson:"source_id"`
	FilteringOption                  FilteringOption                `json:"filtering_option" bson:"filtering_option"`
	DetailedInformationOnFiltering   DetailedInformationFiltering   `jsom:"detailed_information_on_filtering" bson:"detailed_information_on_filtering"`
	DetailedInformationOnDownloading DetailedInformationDownloading `jsom:"detailed_information_on_downloading" bson:"detailed_information_on_downloading"`
	ListFilesResultTaskExecution     []*FilesInformation            `json:"list_files_result_task_execution" bson:"list_files_result_task_execution"`
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
// TaskStatus - состояние задачи
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
