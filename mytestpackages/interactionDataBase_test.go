package mytestpackages_test

import (
	"context"
	"fmt"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master/mytestpackages"
)

type configureDB struct {
	Host, Port, NameDB, User, Password string
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

func createNewFiltrationTask(
	connectDB *mongo.Client,
	taskID, clientID, clientTaskID string,
	tf *configure.FiltrationControlCommonParametersFiltration) error {

	fmt.Println("START function 'createNewFiltrationTask_test'...")

	itf := configure.InformationAboutTaskFiltration{
		TaskID:       taskID,
		ClientID:     clientID,
		ClientTaskID: clientTaskID,
		FilteringOption: configure.FiletringOption{
			ID: tf.ID,
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
			TaskStatus: "wait",
		},
	}

	insertData := make([]interface{}, 0, 1)
	insertData = append(insertData, itf)

	fmt.Printf("------- %v --------\n", insertData)

	//InsertData добавляет все данные
	fmt.Println("===== INSERT DATA ======")

	collection := connectDB.Database("isems-nih").Collection("filter_task_list")
	if _, err := collection.InsertMany(context.TODO(), insertData); err != nil {
		return err
	}

	return nil

}

var _ = Describe("InteractionDataBase", func() {
	taskID := common.GetUniqIDFormatMD5("task_id")
	clientID := common.GetUniqIDFormatMD5("client_id")
	clientTaskID := common.GetUniqIDFormatMD5("client_task_id")

	fmt.Printf("TaskID: %v, clientID: %v, clientTaskID: %v\n", taskID, clientID, clientTaskID)

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
				ID: 189,
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
		It("Информация о параметрах фильтрации должна быть успешно обнавлена", func() {
			/*
			   Написать функцию добавляющую информацию
			   о ходе фильтрации в БД, раздел detailed_information_on_filtering
			    и протестировать ее
			*/
		})
	})
})
