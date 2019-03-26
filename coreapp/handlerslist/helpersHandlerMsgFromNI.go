package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//getConfirmActionSourceListForAPI подгатавливает список источников с выполненными
//над ними действиями и статусом успешности
func getConfirmActionSourceListForAPI(
	chanToAPI chan<- configure.MsgBetweenCoreAndAPI,
	res *configure.MsgBetweenCoreAndNI,
	smt *configure.StoringMemoryTask) {

	fmt.Println("START function 'getConfirmActionSourceListForAPI'")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'getConfirmActionSourceListForAPI'"

	listSource, ok := res.AdvancedOptions.([]configure.ActionTypeListSources)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "type conversion error section type 'error notification'"+funcName)
	}

	//получаем ID клиента API
	st, ok := smt.GetStoringMemoryTask(res.TaskID)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "task with "+res.TaskID+" not found")
	}

	//удаляем задачу из хранилища задач
	smt.DelStoringMemoryTask(res.TaskID)

	msg := configure.SourceControlConfirmActionSource{
		MsgOptions: configure.SourceControlMsgTypeToAPI{
			TaskInfo: configure.MsgTaskInfo{
				State: "end",
			},
			SourceList: listSource,
		},
	}
	msg.MsgType = "information"
	msg.MsgSection = "source control"
	msg.MsgInsturction = "confirm the action"
	msg.ClientTaskID = st.ClientTaskID

	msgjson, _ := json.Marshal(&msg)

	//отправляем данные клиенту
	chanToAPI <- configure.MsgBetweenCoreAndAPI{
		MsgGenerator: "Core module",
		MsgRecipient: "API module",
		IDClientAPI:  st.ClientID,
		MsgJSON:      msgjson,
	}
}
