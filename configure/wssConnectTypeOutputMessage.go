package configure

/*
* Описание типов JSON сообщений отправляемых источникам
*
* Версия 0.11, дата релиза 03.04.2019
* */

//DetailInfoMsgPingPong подробная информация
type DetailInfoMsgPingPong struct {
	MaxCountProcessFiltration int8 `json:"maxCountProcessFiltration"`
	EnableTelemetry           bool `json:"enableTelemetry"`
}

//MsgTypePingPong сообщение типа ping
type MsgTypePingPong struct {
	MsgType string                `json:"messageType"`
	Info    DetailInfoMsgPingPong `json:"info"`
}
