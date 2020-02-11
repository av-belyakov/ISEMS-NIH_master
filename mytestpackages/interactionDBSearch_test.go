package mytestpackages

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"

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

func getShortInformation(connectDB *mongo.Client, sp *configure.SearchParameters) ([]*configure.BriefTaskInformation, error) {

	/* Сформировать запрос к MongoDB */
	lbti := []*configure.BriefTaskInformation{}

	/*
		configure.SearchParameters{
				TaskProcessed: false,
				ID:            1010,
				FilesDownloaded: configure.FilesDownloadedOptions{
					FilesIsDownloaded:    false,
					AllFilesIsDownloaded: false,
				},
				InformationAboutFiltering: configure.InformationAboutFilteringOptions{
					FilesIsFound:     false,
					CountAllFilesMin: 0,
					CountAllFilesMax: 0,
					SizeAllFilesMin:  0,
					SizeAllFilesMax:  0,
				},
				InstalledFilteringOption: configure.SearchFilteringOptions{
					DateTime: configure.DateTimeParameters{
						Start: 1576713600,
						End:   1576886400,
					},
					Protocol: "any",
					NetworkFilters: configure.FiltrationControlParametersNetworkFilters{
						IP: configure.FiltrationControlIPorNetorPortParameters{
							Any: []string{"104.238.175.16", "115.171.23.128"},
							Src: []string{"72.105.58.23"},
							Dst: []string{},
						},
						Port: configure.FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{"8080"},
							Dst: []string{},
						},
						Network: configure.FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{},
							Dst: []string{},
						},
					},
				},
			}

		Нужно подготовить запрос к БД основываясь на применении или исключении
		некоторых частей запроса. Например, если параметры
					CountAllFilesMin: 0,
					CountAllFilesMax: 0,
					SizeAllFilesMin:  0,
					SizeAllFilesMax:  0,
		равны 0 то их не учитывать
	*/

	/*
	   cur, err := qp.Find(bson.D{bson.E{Key: "task_id", Value: taskID}})
	   if err != nil {
	   	return lbti, err
	   }

	   for cur.Next(context.TODO()) {
	   	var model configure.InformationAboutTask
	   	err := cur.Decode(&model)
	   	if err != nil {
	   		return lbti, err
	   	}

	   	lbti = append(lbti, model)
	   }

	   if err := cur.Err(); err != nil {
	   	return lbti, err
	   }

	   cur.Close(context.TODO())
	*/
	return lbti, nil
}

var _ = Describe("InteractionDBSearch", func() {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	conn, err := connectToDB(ctx, configureDB{
		Host:     "127.0.0.1",
		Port:     "37017",
		User:     "module_ISEMS-NIH",
		Password: "tkovomfh&ff93",
		NameDB:   "isems-nih",
	})

	sp := configure.SearchParameters{
		TaskProcessed: false,
		ID:            1010,
		FilesDownloaded: configure.FilesDownloadedOptions{
			FilesIsDownloaded:    false,
			AllFilesIsDownloaded: false,
		},
		InformationAboutFiltering: configure.InformationAboutFilteringOptions{
			FilesIsFound:     false,
			CountAllFilesMin: 0,
			CountAllFilesMax: 0,
			SizeAllFilesMin:  0,
			SizeAllFilesMax:  0,
		},
		InstalledFilteringOption: configure.SearchFilteringOptions{
			DateTime: configure.DateTimeParameters{
				Start: 1576713600,
				End:   1576886400,
			},
			Protocol: "any",
			NetworkFilters: configure.FiltrationControlParametersNetworkFilters{
				IP: configure.FiltrationControlIPorNetorPortParameters{
					Any: []string{"104.238.175.16", "115.171.23.128"},
					Src: []string{"72.105.58.23"},
					Dst: []string{},
				},
				Port: configure.FiltrationControlIPorNetorPortParameters{
					Any: []string{},
					Src: []string{"8080"},
					Dst: []string{},
				},
				Network: configure.FiltrationControlIPorNetorPortParameters{
					Any: []string{},
					Src: []string{},
					Dst: []string{},
				},
			},
		},
	}

	fmt.Println(sp)

	Context("Тест 1: Проверка подключения к БД", func() {
		It("Должно быть установлено подключение с БД", func() {

			fmt.Println(conn)

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 2. Тестируем функцию 'getShortInformation'. Выполняем запрос к БД", func() {
		listTask, err := getShortInformation(conn, &sp)

		It("При выполнение функции 'getShortInformation' не должно быть ошибки", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("При выполнение функции 'getShortInformation' должно быть получено то количество задач которое подпадает под заданные параметры", func() {
			//ВРЕМЕННО!!!
			Expect(listTask).Should(BeNil())

			//Expect(len(listTask)).Should(Equal(0))
		})
	})

	/*
	   Context("", func(){
	   	It("", func(){

	   	})
	   })
	*/
})
