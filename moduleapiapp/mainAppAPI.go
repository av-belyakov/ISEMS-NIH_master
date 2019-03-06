package moduleapiapp

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//MainAppAPI обработчик запросов поступающих через API
func MainAppAPI(appConfig *configure.AppConfig, ism *configure.InformationStoringMemory) (chanOut, chanIn chan configure.MsgBetweenCoreAndAPI) {
	fmt.Println("START module 'MainAppAPI'...")

	chanOut = make(chan configure.MsgBetweenCoreAndAPI, 10)
	chanIn = make(chan configure.MsgBetweenCoreAndAPI, 10)

	/*
			СОЗДАНИЕ СЕРВЕРА WSS ДЛЯ ПОДКЛЮЧЕНИЙ КЛИЕНТОВ

			при подключении клиента запрашиваем у него новый список источников
		------------------------------------------------------------------------
	*/

	/* ПОКА ПРОСТО ТЕСТОВОЕ СООБЩЕНИЕ С НОВЫМ СПИСКОМ ИСТОЧНИКОВ */
	// --- ТЕСТОВЫЙ ОТВЕТ ---
	chanIn <- configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "API module",
		MsgRecipient: "Core module",
		IDClientAPI:  "du68whfh733hjf9393",
MsgJSON: 



/*		
ПОДГОТОВИТЬ ТЕСТОВЫЙ JSON КОТОРЫЙ ДОБАВЛЯЕТСЯ В MsgJSON


ОТПРАВЛЯТЬ СООБЩЕНИЕ С ПРЕКЛКПЛЕННОМ К НЕМУ JSON данными полученными 
от клиента API

MsgType:      "information",
		MsgSection:   "source control",
		IDClientAPI:  "du68whfh733hjf9393",
		AdvancedOptions: configure.MsgInfoChangeStatusSource{
			SourceListIsExist: true,
			SourceList: []configure.MainOperatingParametersSource{
				{9, "127.0.0.1", "fmdif3o444fdf344k0fiif", false, configure.SourceDetailedInformation{}},
				{10, "192.168.0.10", "fmdif3o444fdf344k0fiif", false, configure.SourceDetailedInformation{}},
				{11, "192.168.0.11", "ttrr9gr9r9e9f9fadx94", false, configure.SourceDetailedInformation{}},
				{12, "192.168.0.12", "2n3n3iixcxcc3444xfg0222", false, configure.SourceDetailedInformation{}},
				{13, "192.168.0.13", "osdsoc9c933cc9cn939f9f33", true, configure.SourceDetailedInformation{}},
				{14, "192.168.0.14", "hgdfffffff9333ffodffodofff0", true, configure.SourceDetailedInformation{}},
			},
		},
	}*/
	//------------------------

	//запуск маршрутизатора сообщений от ядра
	go RouteCoreMessage(chanIn, chanOut)

	/*
	   if message := <-*ism.ChannelCollection.ChanMessageToAPI {
	   	*ism.ChannelCollection.ChanMessageFromAPI<- configure.MessageAPI{
	   		MsgID: "2",
	   		MsgType: "response",
	   		MsgDate: 838283,
	   	}
	   }
	   	fmt.Println("MESSAGE TO API:", <-*ism.ChannelCollection.ChanMessageToAPI)
	*/

	return chanOut, chanIn
}
