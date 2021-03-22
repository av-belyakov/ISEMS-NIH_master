package handlers

import (
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/modulenetworkinteractionapp/processrequest"
	"fmt"
)

//SendPing отправка сообщения типа 'ping' содержащего настройки источника
func SendPing(
	sourceIP string,
	sourceID int,
	isl *configure.InformationSourcesList,
	cwt chan<- configure.MsgWsTransmission) error {

	ss, _ := isl.GetSourceSetting(sourceID)

	fmt.Println("___________________ func 'SendPing' _______________________")
	fmt.Printf("func 'SendPing', SEND source ID '%d' new parameters (%v)\n", sourceID, ss.Settings)

	formatJSON, err := processrequest.SendMsgPing(ss)
	if err != nil {
		return err
	}

	//отправляем источнику запрос типа Ping
	cwt <- configure.MsgWsTransmission{
		DestinationHost: sourceIP,
		Data:            &formatJSON,
	}

	return nil
}
