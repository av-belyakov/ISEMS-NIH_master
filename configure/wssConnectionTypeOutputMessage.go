package configure

/*
* Описание типов JSON сообщений отправляемых источникам
*
* Версия 0.1, дата релиза 27.02.2019
* */

//DetailInfoMsgPingPong подробная информация
type DetailInfoMsgPingPong struct {
	MaxCountProcessFiltration int8 `json:"maxCountProcessFiltration"`
}

//MsgTypePingPong сообщение типа ping
type MsgTypePingPong struct {
	MsgType string                `json:"messageType"`
	Info    DetailInfoMsgPingPong `json:"info"`
}
