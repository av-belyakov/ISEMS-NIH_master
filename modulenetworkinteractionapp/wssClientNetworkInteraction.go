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
func WssClientNetworkInteraction(cOut chan<- map[string]string, timeReconnect int, ism *configure.InformationStoringMemory, cIn <-chan struct{}) {
	fmt.Println("START WSS CLIENT...")
}
