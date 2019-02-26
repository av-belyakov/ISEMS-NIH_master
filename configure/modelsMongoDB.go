package configure

/*
* Описание типов хранящихся в БД
* */

//InformationAboutSource информация об источнике в коллекции 'sources_list'
type InformationAboutSource struct {
	ID, MaxCountProcessFiltering int
	IP, Token                    string
	AsServer                     bool
}
