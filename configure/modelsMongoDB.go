package configure

/*
* Описание типов коллекций хранящихся в БД
* */

//InformationAboutSource информация об источнике в коллекции 'sources_list'
type InformationAboutSource struct {
	ID            int                 `json:"id" bson:"id"`
	IP            string              `json:"ip" bson:"ip"`
	Token         string              `json:"token" bson:"token"`
	ShortName     string              `json:"short_name" bson:"short_name"`
	Description   string              `json:"description" bson:"description"`
	AsServer      bool                `json:"as_server" bson:"as_server"`
	NameClientAPI string              `json:"name_client_api" bson:"name_client_api"`
	SourceSetting InfoServiceSettings `json:"source_setting" bson:"source_setting"`
}

//InfoServiceSettings содержит настройки источника
type InfoServiceSettings struct {
	EnableTelemetry           bool     `json:"enable_telemetry" bson:"enable_telemetry"`
	MaxCountProcessFiltration int8     `json:"max_count_process_filtration" bson:"max_count_process_filtration"`
	StorageFolders            []string `json:"storage_folders" bson:"storage_folders"`
}
