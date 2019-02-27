package moduleapiapp

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//MainAppAPI обработчик запросов поступающих через API
func MainAppAPI(appConfig *configure.AppConfig, ism *configure.InformationStoringMemory) (chanOut, chanIn chan configure.MsgBetweenCoreAndAPI) {
	fmt.Println("START module MainAppAPI")

	chanOut = make(chan configure.MsgBetweenCoreAndAPI)
	chanIn = make(chan configure.MsgBetweenCoreAndAPI)

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
