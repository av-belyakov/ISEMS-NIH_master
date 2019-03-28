package handlerslist

import (
	"errors"
	"fmt"

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

	if ti, ok := smt.GetStoringMemoryTask(taskID); ok {
		fmt.Printf("new status task with task ID %q - %v\n", taskID, ti.TaskStatus)
	}

	return nil
}
