package mytestpackages_test

import (
	"context"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master/mytestpackages"
)

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

func connectToDB(ctx context.Context, conf configureDB) (*mongo.Client, error) {
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
}

func getInfoSource(connectDB *mongo.Client, sourceID int) ([]configure.InformationAboutSource, error) {
	qp := QueryParameters{
		NameDB:         "isems-nih",
		CollectionName: "sources_list",
		ConnectDB:      connectDB,
	}

	ias := []configure.InformationAboutSource{}

	cur, err := qp.Find(bson.D{bson.E{Key: "id", Value: sourceID}})
	if err != nil {
		return ias, err
	}

	for cur.Next(context.TODO()) {
		var model configure.InformationAboutSource
		err := cur.Decode(&model)
		if err != nil {
			return ias, err
		}

		ias = append(ias, model)
	}

	if err := cur.Err(); err != nil {
		return ias, err
	}

	cur.Close(context.TODO())

	return ias, nil
}

func createNewFiltrationTask(
	connectDB *mongo.Client,
	taskID, clientID, clientTaskID string,
	tf *configure.FiltrationControlCommonParametersFiltration) error {

	//	fmt.Println("START function 'createNewFiltrationTask_test'...")

	itf := configure.InformationAboutTask{
		TaskID:       taskID,
		ClientID:     clientID,
		ClientTaskID: clientTaskID,
		FilteringOption: configure.FilteringOption{
			DateTime: configure.TimeInterval{
				Start: tf.DateTime.Start,
				End:   tf.DateTime.End,
			},
			Protocol: tf.Protocol,
			Filters: configure.FilteringExpressions{
				IP: configure.FilteringNetworkParameters{
					Any: tf.Filters.IP.Any,
					Src: tf.Filters.IP.Src,
					Dst: tf.Filters.IP.Dst,
				},
				Port: configure.FilteringNetworkParameters{
					Any: tf.Filters.Port.Any,
					Src: tf.Filters.Port.Src,
					Dst: tf.Filters.Port.Dst,
				},
				Network: configure.FilteringNetworkParameters{
					Any: tf.Filters.Network.Any,
					Src: tf.Filters.Network.Src,
					Dst: tf.Filters.Network.Dst,
				},
			},
		},
		DetailedInformationOnFiltering: configure.DetailedInformationFiltering{
			TaskStatus:   "wait",
			WasIndexUsed: true,
			TimeIntervalTaskExecution: configure.TimeInterval{
				Start: time.Now().Unix(),
			},
		},
		DetailedInformationOnDownloading: configure.DetailedInformationDownloading{
			TaskStatus: "not executed",
		},
		ListFilesResultTaskExecution: []*configure.FilesInformation{},
	}

	insertData := make([]interface{}, 0, 1)
	insertData = append(insertData, itf)

	//	fmt.Printf("------- %v --------\n", insertData)

	//InsertData добавляет все данные
	//	fmt.Println("===== INSERT DATA ======")

	collection := connectDB.Database("isems-nih").Collection("task_list")
	if _, err := collection.InsertMany(context.TODO(), insertData); err != nil {
		return err
	}

	return nil

}

