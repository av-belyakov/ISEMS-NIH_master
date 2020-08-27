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
// ConnectionStatus - статус сетевого соединения (true - connect, false - disconnect)
// DateLastConnected - дата последнего соединения
// Description - описание источника
type ShortListSources struct {
	ID                int    `json:"id"`
	IP                string `json:"ip"`
	ShortName         string `json:"sn"`
	ConnectionStatus  bool   `json:"cs"`
	DateLastConnected int64  `json:"dlc"`
	Description       string `json:"d"`
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
// ID - уникальный идентификатор источника
// UserName - имя пользователя инициировавшего задачу (если поле пустое, то считается что выполнение задачи было инициировано автоматически)
// DateTime - интервал времени фильтрации
// Protocol - протокол транспортного уровня
// Filters - параметры фильтрации (ip адреса, сети, порты)
type FiltrationControlCommonParametersFiltration struct {
	ID       int                                       `json:"id"`
	UserName string                                    `json:"un"`
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
// UserName - имя пользователя инициировавшего задачу (если поле пустое, то считается что выполнение задачи было инициировано автоматически)
// FileList - список файлов для скачивания полученных от клиента
type DownloadControlAdditionalOption struct {
	ID        int      `json:"id"`
	TaskIDApp string   `json:"tidapp"`
	UserName  string   `json:"un"`
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
// SearchRequestIsGeneratedAutomatically — был ли запрос на поиск сгенерирован автоматически (TRUE — да, FALSE - нет)
// ID - уникальный цифровой идентификатор источника
// ConsiderParameterTaskProcessed - учитывать параметр TaskProcessed
// TaskProcessed - была ли задача отмечена клиентом API как завершенная
// StatusFilteringTask - статус задачи по фильтрации
// StatusFileDownloadTask - статус задачи по скачиванию файлов
// ConsiderParameterFilesIsDownloaded - учитывать параметр FilesIsDownloaded
// FilesIsDownloaded - выполнялась ли выгрузка файлов
// ConsiderParameterAllFilesIsDownloaded - учитывать параметр AllFilesIsDownloaded
// AllFilesIsDownloaded - были ли выгружены все файлы
// InformationAboutFiltering - поиск по информации о результатах фильтрации
// InstalledFilteringOption - установленные опции фильтрации
type SearchInformationAboutTasksRequestOption struct {
	SearchRequestIsGeneratedAutomatically bool                             `json:"sriga"`
	ID                                    int                              `json:"id"`
	ConsiderParameterTaskProcessed        bool                             `json:"cptp"`
	TaskProcessed                         bool                             `json:"tp"`
	StatusFilteringTask                   string                           `json:"sft"`
	StatusFileDownloadTask                string                           `json:"sfdt"`
	ConsiderParameterFilesIsDownloaded    bool                             `json:"cpfid"`
	FilesIsDownloaded                     bool                             `json:"fid"`
	ConsiderParameterAllFilesIsDownloaded bool                             `json:"cpafid"`
	AllFilesIsDownloaded                  bool                             `json:"afid"`
	InformationAboutFiltering             InformationAboutFilteringOptions `json:"iaf"`
	InstalledFilteringOption              SearchFilteringOptions           `json:"ifo"`
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

// ПОЧТИ ПОЛНАЯ ИНФОРМАЦИЮ ПО ЗАДАЧЕ (нет только списка найденных файлов)

//SearchInformationResponseInformationByTaskID ответ содержащий почти полную информацию о задаче
// поиск которой осуществлялся по ее ID (может быть не весь список файлов)
type SearchInformationResponseInformationByTaskID struct {
	MsgCommon
	MsgOption ResponseInformationByTaskID `json:"o"`
}

//RequestInformationByTaskID запрос на получение информации о задаче по ее ID
type RequestInformationByTaskID struct {
	MsgCommon
	MsgOption ParametersGetInformationByTaskID `json:"o"`
}

//ParametersGetInformationByTaskID содержит параметры для поиска информации о задаче по ее ID
// SearchRequestIsGeneratedAutomatically — был ли запрос на поиск сгенерирован автоматически (TRUE — да, FALSE - нет)
// ReguestTaskID - запрашиваемый уникальный цифровой идентификатор задачи  по фильтрации и скачиванию, выполняемой или выполненной
type ParametersGetInformationByTaskID struct {
	SearchRequestIsGeneratedAutomatically bool   `json:"sriga"`
	RequestTaskID                         string `json:"rtid"`
}

//ResponseInformationByTaskID содержит почти полную информацию (кроме списка найденных файлов) о найденной задаче
// Status — статус выполняемой задачи
// TaskParameter - параметры выполняемой задачи
type ResponseInformationByTaskID struct {
	Status        string                `json:"s"`
	TaskParameter ResponseTaskParameter `json:"tp"`
}

//ResponseTaskParameter содержит основные параметры найденной задачи
// TaskID — внутренний идентификатор задачи
// ClientTaskID — идентификатор задачи присвоенный клиентом
// SourceID -  идентификатор источника
// UserInitiatedFilteringProcess - пользователь инициировавший задачу по фильтрации
// UserInitiatedFileDownloadProcess - пользователь инициировавший задачу по скачиванию файлов
// GeneralInformationAboutTask - основные параметры задачи, не относящиеся ни к одному из разделов
// FilteringOption - параметры фильтрации
// DetailedInformationOnFiltering - результаты фильтрации
// DetailedInformationOnDownloading - результаты выгрузки файлов
// DetailedInformationListFiles - подробная информация о первых 50 найденных файлах (первые 50 и все остальные передаются отдельным запросом)
type ResponseTaskParameter struct {
	TaskID                           string                      `json:"tid"`
	ClientTaskID                     string                      `json:"ctid"`
	SourceID                         int                         `json:"sid"`
	UserInitiatedFilteringProcess    string                      `json:"uifp"`
	UserInitiatedFileDownloadProcess string                      `json:"uifdp"`
	GeneralInformationAboutTask      GeneralInformationAboutTask `json:"giat"`
	FilteringOption                  TaskFilteringOption         `json:"fo"`
	DetailedInformationOnFiltering   InformationOnFiltering      `json:"diof"`
	DetailedInformationOnDownloading InformationOnDownloading    `json:"diod"`
	DetailedInformationListFiles     []FileInformation           `json:"dilf"`
}

//GeneralInformationAboutTask содкржит общие параметры задачи
// TaskProcessed - была ли обработана задача (фильтрация файлов выполнена, какие либо файлы были загружены, тогда клиент API может отметить подобную задачу как отработанную)
// DateTimeProcessed - дата и время завершения задачи
// ClientIDIP - идентификатор клиента API (его ID:IP !!!!!!!!!!!! Надо не забыть !!!!!)
// DetailDescription - детальное описание
type GeneralInformationAboutTask struct {
	TaskProcessed     bool              `json:"tp"`
	DateTimeProcessed int64             `json:"dtp"`
	ClientIDIP        string            `json:"cidip"`
	DetailDescription DetailDescription `json:"dd"`
}

//DetailDescription детальное описание
// UserNameClosedProcess — имя пользователя (внутри клиента API) который закрыл задачу
// DescriptionProcessingResults — не обязательное описание причины закрытия задачи, от клиента API
type DetailDescription struct {
	UserNameClosedProcess        string `json:"uncp"`
	DescriptionProcessingResults string `json:"dpr"`
}

//TaskFilteringOption параметры фильтрации
// DateTime - временной интервал
// Protocol - протокол транспортного уровня
// Filters - сетевые фильтры
type TaskFilteringOption struct {
	DateTime DateTimeParameters                        `json:"dt"`
	Protocol string                                    `json:"p"`
	Filters  FiltrationControlParametersNetworkFilters `json:"f"`
}

//InformationOnFiltering результаты выполнения задачи по фильтрации
// TaskStatus — статус задачи 'not executed'/'wait'/'refused'/'execute'/'not fully completed'/'complete'
// TimeIntervalTaskExecution — интервал выполнения задачи
// WasIndexUsed — использовались ли индексы при выполнении данной задачи
// NumberFilesMeetFilterParameters - кол-во файлов удовлетворяющих параметрам фильтрации
// NumberProcessedFiles - кол-во обработанных файлов
// NumberFilesFoundResultFiltering - кол-во найденных, в результате фильтрации, файлов
// NumberDirectoryFiltartion - кол-во директорий по которым выполняется фильтрация
// NumberErrorProcessedFiles - кол-во не обработанных файлов или файлов обработанных с ошибками
// SizeFilesMeetFilterParameters - общий размер файлов (в байтах) удовлетворяющих параметрам фильтрации
// SizeFilesFoundResultFiltering - общий размер найденных, в результате фильтрации, файлов (в байтах)
// PathDirectoryForFilteredFiles - путь к директории в которой хранятся отфильтрованные файлы
type InformationOnFiltering struct {
	TaskStatus                      string             `json:"ts"`
	TimeIntervalTaskExecution       DateTimeParameters `json:"tte"`
	WasIndexUsed                    bool               `json:"wiu"`
	NumberProcessedFiles            int                `json:"mpf"`
	NumberDirectoryFiltartion       int                `json:"ndf"`
	NumberErrorProcessedFiles       int                `json:"nepf"`
	NumberFilesMeetFilterParameters int                `json:"nfmfp"`
	NumberFilesFoundResultFiltering int                `json:"nffrf"`
	SizeFilesMeetFilterParameters   int64              `json:"sfmfp"`
	SizeFilesFoundResultFiltering   int64              `json:"sffrf"`
	PathDirectoryForFilteredFiles   string             `json:"pdfff"`
}

//InformationOnDownloading результаты выполнения задачи по скачиванию файлов
// TaskStatus -  статус задачи
// TimeIntervalTaskExecution — интервал выполнения задачи
// NumberFilesTotal - общее количество файлов подлежащих скачиванию
// NumberFilesDownloaded - количество уже загруженных файлов
// NumberFilesDownloadedError - количество файлов загруженных с ошибкой
// PathDirectoryStorageDownloadedFiles - путь до директории долговременного хранения скаченных файлов
type InformationOnDownloading struct {
	TaskStatus                          string             `json:"ts"`
	TimeIntervalTaskExecution           DateTimeParameters `json:"tte"`
	NumberFilesTotal                    int                `json:"nft"`
	NumberFilesDownloaded               int                `json:"nfd"`
	NumberFilesDownloadedError          int                `json:"nfde"`
	PathDirectoryStorageDownloadedFiles string             `json:"pdsdf"`
}

//FileInformation информация о файле
// Name - название файла
// Size - размер файла
// IsLoaded - был ли файл загружен (true - загружен)
type FileInformation struct {
	Name     string `json:"n"`
	Size     int64  `json:"s"`
	IsLoaded bool   `json:"isl"`
}

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
// StartTimeTaskExecution - время начала выполнения задачи
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
	StartTimeTaskExecution               int64                       `json:"stte"`
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

// ЗАПРОС ОГРАНИЧЕННОГО СПИСКА НАЙДЕННЫХ В РЕЗУЛЬТАТЕ ФИЛЬТРАЦИИ ФАЙЛОВ

//GetListFoundFilesRequest содержит запрос на получение ограниченного списка найденных файлов
type GetListFoundFilesRequest struct {
	MsgCommon
	MsgOption GetListFoundFilesRequestOption `json:"o"`
}

//GetListFoundFilesRequestOption содержит параметры для запроса списка файлов
// RequestTaskID - ID искомой задачи
// PartSize - количество запрашиваемых файлов
// OffsetListParts - смещение по списку файлов
type GetListFoundFilesRequestOption struct {
	RequestTaskID   string `json:"rtid"`
	PartSize        int    `json:"ps"`
	OffsetListParts int    `json:"olp"`
}

// ОТВЕТ НА ЗАПРОС ОГРАНИЧЕННОГО СПИСКА НАЙДЕННЫХ В РЕЗУЛЬТАТЕ ФИЛЬТРАЦИИ ФАЙЛОВ

//ListFoundFilesResponse содержит ответ с ограниченным списком найденных файлов
type ListFoundFilesResponse struct {
	MsgCommon
	MsgOption ListFoundFilesResponseOption `json:"o"`
}

//ListFoundFilesResponseOption содержит детальный ответ с ограниченным списком найденных файлов
// Status — статус выполняемой задачи
// TaskID — внутренний идентификатор задачи
// ClientTaskID — идентификатор задачи присвоенный клиентом
// SourceID -  идентификатор источника
// FullListSize — полный размер списка файловом
// RequestPartSize — размер запрашиваемой частично
// OffsetListParts — смещение по списку файлов
// ListFiles — список найденных файлов
type ListFoundFilesResponseOption struct {
	Status          string              `json:"s"`
	TaskID          string              `json:"tid"`
	ClientTaskID    string              `json:"ctid"`
	SourceID        int                 `json:"sid"`
	FullListSize    int                 `json:"fls"`
	RequestPartSize int                 `json:"rps"`
	OffsetListParts int                 `json:"olp"`
	ListFiles       []*FilesInformation `json:"lf"`
}

// ЗАПРОС НА ЗАКРЫТИЕ ЗАДАЧИ

//MarkTaskCompletedRequest содержит запрос на закрытие задачи
type MarkTaskCompletedRequest struct {
	MsgCommon
	MsgOption MarkTaskCompletedRequestOption `json:"o"`
}

//MarkTaskCompletedRequestOption содержит параметры для закрытие задачи
// RequestTaskID - внутренний идентификатор задачи
// UserName - имя пользователя
// Description - дополнительное описание
type MarkTaskCompletedRequestOption struct {
	RequestTaskID string `json:"tid"`
	UserName      string `json:"un"`
	Description   string `json:"d"`
}

// ОТВЕТ НА ЗАПРОС ПО ЗАКРЫТИЮ ЗАДАЧИ

//MarkTaskCompletedResponse содержит запрос на закрытие задачи
type MarkTaskCompletedResponse struct {
	MsgCommon
	MsgOption MarkTaskCompletedResponseOption `json:"o"`
}

//MarkTaskCompletedResponseOption содержит параметры для закрытие задачи
// SuccessStatus - статус успешности выполнения
// RequestTaskID - внутренний идентификатор задачи
type MarkTaskCompletedResponseOption struct {
	SuccessStatus bool   `json:"ss"`
	RequestTaskID string `json:"tid"`
}
