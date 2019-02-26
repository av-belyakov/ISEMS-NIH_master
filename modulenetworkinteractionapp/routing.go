package modulenetworkinteractionapp

/*
* Маршрутизация запросов приходящих через websocket
*
* Версия 0.1, дата релиза 26.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/modulenetworkinteractionapp/processrequest"
	"ISEMS-NIH_master/savemessageapp"
)

//Routing маршрутизирует запросы от источников
func Routing(cOut chan<- struct{}, chanColl *configure.ChannelCollection, ism *configure.InformationStoringMemory, cIn <-chan map[string]string) {
	fmt.Println("START 'Routing' module network interaction...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	//обработка данных получаемых через каналы
	for {
		select {
		//изменение статуса источника (от модулей wss)
		case msg := <-cIn:
			wssModule, sourceStatus, sourceIP := msg["wssModule"], msg["sourceStatus"], msg["sourceIP"]
			if wssModule == "client" && sourceStatus == "connect" {
				sourceSettings, _ := ism.GetSourceSetting(sourceIP)
				formatJSON, err := processrequest.SendMsgPing(sourceSettings.MaxCountProcessFilter)
				if err != nil {
					_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))
				}

				chanColl.Cwt <- configure.MsgWsTransmission{
					DestinationHost: sourceIP,
					Data:            formatJSON,
				}
			}

		case msg := <-chanColl.ChannelToMNICommon:
			fmt.Println("RESIVED data TO MNI Common")
			fmt.Println(msg)

			/*
			   !!! ОБРАБОТКА ЗАПРОСОВ ТИПА
			   			PONG, FILTERING, DOWNLOAD
			   !!!
			*/

			/*
				//отправка источнику сообщения типа 'Ping'
				if msg.sourceStatus == "connect" {
					c, err := ism.GetLinkWebsocketConnect(source[0]); err != nil {

					}
				}

				msg := configure.MessageTypeInfoStatusSource{
					IP:               source[0],
					ConnectionStatus: source[1],
					ConnectionTime:   time.Now().Unix(),
				}

				//отправка сообщения ядру
				channelCollection.ChannelFromMNIService <- configure.ServiceMessageInfoStatusSource{
					Type:       "change_sources",
					SourceList: []configure.MessageTypeInfoStatusSource{msg},
				}
			*/
		}
	}
}
