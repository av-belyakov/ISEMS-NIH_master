package configure

/*
* Описание типов JSON сообщений отправляемых источникам
* */

//MsgTypePing сообщение типа ping
type MsgTypePing struct {
	MsgType string            `json:"messageType"`
	Info    DetailInfoMsgPing `json:"info"`
}

//DetailInfoMsgPing подробная информация
// EnableTelemetry - включить телеметрию
// StorageFolders - директории для хранения файлов
// TypeAreaNetwork - тип протокола канального уровня (ip/pppoe)
type DetailInfoMsgPing struct {
	EnableTelemetry bool     `json:"enableTelemetry"`
	StorageFolders  []string `json:"storageFolders"`
	TypeAreaNetwork string   `json:"typeAreaNetwork"`
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

/* ЗАПРОС НА ПОЛУЧЕНИЕ ТЕЛЕМЕТРИИ */

//MsgTypeTelemetryControl сообщение для получения телеметрии
type MsgTypeTelemetryControl struct {
	MsgType string                          `json:"messageType"`
	Info    SettingsTelemetryControlRequest `json:"info"`
}

//SettingsTelemetryControlRequest параметры запроса телеметрии
type SettingsTelemetryControlRequest struct {
	TaskID  string `json:"id"`
	Command string `json:"c"`
}
