package handlerrequestdb

import (
	"context"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"

	"ISEMS-NIH_master/configure"
)

func getInfoTaskForID(qp QueryParameters, taskID string) (*[]configure.InformationAboutTask, error) {
	fmt.Println("START function 'getInfoTaskForID'...")

	itf := []configure.InformationAboutTask{}

	cur, err := qp.Find(bson.D{bson.E{Key: "task_id", Value: taskID}})
	if err != nil {
		return &itf, err
	}

	for cur.Next(context.TODO()) {
		var model configure.InformationAboutTask
		err := cur.Decode(&model)
		if err != nil {
			return &itf, err
		}

		itf = append(itf, model)
	}

	if err := cur.Err(); err != nil {
		return &itf, err
	}

	cur.Close(context.TODO())

	return &itf, nil
}

func getInfoSource(qp QueryParameters, sourceID string) (*configure.InformationAboutSource, error) {
	ias := configure.InformationAboutTask{}

	/*  ПОЛУЧИТЬ ИНФОРМАЦИЮ ПО ИСТОЧНИКУ (ОСОБЕННО КРАТКОЕ НАЗВАНИЕ) */

	return &ias, nil
}