func updateFiltrationTaskParameters(
	connectDB *mongo.Client,
	taskID string,
	ftp *configure.FiltrationTaskParameters) error {

	//обновление основной информации
	commonValueUpdate := bson.D{
		bson.E{Key: "$set", Value: bson.D{
			bson.E{Key: "detailed_information_on_filtering.task_status", Value: ftp.Status},
			bson.E{Key: "detailed_information_on_filtering.time_interval_task_execution.end", Value: time.Now().Unix()},
			bson.E{Key: "detailed_information_on_filtering.number_files_meet_filter_parameters", Value: ftp.NumberFilesMeetFilterParameters},
			bson.E{Key: "detailed_information_on_filtering.number_processed_files", Value: ftp.NumberProcessedFiles},
			bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: ftp.NumberFilesFoundResultFiltering},
			bson.E{Key: "detailed_information_on_filtering.number_directory_filtartion", Value: ftp.NumberDirectoryFiltartion},
			bson.E{Key: "detailed_information_on_filtering.number_error_processed_files", Value: ftp.NumberErrorProcessedFiles},
			bson.E{Key: "detailed_information_on_filtering.size_files_meet_filter_parameters", Value: ftp.SizeFilesMeetFilterParameters},
			bson.E{Key: "detailed_information_on_filtering.size_files_found_result_filtering", Value: ftp.SizeFilesFoundResultFiltering},
			bson.E{Key: "detailed_information_on_filtering.path_directory_for_filtered_files", Value: ftp.PathStorageSource},
			//			bson.E{Key: "detailed_information_on_filtering.list_files_found_result_filtering", Value: ftp.FoundFilesInformation},
		}}}

	//обновляем детальную информацию о ходе фильтрации
	if err := updateOne(connectDB, "isems-nih", "task_list", bson.D{bson.E{Key: "task_id", Value: taskID}}, commonValueUpdate); err != nil {
		return err
	}

	arr := []interface{}{}

	/*for fileName, v := range ftp.FoundFilesInformation {
		arr = append(arr, bson.D{
			bson.E{Key: "file_name", Value: fileName},
			bson.E{Key: "file_size", Value: v.Size},
			bson.E{Key: "file_hex", Value: v.Hex},
			bson.E{Key: "file_loaded", Value: false},
		})
	}*/

	arrayValueUpdate := bson.D{
		bson.E{
			Key: "$addToSet", Value: bson.D{
				bson.E{
					Key: "list_files_result_task_execution",
					Value: bson.D{
						bson.E{
							Key:   "$each",
							Value: arr,
						},
					},
				},
			},
		},
	}

	func() {
		f, err := os.Create("memdumpfile.out")
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()

		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}()

	//обновление информации об отфильтрованном файле
	if err := updateOne(connectDB, "isems-nih", "task_list", bson.D{bson.E{Key: "task_id", Value: taskID}}, arrayValueUpdate); err != nil {
		return err
	}

	return nil
}

