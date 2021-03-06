package modulenetworkinteractionapp

/*
* Модуль сетевого взаимодействия
* Выполняет следующие функции:
* - осуществляет взаимодействие с ядром приложения
* - осуществляет обмен данными и агригирацию данных получаемых от модулей wssServerNI и wssClientNI
* - выгружает файлы сет. трафика и объекты в долговременное хранилище
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/modulenetworkinteractionapp/handlers"
	"ISEMS-NIH_master/savemessageapp"

	"github.com/gorilla/websocket"
)

//при разрыве соединения удаляет дескриптор соединения и изменяет статус клиента
func connClose(
	COut chan<- [2]string,
	c *websocket.Conn,
	isl *configure.InformationSourcesList,
	clientID int,
	ip string,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) {

	c.Close()

	//изменяем статус подключения клиента
	_ = isl.ChangeSourceConnectionStatus(clientID, false)
	//удаляем дескриптор соединения
	isl.DelLinkWebsocketConnection(ip)

	saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
		Description: fmt.Sprintf("disconnect source with ip address '%v'\n", ip),
		FuncName:    "connClose",
	})

	fmt.Printf("SOURCE with ___ %v ___, DISCONNECT\n", ip)

	//отправляем сообщение о разрыве соединения
	COut <- [2]string{ip, "disconnect"}
}

//MainNetworkInteraction осуществляет общее управление
func MainNetworkInteraction(
	appConf *configure.AppConfig,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	isl *configure.InformationSourcesList,
	saveMessageApp *savemessageapp.PathDirLocationLogFiles) (chanOutCore, chanInCore chan *configure.MsgBetweenCoreAndNI) {

	//инициализируем канал для передачи данных через websocket соединение
	cwtRes := make(chan configure.MsgWsTransmission)
	//инициализируем канал для приема данных через websocket соединение
	cwtReq := make(chan configure.MsgWsTransmission)

	//инициализируем каналы для передачи данных между ядром приложения и текущем модулем
	chanOutCore = make(chan *configure.MsgBetweenCoreAndNI)
	chanInCore = make(chan *configure.MsgBetweenCoreAndNI) //, 100)

	//инициализация каналов управления и состояния источников
	chansStatSource := map[string]chan [2]string{
		"outWssModuleServer": make(chan [2]string),
		"outWssModuleClient": make(chan [2]string),
	}

	//обработчик процессов по скачиванию запрошенных файлов
	chanInCRRF := handlers.ControllerReceivingRequestedFiles(smt, qts, isl, saveMessageApp, chanInCore, cwtRes)

	//обработчик сообщений получаемых при фильтрации файлов
	chanInHMRFF := handlers.HandlerMessagesReceivedFilesFiltering(smt, saveMessageApp, chanInCore, cwtRes)

	//маршрутизатор запросов получаемых от CoreApp
	go RouteCoreRequest(cwtRes, chanInCore, chanInCRRF, isl, smt, qts, saveMessageApp, chansStatSource, chanOutCore)
	//маршрутизатор запросов получаемых Wss
	go RouteWssConnectionResponse(cwtRes, chanInCore, chanInCRRF, chanInHMRFF, isl, smt, saveMessageApp, cwtReq)

	//запуск модуля wssServerNI
	go WssServerNetworkInteraction(chansStatSource["outWssModuleServer"], appConf, isl, saveMessageApp, cwtReq)

	//запуск модуля wssClientNI
	go WssClientNetworkInteraction(chansStatSource["outWssModuleClient"], appConf, isl, saveMessageApp, cwtReq)

	go func() {
		for msg := range cwtRes {
			sourceIP, data := msg.DestinationHost, msg.Data
			if conn, ok := isl.GetLinkWebsocketConnect(sourceIP); ok {
				if err := conn.SendWsMessage(1, *data); err != nil {
					saveMessageApp.LogMessage(savemessageapp.TypeLogMessage{
						Description: fmt.Sprint(err),
						FuncName:    "mainNetworkInteraction",
					})
				}
			}
		}
	}()

	return chanOutCore, chanInCore
}
