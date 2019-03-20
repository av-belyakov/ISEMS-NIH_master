package configure

/*
* Описание типа конфигурационных настроек приложения
*
* Версия 0.2, дата релиза 20.03.2019
* */

//

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
	AuthenticationTokenClientAPI            string
	AuthenticationTokenClientsAPI           []SettingsAuthenticationTokenClientsAPI
	ServerAPI                               settingsServerHTTP
	PathCertFile                            string
	PathPrivateKeyFile                      string
	TimeReconnectClient                     int
	ConnectionDB                            settingsConnectionDB
	DirectoryLongTermStorageDownloadedFiles settingsDirectoryStoreFiles
	PathLogFiles                            string
}

/*
//settingsServerHTTP настройки HTTP сервера
type settingsServerHTTP struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	PathCertFile       string `json:"pathCertFile"`
	PathPrivateKeyFile string `json:"pathPrivateKeyFile"`
}

//settingsDBConnection настройки для подключения к БД
type settingsConnectionDB struct {
	Socket         bool   `json:"socket"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	NameDB         string `json:"nameDB"`
	UnixSocketPath string `json:"unixSocketPath"`
}

//settingsDirectoryStoreFiles пути директорий используемых для хранения 'сырых' файлов и объектов
type settingsDirectoryStoreFiles struct {
	Raw    string `json:"raw"`
	Object string `json:"object"`
}

//settingsAuthenticationTokenClientsAPI параметры
type settingsAuthenticationTokenClientsAPI struct {
	Token, Name string
}

//AppConfig хранит настройки из конфигурационного файла приложения
type AppConfig struct {
	VersionApp                              string
	RootDir                                 string
	ServerHTTPS                             settingsServerHTTP `json:"serverHTTPS"`
	AuthenticationTokenClientAPI            string             `json:"authenticationTokenClientAPI"`
	AuthenticationTokenClientsAPI           []settingsAuthenticationTokenClientsAPI
	ServerAPI                               settingsServerHTTP          `json:"serverAPI"`
	PathCertFile                            string                      `json:"pathCertFile"`
	PathPrivateKeyFile                      string                      `json:"pathPrivateKeyFile"`
	TimeReconnectClient                     int                         `json:"timeRecconnectClient"`
	ConnectionDB                            settingsConnectionDB        `json:"connectionDB"`
	DirectoryLongTermStorageDownloadedFiles settingsDirectoryStoreFiles `json:"directoryLongTermStorageDownloadedFiles"`
	PathLogFiles                            string                      `json:"pathLogFiles"`
}
*/
