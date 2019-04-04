package configure

/*
* Описание типов JSON сообщений отправляемых источникам
*
* Версия 0.12, дата релиза 04.04.2019
* */

//DetailInfoMsgPingPong подробная информация
type DetailInfoMsgPingPong struct {
	MaxCountProcessFiltration int8     `json:"maxCountProcessFiltration"`
	EnableTelemetry           bool     `json:"enableTelemetry"`
	StorageFolders            []string `json:"storageFolders"`
}

//MsgTypePingPong сообщение типа ping
type MsgTypePingPong struct {
	MsgType string                `json:"messageType"`
	Info    DetailInfoMsgPingPong `json:"info"`
}
