package handlerrequestdb

import (
	"fmt"

	"ISEMS-NIH_master/configure"
)

//FindingInformationAboutTask поиск в БД информации по ID задачи
func FindingInformationAboutTask(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	fmt.Println("START function 'FindingInformationAboutTask'...")

}
