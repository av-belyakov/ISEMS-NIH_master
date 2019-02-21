package modulenetworkinteractionapp

/*
* Маршрутизация запросов приходящих через websocket
*
* Версия 0.1, дата релиза 21.02.2019
* */

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//Routing маршрутизирует запросы от источников
func Routing(ism *configure.InformationStoringMemory) {
	fmt.Println("START 'Routing' module network interaction...")

	//обработка данных получаемых через каналы
	/*	select {

		}*/
}
