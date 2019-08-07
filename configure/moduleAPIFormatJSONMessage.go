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
// MsgInsturction:
//  - 'get new source list' API->
//  - 'change status source' API->
//  - 'confirm the action' API->
//  - 'send new source list' API<-
//  - 'performing an action' API<-
//  - 'send notification' API<-
type MsgCommon struct {
	MsgType        string `json:"t"`
	MsgSection     string `json:"s"`
	MsgInsturction string `json:"i"`
	ClientTaskID   string `json:"tid"`
}

/*--- ИНФОРМАЦИОННЫЕ СООБЩЕНИЯ ---*/

//MsgNotification информационное сообщение
type MsgNotification struct {
	MsgCommon
	MsgOptions UserNotification `json:"o"`
}

//NotificationParameters детальное описание сообщения
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
//действиями и статусом успешности действия
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
type ActionTypeListSources struct {
	ID             int    `json:"id"`
	Status         string `json:"s"`
	ActionType     string `json:"at"`
	IsSuccess      bool   `json:"is"`
	MessageFailure string `json:"mf"`
}

//SourceListToAPI описание параметров источника API->
//  - ID уникальный числовой идентификатор источника
//  - Status: 'connect'/'disconnect'
//  - ActionType: 'add'/'delete'/'update'/'reconnect'/'none'
//  - IsSuccess: true/false
//  - MessageFailure: <сообщение об ошибке> //пустое если isSuccess = true
type SourceListToAPI struct {
	ID             int    `json:"id"`
	Status         string `json:"status"`
	ActionType     string `json:"actionType"`
	IsSuccess      bool   `json:"isSuccess"`
	MessageFailure string `json:"messageFailure"`
}

//ShortListSources краткие настройки источника
type ShortListSources struct {
	ID          int    `json:"id"`
	IP          string `json:"ip"`
	ShortName   string `json:"sn"`
	Description string `json:"d"`
}

//DetailedListSources весь список источников ->API
//  - ID: уникальный числовой идентификатор источника
//  - ActionType: типа действия над источником
// ('add'/'update'/'delete'/'reconnect'/'status request',
// добавить, обновить, удалить, переподключить, запрос состояния)
//  - ShortName: краткое название источника
//  - Description: полное название источника
//  - Argument: параметры источника, для actionType
// 'delete'/'reconnect'/'status request' это ПОЛЕ ПУСТОЕ
type DetailedListSources struct {
	ID         int             `json:"id"`
	ActionType string          `json:"at"`
	Argument   SourceArguments `json:"arg"`
}

//SourceArguments параметры источников
//  - IP: ip адрес источника
//  - Token: уникальный идентификатор источника
//  - Settings: настройки источника
type SourceArguments struct {
	IP          string         `json:"ip"`
	Token       string         `json:"t"`
	ShortName   string         `json:"sn"`
	Description string         `json:"d"`
	Settings    SourceSettings `json:"s"`
}

//SourceSettings настройки источника
type SourceSettings struct {
	AsServer                  bool     `json:"as"`
	Port                      int      `json:"p"`
	EnableTelemetry           bool     `json:"et"`
	MaxCountProcessFiltration int8     `json:"mcpf"` //<число 1-10>,
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
// Status - статус выполняемой задачи
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
