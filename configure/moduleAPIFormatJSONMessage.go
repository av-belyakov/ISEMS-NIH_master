package configure

/*
* Описание сообщений типа JSON передоваемых между API и клиентами
* */

/*--- ОБЩИЕ ---*/

//MsgCommon общее сообщение
// MsgType:
//  - 'information'
//  - 'command'
// MsgSection:
//  - 'source control'
//  - 'filtration control'
//  - 'download control'
//  - 'information search control'
//  - 'user notification'
// MsgInstruction:
//  - 'get new source list' API->
//  - 'change status source' API->
//  - 'confirm the action' API->
//  - 'send new source list' API<-
//  - 'performing an action' API<-
//  - 'send notification' API<-
type MsgCommon struct {
	MsgType        string `json:"t"`
	MsgSection     string `json:"s"`
	MsgInstruction string `json:"i"`
	ClientTaskID   string `json:"tid"`
}

/*--- ИНФОРМАЦИОННЫЕ СООБЩЕНИЯ ---*/

//MsgNotification информационное сообщение
type MsgNotification struct {
	MsgCommon
	MsgOptions UserNotification `json:"o"`
}

//NotificationParameters детальное описание сообщения
// Type - тип сообщения (success, warning, info, danger)
// Description - описание сообщения
// Sources - список источников к которому данное сообщение применимо
type NotificationParameters struct {
	Type        string `json:"t"`
	Description string `json:"d"`
	Sources     []int  `json:"s"`
}

//UserNotification сообщение пользователю
type UserNotification struct {
	Notification NotificationParameters `json:"n"`
}

/*--- УПРАВЛЕНИЕ ИСТОЧНИКАМИ ---*/

//SourceControlMsgTypeFromAPI подробно по источникам API->
type SourceControlMsgTypeFromAPI struct {
	TaskInfo   MsgTaskInfo           `json:"ti"`
	SourceList []DetailedListSources `json:"sl"`
}

//SourceControlCurrentListSources опции для полного списка источников
type SourceControlCurrentListSources struct {
	MsgCommon
	MsgOptions SourceControlCurrentListSourcesList `json:"o"`
}

//SourceControlConfirmActionSource список источников с выполненными над ними
// действиями и статусом успешности действия
type SourceControlConfirmActionSource struct {
	MsgCommon
	MsgOptions SourceControlMsgTypeToAPI `json:"o"`
}

//SourceControlCurrentListSourcesList описание полного списка источников
type SourceControlCurrentListSourcesList struct {
	TaskInfo   MsgTaskInfo        `json:"ti"`
	SourceList []ShortListSources `json:"sl"`
}

//SourceControlMsgOptions опции при управлении источниками
type SourceControlMsgOptions struct {
	MsgOptions SourceControlMsgTypeFromAPI `json:"o"`
}

//SourceControlActionsTakenSources описание выполненных действий с источниками
type SourceControlActionsTakenSources struct {
	MsgCommon
	MsgOptions SourceControlMsgTypeToAPI `json:"o"`
}

//SourceControlMsgTypeToAPI описание действий над источниками
type SourceControlMsgTypeToAPI struct {
	TaskInfo   MsgTaskInfo             `json:"ti"`
	SourceList []ActionTypeListSources `json:"sl"`
}

//MsgTaskInfo описания состояния задачи
type MsgTaskInfo struct {
	State       string `json:"s"`
	Explanation string `json:"e"`
}

//ActionTypeListSources описание действий над источниками
// ID - уникальный идентификатор источника
// Status - статус подключения источника
// ActionType - тип выполняемого действия
// IsSuccess - успешность действия
// MessageFailure - описание причины неудачи
type ActionTypeListSources struct {
	ID             int    `json:"id"`
	Status         string `json:"s"`
	ActionType     string `json:"at"`
	IsSuccess      bool   `json:"is"`
	MessageFailure string `json:"mf"`
}

//SourceListToAPI описание параметров источника API->
//  ID - уникальный числовой идентификатор источника
//  Status - статус соединения ('connect'/'disconnect')
//  ActionType - тип выполняемого действия ('add'/'delete'/'update'/'reconnect'/'none')
//  IsSuccess - успешность действия (true/false)
//  MessageFailure - описание причины неудачи (пустое если isSuccess = true)
type SourceListToAPI struct {
	ID             int    `json:"id"`
	Status         string `json:"status"`
	ActionType     string `json:"actionType"`
	IsSuccess      bool   `json:"isSuccess"`
	MessageFailure string `json:"messageFailure"`
}

