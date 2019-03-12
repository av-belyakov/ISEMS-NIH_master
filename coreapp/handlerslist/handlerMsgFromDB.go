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
			case "sources_control":

			case "source_telemetry":

			case "filtration":

			case "download":

			case "information_search_results":

			case "error_notification":

			}
		}

		if res.MsgRecipient == "NI module" {
			switch res.MsgSection {
			case "sources_list":
				chanToNI <- configure.MsgBetweenCoreAndNI{
					Section:         "sources_control",
					Command:         "load list",
					AdvancedOptions: res.AdvancedOptions,
				}

			case "sources_control":

			case "filtration":

			case "download":
			}
		}
	}
}
