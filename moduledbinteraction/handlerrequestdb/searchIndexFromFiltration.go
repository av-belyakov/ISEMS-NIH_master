package handlerrequestdb

import (
	"ISEMS-NIH_master/configure"

	"go.mongodb.org/mongo-driver/bson"
)

//searchIndexFromFiltration поиск индексов для выполнения фильтрации
func searchIndexFromFiltration(
	cn string,
	sourceID int,
	tf *configure.QueueTaskInformation,
	qp QueryParameters) (bool, *map[string][]string, error) {

	qp.CollectionName = cn

	//ключ - может быть директория, значение - имя файла
	FoundIndexInformation := map[string][]string{}

	c, err := qp.CountDocuments(bson.D{{}})
	if err != nil {
		return false, &FoundIndexInformation, err
	}

	if c == 0 {
		return false, &FoundIndexInformation, nil
	}

	return true, &FoundIndexInformation, nil
}
