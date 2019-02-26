package moduleapiapp

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//MainAppAPI обработчик запросов поступающих через API
func MainAppAPI(cOut chan<- configure.MessageAPI, appConfig *configure.AppConfig, ism *configure.InformationStoringMemory, cIn <-chan configure.MessageAPI) {
	fmt.Println("START module MainAppAPI")
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
}
