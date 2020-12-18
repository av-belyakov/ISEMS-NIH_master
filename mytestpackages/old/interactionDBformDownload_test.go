package mytestpackages

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
)

/*
type configureDB struct {
	Host, Port, NameDB, User, Password string
}

//QueryParameters параметры для работы с коллекциями БД
type QueryParameters struct {
	NameDB, CollectionName string
	ConnectDB              *mongo.Client
}


//Find найти всю информацию по заданному элементу
func (qp QueryParameters) Find(elem interface{}) (*mongo.Cursor, error) {

	//fmt.Println("\t===== REQUEST TO DB 'FIND' ======")

	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	options := options.Find()

	return collection.Find(context.TODO(), elem, options)
}

/*func connectToDB(ctx context.Context, conf configureDB) (*mongo.Client, error) {
	optAuth := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    conf.NameDB,
		Username:      conf.User,
		Password:      conf.Password,
	}

	opts := options.Client()
	opts.SetAuth(optAuth)

	client, err := mongo.NewClient(opts.ApplyURI("mongodb://" + conf.Host + ":" + conf.Port + "/" + conf.NameDB))
	if err != nil {
		return nil, err
	}

	client.Connect(ctx)

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}*/

func getInfoFiltrationTaskForClientTaskID(connectDB *mongo.Client, taskID string) ([]configure.InformationAboutTask, error) {
	qp := QueryParameters{
		NameDB:         "isems-nih",
		CollectionName: "task_list",
		ConnectDB:      connectDB,
	}

	itf := []configure.InformationAboutTask{}

	cur, err := qp.Find(bson.D{bson.E{Key: "task_id", Value: taskID}})
	if err != nil {
		return itf, err
	}

	for cur.Next(context.TODO()) {
		var model configure.InformationAboutTask
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

func updateOne(
	connectDB *mongo.Client,
	nameDB, nameCollection string,
	searchElem, update interface{}) error {

	//	fmt.Println("===== UPDATE ONE ======")

	collection := connectDB.Database(nameDB).Collection(nameCollection)
	if _, err := collection.UpdateOne(context.TODO(), searchElem, update); err != nil {
		return err
	}

	return nil
}

func updateOneArrayFilters(
	connectDB *mongo.Client,
	nameDB, nameCollection string,
	filter, update interface{},
	uo *options.UpdateOptions) error {

	collection := connectDB.Database(nameDB).Collection(nameCollection)
	if _, err := collection.UpdateOne(context.TODO(), filter, update, uo); err != nil {
		return err
	}

	return nil
}

//UpdateMany обновляет множественные параметры в элементе
func (qp *QueryParameters) UpdateMany(searchElem, update []interface{}) error {

	fmt.Println("\t===== REQUEST TO DB 'UPDATE MANY' ======")

	collection := qp.ConnectDB.Database(qp.NameDB).Collection(qp.CollectionName)
	if _, err := collection.UpdateMany(context.TODO(), searchElem, update); err != nil {
		return err
	}

	return nil
}

//UpdateFinishedInformationAboutTask запись информации по задаче (задача завершена)
func UpdateFinishedInformationAboutTask(
	qp QueryParameters,
	smt *configure.StoringMemoryTask,
	req configure.MsgBetweenCoreAndDB) error {

	//при добавлении информации в БД не забыть изменить статус на 'complete'
	ti, ok := smt.GetStoringMemoryTask(req.TaskID)
	if !ok {
		return fmt.Errorf("task with ID '%v' not found (DB module)", req.TaskID)
	}

	const timeUpdate = 30
	//выполнять обновление информации в БД для сообщения типа 'complete' всегда,
	// для сообщения типа 'execute' только раз 31 секунду
	/*	if (ti.TaskParameter.DownloadTask.Status == "execute") && ((time.Now().Unix() - ti.TimeInsertDB) < timeUpdate) {
		return nil
	}*/

	//обновление основной информации
	commonValueUpdate := bson.D{
		bson.E{Key: "$set", Value: bson.D{
			bson.E{Key: "detailed_information_on_downloading.task_status", Value: "complete"},
			bson.E{Key: "detailed_information_on_downloading.time_interval_task_execution.end", Value: time.Now().Unix()},
			bson.E{Key: "detailed_information_on_downloading.number_files_total", Value: ti.TaskParameter.DownloadTask.NumberFilesTotal},
			bson.E{Key: "detailed_information_on_downloading.number_files_downloaded", Value: ti.TaskParameter.DownloadTask.NumberFilesDownloaded},
			bson.E{Key: "detailed_information_on_downloading.number_files_downloaded_error", Value: ti.TaskParameter.DownloadTask.NumberFilesDownloadedError},
			bson.E{Key: "detailed_information_on_downloading.path_directory_storage_downloaded_files", Value: ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles},
		}}}

	//обновляем детальную информацию о ходе фильтрации
	if err := updateOne(qp.ConnectDB, "isems-nih", "task_list", bson.D{bson.E{Key: "task_id", Value: req.TaskID}}, commonValueUpdate); err != nil {
		return err
	}

	var arrayFiles []interface{}
	for fn, fi := range ti.TaskParameter.ListFilesDetailedInformation {
		t := time.Now().Unix() - (timeUpdate * 2)

		fmt.Printf("Time download:%v, time now:%v, equal: %v\n", fi.TimeDownload, t, fi.TimeDownload >= t)

		if fi.IsLoaded && ( /*fi.TimeDownload*/ int64(0) >= time.Now().Unix()-(timeUpdate*2)) {
			arrayFiles = append(arrayFiles, bson.D{bson.E{Key: "elem.file_name", Value: fn}})
		}
	}

	fmt.Printf("Count files update = %v\n", len(arrayFiles))

	if len(arrayFiles) == 0 {
		return nil
	}

	//обновляем информацию по загруженным файлам
	if err := updateOneArrayFilters(
		qp.ConnectDB,
		"isems-nih",
		"task_list",
		bson.D{
			bson.E{Key: "task_id", Value: req.TaskID}},
		bson.D{
			bson.E{Key: "$set", Value: bson.D{
				bson.E{Key: "list_files_result_task_execution.$[elem].file_loaded", Value: true},
			}}},
		&options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: []interface{}{bson.D{
					bson.E{
						Key: "$or", Value: arrayFiles,
					},
				}},
			},
		}); err != nil {
		return err
	}

	return nil
}