//ShortListSources краткие настройки источника
// ID - уникальный числовой идентификатор источника
// IP - ip адрес источника
// ShortName - краткое название источника
// Description - описание источника
type ShortListSources struct {
	ID          int    `json:"id"`
	IP          string `json:"ip"`
	ShortName   string `json:"sn"`
	Description string `json:"d"`
}

//DetailedListSources весь список источников ->API
// ID - уникальный числовой идентификатор источника
// ActionType - типа действия над источником ('add'/'update'/'delete'/'reconnect'/'status request',
//  добавить, обновить, удалить, переподключить, запрос состояния)
// ShortName - краткое название источника
// Description - полное название источника
// Argument - параметры источника, для actionType 'delete'/'reconnect'/'status request' это ПОЛЕ ПУСТОЕ
type DetailedListSources struct {
	ID         int             `json:"id"`
	ActionType string          `json:"at"`
	Argument   SourceArguments `json:"arg"`
}

//SourceArguments параметры источников
// IP - ip адрес источника
// Token - уникальный идентификатор источника
// Settings - настройки источника
// Description - описание источника
// Settings - настройки источника
type SourceArguments struct {
	IP          string         `json:"ip"`
	Token       string         `json:"t"`
	ShortName   string         `json:"sn"`
	Description string         `json:"d"`
	Settings    SourceSettings `json:"s"`
}

//SourceSettings настройки источника
// AsServer - запуск источника как сервер или как клиент (server/client)
// Port - сетевой порт
// EnableTelemetry - включить передачу телеметрии
// MaxCountProcessFiltration - максимальное количество одновременно запущенных процессов фильтрации (число 1-10),
// StorageFolders - список директорий с файлами по которым выполняется фильтрация
// TypeAreaNetwork - тип протокола канального уровня (ip/pppoe)
type SourceSettings struct {
	AsServer                  bool     `json:"as"`
	Port                      int      `json:"p"`
	EnableTelemetry           bool     `json:"et"`
	MaxCountProcessFiltration int8     `json:"mcpf"`
	StorageFolders            []string `json:"sf"`
	TypeAreaNetwork           string   `json:"tan"`
}

/*--- ИНФОРМАЦИЯ ПО ТЕЛЕМЕТРИИ ---*/

//Telemetry телеметрия
type Telemetry struct {
	MsgCommon
	MsgOptions TelemetryOptions `json:"o"`
}

//TelemetryOptions дополнительные опции
type TelemetryOptions struct {
	SourceID    int                  `json:"id"`
	Information TelemetryInformation `json:"i"`
}

/*--- ИНФОРМАЦИЯ О ВЕРСИИ УСТАНОВЛЕННОГО НА ИСТОЧНИКЕ ПРОГРАММНОГО ОБЕСПЕЧЕНИЯ И ДАТЕ ЕГО РЕЛИЗА ---*/

//SourceVersionApp информация о версии приложения на источнике
type SourceVersionApp struct {
	MsgCommon
	MsgOptions SourceVersionAppOptions `json:"o"`
}

//SourceVersionAppOptions дополнительные опции
// SourceID - уникальный идентификатор источника
// AppVersion - версия установленного на источнике приложения
// AppReleaseDate - дата релиза установленного на источнике приложения
type SourceVersionAppOptions struct {
	SourceID       int    `json:"id"`
	AppVersion     string `json:"av"`
	AppReleaseDate string `json:"ard"`
}

/*--- УПРАВЛЕНИЕ ФИЛЬТРАЦИЕЙ ---*/

//FiltrationControlTypeStart общее описание запроса на начало фильтрации
type FiltrationControlTypeStart struct {
	MsgCommon
	MsgOption FiltrationControlCommonParametersFiltration `json:"o"`
}

//FiltrationControlCommonParametersFiltration описание параметров фильтрации
type FiltrationControlCommonParametersFiltration struct {
	ID       int                                       `json:"id"`
	DateTime DateTimeParameters                        `json:"dt"`
	Protocol string                                    `json:"p"`
	Filters  FiltrationControlParametersNetworkFilters `json:"f"`
}

