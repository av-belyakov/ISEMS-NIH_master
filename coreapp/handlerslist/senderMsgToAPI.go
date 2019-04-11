package handlerslist

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

func senderMsgToAPI(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	smt *configure.StoringMemoryTask,
	taskID, clientID string,
	msgjson []byte) error {

	//отправляем данные клиенту
	chanToAPI <- &configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  clientID,
		MsgJSON:      msgjson,
	}

	//устанавливаем статус задачи как выполненую
	smt.StoringMemoryTaskComplete(taskID)

	if ti, ok := smt.GetStoringMemoryTask(taskID); ok {
		fmt.Printf("new status task with task ID %q - %v\n", taskID, ti.TaskStatus)
	}

	return nil
}
