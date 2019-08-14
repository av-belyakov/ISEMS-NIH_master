package handlers

import (
	"ISEMS-NIH_master/configure"
)

//fileDownloadProcessing обработчик выполняющий процесс по скачиванию файлов
func fileDownloadProcessing(
	cwt chan<- configure.MsgWsTransmission,
	isl *configure.InformationSourcesList,
	msg *configure.MsgBetweenCoreAndNI,
	smt *configure.StoringMemoryTask,
	qts *configure.QueueTaskStorage,
	chanInCore chan<- *configure.MsgBetweenCoreAndNI) {

	//msg.TaskID
	//msg.ClientName
	//msg.SourceID

	/*
	   Непосредственно выполняет скачивание файлов с источника
	   отправляя источнику задачи на скачивания по очередно,
	   в каждой задаче свой файл
	*/
}