//DateTimeParameters параметры времени
type DateTimeParameters struct {
	Start int64 `json:"s"`
	End   int64 `json:"e"`
}

//FiltrationControlParametersNetworkFilters параметры сетевых фильтров
type FiltrationControlParametersNetworkFilters struct {
	IP      FiltrationControlIPorNetorPortParameters `json:"ip"`
	Port    FiltrationControlIPorNetorPortParameters `json:"pt"`
	Network FiltrationControlIPorNetorPortParameters `json:"nw"`
}

//FiltrationControlIPorNetorPortParameters параметры для ip или network
type FiltrationControlIPorNetorPortParameters struct {
	Any []string `json:"any"`
	Src []string `json:"src"`
	Dst []string `json:"dst"`
}

//FiltrationControlTypeInfo общее описание сообщения о ходе фильтрации
type FiltrationControlTypeInfo struct {
	MsgCommon
	MsgOption FiltrationControlMsgTypeInfo `json:"o"`
}

//FiltrationControlMsgTypeInfo информационные сообщения о ходе фильтрации
// ID - уникальный цифровой идентификатор источника
// TaskIDApp - уникальный цифровой идентификатор задачи присвоенный приложением
// Status - статус выполняемой задачи
// NumberFilesMeetFilterParameters - кол-во файлов удовлетворяющих параметрам фильтрации
// NumberProcessedFiles - кол-во обработанных файлов
// NumberFilesFoundResultFiltering - кол-во найденных, в результате фильтрации, файлов
// NumberErrorProcessedFiles - кол-во не обработанных файлов или файлов обработанных с ошибками
// NumberDirectoryFiltartion - кол-во директорий по которым выполняется фильтрация
// SizeFilesMeetFilterParameters - общий размер файлов (в байтах) удовлетворяющих параметрам фильтрации
// SizeFilesFoundResultFiltering - общий размер найденных, в результате фильтрации, файлов (в байтах)
// PathStorageSource — путь до директории в которой сохраняются файлы при
// NumberMessagesParts - коичество частей сообщения
// FoundFilesInformation - информация о файлах, ключ - имя файла
type FiltrationControlMsgTypeInfo struct {
	ID                              int                               `json:"id"`
	TaskIDApp                       string                            `json:"tidapp"`
	Status                          string                            `json:"s"`
	NumberFilesMeetFilterParameters int                               `json:"nfmfp"`
	NumberProcessedFiles            int                               `json:"npf"`
	NumberFilesFoundResultFiltering int                               `json:"nffrf"`
	NumberErrorProcessedFiles       int                               `json:"nepf"`
	NumberDirectoryFiltartion       int                               `json:"ndf"`
	SizeFilesMeetFilterParameters   int64                             `json:"sfmfp"`
	SizeFilesFoundResultFiltering   int64                             `json:"sffrf"`
	PathStorageSource               string                            `json:"pss"`
	FoundFilesInformation           map[string]*InputFilesInformation `json:"ffi"`
}

/*--- УПРАВЛЕНИЕ СКАЧИВАНИЕМ ФАЙЛОВ ---*/

//DownloadControlTypeStart общее описание запроса на начало скачивания файлов
type DownloadControlTypeStart struct {
	MsgCommon
	MsgOption DownloadControlAdditionalOption `json:"o"`
}

//DownloadControlAdditionalOption список файлов на скачивание
// ID - уникальный идентификатор источника
// TaskIDApp - уникальный идентификатор задачи по фильтрации
// FileList - список файлов для скачивания полученных от клиента
type DownloadControlAdditionalOption struct {
	ID        int      `json:"id"`
	TaskIDApp string   `json:"tidapp"`
	FileList  []string `json:"fl"`
}

//DownloadControlTypeInfo общее описание сообщения о ходе скачивания файлов
type DownloadControlTypeInfo struct {
	MsgCommon
	MsgOption DownloadControlMsgTypeInfo `json:"o"`
}

