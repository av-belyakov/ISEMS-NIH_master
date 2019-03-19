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
	MsgType        string `json:"msgType"`
	MsgSection     string `json:"msgSection"`
	MsgInsturction string `json:"msgInsturction"`
	ClientTaskID   string `json:"taskID"`
	//MsgOptions     interface{} `json:"msgOptions"`
}

//MsgNotification информационное сообщение
type MsgNotification struct {
	MsgCommon
	MsgOptions UserNotification `json:"msgOptions"`
}

//NotificationParameters детальное описание сообщения
type NotificationParameters struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Sources     []int  `json:"sources"`
}

//UserNotification сообщение пользователю
type UserNotification struct {
	Notification NotificationParameters `json:"notification"`
}

//SourceControlMsgOptions опции настройки источников
type SourceControlMsgOptions struct {
	MsgOptions SourceControlMsgTypeFromAPI `json:"msgOptions"`
}

//SourceControlMsgTypeFromAPI подробно по источникам API->
type SourceControlMsgTypeFromAPI struct {
	SourceList []SourceListFromAPI `json:"sourceList"`
}

//SourceControlMsgTypeToAPI подробно по источникам ->API
type SourceControlMsgTypeToAPI struct {
	SourceList []SourceListToAPI `json:"sourceList"`
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

//SourceListInfoToAPI весь список источников ->API
//  - ID: уникальный числовой идентификатор источника
type SourceListInfoToAPI struct {
	ID int `json:"id"`
	SourceArguments
}

//SourceListFromAPI весь список источников ->API
//  - ID: уникальный числовой идентификатор источника
//  - ActionType: типа действия над источником
// ('add'/'update'/'delete'/'reconnect'/'status request',
// добавить, обновить, удалить, переподключить, запрос состояния)
//  - Argument: параметры источника, для actionType
// 'delete'/'reconnect'/'status request' это ПОЛЕ ПУСТОЕ
type SourceListFromAPI struct {
	ID         int             `json:"id"`
	ActionType string          `json:"actionType"`
	Argument   SourceArguments `json:"argument"`
	//map[string]SourceArguments `json:"argument"`
}

//SourceArguments параметры источников
//  - IP: ip адрес источника
//  - Token: уникальный идентификатор источника
//  - Settings: настройки источника
type SourceArguments struct {
	IP       string         `json:"ip"`
	Token    string         `json:"token"`
	Settings SourceSettings `json:"settings"`
}

//SourceSettings настройки источника
type SourceSettings struct {
	AsServer                  bool     `json:"asServer"`
	EnableTelemetry           bool     `json:"enableTelemetry"`
	MaxCountProcessFiltration int8     `json:"maxCountProcessFiltration"` //<число 1-10>,
	StorageFolders            []string `json:"storageFolders"`
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
