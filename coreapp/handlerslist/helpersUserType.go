package handlerslist

import (
	"ISEMS-NIH_master/configure"
)

//HandlersStoringMemory вспомогательные хранилища
type HandlersStoringMemory struct {
	SMT  *configure.StoringMemoryTask
	QTS  *configure.QueueTaskStorage
	ISL  *configure.InformationSourcesList
	TSSQ *configure.TemporaryStorageSearchQueries
}

//HandlerOutChans каналы вывода из ядра информации
type HandlerOutChans struct {
	OutCoreChanAPI chan<- *configure.MsgBetweenCoreAndAPI
	OutCoreChanDB  chan<- *configure.MsgBetweenCoreAndDB
	OutCoreChanNI  chan<- *configure.MsgBetweenCoreAndNI
}
