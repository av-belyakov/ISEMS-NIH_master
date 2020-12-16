package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
)

//ErrorMessageType параметры для отработки сообщений об ошибках
// SourceID - ID источника
// TaskID - ID задачи
// TaskIDClientAPI - ID задачи клиента API
// IDClientAPI - ID клиента API
// Section - секция (filtering, download and etc.)
// Instruction - инструкция
// MsgType - тип сообщения (danger, warning, success, info)
// MsgHuman - описание сообщения
// Sources - список источников
// SearchRequestIsGeneratedAutomatically — был ли запрос на поиск сгенерирован автоматически (TRUE — да, FALSE - нет)
type ErrorMessageType struct {
	SourceID                              int
	TaskID                                string
	TaskIDClientAPI                       string
	IDClientAPI                           string
	Section                               string
	Instruction                           string
	MsgType                               string
	MsgHuman                              string
	Sources                               []int
	SearchRequestIsGeneratedAutomatically bool
	ChanToAPI                             chan<- *configure.MsgBetweenCoreAndAPI
}

//ErrorMessage формирует и отправляет клиенту API два сообщения, информационное сообщение и сообщение с откланенным статусом задачи
func ErrorMessage(emt ErrorMessageType) error {
	//если запрос не был сгенерирован автоматически
	if !emt.SearchRequestIsGeneratedAutomatically {

		fmt.Printf("func 'ErrorMessage', REGUESt renerated is not automatically")

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
	}

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
