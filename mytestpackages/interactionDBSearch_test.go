package mytestpackages

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mongodb/mongo-go-driver/bson"
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

func getShortInformation(qp QueryParameters, sp *configure.SearchParameters) ([]*configure.BriefTaskInformation, error) {
	queryTemplate := map[string]bson.E{
		"sourceID":          bson.E{Key: "source_id", Value: bson.D{{Key: "$eq", Value: sp.ID}}},
		"filesIsFound":      bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: bson.D{{Key: "$gt", Value: 0}}},
		"taskProcessed":     bson.E{Key: "general_information_about_task.task_processed", Value: sp.TaskProcessed},
		"filesIsDownloaded": bson.E{Key: "detailed_information_on_downloading.number_files_downloaded", Value: bson.D{{Key: "$gt", Value: 0}}},
		"allFilesIsDownloaded": bson.E{Key: "$expr", Value: bson.D{
			{Key: "$eq", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}},
	}

	var querySourceID bson.E
	var queryFilesIsFound bson.E
	var queryTaskProcessed bson.E
	var queryFilesIsDownloaded bson.E
	var queryAllFilesIsDownloaded bson.E

	//поиск по ID источника
	if sp.ID > 0 {
		querySourceID = queryTemplate["sourceID"]
	}

	//была ли задача обработана
	/* !!! Пока закоментированно так как в коллекции только ОДИН документ с полем "general_information_about_task.task_processed"
	if sp.TaskProcessed {
		queryTaskProcessed := bson.E{Key: "general_information_about_task.task_processed", Value: sp.TaskProcessed}
	}
	*/
	//выполнялась ли выгрузка файлов
	if sp.FilesDownloaded.FilesIsDownloaded {
		queryFilesIsDownloaded = queryTemplate["filesIsDownloaded"]
	}

	//все ли файлы были выгружены
	if sp.FilesDownloaded.AllFilesIsDownloaded {
		queryFilesIsDownloaded = queryTemplate["filesIsDownloaded"]
		queryAllFilesIsDownloaded = queryTemplate["allFilesIsDownloaded"]
	}

	//были ли найденны какие либо файлы в результате фильтрации
	if sp.InformationAboutFiltering.FilesIsFound {
		queryFilesIsFound = queryTemplate["filesIsFound"]
	}

	/*
	   Продолжить с параметров:
	   					CountAllFilesMin: 0,
	   					CountAllFilesMax: 0,
	   					SizeAllFilesMin:  0,
	   					SizeAllFilesMax:  0,

	*/

	/*	queryAllFilesIsDownloaded = bson.E{Key: "$expr", Value: bson.D{
			{Key: "$eq", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}}


		queryFilesIsDownloaded = bson.E{Key: "detailed_information_on_downloading.number_files_total", Value: bson.D{{Key: "$gt", Value: 0}}}
		queryAllFilesIsDownloaded = bson.E{Key: "$expr", Value: bson.D{
			{Key: "$eq", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}}

		/*queryAllFilesIsDownloaded = bson.E{Key: "$or", Value: bson.A{bson.E{Key: "detailed_information_on_downloading.number_files_total", Value: bson.D{{Key: "$gt", Value: 0}}},
			bson.E{Key: "$expr", Value: bson.D{
				{Key: "$eq", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}}}}

		/*	bson.E{Key: "$expr", Value: bson.D{
			{Key: "$eq", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}}

			queryAllFilesIsDownloaded = bson.E{
			Key: "detailed_information_on_downloading.number_files_total", Value: bson.D{{Key: "$gt", Value: 0}},
			Key: "$expr", Value: bson.D{
				{Key: "$eq", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}}
	*/
	//	fmt.Printf("querySourceID: %v, %v\n", querySourceID, queryFilesIsDownloaded)

	lbti := []*configure.BriefTaskInformation{}

	cur, err := qp.Find(bson.D{
		querySourceID,
		queryTaskProcessed,
		queryFilesIsDownloaded,
		queryAllFilesIsDownloaded,
		queryFilesIsFound})
	if err != nil {
		return lbti, err
	}

	for cur.Next(context.Background()) {
		var model configure.BriefTaskInformation
		err := cur.Decode(&model)
		if err != nil {
			return lbti, err
		}

		lbti = append(lbti, &model)
	}

	if err := cur.Err(); err != nil {
		return lbti, err
	}

	cur.Close(context.Background())

	return lbti, nil
	/*
		configure.SearchParameters{
				TaskProcessed: false, //обрабатывалась ли задача
				ID:            1010,
				FilesDownloaded: configure.FilesDownloadedOptions{
					FilesIsDownloaded:    false, //выполнялась ли выгрузка файлов
					AllFilesIsDownloaded: false, //все ли файлы были выгружены
				},
				InformationAboutFiltering: configure.InformationAboutFilteringOptions{
					FilesIsFound:     false, //были ли найдены в результате фильтрации какие либо файлы
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
	*/
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

	qp := QueryParameters{
		NameDB:         "isems-nih",
		CollectionName: "task_list",
		ConnectDB:      conn,
	}

	sp := configure.SearchParameters{
		TaskProcessed: false,
		ID:            0,
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
				Start: 0,
				End:   0,
			},
			Protocol: "any",
			NetworkFilters: configure.FiltrationControlParametersNetworkFilters{
				IP: configure.FiltrationControlIPorNetorPortParameters{
					Any: []string{},
					Src: []string{},
					Dst: []string{},
				},
				Port: configure.FiltrationControlIPorNetorPortParameters{
					Any: []string{},
					Src: []string{},
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

	Context("Тест 1: Проверка подключения к БД", func() {
		It("Должно быть установлено подключение с БД", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Тест 2. Тестируем функцию 'getShortInformation'. Запрос к БД для получения всех задач (когда в запросе ничего не задано)", func() {
		It("При выполнения запроса должно быть получено 14 задач", func() {
			listTask, err := getShortInformation(qp, &sp)

			Expect(err).ToNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(14))
		})
	})

	Context("Тест 3. Тестируем функцию 'getShortInformation'. Добавляем ID источника которого НЕТ в базе.", func() {
		spt1 := configure.SearchParameters{}
		spt1.ID = 1000

		listTask, err := getShortInformation(qp, &spt1)

		It("Не должно быть ошибки", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должно быть '0' совпадений", func() {
			Expect(len(listTask)).Should(Equal(0))
		})
	})

	Context("Тест 4. Тестируем функцию 'getShortInformation'. Добавляем ID источника который ПРИСУТСТВУЕТ в базе.", func() {
		spt2 := configure.SearchParameters{}
		spt2.ID = 1221

		listTask, err := getShortInformation(qp, &spt2)

		It("Не должно быть ошибки", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должно быть '14' совпадений", func() {
			Expect(len(listTask)).Should(Equal(14))
		})
	})

	Context("Тест 5. Тестируем функцию 'getShortInformation'. Ищем выполнялась ли выгрузка файлов.", func() {
		spt3 := configure.SearchParameters{}
		spt3.FilesDownloaded.FilesIsDownloaded = true

		listTask, err := getShortInformation(qp, &spt3)

		It("Не должно быть ошибки", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должно быть '8' совпадений", func() {
			Expect(len(listTask)).Should(Equal(8))
		})
	})

	Context("Тест 6. Тестируем функцию 'getShortInformation'. Были ли выгружены ВСЕ файлы.", func() {
		spt4 := configure.SearchParameters{}
		spt4.FilesDownloaded.AllFilesIsDownloaded = true

		listTask, err := getShortInformation(qp, &spt4)

		It("Не должно быть ошибки", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должно быть '7' совпадений", func() {
			Expect(len(listTask)).Should(Equal(7))
		})
	})

	Context("Тест 7. Тестируем функцию 'getShortInformation'. Были ли найдены какие либо файлы найденные в результате фильтрации.", func() {
		spt5 := configure.SearchParameters{}
		spt5.InformationAboutFiltering.FilesIsFound = true

		listTask, err := getShortInformation(qp, &spt5)

		It("Не должно быть ошибки", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должно быть '9' совпадений", func() {
			Expect(len(listTask)).Should(Equal(9))
		})
	})

	/*
	   Context("", func(){
	   	It("", func(){

	   	})
	   })
	*/
})
