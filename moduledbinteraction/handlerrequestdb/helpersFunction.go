package handlerrequestdb

import (
	"context"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"

	"ISEMS-NIH_master/configure"
)

func getInfoFiltrationTaskForID(qp QueryParameters, taskID string) ([]configure.InformationAboutTaskFiltration, error) {
	fmt.Println("START function 'getInfoFiltrationTaskForID'...")

	itf := []configure.InformationAboutTaskFiltration{}

	cur, err := qp.Find(bson.D{bson.E{Key: "task_id", Value: taskID}})
	if err != nil {
		return itf, err
	}

	for cur.Next(context.TODO()) {
		var model configure.InformationAboutTaskFiltration
		err := cur.Decode(&model)
		if err != nil {
			return itf, err
		}

		itf = append(itf, model)
	}

	if err := cur.Err(); err != nil {
		return itf, err
	}

	cur.Close(context.TODO())

	return itf, nil
}
