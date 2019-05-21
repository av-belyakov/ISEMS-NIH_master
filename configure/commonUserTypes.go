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

//TypeFiltrationMsgParameters параметры фильтрации файлов
// ID - уникальный цифровой идентификатор источника
// Status - статус задачи 'wait'/'refused'/'execute'/'completed'/'stop' ('ожидает' / 'отклонена' / 'выполняется' / 'завершена' / 'остановлена')
// UseIndex - используется ли индекс для поиска файлов
// CountFilteFiltration - количество файлов подходящих под параметры фильтрации
// SizeFileFiltration — общий размер файлов подходящих под параметры фильтрации
// CountDirectoryFiltartion — количество директорий по которым выполняется фильт.
// CountFilesProcessed — количество обработанных файлов
// NumberFilesFound — количество найденных файлов
// SizeFilesProcessed — общий размер обработанных файлов
// PathStorageSource — путь до директории в которой сохраняются файлы при
type TypeFiltrationMsgParameters struct {
	ID                       int
	Status                   string
	UseIndex                 bool
	CountFilteFiltration     int
	SizeFileFiltration       uint64
	CountDirectoryFiltartion int
	CountFilesProcessed      int
	NumberFilesFound         int
	SizeFilesProcessed       uint64
	PathStorageSource        string
}