func getInfoTaskForID(connectDB *mongo.Client, taskID string) ([]configure.InformationAboutTask, error) {
	qp := QueryParameters{
		NameDB:         "isems-nih",
		CollectionName: "task_list",
		ConnectDB:      connectDB,
	}

	itf := []configure.InformationAboutTask{}

	cur, err := qp.Find(bson.D{bson.E{Key: "task_id", Value: "ff644a6faed4bbfec3a31cf8826a5587"}})
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

func updateMany(
	connectDB *mongo.Client,
	nameDB, nameCollection string,
	searchElem, update []interface{}) error {

	//	fmt.Println("===== UPDATE MANY ======")

	collection := connectDB.Database(nameDB).Collection(nameCollection)
	if _, err := collection.UpdateMany(context.TODO(), searchElem, update); err != nil {
		return err
	}

	return nil
}

var _ = Describe("InteractionDataBase", func() {
	taskID := common.GetUniqIDFormatMD5("task_id")
	clientID := common.GetUniqIDFormatMD5("client_id")
	clientTaskID := common.GetUniqIDFormatMD5("client_task_id")

	//	fmt.Printf("TaskID: %v, clientID: %v, clientTaskID: %v\n", taskID, clientID, clientTaskID)

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	conn, err := connectToDB(ctx, configureDB{
		Host:     "127.0.0.1",
		Port:     "37017",
		User:     "module_ISEMS-NIH",
		Password: "tkovomfh&ff93",
		NameDB:   "isems-nih",
	})

	Context("Тест 1: Проверка подключения к БД", func() {
		It("Должно быть установлено подключение с БД", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 2: Создание в БД записи о новой задаче по фильтрации сет. трафика", func() {
		It("Должна быть успешно создана новая запись по задаче фильтрации сет. трафика", func() {
			tf := configure.FiltrationControlCommonParametersFiltration{
				DateTime: configure.DateTimeParameters{
					Start: time.Now().Unix(),
					End:   time.Now().Unix(),
				},
				Protocol: "tcp",
				Filters: configure.FiltrationControlParametersNetworkFilters{
					IP: configure.FiltrationControlIPorNetorPortParameters{
						Any: []string{"240.45.56.23", "89.100.23.24"},
					},
					Port: configure.FiltrationControlIPorNetorPortParameters{
						Dst: []string{"80"},
						Any: []string{"22", "23"},
					},
				},
			}

			err := createNewFiltrationTask(conn, taskID, clientID, clientTaskID, &tf)

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 3: Обновление информации о параметрах по задачи на фильтрацию сет. трафика", func() {
		It("Информация о параметрах фильтрации должна быть успешно обновлена", func() {
			parameters := configure.FiltrationTaskParameters{
				ID:                              189,
				Status:                          "execute",
				NumberFilesMeetFilterParameters: 331,
				SizeFilesMeetFilterParameters:   472435353569055,
				NumberDirectoryFiltartion:       4,
				NumberProcessedFiles:            22,
				NumberFilesFoundResultFiltering: 5,
				NumberErrorProcessedFiles:       0,
				SizeFilesFoundResultFiltering:   32455311111,
				PathStorageSource:               "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
				/*FoundFilesInformation: map[string]*configure.FoundFilesInformation{
					"1438535410_2015_08_02____20_10_10_644263.tdp": &configure.FoundFilesInformation{
						Size: 456577876,
						Hex:  "fj933r9fff99g9gd32",
					},
					"1438535410_2015_08_02____20_10_11_34435.tdp": &configure.FoundFilesInformation{
						Size: 1448375,
						Hex:  "fj9j939j9t88232",
					},
					"1438535410_2015_08_02____20_10_12_577263.tdp": &configure.FoundFilesInformation{
						Size: 332495596,
						Hex:  "jifj9e9r33FH8",
					},
					"1438535410_2015_08_02____20_10_13_535663.tdp": &configure.FoundFilesInformation{
						Size: 56239090546,
						Hex:  "afg74y777dff7",
					},
				},*/
			}

			err := updateFiltrationTaskParameters(conn, taskID, &parameters)

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 4: Проверка записи информации о найденном файле, если информация о нем уже существует в БД", func() {
		It("Информация о файле должна быть успешно обновлена", func() {
			parameters := configure.FiltrationTaskParameters{
				ID:                              189,
				Status:                          "execute",
				NumberFilesMeetFilterParameters: 331,
				SizeFilesMeetFilterParameters:   472435353569055,
				NumberDirectoryFiltartion:       4,
				NumberProcessedFiles:            22,
				NumberFilesFoundResultFiltering: 5,
				NumberErrorProcessedFiles:       0,
				SizeFilesFoundResultFiltering:   32455311111,
				PathStorageSource:               "/home/ISEMS_NIH_slave/ISEMS_NIH_slave_RAW/2019_May_14_23_36_3a5c3b12a1790153a8d55a763e26c58e/",
				/*FoundFilesInformation: map[string]*configure.FoundFilesInformation{
					"1438535555_2015_08_02____20_10_11_644263.tdp": &configure.FoundFilesInformation{
						Size: 98765432100,
						Hex:  "ffffffff9339993",
					},
				},*/
			}

			err := updateFiltrationTaskParameters(conn, "ea9e9a0d2e9706bce846171379cbe020", &parameters)

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 5: Получаем всю информацию по выполняемой задачи", func() {
		It("В результате должна быть получена вся информация о задачи фильтрации по ее ID", func() {
			_, err := getInfoTaskForID(conn, taskID)

			//			fmt.Println(err)
			//			fmt.Printf("---------- All information about task -----\n%v\n", ti)

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 6: Получаем всю информацию об источнике", func() {
		It("Должна вернутся вся информация об источнике, интересует его краткое название", func() {
			sourceInfo, err := getInfoSource(conn, 1111)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(sourceInfo[0].ShortName).Should(Equal("ISEMS-NIH_slave 2"))

			/*
				Получить КРАТКОЕ название источника
				в тестах УСПЕШНО, добавить в helpersFunction.getInfoSource()
			*/

		})
	})
})
