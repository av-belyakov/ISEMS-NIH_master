package handlerrequestdb

import (
	"fmt"

	"ISEMS-NIH_master/configure"

	"github.com/mongodb/mongo-go-driver/bson"
)

//searchIndexFormFiltration поиск индексов для выполнения фильтрации
func searchIndexFormFiltration(
	cn string,
	tf *configure.FiltrationControlCommonParametersFiltration,
	qp QueryParameters) (bool, *map[string]string, error) {

	fmt.Println("START function 'searchIndexFormFiltration'...")

	qp.CollectionName = cn

	//ключ - может быть директория, значение - имя файла
	FoundIndexInformation := map[string]string{}

	c, err := qp.CountDocuments(bson.D{{}})
	if err != nil {
		return false, &FoundIndexInformation, err
	}

	if c == 0 {
		return false, &FoundIndexInformation, nil
	}

	return true, &FoundIndexInformation, nil
}
