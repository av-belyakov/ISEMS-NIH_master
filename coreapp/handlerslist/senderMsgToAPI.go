package handlerslist

import (
	"errors"

	"ISEMS-NIH_master/configure"
)

func senderMsgToAPI(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	smt *configure.StoringMemoryTask,
	taskID string,
	msgjson []byte) error {

	st, ok := smt.GetStoringMemoryTask(taskID)
	if !ok {
		return errors.New("task with " + taskID + " not found")
	}

	//отправляем данные клиенту
	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  st.ClientID,
		MsgJSON:      msgjson,
	}

	//устанавливаем статус задачи как выполненую
	smt.StoringMemoryTaskComplete(taskID)

	return nil
}
