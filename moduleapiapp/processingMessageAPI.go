package handlermessageapi

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//ProcessingMessageAPI обработчик запросов поступающих через API
func ProcessingMessageAPI(appConfig *configure.AppConfig, ism *configure.InformationStoringMemory) {
	fmt.Println("START func ProcessingMessageAPI")
	/*
	   if message := <-*ism.ChannelCollection.ChanMessageToAPI {
	   	*ism.ChannelCollection.ChanMessageFromAPI<- configure.MessageAPI{
	   		MsgID: "2",
	   		MsgType: "response",
	   		MsgDate: 838283,
	   	}
	   }
	*/
	fmt.Println("MESSAGE TO API:", <-*ism.ChannelCollection.ChanMessageToAPI)
}
