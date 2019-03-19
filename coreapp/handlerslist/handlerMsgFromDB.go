package handlerslist

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//HandlerMsgFromDB обработчик сообщений приходящих от модуля взаимодействия с базой данных
func HandlerMsgFromDB(chanToAPI chan<- configure.MsgBetweenCoreAndAPI, res *configure.MsgBetweenCoreAndDB, chanToNI chan<- configure.MsgBetweenCoreAndNI) {
	fmt.Println("START function 'HandlerMsgFromDB' module coreapp")

	fmt.Printf("%v", res)

	if res.MsgGenerator == "DB module" {
		if res.MsgRecipient == "API module" {
			switch res.MsgSection {
			case "source control":

			case "source telemetry":

			case "filtration":

			case "download":

			case "information search results":

			case "error notification":

			}
		} else if res.MsgRecipient == "NI module" {
			switch res.MsgSection {
			case "source list":
				chanToNI <- configure.MsgBetweenCoreAndNI{
					Section:         "source control",
					Command:         "create list",
					AdvancedOptions: res.AdvancedOptions,
				}

			case "source control":

			case "filtration":

			case "download":
			}
		} else if res.MsgRecipient == "Core module" {
			fmt.Printf("RESIPENT MSG FOR CORE %v", res)
		}
	}
}
