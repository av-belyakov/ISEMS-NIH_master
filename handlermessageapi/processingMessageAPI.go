package handlermessageapi

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//ProcessingMessageAPI обработчик запросов поступающих через API
func ProcessingMessageAPI(cmf chan<- configure.MessageAPI, ism *configure.InformationStoringMemory, cmt <-chan configure.MessageAPI) {
	fmt.Println("START func ProcessingMessageAPI")
}