//DownloadControlMsgTypeInfo информационные сообщения о ходе скачивания файлов
// ID - уникальный цифровой идентификатор источника
// TaskIDApp - уникальный цифровой идентификатор задачи присвоенный приложением при выполнении фильтрации
// Status - статус выполняемой задачи 'wait'/'refused'/'execute'/'complete'/'stoped'
// NumberFilesTotal — общее количество скачиваемых файлов
// NumberFilesDownloaded — количество успешно скаченных файлов
// NumberFilesDownloadedError — количество файлов скаченных с ошибкой
// PathDirectoryStorageDownloadedFiles — путь до директории в файловом хранилище
// DetailedFileInformation — подробная информация о скачиваемом файле
type DownloadControlMsgTypeInfo struct {
	ID                                  int                 `json:"id"`
	TaskIDApp                           string              `json:"tidapp"`
	Status                              string              `json:"s"`
	NumberFilesTotal                    int                 `json:"nft"`
	NumberFilesDownloaded               int                 `json:"nfd"`
	NumberFilesDownloadedError          int                 `json:"nfde"`
	PathDirectoryStorageDownloadedFiles string              `json:"pdsdf"`
	DetailedFileInformation             MoreFileInformation `json:"dfi"`
}

//MoreFileInformation подробная информация о скачиваемом файле
// Name — название файла
// Hex — хеш сумма
// FullSizeByte — полный размер файла в байтах
// AcceptedSizeByte — скаченный размер файла в байтах
// AcceptedSizePercent — скаченный размер файла в процентах
type MoreFileInformation struct {
	Name                string `json:"n"`
	Hex                 string `json:"h"`
	FullSizeByte        int64  `json:"fsb"`
	AcceptedSizeByte    int64  `json:"asb"`
	AcceptedSizePercent int    `json:"asp"`
}

/*--- ПОИСК ИНФОРМАЦИИ О ВЫПОЛНЯЕМЫХ ИЛИ ВЫПОЛНЕННЫХ ЗАДАЧАХ ---*/

// ПОИСК ИНФОРМАЦИИ ПО ЗАДАННЫМ ПАРАМЕТРАМ

//SearchInformationAboutTasksRequest общее описание запроса на поиск информации по задачам
type SearchInformationAboutTasksRequest struct {
	MsgCommon
	MsgOption SearchInformationAboutTasksRequestOption `json:"o"`
}

//SearchInformationAboutTasksRequestOption дополнительные опции для поиска информации по задаче
// TaskProcessed - была ли задача отмечена клиентом API как завершенная
// ID - уникальный цифровой идентификатор источника
// StatusFilteringTask - статус задачи по фильтрации
// StatusFileDownloadTask - статус задачи по скачиванию файлов
// FilesDownloaded - опции выгрузки файлов
// InformationAboutFiltering - поиск по информации о результатах фильтрации
// InstalledFilteringOption - установленные опции фильтрации
type SearchInformationAboutTasksRequestOption struct {
	TaskProcessed             bool                             `json:"tp"`
	ID                        int                              `json:"id"`
	StatusFilteringTask       string                           `json:"sft"`
	StatusFileDownloadTask    string                           `json:"sfdt"`
	FilesDownloaded           FilesDownloadedOptions           `json:"fd"`
	InformationAboutFiltering InformationAboutFilteringOptions `json:"iaf"`
	InstalledFilteringOption  SearchFilteringOptions           `json:"ifo"`
}

//FilesDownloadedOptions опции выгрузки файлов
// FilesIsDownloaded - выполнялась ли выгрузка файлов
// AllFilesIsDownloaded - были ли выгружены все файлы
type FilesDownloadedOptions struct {
	FilesIsDownloaded    bool `json:"fid"`
	AllFilesIsDownloaded bool `json:"afid"`
}

//InformationAboutFilteringOptions опции для поиска по информации о результатах фильтрации
// FilesIsFound - были ли найдены файлы
// CountAllFilesMin - минимальное общее количество всех найденных файлов
// CountAllFilesMax - максимальное общее количество всех найденных файлов
// SizeAllFilesMin - общий минимальный размер всех найденных файлов
// SizeAllFilesMax - общий максимальный размер всех найденных файлов
type InformationAboutFilteringOptions struct {
	FilesIsFound     bool  `json:"fif"`
	CountAllFilesMin int   `json:"cafmin"`
	CountAllFilesMax int   `json:"cafmax"`
	SizeAllFilesMin  int64 `json:"safmin"`
	SizeAllFilesMax  int64 `json:"safmax"`
}

