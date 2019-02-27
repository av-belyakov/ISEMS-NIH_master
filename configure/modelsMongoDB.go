package configure

/*
* Описание типов хранящихся в БД
* */

//InformationAboutSource информация об источнике в коллекции 'sources_list'
type InformationAboutSource struct {
	ID            int
	IP            string
	Token         string
	AsServer      bool
	SourceSetting InfoServiceSettings
}

//InfoServiceSettings содержит настройки источника
type InfoServiceSettings struct {
	EnableTelemetry          bool
	MaxCountProcessFiltering int
}
