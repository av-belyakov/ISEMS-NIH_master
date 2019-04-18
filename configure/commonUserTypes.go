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
	IndexData       map[string]string
}
