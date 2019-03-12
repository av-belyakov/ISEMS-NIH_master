package handlerslist

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//HandlerMsgFromAPI обработчик сообщений приходящих от модуля API
func HandlerMsgFromAPI(chanToDB chan<- configure.MsgBetweenCoreAndDB, clientID, msgGeneration string, msgCommon *configure.MsgCommon, chanToNI chan<- configure.MsgBetweenCoreAndNI) {

	fmt.Println("START function 'HandlerMsgFromAPI'...")
	fmt.Println("ID client:", clientID)
	fmt.Println("message generation:", msgGeneration)
	fmt.Printf("message options %v", msgCommon)

	fmt.Println("MsgOptions = ")
	fmt.Printf("%v", msgCommon.MsgOptions)

	//	MsgCommon общее сообщение
	// MsgType:
	//  - 'information'
	//  - 'command'
	// MsgSection:
	//  - 'source control'
	//  - 'filtration control'
	//  - 'download control'
	//  - 'information search control'
	//  - 'user notification'
	// MsgInsturction:
	//  - 'get new source list' API->
	//  - 'change status source' API->
	//  - 'confirm the action' API->
	//  - 'send new source list' API<-
	//  - 'performing an action' API<-
	//  - 'send notification' API<-
}
