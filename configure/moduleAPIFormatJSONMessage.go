package configure

/*
* Описание сообщений типа JSON передоваемых между API и клиентами
* */

//MsgType общее сообщение
// MsgType:
//  - 'information'
//  - 'command'
// MsgSection:
//  - 'source control'
//  - 'filtration control'
//  - 'download control'
//  - 'information search control'
// MsgInsturction:
//  - 'get new source list' API->
//  - 'change status source' API->
//  - 'confirm the action' API->
//  - 'send new source list' API<-
//  - 'performing an action' API<-
type MsgType struct {
	MsgType        string `json:"msgType"`
	MsgSection     string `json:"msgSection"`
	MsgInsturction string `json:"msgInsturction"`
	MsgOptions     []byte `json:"msgOptions"`
}

//SourceControlMsgTypeInfoFromAPI подробно по источникам API->
type SourceControlMsgTypeInfoFromAPI struct {
	SourceList []SourceListInfoFromAPI `json:"sourceList"`
}

//SourceControlMsgTypeCommandFromAPI при запросе списка источников API->
type SourceControlMsgTypeCommandFromAPI map[string]string

//SourceControlMsgTypeInfoToAPI подробно по источникам ->API
type SourceControlMsgTypeInfoToAPI struct {
	SourceList []SourceListInfoToAPI `json:"sourceList"`
}

//SourceListInfoFromAPI описание параметров источника API->
//  - ID уникальный числовой идентификатор источника
//  - Status: 'connect'/'disconnect'
//  - ActionType: 'add'/'delete'/'update'/'reconnect'/'none'
//  - IsSuccess: true/false
//  - MessageFailure: <сообщение об ошибке> //пустое если isSuccess = true
type SourceListInfoFromAPI struct {
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

//SourceListCommandToAPI весь список источников ->API
//  - ID: уникальный числовой идентификатор источника
//  - ActionType: типа действия над источником
// ('add'/'update'/'delete'/'reconnect'/'status request',
// добавить, обновить, удалить, переподключить, запрос состояния)
//  - Argument: параметры источника, для actionType
// 'delete'/'reconnect'/'status request' это ПОЛЕ ПУСТОЕ
type SourceListCommandToAPI struct {
	ID         int             `json:"id"`
	ActionType string          `json:"actionType"`
	Argument   SourceArguments `json:"argument"`
}

//SourceArguments параметры источников
//  - IP: ip адрес источника
//  - Token: уникальный идентификатор источника
type SourceArguments struct {
	IP    string `json:"ip"`
	Token string `json:"token"`
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
