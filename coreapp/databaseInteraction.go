package coreapp

/*
* Ядро приложения
* Модуль взаимодействия с БД
*
* Версия 0.1, дата релиза 20.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/handlerrequestdb"
)

//DatabaseInteraction обрабатывает запросы БД
func DatabaseInteraction(nameDB string, ism *configure.InformationStoringMemory, linkConnection *configure.MongoDBConnect) {
	fmt.Println("START module 'CoreAppDBInteraction'...")

	/* ____ ЗАПИСЬ ТЕСТОВОЙ КОЛЛЕКЦИИ ____ */
	//--------------------------------------
	qcs := handlerrequestdb.QueryCollectionSources{
		NameDB:         nameDB,
		CollectionName: "sources_list",
		ConnectDB:      linkConnection.Connect,
	}

	listSources := []interface{}{
		configure.InformationAboutSource{9, "127.0.0.1", "fmdif3o444fdf344k0fiif", false},
		configure.InformationAboutSource{10, "192.168.0.10", "fmdif3o444fdf344k0fiif", false},
		configure.InformationAboutSource{11, "192.168.0.11", "ttrr9gr9r9e9f9fadx94", true},
		configure.InformationAboutSource{12, "192.168.0.12", "2n3n3iixcxcc3444xfg0222", true},
		configure.InformationAboutSource{13, "192.168.0.13", "osdsoc9c933cc9cn939f9f33", false},
		configure.InformationAboutSource{14, "192.168.0.14", "hgdfffffff9333ffodffodofff0", false},
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
