package configure

/*
* Описание сообщений типа JSON передоваемых между API и клиентами
* */

//MsgCommonTask общее сообщение которое должно содержать ID задачи
/*type MsgCommonTask struct {
	TaskID string `json:"taskID"`
	MsgCommon
}*/

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

//SourceControlMsgTypeFromAPI подробно по источникам API->
type SourceControlMsgTypeFromAPI struct {
	SourceList []DetailedListSources `json:"sl"`
}

//SourceControlCurrentListSources опции для полного списка источников
type SourceControlCurrentListSources struct {
	MsgCommon
	MsgOptions SourceControlCurrentListSourcesList `json:"o"`
}

//SourceControlCurrentListSourcesList описание полного списка источников
type SourceControlCurrentListSourcesList struct {
	SourceList []ShortListSources `json:"sl"`
}

//SourceControlMsgOptions опции при управлении источниками
type SourceControlMsgOptions struct {
	MsgOptions SourceControlMsgTypeFromAPI `json:"o"`
}

//SourceControlActionsTakenSources описание выполненных действий с источниками
type SourceControlActionsTakenSources struct {
	MsgOptions SourceControlMsgTypeToAPI `json:"o"`
}

//SourceControlMsgTypeToAPI описание действий над источниками
type SourceControlMsgTypeToAPI struct {
	SourceList []ActionTypeListSources `json:"sl"`
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
	EnableTelemetry           bool     `json:"et"`
	MaxCountProcessFiltration int8     `json:"mcpf"` //<число 1-10>,
	StorageFolders            []string `json:"sf"`
}

//FiltrationControlMsgTypeInfo информационные сообщения о ходе фильтрации
type FiltrationControlMsgTypeInfo struct{}

//FiltrationControlMsgTypeCommand командные сообщения связанные с фильтрацией
//MsgInsturction:
//  - 'filtration start'
//  - 'filtration stop'
type FiltrationControlMsgTypeCommand struct {
	MsgInsturction string `json:"msgInsturction"`
}

//DownloadControlMsgTypeInfo информационные сообщения о ходе скачивания файлов
type DownloadControlMsgTypeInfo struct{}

//DownloadControlMsgTypeCommand командные сообщения связанные со скачиванием файлов
//MsgInsturction:
//  - 'download start'
//  - 'download stop'
//  - 'download resume'
type DownloadControlMsgTypeCommand struct {
	MsgInsturction string `json:"msgInsturction"`
}
