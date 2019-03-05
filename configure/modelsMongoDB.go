package configure

/*
* Описание типов хранящихся в БД
* */

//InformationAboutSource информация об источнике в коллекции 'sources_list'
type InformationAboutSource struct {
	ID            int                 `json:"id" bson:"id"`
	IP            string              `json:"ip" bson:"ip"`
	Token         string              `json:"token" bson:"token"`
	AsServer      bool                `json:"as_server" bson:"as_server"`
	IDClientAPI   string              `json:"id_client_api" bson:"id_client_api"`
	SourceSetting InfoServiceSettings `json:"source_setting" bson:"source_setting"`
}

//InfoServiceSettings содержит настройки источника
type InfoServiceSettings struct {
	EnableTelemetry           bool `json:"enable_telemetry" bson:"enable_telemetry"`
	MaxCountProcessfiltration int  `json:"max_count_process_filtration" bson:"max_count_process_filtration"`
}
