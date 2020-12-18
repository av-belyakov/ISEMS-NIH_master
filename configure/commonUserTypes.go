package configure

/*
* Описание общих пользовательских типов
* */

//SettingsChangeConnectionStatusSource тип с информацией об источнике изменившем
// свое состояние подключения
type SettingsChangeConnectionStatusSource struct {
	ID     int
	Status string
}

//TypeFiltrationMsgFoundIndex тип 'фильтрация', информация о найденных индексах
type TypeFiltrationMsgFoundIndex struct {
	FilteringOption FiltrationControlCommonParametersFiltration
	IndexIsFound    bool
	IndexData       map[string][]string
}

//TypeFiltrationMsgFoundFileInformationAndTaskStatus тип 'фильтрация', информация о найденных файлах и статусе задачи
// TaskStatus - статус задачи
// ListFoundFile - список найденных файлов
// NumberFilesMeetFilterParameters - количество файлов подходящих под заданные условия
// NumberProcessedFiles - количество обработанных файлов
// NumberFilesFoundResultFiltering - количество файлов найденных в результате фильтрации
// NumberDirectoryFiltartion - количество директорий подлежащих фильтрации
// NumberErrorProcessedFiles - количество файлов обработанных с ошибкой
// SizeFilesMeetFilterParameters - общий размер файлов подходящих под заданные условия
// SizeFilesFoundResultFiltering - общий размер найденных файлов
// PathStorageSource - директория для хранения файлов
type TypeFiltrationMsgFoundFileInformationAndTaskStatus struct {
	TaskStatus                      string
	ListFoundFile                   map[string]DetailedFilesInformation
	NumberFilesMeetFilterParameters int
	NumberProcessedFiles            int
	NumberFilesFoundResultFiltering int
	NumberDirectoryFiltartion       int
	NumberErrorProcessedFiles       int
	SizeFilesMeetFilterParameters   int64
	SizeFilesFoundResultFiltering   int64
	PathStorageSource               string
}

//TypeGetInfoTaskFromMarkTaskCompleteProcess описание типа при обработке задачи по маркированию задачи как завершенная
// TaskIsExist - найдена ли информация по задаче
// UserName - имя пользователя полученное от клиента API
// Description - общее описание причины закрытия задачи
// FiltrationTaskStatus - была ли выполнена задача по фильтрации сет. трафика
// FilesDownloaded - выполнялась ли загрузка хотя бы одного файла
type TypeGetInfoTaskFromMarkTaskCompleteProcess struct {
	TaskIsExist          bool
	UserName             string
	Description          string
	FiltrationTaskStatus bool
	FilesDownloaded      bool
}
