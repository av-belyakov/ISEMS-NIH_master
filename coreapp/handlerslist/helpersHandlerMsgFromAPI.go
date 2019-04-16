package handlerslist

import (
	"fmt"
	"strings"
	"time"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/notifications"
	"ISEMS-NIH_master/savemessageapp"
)

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

	circle := func(fp map[string]map[string]*[]string, f func(string, *[]string) error) error {
		for pn, pv := range fp {
			var err error

			for _, v := range pv {
				if err = f(pn, v); err != nil {
					return err
				}
			}

			if err != nil {
				return err
			}
		}

		return nil
	}

	checkIPOrPortOrNetwork := func(paramType string, param *[]string) error {
		switch paramType {
		case "IP":

		case "Port":

		case "Network":

		}

		return nil
	}

	filterParameters := map[string]map[string]*[]string{
		"IP": map[string]*[]string{
			"Any": &fccpf.Filters.IP.Any,
			"Src": &fccpf.Filters.IP.Src,
			"Dst": &fccpf.Filters.IP.Dst,
		},
		"Port": map[string]*[]string{
			"Any": &fccpf.Filters.Port.Any,
			"Src": &fccpf.Filters.Port.Src,
			"Dst": &fccpf.Filters.Port.Dst,
		},
		"Network": map[string]*[]string{
			"Any": &fccpf.Filters.Network.Any,
			"Src": &fccpf.Filters.Network.Src,
			"Dst": &fccpf.Filters.Network.Dst,
		},
	}

	//проверка ip адресов, портов и подсетей
	if err := circle(filterParameters, checkIPOrPortOrNetwork); err != nil {
		return "невозможно начать фильтрацию, задан некорректный ip адрес, порт или подсеть", false
	}

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
