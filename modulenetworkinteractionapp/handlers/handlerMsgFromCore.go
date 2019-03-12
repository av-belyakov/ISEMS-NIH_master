package handlers

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//HandlerMsgFromCore обработчик сообщений от ядра приложения
func HandlerMsgFromCore(cwt chan<- configure.MsgWsTransmission, isl *configure.InformationSourcesList, msg configure.MsgBetweenCoreAndNI) {
	fmt.Println("START func HandlerMsgFromCore...")

	//инициализируем функцию конструктор для записи лог-файлов
	//saveMessageApp := savemessageapp.New()

	switch msg.Section {
	case "sources_control":
		if msg.Command == "load list" {
			if sl, ok := msg.AdvancedOptions.([]configure.InformationAboutSource); ok {
				loadSources(isl, sl)

				fmt.Println("create source list for memory success")
				fmt.Printf("\n%T%v\n", isl, isl)
			}
		}

		if msg.Command == "add" {

		}

		if msg.Command == "del" {

		}

		if msg.Command == "update" {

		}

		if msg.Command == "reconnect" {

		}

	case "filtration_control":
		if msg.Command == "start" {

		}

		if msg.Command == "stop" {

		}

	case "download_control":
		if msg.Command == "start" {

		}

		if msg.Command == "stop" {

		}

	}
}

func loadSources(isl *configure.InformationSourcesList, list []configure.InformationAboutSource) {
	for _, source := range list {
		isl.AddSourceSettings(source.IP, configure.SourceSetting{
			ID:       source.ID,
			Token:    source.Token,
			AsServer: source.AsServer,
			Settings: source.SourceSetting,
		})
	}
}
