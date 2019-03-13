package notifications

import (
	"encoding/json"

	"ISEMS-NIH_master/configure"
)

//SendNotificationToClientAPI отправить сообщение клиенту API
func SendNotificationToClientAPI(c chan<- configure.MsgBetweenCoreAndAPI, msgType, msgDescription, clientID string) {
	msgjson, _ := json.Marshal(&configure.MsgCommon{
		MsgType:        "information",
		MsgSection:     "user notification",
		MsgInsturction: "send notification",
		MsgOptions: configure.UserNotification{
			Notification: configure.NotificationParameters{
				Type:        msgType,
				Description: msgDescription,
			},
		},
	})

	//отправляем ошибку
	c <- configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  clientID,
		MsgJSON:      msgjson,
	}
}
