package modulenetworkinteractionapp

/*
* Клиент для взаимодействия с источниками
* осуществляет соединение с источниками если те выступают в роли сервера
*
* Версия 0.1, дата релиза 20.02.2019
* */

import (
	"log"

	"ISEMS-NIH_master/configure"
)

//WssClientNetworkInteraction обрабатывает запросы с источников
func WssClientNetworkInteraction(cOut chan<- [2]string, timeReconnect int, ism *configure.InformationSourcesList, cIn <-chan [2]string) {
	log.Println("\tThe HTTPS client Network Integration is running")

	/*
		в cOut chan<- [2]string отправляем сообщения о установленных или
		разорванных соединениях

		из cIn <-chan [2]string получаем информацию о добавленных, удаленных
		или измененных параметрах для источников что бы отключить, переподключится
		или добавить источник в список систочников с которыми необходимо выполнить соединение
	*/
}
