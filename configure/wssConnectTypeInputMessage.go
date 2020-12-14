package configure

/*
* Описание типов JSON сообщений принимаемых от источников
* */

//MsgTypePong сообщение типа pong
type MsgTypePong struct {
	MsgType string            `json:"messageType"`
	Info    DetailInfoMsgPong `json:"info"`
}

//DetailInfoMsgPong подробная информация
// AppVersion - версия приложения
// AppReleaseDate - дата релиза версии приложения
type DetailInfoMsgPong struct {
	AppVersion     string `json:"av"`
	AppReleaseDate string `json:"ard"`
}

//SourceTelemetry полная системная информация подготовленная к отправке
type SourceTelemetry struct {
	MessageType string               `json:"messageType"`
	TaskID      string               `json:"taskID"`
	Info        TelemetryInformation `json:"info"`
}

//TelemetryInformation системная информация
type TelemetryInformation struct {
	CurrentDateTime    int64                     `json:"currentDateTime"`
	DiskSpace          []map[string]string       `json:"diskSpace"`
	TimeInterval       map[string]map[string]int `json:"timeInterval"`
	RandomAccessMemory Memory                    `json:"randomAccessMemory"`
	LoadCPU            float64                   `json:"loadCPU"`
	LoadNetwork        map[string]map[string]int `json:"loadNetwork"`
}

//Memory содержит информацию об используемой ПО
type Memory struct {
	Total int `json:"total"`
	Used  int `json:"used"`
	Free  int `json:"free"`
}

//CurrentDisk начальное и конечное время для файлов сет. трафика
type CurrentDisk struct {
	DateMin int `json:"dateMin"`
	DateMax int `json:"dateMax"`
}

//FoundFilesInfo содержит информацию о файле
type FoundFilesInfo struct {
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
}

//MsgTypeNotification информационное сообщение
type MsgTypeNotification struct {
	MsgType string                    `json:"messageType"`
	Info    DetailInfoMsgNotification `json:"info"`
}

//DetailInfoMsgNotification информационное сообщение, подробная информация
// TaskID - id задачи
// Section - раздел обработки заданий
// TypeActionPerformed - тип выполняемого действия
// CriticalityMessage - тип сообщения ('info'/'success'/'warning'/'danger')
// Description - описание сообщения
type DetailInfoMsgNotification struct {
	TaskID              string `json:"tid"`
	Section             string `json:"s"`
	TypeActionPerformed string `json:"tap"`
	CriticalityMessage  string `json:"cm"`
	Description         string `json:"d"`
}

//MsgTypeError сообщение отправляемое при возникновении ошибки
type MsgTypeError struct {
	MsgType string             `json:"messageType"`
	Info    DetailInfoMsgError `json:"info"`
}

//DetailInfoMsgError детальное описание ошибки
// TaskID - id задачи
// ErrorName - наименование ошибки
// ErrorDescription - детальное описание ошибки
type DetailInfoMsgError struct {
	TaskID           string `json:"tid"`
	ErrorName        string `json:"en"`
	ErrorDescription string `json:"ed"`
}

//MsgTypeFiltration сообщение типа 'filtration'
type MsgTypeFiltration struct {
	MsgType string                  `json:"messageType"`
	Info    DetailInfoMsgFiltration `json:"info"`
}

//DetailInfoMsgFiltration подробная информация
// TaskID - ID задачи
// TaskStatus - статус выполняемой задачи
// NumberFilesMeetFilterParameters - кол-во файлов удовлетворяющих параметрам фильтрации
// NumberProcessedFiles - кол-во обработанных файлов
// NumberFilesFoundResultFiltering - кол-во найденных, в результате фильтрации, файлов
// NumberDirectoryFiltartion - кол-во директорий по которым выполняется фильтрация
// NumberErrorProcessedFiles - кол-во не обработанных файлов или файлов обработанных с ошибками
// SizeFilesMeetFilterParameters - общий размер файлов (в байтах) удовлетворяющих параметрам фильтрации
// SizeFilesFoundResultFiltering - общий размер найденных, в результате фильтрации, файлов (в байтах)
// PathStorageSource — путь до директории в которой сохраняются файлы при
// NumberMessagesParts - коичество частей сообщения
// FoundFilesInformation - информация о файлах, ключ - имя файла
type DetailInfoMsgFiltration struct {
	TaskID                          string                            `json:"tid"`
	TaskStatus                      string                            `json:"ts"`
	NumberFilesMeetFilterParameters int                               `json:"nfmfp"`
	NumberProcessedFiles            int                               `json:"npf"`
	NumberFilesFoundResultFiltering int                               `json:"nffrf"`
	NumberDirectoryFiltartion       int                               `json:"ndf"`
	NumberErrorProcessedFiles       int                               `json:"nepf"`
	SizeFilesMeetFilterParameters   int64                             `json:"sfmfp"`
	SizeFilesFoundResultFiltering   int64                             `json:"sffrf"`
	PathStorageSource               string                            `json:"pss"`
	NumberMessagesParts             [2]int                            `json:"nmp"`
	FoundFilesInformation           map[string]*InputFilesInformation `json:"ffi"`
}

//InputFilesInformation подробная информация о файлах
// Size - размер файла
// Hex - хеш сумма файла
type InputFilesInformation struct {
	Size int64  `json:"s"`
	Hex  string `json:"h"`
}

//MsgTypeDownload сообщение типа 'download'
type MsgTypeDownload struct {
	MsgType string                `json:"messageType"`
	Info    DetailInfoMsgDownload `json:"info"`
}

//DetailInfoMsgDownload подробная информация
// TaskID - ID задачи
// Command - статус выполняемой задачи
//  - 'give me the file' (master -> slave), запрос файла
//  - 'ready to receive file' (master -> salve), подтверждение готовности приема файла
//  - 'ready for the transfer' (slave -> master), подтверждение готовности передачи
//  - 'file transfer not possible' (slave -> master), сообщение о невозможности передачи
//  - 'file transfer complete' (slave -> master), сообщение о завершении передачи
// PathDirStorage - директория в которой хранятся файлы на источнике
// FileOptions - параметры файла
type DetailInfoMsgDownload struct {
	TaskID         string              `json:"tid"`
	Command        string              `json:"c"`
	PathDirStorage string              `json:"pds"`
	FileOptions    DownloadFileOptions `json:"fo"`
}

//DownloadFileOptions параметры загружаемого файла
// Name - название файла
// Size - размер файла
// Hex - контрольная сумма файла
// NumChunk - кол-во передаваемых кусочков
// ChunkSize - размер передаваемого кусочка
type DownloadFileOptions struct {
	Name      string `json:"n"`
	Size      int64  `json:"sz"`
	Hex       string `json:"hex"`
	NumChunk  int    `json:"nc"`
	ChunkSize int    `json:"cs"`
}
