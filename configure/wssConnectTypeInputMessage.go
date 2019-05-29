package configure

/*
* Описание типов JSON сообщений принимаемых от источников
*
* Версия 0.23, дата релиза 27.05.2019
* */

//SourceTelemetry полная системная информация подготовленная к отправке
type SourceTelemetry struct {
	MessageType string               `json:"messageType"`
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

//MsgTypeFiltration сообщение типа ping
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

/*
//ChunkListParameters хранит набор параметров для разделения среза имен файлов на отдельные части
type ChunkListParameters struct {
	NumPart, CountParts, SizeChunk int
	ListFoundFiles                 []FoundFilesInfo
}

//FilterInfoPattern является шаблоном типа Info
type FilterInfoPattern struct {
	Processing string `json:"processing"`
	TaskIndex  string `json:"taskIndex"`
	IPAddress  string `json:"ipAddress"`
}

//FilterCountPattern шаблон для частей учета некоторого количества
type FilterCountPattern struct {
	CountCycleComplete    int   `json:"countCycleComplete"`
	CountFilesFound       int   `json:"countFilesFound"`
	CountFoundFilesSize   int64 `json:"countFoundFilesSize"`
	CountFilesProcessed   int   `json:"countFilesProcessed"`
	CountFilesUnprocessed int   `json:"countFilesUnprocessed"`
}

//InfoProcessingFile информация об обработанном файле
type InfoProcessingFile struct {
	FileName          string `json:"fileName"`
	DirectoryLocation string `json:"directoryLocation"`
	StatusProcessed   bool   `json:"statusProcessed"`
}

//MessageTypefiltrationStopInfo сообщение при ОСТАНОВ выполнения фильтрации
type MessageTypefiltrationStopInfo struct {
	FilterInfoPattern
}

//MessageTypefiltrationCompleteInfoFirstPart детальная информация при ЗАВЕРШЕНИИ выполнения фильтрации (первая часть)
type MessageTypefiltrationCompleteInfoFirstPart struct {
	FilterInfoPattern
	FilterCountPattern
	NumberMessageParts [2]int `json:"numberMessageParts"`
}

//MessageTypefiltrationCompleteInfoSecondPart информация при ЗАВЕРШЕНИИ выполнения фильтрации (вторая часть)
type MessageTypefiltrationCompleteInfoSecondPart struct {
	FilterInfoPattern
	NumberMessageParts             [2]int           `json:"numberMessageParts"`
	ListFilesFoundDuringfiltration []FoundFilesInfo `json:"listFilesFoundDuringfiltration"`
}

//MessageTypefiltrationStartInfoFirstPart детальная информаци, первый фрагмент (без имен файлов)
type MessageTypefiltrationStartInfoFirstPart struct {
	FilterInfoPattern
	Directoryfiltration      string         `json:"directoryfiltration"`
	CountDirectoryfiltration int            `json:"countDirectoryfiltration"`
	CountFullCycle           int            `json:"countFullCycle"`
	CountFilesfiltration     int            `json:"countFilesfiltration"`
	CountMaxFilesSize        int64          `json:"countMaxFilesSize"`
	UseIndexes               bool           `json:"useIndexes"`
	NumberMessageParts       [2]int         `json:"numberMessageParts"`
	ListCountFilesFilter     map[string]int `json:"listCountFilesFilter"`
}

//MessageTypefiltrationStartInfoSecondPart детальная информация с именами файлов
type MessageTypefiltrationStartInfoSecondPart struct {
	FilterInfoPattern
	UseIndexes         bool                `json:"useIndexes"`
	NumberMessageParts [2]int              `json:"numberMessageParts"`
	ListFilesFilter    map[string][]string `json:"listFilesFilter"`
}

//MessageTypefiltrationExecuteOrUnexecuteInfo детальная информация при выполнении или не выполнении фильтрации
type MessageTypefiltrationExecuteOrUnexecuteInfo struct {
	FilterInfoPattern
	FilterCountPattern
	InfoProcessingFile `json:"infoProcessingFile"`
}

//MessageTypefiltrationStartFirstPart при начале фильтрации (первая часть)
type MessageTypefiltrationStartFirstPart struct {
	MessageType string                                  `json:"messageType"`
	Info        MessageTypefiltrationStartInfoFirstPart `json:"info"`
}

//MessageTypefiltrationStartSecondPart при начале фильтрации (первая часть)
type MessageTypefiltrationStartSecondPart struct {
	MessageType string                                   `json:"messageType"`
	Info        MessageTypefiltrationStartInfoSecondPart `json:"info"`
}

//MessageTypefiltrationStop отправляется для подтверждения остановки фильтрации
type MessageTypefiltrationStop struct {
	MessageType string                        `json:"messageType"`
	Info        MessageTypefiltrationStopInfo `json:"info"`
}

//MessageTypefiltrationCompleteFirstPart отправляется при завершении фильтрации
type MessageTypefiltrationCompleteFirstPart struct {
	MessageType string                                     `json:"messageType"`
	Info        MessageTypefiltrationCompleteInfoFirstPart `json:"info"`
}

//MessageTypefiltrationCompleteSecondPart отправляется при завершении фильтрации
type MessageTypefiltrationCompleteSecondPart struct {
	MessageType string                                      `json:"messageType"`
	Info        MessageTypefiltrationCompleteInfoSecondPart `json:"info"`
}

//MessageTypefiltrationExecutedOrUnexecuted при выполнении или не выполнении фильтрации
type MessageTypefiltrationExecutedOrUnexecuted struct {
	MessageType string                                      `json:"messageType"`
	Info        MessageTypefiltrationExecuteOrUnexecuteInfo `json:"info"`
}

//MessageTypeDownloadFilesInfoReadyOrCompleted содержит информацию передоваемую при сообщениях о готовности или завершении передачи
type MessageTypeDownloadFilesInfoReadyOrCompleted struct {
	Processing string `json:"processing"`
	TaskIndex  string `json:"taskIndex"`
}

//MessageTypeDownloadFilesInfoExecute содержит информацию передоваемую при сообщениях о передаче информации о файле
type MessageTypeDownloadFilesInfoExecute struct {
	MessageTypeDownloadFilesInfoReadyOrCompleted
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	FileHash string `json:"fileHash"`
}

//MessageTypeDownloadFilesReadyOrCompleted применяется для отправки сообщений о готовности или завершении передачи
type MessageTypeDownloadFilesReadyOrCompleted struct {
	MessageType string                                       `json:"messageType"`
	Info        MessageTypeDownloadFilesInfoReadyOrCompleted `json:"info"`
}

//MessageTypeDownloadFilesExecute применяется для отправки сообщений о передаче файлов
type MessageTypeDownloadFilesExecute struct {
	MessageType string                              `json:"messageType"`
	Info        MessageTypeDownloadFilesInfoExecute `json:"info"`
}
*/
