package notifications

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
)

//NotificationSettingsToClientAPI параметры сообщения отправляемые клиенту API
// MsgType - тип сообщения ('info'/'success'/'warning'/'danger')
// MsgDescription - подробное описание сообщения
// Sources - список ID источников к которым оно относится
type NotificationSettingsToClientAPI struct {
	MsgType, MsgDescription string
	Sources                 []int
}

//SendNotificationToClientAPI отправить сообщение клиенту API
func SendNotificationToClientAPI(
	c chan<- *configure.MsgBetweenCoreAndAPI,
	ns NotificationSettingsToClientAPI,
	clientTaskID, clientID string) {

	notify := configure.MsgNotification{
		MsgOptions: configure.UserNotification{
			Notification: configure.NotificationParameters{
				Type:        ns.MsgType,
				Description: ns.MsgDescription,
				Sources:     ns.Sources,
			},
		},
	}

	notify.MsgType = "information"
	notify.MsgSection = "user notification"
	notify.MsgInsturction = "send notification"
	notify.ClientTaskID = clientTaskID

	fmt.Printf("___ ___ ___ SEND NOTIFY ___ ___ ___\n%v\n", notify)

	msgjson, _ := json.Marshal(&notify)

	//отправляем сообщение
	c <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  clientID,
		MsgJSON:      msgjson,
	}
}
