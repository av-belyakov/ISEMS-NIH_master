package coreapp

/*
* Ядро приложения
* Модуль взаимодействия с БД
*
* Версия 0.1, дата релиза 26.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/handlerrequestdb"
	"ISEMS-NIH_master/savemessageapp"
)

//DatabaseInteraction обрабатывает запросы БД
func DatabaseInteraction(nameDB string, linkConnection *configure.MongoDBConnect, ism *configure.InformationStoringMemory) {
	fmt.Println("START module 'CoreAppDBInteraction'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	qcs := handlerrequestdb.QueryCollectionSources{
		NameDB:         nameDB,
		CollectionName: "sources_list",
		ConnectDB:      linkConnection.Connect,
	}

	/* ____ ЗАПИСЬ ТЕСТОВОЙ КОЛЛЕКЦИИ ____ */
	recordTestSourceList(nameDB, &qcs)

	//при старте подпрограммы получаем список источников
	sourcesList, err := qcs.FindAll()
	if err != nil {
		_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
	}

	fmt.Println("--- sources list ---")
	fmt.Println(sourcesList)

//записиываем настройки источников в память
for _, source := range sourcesList {
	fmt.Printf("%v", source)

 ism.AddSourceSettings(source.IP, configure.ServiceSettings{
	 ID: source.ID,
Token: source.Token,
AsServer: source.AsServer,
MaxCountProcessFilter: source.MaxCountProcessFiltering
 }) {
	
}

}

//запись тестовой коллекции
func recordTestSourceList(nameDB string, qcs *handlerrequestdb.QueryCollectionSources) {
	listSources := []interface{}{
		configure.InformationAboutSource{9, 3, "127.0.0.1", "fmdif3o444fdf344k0fiif", false},
		configure.InformationAboutSource{10, 3, "192.168.0.10", "fmdif3o444fdf344k0fiif", false},
		configure.InformationAboutSource{11, 3, "192.168.0.11", "ttrr9gr9r9e9f9fadx94", true},
		configure.InformationAboutSource{12, 4, "192.168.0.12", "2n3n3iixcxcc3444xfg0222", true},
		configure.InformationAboutSource{13, 5, "192.168.0.13", "osdsoc9c933cc9cn939f9f33", false},
		configure.InformationAboutSource{14, 3, "192.168.0.14", "hgdfffffff9333ffodffodofff0", false},
	}

	fmt.Printf("%T%v", listSources, listSources)
	fmt.Println("...\n")

	res, err := qcs.InserListSourcesTMP(listSources)
	if err != nil {
		fmt.Println(err)
	}

	isSeccess := "NO"
	if res {
		isSeccess = "YES"
	}

	fmt.Println("\vInser data is ", isSeccess)
}
