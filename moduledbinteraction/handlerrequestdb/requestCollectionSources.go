package handlerrequestdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson"

	"ISEMS-NIH_master/configure"
)

//GetAllSourcesList получить весь список источников
func GetAllSourcesList(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgSection:   "source control",
		IDClientAPI:  req.IDClientAPI,
		TaskID:       req.TaskID,
	}

	sourcesList, err := findAll(qp)
	if err != nil {
		msgRes.MsgRecipient = "Core module"
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "an error occurred while processing request get the list of sources",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	msgRes.MsgSection = "source list"
	msgRes.AdvancedOptions = sourcesList

	//отправка списка источников маршрутизатору ядра приложения
	chanIn <- &msgRes
}

//InsertListSources добавить информацию об источниках
//которых нет в БД или параметры по которым отличаются
func InsertListSources(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgSection:   "source control",
		IDClientAPI:  req.IDClientAPI,
		TaskID:       req.TaskID,
	}

	//получаем список источников
	listSources, err := findAll(qp)
	if err != nil {
		msgRes.MsgRecipient = "Core module"
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: "an error occurred while processing request get the list of sources",
			ErrorBody:             err,
		}

		chanIn <- &msgRes

		return
	}

	l, ok := req.AdvancedOptions.(*[]configure.InformationAboutSource)
	if !ok {
		errMsg := "incorrect list of sources received"

		msgRes.MsgRecipient = "Core module"
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             errors.New(errMsg),
		}

		chanIn <- &msgRes

		return
	}

	insertData := make([]interface{}, 0, len(*l))

	//если список источников в БД пуст, добавляем все что есть
	if len(listSources) == 0 {
		for _, v := range *l {
			insertData = append(insertData, v)
		}

		qp.InsertData(insertData)

		return
	}

	//список который пришел от клиента API
	for _, itemAddList := range *l {
		//список из БД
		for _, itemFindList := range listSources {
			//если источник с таким ID существует, удаляем его и заменяем новым
			if itemFindList.ID == itemAddList.ID {
				_ = qp.DeleteOneData(bson.D{bson.E{Key: "id", Value: itemAddList.ID}})
			}
		}

		insertData = append(insertData, itemAddList)
	}

	qp.InsertData(insertData)
}

//UpdateSourceToSourcesList обновить информацию об источниках
func UpdateSourceToSourcesList(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgSection:   "source control",
		IDClientAPI:  req.IDClientAPI,
		TaskID:       req.TaskID,
	}

	l, ok := req.AdvancedOptions.(*[]configure.InformationAboutSource)
	if !ok {
		errMsg := "incorrect list of sources received"

		msgRes.MsgRecipient = "Core module"
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             errors.New(errMsg),
		}

		chanIn <- &msgRes

		return
	}

	for _, i := range *l {
		if err := qp.UpdateOne(bson.D{bson.E{Key: "id", Value: i.ID}}, bson.D{
			bson.E{Key: "$set", Value: bson.D{
				bson.E{Key: "id", Value: i.ID},
				bson.E{Key: "ip", Value: i.IP},
				bson.E{Key: "token", Value: i.Token},
				bson.E{Key: "short_name", Value: i.ShortName},
				bson.E{Key: "description", Value: i.Description},
				bson.E{Key: "as_server", Value: i.AsServer},
				bson.E{Key: "name_client_api", Value: i.NameClientAPI},
				bson.E{Key: "source_setting", Value: bson.D{
					bson.E{Key: "enable_telemetry", Value: i.SourceSetting.EnableTelemetry},
					bson.E{Key: "max_count_process_filtration", Value: i.SourceSetting.MaxCountProcessFiltration},
					bson.E{Key: "storage_folders", Value: i.SourceSetting.StorageFolders},
					bson.E{Key: "if_as_server_then_port", Value: i.SourceSetting.IfAsServerThenPort},
				}}}},
		}); err != nil {
			msgRes.MsgRecipient = "Core module"
			msgRes.MsgSection = "error notification"
			msgRes.AdvancedOptions = configure.ErrorNotification{
				SourceReport:          "DB module",
				HumanDescriptionError: "error writing list of sources in the database",
				ErrorBody:             errors.New(fmt.Sprint(err)),
			}

			chanIn <- &msgRes

			return
		}
	}
}

//DelSourceToSourcesList удалить источники
func DelSourceToSourcesList(
	chanIn chan<- *configure.MsgBetweenCoreAndDB,
	req *configure.MsgBetweenCoreAndDB,
	qp QueryParameters) {

	msgRes := configure.MsgBetweenCoreAndDB{
		MsgGenerator: req.MsgRecipient,
		MsgRecipient: req.MsgGenerator,
		MsgSection:   "source control",
		IDClientAPI:  req.IDClientAPI,
		TaskID:       req.TaskID,
	}

	l, ok := req.AdvancedOptions.(*[]int)
	if !ok {
		errMsg := "incorrect list of sources received"

		msgRes.MsgRecipient = "Core module"
		msgRes.MsgSection = "error notification"
		msgRes.AdvancedOptions = configure.ErrorNotification{
			SourceReport:          "DB module",
			HumanDescriptionError: errMsg,
			ErrorBody:             errors.New(errMsg),
		}

		chanIn <- &msgRes

		return
	}

	for _, id := range *l {
		_ = qp.DeleteOneData(bson.D{bson.E{Key: "id", Value: id}})
	}
}

//findAll найти всю информацию по всем источникам
func findAll(qp QueryParameters) ([]configure.InformationAboutSource, error) {
	cur, err := qp.FindAlltoCollection()
	if err != nil {
		return nil, err
	}

	listSources := []configure.InformationAboutSource{}
	//получаем все ID источников
	for cur.Next(context.TODO()) {
		var model configure.InformationAboutSource
		err := cur.Decode(&model)
		if err != nil {
			return nil, err
		}

		listSources = append(listSources, model)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	cur.Close(context.TODO())

	return listSources, nil
}
