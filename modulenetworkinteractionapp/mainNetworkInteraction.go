package modulenetworkinteractionapp

/*
* Модуль сетевого взаимодействия
* Выполняет следующие функции:
* - осуществляет взаимодействие с ядром приложения
* - осуществляет обмен данными и агригирацию данных получаемых от модулей wssServerNI и wssClientNI
* - выгружает файлы сет. трафика и объекты в долговременное хранилище
*
* Версия 0.11, дата релиза 27.02.2019
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
	ip string) {

	/*
			   Обработка разрыва соединения с источником
			   до отправки сообщения в модуль Ядра

		Изменение статуса соединений для источника,
		Изменить AvailabilityConnection в StoringMemoryQueueTask для всех
		задач выполняемых на данном источнике с true на false

			   здесь можно сделать отправку сообщения об
			   аварийном останове работы модуля отвечающего
			   за прием файлов
	*/

	c.Close()

	//изменяем статус подключения клиента
	_ = isl.ChangeSourceConnectionStatus(clientID, false)
	//удаляем дескриптор соединения
	isl.DelLinkWebsocketConnection(ip)

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
	chanInCore = make(chan *configure.MsgBetweenCoreAndNI)

	//инициализация каналов управления и состояния источников
	chansStatSource := map[string]chan [2]string{
		"outWssModuleServer": make(chan [2]string, 10),
		"outWssModuleClient": make(chan [2]string, 10),
	}

	//обработчик процессов по скачиванию запрошенных файлов
	chanInCRRF := handlers.ControllerReceivingRequestedFiles(smt, qts, isl, saveMessageApp, chanInCore, cwtRes)

	//маршрутизатор запросов получаемых от CoreApp
	go RouteCoreRequest(cwtRes, chanInCore, chanInCRRF, isl, smt, qts, saveMessageApp, chansStatSource, chanOutCore)
	//маршрутизатор запросов получаемых Wss
	go RouteWssConnectionResponse(cwtRes, isl, smt, saveMessageApp, chanInCore, chanInCRRF, cwtReq)

	//запуск модуля wssServerNI
	go WssServerNetworkInteraction(chansStatSource["outWssModuleServer"], appConf, isl, cwtReq)

	//запуск модуля wssClientNI
	go WssClientNetworkInteraction(chansStatSource["outWssModuleClient"], appConf, isl, cwtReq)

	go func() {
		for msg := range cwtRes {
			sourceIP, data := msg.DestinationHost, msg.Data
			if conn, ok := isl.GetLinkWebsocketConnect(sourceIP); ok {
				if err := conn.SendWsMessage(1, *data); err != nil {
					_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))
				}
			}
		}
	}()

	return chanOutCore, chanInCore
}
