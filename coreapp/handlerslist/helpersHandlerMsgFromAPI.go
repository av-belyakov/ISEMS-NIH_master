package handlerslist

import (
	"fmt"
	"strings"
	"time"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

/*
type FiltrationControlCommonParametersFiltration struct {
	ID       int                                       `json:"id"`
	DateTime DateTimeParameters                        `json:"dt"`
	Protocol string                                    `json:"p"`
	Filters  FiltrationControlParametersNetworkFilters `json:"f"`
}

type DateTimeParameters struct {
	Start int64 `json:"s"`
	End   int64 `json:"e"`
}

type FiltrationControlParametersNetworkFilters struct {
	IP      FiltrationControlIPorNetParameters `json:"ip"`
	Port    FiltrationControlPortParameters    `json:"pt"`
	Network FiltrationControlIPorNetParameters `json:"nw"`
}

type FiltrationControlIPorNetParameters struct {
	Any []string `json:"any"`
	Src []string `json:"src"`
	Dst []string `json:"dst"`
}

type FiltrationControlPortParameters struct {
	Any []int `json:"any"`
	Src []int `json:"src"`
	Dst []int `json:"dst"`
}
*/

//checkParametersFiltration проверяет параметры фильтрации
func checkParametersFiltration(fccpf *configure.FiltrationControlCommonParametersFiltration) (string, bool) {
	fmt.Println("START function 'checkParametersFiltration'...")

	//проверяем наличие ID источника
	if fccpf.ID == 0 {
		return "отсутствует идентификатор источника", false
	}

	//проверяем временной интервал
	isZero := ((fccpf.DateTime.Start == 0) || (fccpf.DateTime.End == 0))
	if isZero || (fccpf.DateTime.Start > fccpf.DateTime.End) {
		return "задан неверный временной интервал", false
	}

	//проверяем тип протокола
	if strings.EqualFold(fccpf.Protocol, "") {
		fccpf.Protocol = "any"
	}

	isProtoTCP := strings.EqualFold(fccpf.Protocol, "tcp")
	isProtoUDP := strings.EqualFold(fccpf.Protocol, "udp")
	isProtoANY := strings.EqualFold(fccpf.Protocol, "any")

	if !isProtoTCP && !isProtoUDP && !isProtoANY {
		return "задан неверный идентификатор транспортного протокола", false
	}

	isEmpty := true

	//проверка ip адресов

	//проверяем параметры свойства 'Filters' на пустоту
	if isEmpty {
		return "невозможно начать фильтрацию, необходимо указать хотябы один искомый ip адрес, порт или подсеть", false
	}

	return "", true
}

//handlerFiltrationControlTypeStart обработчик запроса на фильтрацию
func handlerFiltrationControlTypeStart(
	chanToDB chan<- *configure.MsgBetweenCoreAndDB,
	fcts *configure.FiltrationControlTypeStart,
	smt *configure.StoringMemoryTask,
	clientID string,
	chanToAPI chan<- *configure.MsgBetweenCoreAndAPI) {

	fmt.Println("START function 'handlerFiltrationControlTypeStart'...")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	funcName := ", function 'handlerFiltrationControlTypeStart'"

	if msg, ok := checkParametersFiltration(&fcts.MsgOption); !ok {
		notifications.SendNotificationToClientAPI(
			chanToAPI,
			notifications.NotificationSettingsToClientAPI{
				MsgType:        "danger",
				MsgDescription: msg,
			},
			fcts.ClientTaskID,
			clientID)

		_ = saveMessageApp.LogMessage("error", "incorrect parameters for filtering are set"+funcName)

		return
	}

	//добавляем новую задачу
	taskID := smt.AddStoringMemoryTask(configure.TaskDescription{
		ClientID:                        clientID,
		ClientTaskID:                    fcts.ClientTaskID,
		TaskType:                        fcts.MsgSection,
		ModuleThatSetTask:               "API module",
		ModuleResponsibleImplementation: "NI module",
		TimeUpdate:                      time.Now().Unix(),
	})

	fmt.Printf("JSON: %v\nTaskID: %v\n", fcts, taskID)

	//сохранение параметров задачи в БД

	//запрос на наличие индексов по заданным параметрам

}
