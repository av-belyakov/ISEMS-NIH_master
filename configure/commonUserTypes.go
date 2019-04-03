package configure

/*
* Описание общих пльзовательский типов
* */

//SettingsChangeConnectionStatusSource тип с информацией об источнике изменившем
// свое состояние подключения
type SettingsChangeConnectionStatusSource struct {
	ID     int
	Status string
}
