package configure

/*
* Коллекция типов сообщений
*
* Версия 0.1, дата релиза 05.03.2019
* */

/*
//  - change_status_source (info)
//  - source_telemetry (info)
//  - filtration (info)
//  - download (info)
//  - information_search_results (info)
//  - error_notification (info)
//  - source_control (command)
//  - filtration (command)
//  - download (command)
//  - information_search (command)
*/

/* --- ИНФОРМАЦИОННЫЕ --- */

//MsgInfoChangeStatusSource изменение статуса источника:
//  - передача нового списка источников
//  - изменение состояния источника
type MsgInfoChangeStatusSource struct {
	SourceListIsExist bool
	SourceList        []MainOperatingParametersSource
}

//MsgInfoSourceTelemetry данные по телеметрии
type MsgInfoSourceTelemetry struct {
}

//MsgInfofiltration информация о фильтрации
//
type MsgInfofiltration struct {
}

//MsgInfoDownload информация о скачивании файлов
//
type MsgInfoDownload struct {
}

//MsgInfoErrorNotification сообщения об ошибках
//
type MsgInfoErrorNotification struct {
}

//MsgInformationSearchResults информация о результатах поиска
//
type MsgInformationSearchResults struct {
}

/* --- КОМАНДНЫЕ --- */

//MsgCommandSourceControl команды по источникам:
//  - получить список источников
//  - добавлен новый источник
//  - удален источник
//  - настройки источника изменены
//  - выполнить переподключение источника
type MsgCommandSourceControl struct {
	ListRequest bool //получить список источников
	Sources     []SourceParameter
}

//SourceParameter параметры источника
type SourceParameter struct {
	SourceID          int
	ToReconnectSource bool   //true переподключить источник
	OccurredEvent     string //произошедшие события (добавлен, удален, настройки изменены)
	//add, delete, update
	NewCharacteristic MainOperatingParametersSource
}

//MainOperatingParametersSource параметры для подключения источника
type MainOperatingParametersSource struct {
	ID        int
	IP, Token string
	AsServer  bool
	Options   SourceDetailedInformation
}

//SourceDetailedInformation детальная информация по настройке источника
type SourceDetailedInformation struct {
	EnableTelemetry          bool //true - включить
	MaxCountProcessFiltering int
}

//MsgCommandfiltration команды по фильтрации:
//  - начать фильтрацию
//  - обработка задачи
//  - остановить фильтрацию
//  - фильтрация начата/отклонена/остановленна
type MsgCommandfiltration struct {
}

//MsgCommandDownload команды по скачиванию файлов
//  - начать выгрузку файлов
//  - остановить выгрузку
//  - возобновить выгрузку
type MsgCommandDownload struct {
}

//MsgCommandInformationSearch команды для поиска информации о задачах
//  - выполнить поиск задачи по заданным параметрам
type MsgCommandInformationSearch struct {
}
