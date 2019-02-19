package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"

	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/handlermessageapi"
	"ISEMS-NIH_master/handlerrequestdb"
	"ISEMS-NIH_master/savemessageapp"
)

var appConfig configure.AppConfig
var ism configure.InformationStoringMemory

//ReadConfig читает конфигурационный файл и сохраняет данные в appConfig
func readConfigApp(fileName string, appc *configure.AppConfig) error {
	var err error
	row, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(row, &appc)
	if err != nil {
		return err
	}

	return err
}

//getVersionApp получает версию приложения из файла README.md
func getVersionApp(appc *configure.AppConfig) error {
	failureMessage := "version not found"
	content, err := ioutil.ReadFile(appc.RootDir + "README.md")
	if err != nil {
		return err
	}

	//Application ISEMS-NIH master, v0.1
	pattern := `^Application\sISEMS-NIH\s(master|slave),\sv\d+\.\d+`
	rx := regexp.MustCompile(pattern)
	numVersion := rx.FindString(string(content))

	if len(numVersion) == 0 {
		appc.VersionApp = failureMessage
		return nil
	}

	s := strings.Split(numVersion, " ")
	if len(s) < 3 {
		appc.VersionApp = failureMessage
		return nil
	}

	appc.VersionApp = s[3]

	return nil
}

//connectToDB устанавливает соединение с БД
func connectToDB(ctx context.Context, appc *configure.AppConfig) (*mongo.Client, error) {
	host := appc.ConnectionDB.Host
	port := appc.ConnectionDB.Port
	/*
		user := appc.ConnectionDB.User
		pwd := appc.ConnectionDB.Password
	*/
	optAuth := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    appc.ConnectionDB.NameDB,
		Username:      appc.ConnectionDB.User,
		Password:      appc.ConnectionDB.Password,
	}

	opts := options.Client()
	opts.SetAuth(optAuth)

	client, err := mongo.NewClientWithOptions("mongodb://"+host+":"+strconv.Itoa(port)+"/"+appc.ConnectionDB.NameDB, opts)
	if err != nil {
		return nil, err
	}

	/*client, err := mongo.NewClient("mongodb://" + user + ":" + pwd + "@" + host + ":" + strconv.Itoa(port) + "/" + appc.ConnectionDB.NameDB)
	if err != nil {
		return nil, err
	}*/

	client.Connect(ctx)

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}

func init() {
	fmt.Println("START func init")

	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	ism.MongoConnect.CTX = ctx

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	//читаем конфигурационный файл приложения
	err = readConfigApp(dir+"/config.json", &appConfig)
	if err != nil {
		fmt.Println("Error reading configuration file", err)
		os.Exit(1)
	}

	appConfig.RootDir = dir + "/"

	appConfig.PathKeyFile = appConfig.RootDir + appConfig.PathKeyFile
	appConfig.PathCertFile = appConfig.RootDir + appConfig.PathCertFile

	//соединяемся с БД
	mongoConnect, err := connectToDB(ctx, &appConfig)
	if err != nil {
		_ = saveMessageApp.LogMessage("err", fmt.Sprint(err))

		fmt.Println("Database connection error", err)
		os.Exit(1)
	}

	ism.MongoConnect.Connect = mongoConnect

	//получаем номер версии приложения
	if err = getVersionApp(&appConfig); err != nil {
		_ = saveMessageApp.LogMessage("err", "it is impossible to obtain the version number of the application")
	}

	//инициализируем каналы для взаимодействия с API
	chanMessageToAPI := make(chan configure.MessageAPI)   //к API
	chanMessageFromAPI := make(chan configure.MessageAPI) //из API

	//инициализация модуля для взаимодействия с API (обработчик внешних запросов)
	go handlermessageapi.ProcessingMessageAPI(chanMessageFromAPI, &ism, chanMessageToAPI)

	/* ____ ЗАПИСЬ ТЕСТОВОЙ КОЛЛЕКЦИИ ____ */
	qcs := handlerrequestdb.QueryCollectionSources{
		NameDB:         appConfig.ConnectionDB.NameDB,
		CollectionName: "sources_list",
		ConnectDB:      ism.MongoConnect.Connect,
	}

	res, err := qcs.InsertListSource()
	if err != nil {
		fmt.Println(err)
	}

	isSeccess := "NO"
	if res {
		isSeccess = "YES"
	}

	fmt.Println("\vInser data is ", isSeccess)
	/* __________________________________ */
}

func main() {
	fmt.Println("!!! START func main !!!")

	fmt.Printf("%T%v\n", appConfig, appConfig)

}
