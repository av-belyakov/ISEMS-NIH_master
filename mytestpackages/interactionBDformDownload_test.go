package mytestpackages

import (
	"context"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/configure"
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

	client, err := mongo.NewClientWithOptions("mongodb://"+conf.Host+":"+conf.Port+"/"+conf.NameDB, opts)
	if err != nil {
		return nil, err
	}

	client.Connect(ctx)

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}

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

var _ = Describe("InteractionDataBaseFromDownloadFiles", func() {
	taskID := "c440d417dc4386d63403eaed222dc4de"

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

	Context("Тест 2: Запрос к БД для получения списка файлов для скачивания", func() {
		It("Для выбранной в тесте задаче должно быть найдено 32 файла", func() {
			ti, err := getInfoFiltrationTaskForClientTaskID(conn, taskID)

			//			fmt.Println(err)
			//			fmt.Printf("---------- All information about task -----\n%v\n", ti)

			Expect(err).ToNot(HaveOccurred())

			//			fmt.Printf("---- INFORMATION ----\n%v\n", ti)

			numFiles := ti[0].ListFilesResultTaskExecution

			Expect(len(numFiles)).Should(Equal(32))
		})
	})
})
