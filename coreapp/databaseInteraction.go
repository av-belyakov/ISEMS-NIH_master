package coreapp

/*
* Ядро приложения
* Модуль взаимодействия с БД
*
* Версия 0.2, дата релиза 27.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/handlerrequestdb"
	"ISEMS-NIH_master/savemessageapp"
)

//DatabaseInteraction обрабатывает запросы БД
func DatabaseInteraction(nameDB string, linkConnection *configure.MongoDBConnect, ism *configure.InformationStoringMemory) (chanOutput, chanInput chan configure.MsgBetweenCoreAndDB) {
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

	//при старте получаем список источников
	sourcesList, err := qcs.FindAll()
	if err != nil {
		_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
	}

	fmt.Println("--- sources list ---")
	fmt.Println(sourcesList)

	//записиываем настройки источников в память
	for _, source := range sourcesList {
		fmt.Printf("%v", source)

		/*ism.AddSourceSettings(source.IP, configure.SourceSetting{
			ID:       source.ID,
			Token:    source.Token,
			AsServer: source.AsServer,
			Settings: configure.SourceServiceSettings{
				MaxCountProcessFilter: source.SourceSetting.MaxCountProcessFiltering,
			},
		})*/
	}

	//обработка запросов к БД приходящих из CoreApp
	go func() {
		for msg := range chanInput {
			fmt.Println("resived message from CoreApp to BD")
			fmt.Println(msg)

			go wrapperFunc(chanOutput, msg)
		}
	}()

	return chanOutput, chanInput
}

func wrapperFunc(chanOut chan<- configure.MsgBetweenCoreAndDB, msg configure.MsgBetweenCoreAndDB) {

}

//ЗАПИСЬ ТЕСТОВОЙ КОЛЛЕКЦИИ
func recordTestSourceList(nameDB string, qcs *handlerrequestdb.QueryCollectionSources) {
	listSources := []configure.InformationAboutSource{
		{9, "127.0.0.1", "fmdif3o444fdf344k0fiif", false, configure.InfoServiceSettings{false, 3}},
		{10, "192.168.0.10", "fmdif3o444fdf344k0fiif", false, configure.InfoServiceSettings{false, 3}},
		{11, "192.168.0.11", "ttrr9gr9r9e9f9fadx94", true, configure.InfoServiceSettings{false, 3}},
		{12, "192.168.0.12", "2n3n3iixcxcc3444xfg0222", true, configure.InfoServiceSettings{false, 3}},
		{13, "192.168.0.13", "osdsoc9c933cc9cn939f9f33", false, configure.InfoServiceSettings{false, 3}},
		{14, "192.168.0.14", "hgdfffffff9333ffodffodofff0", false, configure.InfoServiceSettings{false, 3}},
	}

	fmt.Printf("%T%v\n", listSources, listSources)

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
