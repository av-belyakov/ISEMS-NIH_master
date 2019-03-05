package configure

/*
* Описание сообщений типа JSON передоваемых между API и клиентами
* */

//MsgType общее сообщение
// MsgType:
//  - 'information'
//  - 'command'
// MsgSection:
//  - 'sensor control'
//  - 'filtration control'
//  - 'download control'
//  - 'information search control'
type MsgType struct {
	MsgType    string `json:"msgType"`
	MsgSection string `json:"msgSection"`
}

//SourceControlMsgTypeInfo информационные сообщения о источникам
type SourceControlMsgTypeInfo struct {
}

//SourceControlMsgTypeCommand командные сообщения по источникам
//MsgInsturction:
//  - 'add new sources'
//  - 'update sources'
//  - 'delete sources'
//  - 'reconnect sources'
type SourceControlMsgTypeCommand struct {
	MsgInsturction string `json:"msgInsturction"`
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
