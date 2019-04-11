package handlerslist

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
)

//getCurrentSourceListForAPI подготавливает список актуальных источников для передаче клиенту API
func getCurrentSourceListForAPI(
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI,
	res *configure.MsgBetweenCoreAndDB,
	smt *configure.StoringMemoryTask) {

	fmt.Println("START function 'getCurrentSourceListForAPI'")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()
	funcName := ", function 'getCurrentSourceListForAPI'"

	listSource, ok := res.AdvancedOptions.([]configure.InformationAboutSource)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "type conversion error section type 'error notification'"+funcName)

		return
	}

	list := make([]configure.ShortListSources, 0, len(listSource))

	//формируем ответ клиенту API
	for _, s := range listSource {
		list = append(list, configure.ShortListSources{
			ID:          s.ID,
			IP:          s.IP,
			ShortName:   s.ShortName,
			Description: s.Description,
		})
	}

	st, ok := smt.GetStoringMemoryTask(res.TaskID)
	if !ok {
		_ = saveMessageApp.LogMessage("error", "task with "+res.TaskID+" not found")

		return
	}

	msg := configure.SourceControlCurrentListSources{
		MsgOptions: configure.SourceControlCurrentListSourcesList{
			TaskInfo: configure.MsgTaskInfo{
				State: "end",
			},
			SourceList: list,
		},
	}
	msg.MsgType = "information"
	msg.MsgSection = "source control"
	msg.MsgInsturction = "send current source list"
	msg.ClientTaskID = st.ClientTaskID

	msgjson, _ := json.Marshal(&msg)

	if err := senderMsgToAPI(chanToAPI, smt, res.TaskID, st.ClientID, msgjson); err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
	}
}
