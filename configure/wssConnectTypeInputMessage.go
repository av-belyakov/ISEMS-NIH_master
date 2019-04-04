package configure

/*
* Описание типов JSON сообщений принимаемых от источников
*
* Версия 0.2, дата релиза 04.04.2019
* */

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

//DetailedInformation системная информация
type DetailedInformation struct {
	CurrentDateTime    int64                     `json:"currentDateTime"`
	DiskSpace          []map[string]string       `json:"diskSpace"`
	TimeInterval       map[string]map[string]int `json:"timeInterval"`
	RandomAccessMemory Memory                    `json:"randomAccessMemory"`
	LoadCPU            float64                   `json:"loadCPU"`
	LoadNetwork        map[string]map[string]int `json:"loadNetwork"`
}

//SysInfo полная системная информация подготовленная к отправке
type SysInfo struct {
	MessageType string              `json:"messageType"`
	Info        DetailedInformation `json:"info"`
}

//FoundFilesInfo содержит информацию о файле
type FoundFilesInfo struct {
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
}

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