//SearchFilteringOptions искомые опции фильтрации
// DateTime - временной диапазон по которому осуществлялась фильтрация
// Protocol - тип транспортного протокола
// NetworkFilters - сетевые фильтры
type SearchFilteringOptions struct {
	DateTime       DateTimeParameters                        `json:"dt"`
	Protocol       string                                    `json:"p"`
	NetworkFilters FiltrationControlParametersNetworkFilters `json:"nf"`
}

// ПОЛУЧИТЬ ВЫБРАННУЮ ЧАСТЬ КРАТКОЙ ИНФОРМАЦИИ ИЗ СПИСКА НАЙДЕННЫХ ЗАДАЧ

// ПОЛУЧИТЬ ПОЛНУЮ ИНФОРМАЦИЮ ПО ЗАДАЧЕ

// ОТВЕТ ПРИ ПОИСКЕ ИНФОРМАЦИИ О ЗАДАЧАХ

//SearchInformationResponseCommanInfo общее описание ответа при поиске информации о задачах
type SearchInformationResponseCommanInfo struct {
	MsgCommon
	MsgOption SearchInformationResponseOptionCommanInfo `json:"o"`
}

//SearchInformationResponseOptionCommanInfo общее описание ответа при поиске информации о задачах
// TaskIDApp - уникальный цифровой идентификатор задачи присвоенный приложением
// Status - статус выполняемой задачи
// TotalNumberTasksFound - общее количество найденных задач
// PaginationOptions - параметры разбиения по частям
// ShortListFoundTasks - краткий список найденных задач
type SearchInformationResponseOptionCommanInfo struct {
	TaskIDApp             string                  `json:"tidapp"`
	Status                string                  `json:"s"`
	TotalNumberTasksFound int                     `json:"tntf"`
	PaginationOptions     PaginationOption        `json:"p"`
	ShortListFoundTasks   []*BriefTaskInformation `json:"slft"`
}

//PaginationOption параметры разбиения по частям
// ChunkSize - размер сегмента (кол-во задач в сегменте)
// ChunkNumber - общее количество сегментов
// ChunkCurrentNumber - номер текущего фрагмента
type PaginationOption struct {
	ChunkSize          int `json:"cs"`
	ChunkNumber        int `json:"cn"`
	ChunkCurrentNumber int `json:"ccn"`
}

//BriefTaskInformation краткая информация о найденной задаче
// TaskID - ID задачи присвоенный приложением
// ClientTaskID - ID задачи переданный клиентом API
// SourceID - ID источника
// ParametersFiltrationOptions - параметры фильтрации
// FilteringTaskStatus - статус задачи по фильтрации
// FileDownloadTaskStatus - статус задачи по скачиванию файлов
// NumberFilesFoundAsResultFiltering - кол-во файлов найденных в результате фильтрации
// TotalSizeFilesFoundAsResultFiltering - общий размер файлов найденных в результате фильтрации
// NumberFilesDownloaded - кол-во принятых файлов
type BriefTaskInformation struct {
	TaskID                               string                      `json:"tid"`
	ClientTaskID                         string                      `json:"ctid"`
	SourceID                             int                         `json:"sid"`
	ParametersFiltration                 ParametersFiltrationOptions `json:"pf"`
	FilteringTaskStatus                  string                      `json:"fts"`
	FileDownloadTaskStatus               string                      `json:"fdts"`
	NumberFilesFoundAsResultFiltering    int                         `json:"nffarf"`
	TotalSizeFilesFoundAsResultFiltering int64                       `json:"tsffarf"`
	NumberFilesDownloaded                int                         `json:"nfd"`
}

//ParametersFiltrationOptions параметры фильтрации
// DateTime - временной диапазон для фильтрации
// Protocol - сетевой протокол
// Filters - фильтры для фильтрации
type ParametersFiltrationOptions struct {
	DateTime DateTimeParameters                        `json:"dt"`
	Protocol string                                    `json:"p"`
	Filters  FiltrationControlParametersNetworkFilters `json:"f"`
}

// ОТВЕТ ПРИ ЗАПРОСЕ СЛЕДУЮЩЕЙ ЧАСТИ НАЙДЕННЫХ ЗАДАЧ

// ОТВЕТ НА ЗАПРОС ПОЛНОЙ ИНФОРМАЦИИ О ЗАДАЧЕ
