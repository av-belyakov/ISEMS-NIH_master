package handlers

import (
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/modulenetworkinteractionapp/processrequest"
)

//SendPing отправка сообщения типа 'ping' содержащего настройки источника
func SendPing(
	sourceIP string,
	sourceID int,
	isl *configure.InformationSourcesList,
	cwt chan<- configure.MsgWsTransmission) error {

	ss, _ := isl.GetSourceSetting(sourceID)

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
