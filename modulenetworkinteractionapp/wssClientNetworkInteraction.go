package modulenetworkinteractionapp

/*
* Клиент для взаимодействия с источниками
* осуществляет соединение с источниками если те выступают в роли сервера
*
* Версия 0.1, дата релиза 20.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//WssClientNetworkInteraction обрабатывает запросы с источников
func WssClientNetworkInteraction(timeReconnect int, ism *configure.InformationStoringMemory) {
	fmt.Println("START WSS CLIENT...")
}
