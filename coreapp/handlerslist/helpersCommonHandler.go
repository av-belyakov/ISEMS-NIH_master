package handlerslist

import (
	"encoding/json"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//ErrorMessageType параметры для отработки сообщений об ошибках
type ErrorMessageType struct {
	SourceID        int
	TaskID          string
	TaskIDClientAPI string
	IDClientAPI     string
	Section         string
	Instruction     string
	MsgType         string
	MsgHuman        string
	Sources         []int
	ChanToAPI       chan<- *configure.MsgBetweenCoreAndAPI
}

//ErrorMessage формирует и отправляет клиенту API два сообщения, информационное сообщение и сообщение с откланенным статусом задачи
func ErrorMessage(emt ErrorMessageType) error {
	//отправляем информационное сообщение
	notifications.SendNotificationToClientAPI(
		emt.ChanToAPI,
		notifications.NotificationSettingsToClientAPI{
			MsgType:        emt.MsgType,
			MsgDescription: emt.MsgHuman,
			Sources:        emt.Sources,
		},
		emt.TaskIDClientAPI,
		emt.IDClientAPI)

	//отправляем сообщение о том что задача была отклонена
	resMsg := configure.DownloadControlTypeInfo{
		MsgOption: configure.DownloadControlMsgTypeInfo{
			ID:        emt.SourceID,
			TaskIDApp: emt.TaskID,
			Status:    "refused",
		},
	}

	resMsg.MsgType = "information"
	resMsg.MsgSection = emt.Section
	resMsg.MsgInstruction = emt.Instruction
	resMsg.ClientTaskID = emt.TaskIDClientAPI

	msgJSON, err := json.Marshal(resMsg)
	if err != nil {
		return err
	}

	emt.ChanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  emt.IDClientAPI,
		MsgJSON:      msgJSON,
	}

	return nil
}
