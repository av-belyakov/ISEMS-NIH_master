package configure

/*
* Описание типа конфигурационных настроек приложения
* */

//settingsServerHTTP настройки HTTP сервера
type settingsServerHTTP struct {
	Host               string
	Port               int
	PathCertFile       string
	PathPrivateKeyFile string
}

//settingsDBConnection настройки для подключения к БД
type settingsConnectionDB struct {
	Socket         bool
	Host           string
	Port           int
	User           string
	Password       string
	NameDB         string
	UnixSocketPath string
}

//settingsDirectoryStoreFiles пути директорий используемых для хранения 'сырых' файлов и объектов
type settingsDirectoryStoreFiles struct {
	Raw    string
	Object string
}

//SettingsAuthenticationTokenClientsAPI параметры
type SettingsAuthenticationTokenClientsAPI struct {
	Token, Name string
}

//AppConfig хранит настройки из конфигурационного файла приложения
// VersionApp - версия приложения
// RootDir - корневая директория приложения
// ServerHTTPS - параметры для запуска сервера к которому подключаются сенсора
// AuthenticationTokenClientsAPI - идентификационные токены для подключения клиентов
// ServerAPI - параметры запуска сервера для взаимодействия с клиентом API
// PathRootCA - путь до ключа с корневым сертификатом
// TimeReconnectClient - актуально только в режиме isServer = false, тогда с заданным интервалом времени будут попытки соединения с адресами slave
// ConnectionDB - настройки для доступа к БД
// DirectoryLongTermStorageDownloadedFiles - директория для долговременного хранения скаченных файлов
// MaximumTotalSizeFilesDownloadedAutomatically - максимальный, общий размер файлов скачиваемых в автоматическом режиме (в Мб)
// PathLogFiles - место расположение лог-файла приложения
type AppConfig struct {
	VersionApp                                   string
	RootDir                                      string
	ServerHTTPS                                  settingsServerHTTP
	AuthenticationTokenClientsAPI                []SettingsAuthenticationTokenClientsAPI
	ServerAPI                                    settingsServerHTTP
	PathRootCA                                   string
	TimeReconnectClient                          int
	ConnectionDB                                 settingsConnectionDB
	DirectoryLongTermStorageDownloadedFiles      settingsDirectoryStoreFiles
	MaximumTotalSizeFilesDownloadedAutomatically int64
	PathLogFiles                                 string
}
