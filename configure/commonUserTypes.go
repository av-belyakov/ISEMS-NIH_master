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
type TypeFiltrationMsgFoundFileInformationAndTaskStatus struct {
	TaskStatus    string
	ListFoundFile map[string]*FoundFilesInformation
}

//TypeGetInfoTaskFromMarkTaskCompleteProcess описание типа при обработке задачи по маркированию задачи как завершенная
type TypeGetInfoTaskFromMarkTaskCompleteProcess struct {
	UserName             string
	Description          string
	FiltrationTaskStatus bool
	FilesDownloaded      bool
}
