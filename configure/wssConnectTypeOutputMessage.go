package configure

/*
* Описание типов JSON сообщений отправляемых источникам
*
* Версия 0.14, дата релиза 15.10.2019
* */

//DetailInfoMsgPingPong подробная информация
type DetailInfoMsgPingPong struct {
	EnableTelemetry bool     `json:"enableTelemetry"`
	StorageFolders  []string `json:"storageFolders"`
}

//MsgTypePingPong сообщение типа ping
type MsgTypePingPong struct {
	MsgType string                `json:"messageType"`
	Info    DetailInfoMsgPingPong `json:"info"`
}

/* ПАРАМЕТРЫ ФИЛЬТРАЦИИ */

//MsgTypeFiltrationControl сообщение для запуска процесса фильтрации
type MsgTypeFiltrationControl struct {
	MsgType string                    `json:"messageType"`
	Info    SettingsFiltrationControl `json:"info"`
}

//SettingsFiltrationControl описание параметров для запуска процесса фильтрации
// TaskID - идентификатор задачи
// Command - команда 'start'/'stop'
// IndexIsFound - найдены ли индексы
// CountIndexFiles - количество файлов найденных в результате поиска по индексу
// NumberMessagesFrom - количество сообщений из... например, 1 из 3
// Options - параметры фильтрации, заполняются ТОЛЬКО в сообщении где NumberMessageFrom[0,N]
type SettingsFiltrationControl struct {
	TaskID                 string                                      `json:"id"`
	Command                string                                      `json:"c"`
	Options                FiltrationControlCommonParametersFiltration `json:"o"`
	IndexIsFound           bool                                        `json:"iif"`
	CountIndexFiles        int                                         `json:"cif"`
	NumberMessagesFrom     [2]int                                      `json:"nmf"`
	ListFilesReceivedIndex map[string][]string                         `json:"lfri"`
}
