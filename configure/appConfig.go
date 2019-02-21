package configure

/*
* Описание типа конфигурационных настроек приложения
*
* Версия 0.1, дата релиза 18.02.2019
* */

//settingsServerHTTP хранит настройки HTTP сервера
type settingsServerHTTP struct {
	Host string `json:"host"`
	Port int    `josn:"port"`
}

//settingsDBConnection хранит настройки для подключения к БД
type settingsConnectionDB struct {
	Socket         bool   `json:"socket"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	NameDB         string `json:"nameDB"`
	UnixSocketPath string `json:"unixSocketPath"`
}

//settingsDirectoryStoreFiles хранит пути директорий используемых для хранения 'сырых' файлов и объектов
type settingsDirectoryStoreFiles struct {
	Raw    string `json:"raw"`
	Object string `json:"object"`
}

//AppConfig хранит настройки из конфигурационного файла приложения
type AppConfig struct {
	VersionApp                              string
	RootDir                                 string
	AuthenticationToken                     string                      `json:"authenticationToken"`
	ServerHTTP                              settingsServerHTTP          `json:"serverHTTP"`
	PathCertFile                            string                      `json:"pathCertFile"`
	PathPrivateKeyFile                      string                      `json:"pathKeyFile"`
	TimeReconnectClient                     int                         `json:"timeRecconnectClient"`
	ConnectionDB                            settingsConnectionDB        `json:"connectionDB"`
	DirectoryLongTermStorageDownloadedFiles settingsDirectoryStoreFiles `json:"directoryLongTermStorageDownloadedFiles"`
	PathLogFiles                            string                      `json:"pathLogFiles"`
}