var _ = Describe("InteractionDataBaseFromDownloadFiles", func() {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()
	conn, err := connectToDB(ctx, configureDB{
		Host:     "127.0.0.1",
		Port:     37017,
		User:     "module_ISEMS-NIH",
		Password: "tkovomfh&ff93",
		NameDB:   "isems-nih",
	})

	Context("Тест 1: Проверка подключения к БД", func() {
		It("Должно быть установлено подключение с БД", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 2: Запрос к БД для получения списка файлов для скачивания", func() {
		It("Для выбранной в тесте задаче должно быть найдено 32 файла", func() {
			taskID := "cd42ebd8e8f260a1f713e97379fa135e"

			ti, err := getInfoFiltrationTaskForClientTaskID(conn, taskID)

			//fmt.Println(err)
			//fmt.Printf("---------- All information about task -----\n%v\n", ti)

			Expect(err).ToNot(HaveOccurred())

			//			fmt.Printf("---- INFORMATION ----\n%v\n", ti)

			numFiles := ti[0].ListFilesResultTaskExecution

			Expect(len(numFiles)).Should(Equal(32))
		})
	})

	Context("Тест 3: Запись информации о скачивании файлов в БД", func() {
		smt := configure.NewRepositorySMT()

		taskID := "239379934af7a56f5e86e90894811018"
		clientID := "b73aaca054c920d13500a6ad9beb0c3b"
		clientTaskID := "58c5eb8c72e64fedfe34380d9be2b07668fc6807"

		qp := QueryParameters{
			NameDB:         "isems-nih",
			CollectionName: "task_list",
			ConnectDB:      conn,
		}

		tf1 := "1560801329_2019_06_17____22_55_29_29140.tdp"
		tf2 := "1560803273_2019_06_17____23_27_53_59020.tdp"
		tf3 := "1560801977_2019_06_17____23_06_17_36143.tdp"
		tf4 := "1560800357_2019_06_17____22_39_17_7593.tdp"
		tf5 := "1560802301_2019_06_17____23_11_41_969.tdp"

		taskDescription := configure.TaskDescription{
			ClientID:                        clientID,
			ClientTaskID:                    clientTaskID,
			TaskType:                        "download control",
			ModuleThatSetTask:               "API module",
			ModuleResponsibleImplementation: "NI module",
			TimeUpdate:                      time.Now().Unix(),
			TimeInterval: configure.TimeIntervalTaskExecution{
				Start: (time.Now().Unix() - 2500),
				End:   time.Now().Unix(),
			},
			TaskParameter: configure.DescriptionTaskParameters{
				DownloadTask: &configure.DownloadTaskParameters{
					ID:                                  1221,
					Status:                              "executed",
					NumberFilesTotal:                    3,
					NumberFilesDownloaded:               2,
					PathDirectoryStorageDownloadedFiles: "/__TMP/write",
					FileInformation: configure.DetailedFileInformation{
						Name:                "26_04_2016___01_02_59.tdp",
						Hex:                 "ld0jf9jg9j9434884848hg8h8",
						FullSizeByte:        54793063,
						AcceptedSizeByte:    82347,
						AcceptedSizePercent: 12,
						NumChunk:            34583,
						ChunkSize:           4096,
						NumAcceptedChunk:    3123,
					},
					/*					DownloadingFilesInformation: map[string]*configure.DownloadFilesInformation{
										tf1: &configure.DownloadFilesInformation{IsLoaded: true},
										tf2: &configure.DownloadFilesInformation{IsLoaded: true},
										tf3: &configure.DownloadFilesInformation{},
										tf4: &configure.DownloadFilesInformation{IsLoaded: true},
										tf5: &configure.DownloadFilesInformation{IsLoaded: true},
									},*/
				},
			},
		}

		/*
			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf1].Size = 3081429
			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf1].Hex = "a86b143391a1eeae4078786f624b5257"

			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf2].Size = 3137245
			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf2].Hex = "3ab19032a4a3d990a5a0b92042a93ef4"

			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf3].Size = 70951216
			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf3].Hex = "8b95f4e9454e5fe755bc7d6cfbe1f4a1"

			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf4].Size = 2948350
			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf4].Hex = "ead26b5d302e53961b75a7e92c080187"

			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf5].Size = 3421341
			taskDescription.TaskParameter.DownloadTask.DownloadingFilesInformation[tf5].Hex = "6654b432096dab1f4f0818a42143e341"
		*/

		//добавляем новую задачу
		smt.AddStoringMemoryTask(taskID, taskDescription)

		smt.UpdateListFilesDetailedInformation(taskID, map[string]configure.DetailedFilesInformation{
			tf1: (configure.DetailedFilesInformation{IsLoaded: true}),
			tf2: (configure.DetailedFilesInformation{IsLoaded: true}),
			tf3: (configure.DetailedFilesInformation{}),
			tf4: (configure.DetailedFilesInformation{IsLoaded: true}),
			tf5: (configure.DetailedFilesInformation{IsLoaded: true}),
		})

		ti, ok := smt.GetStoringMemoryTask(taskID)

		It("В storingMemoryTask должна быть добавлена информация о задаче", func() {

			fmt.Println(ti.TaskParameter.DownloadTask)

			Expect(ok).Should(BeTrue())
			Expect(ti.TaskParameter.DownloadTask.PathDirectoryStorageDownloadedFiles).Should(Equal("/__TMP/write"))
		})

		It("Запись в БД должна быть выполнена без ошибок", func() {
			err := UpdateFinishedInformationAboutTask(qp, smt, configure.MsgBetweenCoreAndDB{
				TaskID:          taskID,
				IDClientAPI:     clientID,
				TaskIDClientAPI: clientTaskID,
			})

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должны записаться параметры о выполнении задачи по скачиванию", func() {

		})
	})
})
