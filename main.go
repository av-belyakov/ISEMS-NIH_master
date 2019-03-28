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
	"ISEMS-NIH_master/coreapp"
	"ISEMS-NIH_master/savemessageapp"
)

var appConfig configure.AppConfig
var mongoDBConnect configure.MongoDBConnect

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
	//инициализируем функцию конструктор для записи лог-файлов
	saveMessageApp := savemessageapp.New()

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	mongoDBConnect.CTX = ctx

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

	//для сервера API
	appConfig.ServerAPI.PathCertFile = appConfig.RootDir + appConfig.ServerAPI.PathCertFile
	appConfig.ServerAPI.PathPrivateKeyFile = appConfig.RootDir + appConfig.ServerAPI.PathPrivateKeyFile

	//для сервера обеспечивающего подключение источников
	appConfig.ServerHTTPS.PathCertFile = appConfig.RootDir + appConfig.ServerHTTPS.PathCertFile
	appConfig.ServerHTTPS.PathPrivateKeyFile = appConfig.RootDir + appConfig.ServerHTTPS.PathPrivateKeyFile

	//соединяемся с БД
	mongoConnect, err := connectToDB(ctx, &appConfig)
	if err != nil {
		_ = saveMessageApp.LogMessage("error", fmt.Sprint(err))

		fmt.Println("Database connection error", err)
		os.Exit(1)
	}

	mongoDBConnect.Connect = mongoConnect

	//получаем номер версии приложения
	if err = getVersionApp(&appConfig); err != nil {
		_ = saveMessageApp.LogMessage("error", "it is impossible to obtain the version number of the application")
	}
}

func main() {
	log.Printf("START application ISEMS-NIH_master version %q\n", appConfig.VersionApp)

	//запуск ядра приложения
	coreapp.CoreApp(&appConfig, &mongoDBConnect)
}
