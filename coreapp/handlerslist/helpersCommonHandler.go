package handlerslist

import (
	"encoding/json"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//ErrorMessage формирует и отправляет клиенту API два сообщения, информационное сообщение и сообщение с откланенным статусом задачи
func ErrorMessage(res *configure.MsgBetweenCoreAndDB, sourceID int, msgType, msg string, chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) error {
	//отправляем информационное сообщение
	notifications.SendNotificationToClientAPI(
		chanToAPI,
		notifications.NotificationSettingsToClientAPI{
			MsgType:        msgType,
			MsgDescription: msg,
		},
		res.TaskIDClientAPI,
		res.IDClientAPI)

	//отправляем сообщение о том что задача была отклонена
	resMsg := configure.DownloadControlTypeInfo{
		MsgOption: configure.DownloadControlMsgTypeInfo{
			ID:        sourceID,
			TaskIDApp: res.TaskID,
			Status:    "refused",
		},
	}

	resMsg.MsgType = "information"
	resMsg.MsgSection = res.MsgSection
	resMsg.MsgInsturction = res.Instruction
	resMsg.ClientTaskID = res.TaskIDClientAPI

	msgJSON, err := json.Marshal(resMsg)
	if err != nil {
		return err
	}

	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  res.IDClientAPI,
		MsgJSON:      msgJSON,
	}

	return nil
}
