package configure

/*
* Описание типа конфигурационных настроек приложения
*
* Версия 0.2, дата релиза 20.03.2019
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
type AppConfig struct {
	VersionApp                              string
	RootDir                                 string
	ServerHTTPS                             settingsServerHTTP
	AuthenticationTokenClientsAPI           []SettingsAuthenticationTokenClientsAPI
	ServerAPI                               settingsServerHTTP
	PathRootCA                              string
	TimeReconnectClient                     int
	ConnectionDB                            settingsConnectionDB
	DirectoryLongTermStorageDownloadedFiles settingsDirectoryStoreFiles
	PathLogFiles                            string
}
