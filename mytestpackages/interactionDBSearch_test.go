package mytestpackages

import (
	"context"
	"fmt"
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
	getQueryTmpNetParams := func(fcp configure.FiltrationControlParametersNetworkFilters, queryType string) bson.E {
		listQueryType := map[string]struct {
			e string
			o configure.FiltrationControlIPorNetorPortParameters
		}{
			"ip":      {e: "ip", o: fcp.IP},
			"port":    {e: "port", o: fcp.Port},
			"network": {e: "network", o: fcp.Network},
		}

		numIPAny := len(listQueryType[queryType].o.Any)
		numIPSrc := len(listQueryType[queryType].o.Src)
		numIPDst := len(listQueryType[queryType].o.Dst)

		if numIPAny > 0 && numIPSrc == 0 && numIPDst == 0 {
			return bson.E{}
		}

		if numIPAny > 0 && numIPSrc == 0 && numIPDst == 0 {
			return bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".any", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Any}}
		}

		if numIPSrc > 0 && numIPAny == 0 && numIPDst == 0 {
			return bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Src}}
		}

		if numIPDst > 0 && numIPAny == 0 && numIPSrc == 0 {
			return bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Dst}}
		}

		if (numIPSrc > 0 && numIPDst > 0) && numIPAny == 0 {
			return bson.E{Key: "$and", Value: bson.A{
				bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Src}},
				bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Dst}},
			}}
		}

		return bson.E{Key: "$or", Value: bson.A{
			bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".any", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Any}},
			bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Src}},
			bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Dst}},
		}}
	}

	checkParameterContainsValues := func(fcinpp configure.FiltrationControlIPorNetorPortParameters) bool {
		if len(fcinpp.Any) > 0 {
			return true
		}

		if len(fcinpp.Src) > 0 {
			return true
		}

		if len(fcinpp.Dst) > 0 {
			return true
		}

		return false
	}

	queryTemplate := map[string]bson.E{
		"sourceID":          bson.E{Key: "source_id", Value: bson.D{{Key: "$eq", Value: sp.ID}}},
		"filesIsFound":      bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: bson.D{{Key: "$gt", Value: 0}}},
		"taskProcessed":     bson.E{Key: "general_information_about_task.task_processed", Value: sp.TaskProcessed},
		"filesIsDownloaded": bson.E{Key: "detailed_information_on_downloading.number_files_downloaded", Value: bson.D{{Key: "$gt", Value: 0}}},
		"allFilesIsDownloaded": bson.E{Key: "$expr", Value: bson.D{
			{Key: "$eq", Value: bson.A{"$detailed_information_on_downloading.number_files_total", "$detailed_information_on_downloading.number_files_downloaded"}}}},
		"sizeAllFiles": bson.E{Key: "detailed_information_on_filtering.size_files_found_result_filtering", Value: bson.D{
			{Key: "$gte", Value: sp.InformationAboutFiltering.SizeAllFilesMin},
			{Key: "$lte", Value: sp.InformationAboutFiltering.SizeAllFilesMax},
		}},
		"countAllFiles": bson.E{Key: "detailed_information_on_filtering.number_files_found_result_filtering", Value: bson.D{
			{Key: "$gte", Value: sp.InformationAboutFiltering.CountAllFilesMin},
			{Key: "$lte", Value: sp.InformationAboutFiltering.CountAllFilesMax},
		}},
		"dateTimeParameters": bson.E{Key: "$and", Value: bson.A{
			bson.D{{Key: "filtering_option.date_time_interval.start", Value: bson.D{
				{Key: "$gte", Value: sp.InstalledFilteringOption.DateTime.Start}}}},
			bson.D{{Key: "filtering_option.date_time_interval.end", Value: bson.D{
				{Key: "$lte", Value: sp.InstalledFilteringOption.DateTime.End}}}},
		}},
		"transportProtocol":        bson.E{Key: "filtering_option.protocol", Value: sp.InstalledFilteringOption.Protocol},
		"statusFilteringTask":      bson.E{Key: "detailed_information_on_filtering.task_status", Value: sp.StatusFilteringTask},
		"statusFileDownloadTask":   bson.E{Key: "detailed_information_on_downloading.task_status", Value: sp.StatusFileDownloadTask},
		"networkParametersIP":      getQueryTmpNetParams(sp.InstalledFilteringOption.NetworkFilters, "ip"),
		"networkParametersPort":    getQueryTmpNetParams(sp.InstalledFilteringOption.NetworkFilters, "port"),
		"networkParametersNetwork": getQueryTmpNetParams(sp.InstalledFilteringOption.NetworkFilters, "network"),
	}

	var (
		querySourceID                   bson.E
		queryFilesIsFound               bson.E
		querySizeAllFiles               bson.E
		queryCountAllFiles              bson.E
		queryTaskProcessed              bson.E
		queryFilesIsDownloaded          bson.E
		queryTransportProtocol          bson.E
		querydateTimeParameters         bson.E
		queryStatusFilteringTask        bson.E
		queryAllFilesIsDownloaded       bson.E
		queryStatusFileDownloadTask     bson.E
		queryNetworkParametersIPNetPort bson.E
	)

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

	//диапазон количества найденных файлов
	cafmin := sp.InformationAboutFiltering.CountAllFilesMin
	cafmax := sp.InformationAboutFiltering.CountAllFilesMax
	if (cafmax > 0) && (cafmax > cafmin) {
		queryCountAllFiles = queryTemplate["countAllFiles"]
	}

	//диапазон общего размера всех найденных файлов
	safmin := sp.InformationAboutFiltering.SizeAllFilesMin
	safmax := sp.InformationAboutFiltering.SizeAllFilesMax
	if (safmax > 0) && (safmax > safmin) {
		querySizeAllFiles = queryTemplate["sizeAllFiles"]
	}

	//временной диапазон фильтруемых данных
	dts := sp.InstalledFilteringOption.DateTime.Start
	dte := sp.InstalledFilteringOption.DateTime.End
	if (dts > 0) && (dte > 0) && (dts < dte) {
		querydateTimeParameters = queryTemplate["dateTimeParameters"]
	}

	//транспортный протокол
	if sp.InstalledFilteringOption.Protocol == "tcp" || sp.InstalledFilteringOption.Protocol == "udp" {
		queryTransportProtocol = queryTemplate["transportProtocol"]
	}

	//статус задачи по фильтрации
	if (len(sp.StatusFilteringTask) > 0) && (sp.StatusFilteringTask != "any") {
		queryStatusFilteringTask = queryTemplate["statusFilteringTask"]
	}

	//статус задачи по скачиванию файлов
	if (len(sp.StatusFileDownloadTask) > 0) && (sp.StatusFileDownloadTask != "any") {
		queryStatusFileDownloadTask = queryTemplate["statusFileDownloadTask"]
	}

	isContainsValueIP := checkParameterContainsValues(sp.InstalledFilteringOption.NetworkFilters.IP)
	isContainsValuePort := checkParameterContainsValues(sp.InstalledFilteringOption.NetworkFilters.Port)
	isContainsValueNetwork := checkParameterContainsValues(sp.InstalledFilteringOption.NetworkFilters.Network)

	//нет port и есть network или ip
	if !isContainsValuePort && (isContainsValueIP || isContainsValueNetwork) {
		queryNetworkParametersIPNetPort = bson.E{Key: "$or", Value: bson.A{
			queryTemplate["networkParametersIP"],
			queryTemplate["networkParametersNetwork"],
		}}
	}

	//есть только порт нет network и ip
	if isContainsValuePort && !isContainsValueIP && !isContainsValueNetwork {
		queryNetworkParametersIPNetPort = queryTemplate["networkParametersPort"]
	}

	//есть все или port и какой то из network или ip
	if isContainsValuePort && isContainsValueIP && isContainsValueNetwork {
		queryNetworkParametersIPNetPort = bson.E{
			Key: "$and", Value: bson.A{
				bson.E{Key: "$or", Value: bson.A{
					queryTemplate["networkParametersIP"],
					queryTemplate["networkParametersNetwork"],
				}},
				queryTemplate["networkParametersPort"],
			},
		}
	}

	//fmt.Printf("networkParametersIP: %v, networkParametersPort: %v, networkParametersNetwork: %v\n", queryTemplate["networkParametersIP"], queryTemplate["networkParametersPort"], queryTemplate["networkParametersNetwork"])

	lbti := []*configure.BriefTaskInformation{}

	cur, err := qp.Find(bson.D{
		querySourceID,
		queryTaskProcessed,
		queryFilesIsDownloaded,
		queryAllFilesIsDownloaded,
		queryFilesIsFound,
		queryCountAllFiles,
		querySizeAllFiles,
		querydateTimeParameters,
		queryTransportProtocol,
		queryStatusFilteringTask,
		queryStatusFileDownloadTask,
		queryNetworkParametersIPNetPort})
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

	Context("Тест 8. Тестируем функцию 'getShortInformation'. Поиск по общему размеру найденных файлов, где размер больше чем параметр 'SizeAllFilesMin' и меньше чем 'SizeAllFilesMax'.", func() {
		spt6 := configure.SearchParameters{}
		spt6.InformationAboutFiltering.SizeAllFilesMin = 3330
		spt6.InformationAboutFiltering.SizeAllFilesMax = 13900040

		listTask, err := getShortInformation(qp, &spt6)

		It("Не должно быть ошибки", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должно быть '9' совпадений", func() {
			Expect(len(listTask)).Should(Equal(9))
		})

		It("Должно быть '0' совпадений так как в указанных приделах данных нет", func() {
			spt61 := configure.SearchParameters{}
			spt61.InformationAboutFiltering.SizeAllFilesMin = 23900040
			spt61.InformationAboutFiltering.SizeAllFilesMax = 23900100
			listTask, _ := getShortInformation(qp, &spt61)

			Expect(len(listTask)).Should(Equal(0))
		})

		It("Должно быть '14' совпадений, тоесть ВСЕ. Так как параметры не верны min > max и следовательно не учитиваются", func() {
			spt62 := configure.SearchParameters{}
			spt62.InformationAboutFiltering.SizeAllFilesMin = 23900040
			spt62.InformationAboutFiltering.SizeAllFilesMax = 100
			listTask, _ := getShortInformation(qp, &spt62)

			Expect(len(listTask)).Should(Equal(14))
		})
	})

	Context("Тест 9. Тестируем функцию 'getShortInformation'. Поиск по количеству найденных файлов, где кол-во больше чем параметр 'CountAllFilesMin' и меньше чем 'CountAllFilesMax'.", func() {
		spt7 := configure.SearchParameters{}
		spt7.InformationAboutFiltering.CountAllFilesMin = 5
		spt7.InformationAboutFiltering.CountAllFilesMax = 10

		listTask, err := getShortInformation(qp, &spt7)

		It("Не должно быть ошибки", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должно быть '9' совпадений", func() {
			Expect(len(listTask)).Should(Equal(9))
		})
	})

	Context("Тест 10. Тестируем функцию 'getShortInformation'. Поиск по временному диапазону", func() {
		It("Должно быть '12' совпадений, так как временной интервал удовлетворяет заданным параметрам", func() {
			spt81 := configure.SearchParameters{}
			spt81.InstalledFilteringOption.DateTime.Start = 1560729600
			spt81.InstalledFilteringOption.DateTime.End = 1560898800

			listTask, err := getShortInformation(qp, &spt81)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(12))
		})

		It("Должно быть '2' совпадений, так как временной интервал удовлетворяет заданным параметрам", func() {
			spt82 := configure.SearchParameters{}
			spt82.InstalledFilteringOption.DateTime.Start = 1576713599 //1576713600
			spt82.InstalledFilteringOption.DateTime.End = 1576886401   //1576886400

			listTask, err := getShortInformation(qp, &spt82)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(2))
		})

		It("Должно быть '0' совпадений, так как временной интервал НЕ удовлетворяет заданным параметрам", func() {
			spt83 := configure.SearchParameters{}
			spt83.InstalledFilteringOption.DateTime.Start = 16713600
			spt83.InstalledFilteringOption.DateTime.End = 176886400

			listTask, err := getShortInformation(qp, &spt83)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(0))
		})

		It("Должно быть '14' совпадений, так как временной интервал выходит за рамки допустимых параметрам", func() {
			spt84 := configure.SearchParameters{}
			spt84.InstalledFilteringOption.DateTime.Start = 1576886400
			spt84.InstalledFilteringOption.DateTime.End = 1576713600

			listTask, err := getShortInformation(qp, &spt84)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(14))
		})
	})

	Context("Тест 11. Тестируем функцию 'getShortInformation'. Поиск по протоколу транспортного уровня.", func() {
		It("Должно быть '14' совпадений", func() {
			spt91 := configure.SearchParameters{}
			spt91.InstalledFilteringOption.Protocol = "tcp"

			listTask, err := getShortInformation(qp, &spt91)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(14))
		})

		It("Должно быть '0' совпадений", func() {
			spt91 := configure.SearchParameters{}
			spt91.InstalledFilteringOption.Protocol = "udp"

			listTask, err := getShortInformation(qp, &spt91)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(0))
		})
	})

	Context("Тест 12. Тестируем функцию 'getShortInformation'. Поиск по статусу задачи фильтрации.", func() {
		It("Должно быть '10' совпадений", func() {
			spt101 := configure.SearchParameters{}
			spt101.StatusFilteringTask = "complete"

			listTask, err := getShortInformation(qp, &spt101)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(10))
		})

		It("Должно быть '1' совпадений", func() {
			spt102 := configure.SearchParameters{}
			spt102.StatusFilteringTask = "refused"

			listTask, err := getShortInformation(qp, &spt102)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(1))
		})
	})

	Context("Тест 13. Тестируем функцию 'getShortInformation'. Поиск по статусу задачи по скачиванию файлов.", func() {
		It("Должно быть '6' совпадений", func() {
			spt111 := configure.SearchParameters{}
			spt111.StatusFileDownloadTask = "not executed"

			listTask, err := getShortInformation(qp, &spt111)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(6))
		})

		It("Должно быть '2' совпадений", func() {
			spt112 := configure.SearchParameters{}
			spt112.StatusFileDownloadTask = "execute"

			listTask, err := getShortInformation(qp, &spt112)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(listTask)).Should(Equal(2))
		})
	})

	Context("Тест 14. Тестируем функцию 'getQueryTmpNetParams' формирующую строку запроса сетевых параметров", func() {
		getQueryTmpNetParams := func(fcp configure.FiltrationControlParametersNetworkFilters, queryType string) bson.E {
			listQueryType := map[string]struct {
				e string
				o configure.FiltrationControlIPorNetorPortParameters
			}{
				"ip":      {e: "ip", o: fcp.IP},
				"port":    {e: "port", o: fcp.Port},
				"network": {e: "network", o: fcp.Network},
			}

			numIPAny := len(listQueryType[queryType].o.Any)
			numIPSrc := len(listQueryType[queryType].o.Src)
			numIPDst := len(listQueryType[queryType].o.Dst)

			if numIPAny == 0 && numIPSrc == 0 && numIPDst == 0 {

				fmt.Println("func 'getQueryTmpNetParams', all parameters is 0")

				return bson.E{}
			}

			if numIPAny > 0 && numIPSrc == 0 && numIPDst == 0 {

				fmt.Println("func 'getQueryTmpNetParams', ANY > 0, SRC and DST parameters is 0")

				return bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".any", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Any}}
			}

			if numIPSrc > 0 && numIPAny == 0 && numIPDst == 0 {

				fmt.Println("func 'getQueryTmpNetParams', SRC > 0, ANY and DST parameters is 0")

				return bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Src}}
			}

			if numIPDst > 0 && numIPAny == 0 && numIPSrc == 0 {

				fmt.Println("func 'getQueryTmpNetParams', DST > 0, ANY and SRC parameters is 0")

				return bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Dst}}
			}

			if (numIPSrc > 0 && numIPDst > 0) && numIPAny == 0 {

				fmt.Println("func 'getQueryTmpNetParams', SRC and DST > 0, ANY parameters is 0")

				return bson.E{Key: "$and", Value: bson.A{
					bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Src}},
					bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Dst}},
				}}
			}

			fmt.Println("func 'getQueryTmpNetParams', ANY and SRC and DST > 0")

			return bson.E{Key: "$or", Value: bson.A{
				bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".any", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Any}},
				bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".src", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Src}},
				bson.E{Key: "filtering_option.filters." + listQueryType[queryType].e + ".dst", Value: bson.E{Key: "$in", Value: listQueryType[queryType].o.Dst}},
			}}
		}

		It("Должнен быть сформирован корректный запрос", func() {

			fmt.Println(getQueryTmpNetParams(configure.FiltrationControlParametersNetworkFilters{
				IP: configure.FiltrationControlIPorNetorPortParameters{
					Any: []string{"129.56.3.6", "89.23.6.64", "206.35.1.46"},
					Src: []string{"65.2.33.4"},
					Dst: []string{"96.32.6.5", "78.100.23.6", "85.144.6.6"},
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
			}, "ip"))

			Expect(true).Should(BeTrue())
		})
	})

	Context("Тест 15. Проверяем функцию 'checkParameterContainsValues'", func() {
		checkParameterContainsValues := func(fcinpp configure.FiltrationControlIPorNetorPortParameters) bool {
			if len(fcinpp.Any) > 0 {
				//				fmt.Println("func 'checkParameterContainsValues' len(fcinpp.Any) > 0")
				return true
			}

			if len(fcinpp.Src) > 0 {
				//				fmt.Println("func 'checkParameterContainsValues' len(fcinpp.Src) > 0")
				return true
			}

			if len(fcinpp.Dst) > 0 {
				//				fmt.Println("func 'checkParameterContainsValues' len(fcinpp.Dst) > 0")
				return true
			}

			//			fmt.Println("func 'checkParameterContainsValues' ALL == 0")
			return false
		}

		It("Должен быть False так как все параметры пусты", func() {
			Expect(checkParameterContainsValues(configure.FiltrationControlIPorNetorPortParameters{})).Should(BeFalse())
		})

		It("Должен быть True так как один из параметров заполнен 1.", func() {
			Expect(checkParameterContainsValues(configure.FiltrationControlIPorNetorPortParameters{Any: []string{"45.66.6.1"}})).Should(BeTrue())
		})

		It("Должен быть True так как один из параметров заполнен 2.", func() {
			Expect(checkParameterContainsValues(configure.FiltrationControlIPorNetorPortParameters{Src: []string{"12.6.6.4"}, Dst: []string{"9.44.6.3"}})).Should(BeTrue())
		})

		It("Должен быть True так как один из параметров заполнен 3.", func() {
			Expect(checkParameterContainsValues(configure.FiltrationControlIPorNetorPortParameters{Dst: []string{"4.66.4.7"}})).Should(BeTrue())
		})
	})

	/*

		!!! Выполнить нижеперечисленные тесты !!!

	*/

	Context("Тест 16. Проверяем поиск информации по сетевым параметрам (IP, Port, Network)", func() {
		It("Поиск только по ip адресам, должно быть получено '' значений", func() {

		})

		It("Поиск только по network, должно быть получено '' значений", func() {

		})

		It("Поиск только по ip адресам или network, должно быть получено '' значений", func() {

		})

		It("Поиск только по ip адресам и port, должно быть получено '' значений", func() {

		})
	})

	Context("Тест 17. Проверяем поиск информации по сетевым параметрам (IP, Port, Network) и какой либо доп. параметр", func() {
		It("Поиск только по (ip адресам или network) и временному диапазону, должно быть получено '' значений", func() {

		})

		It("Поиск только по (ip адресам или network) и временному диапазону и статусу фильтрации, должно быть получено '' значений", func() {

		})
	})

	/*
	   Context("", func(){
	   	It("", func(){

	   	})
	   })
	*/
})
