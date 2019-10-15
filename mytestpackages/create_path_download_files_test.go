package mytestpackages

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	//. "ISEMS-NIH_master"
)

//NecessaryParametersFiltrationProblem содержит набор параметров для формирования директории
type NecessaryParametersFiltrationProblem struct {
	SourceID                        int
	SourceShortName                 string
	TaskID                          string
	PathRoot                        string
	DateTimeCreateTask              int64
	UseIndex                        bool
	NumberFilesMeetFilterParameters int
	NumberProcessedFiles            int
	NumberFilesFoundResultFiltering int
	NumberDirectoryFiltartion       int
	NumberErrorProcessedFiles       int
	SizeFilesMeetFilterParameters   int64
	SizeFilesFoundResultFiltering   int64
	FiltrationOption                configure.FilteringOption
}

func CreatePathDirectory(npfp *NecessaryParametersFiltrationProblem) (string, error) {
	type timeParameters struct {
		yearStr, mothStr, mothInt, dayStr, timeStr string
	}

	timeConv := func(t time.Time) timeParameters {
		addZero := func(s string) string {
			if len(s) > 1 {
				return s
			}

			return fmt.Sprintf("0%v", s)
		}

		t = t.UTC()

		hour := addZero(strconv.Itoa(t.Hour()))
		min := addZero(strconv.Itoa(t.Minute()))

		return timeParameters{
			yearStr: strconv.Itoa(t.Year()),
			mothStr: t.Month().String(),
			mothInt: common.MothPrintIntAsString(t.Month()),
			dayStr:  strconv.Itoa(t.Day()),
			timeStr: fmt.Sprintf("T%v:%v", hour, min),
		}
	}

	dts := timeConv(time.Unix(npfp.FiltrationOption.DateTime.Start, 0))
	dte := timeConv(time.Unix(npfp.FiltrationOption.DateTime.End, 0))

	sourceName := fmt.Sprintf("%v-%v", npfp.SourceID, strings.ReplaceAll(npfp.SourceShortName, " ", "_"))
	dtStart := fmt.Sprintf("%v.%v.%v%v", dts.dayStr, dts.mothInt, dts.yearStr, dts.timeStr)
	dtEnd := fmt.Sprintf("%v.%v.%v%v", dte.dayStr, dte.mothInt, dte.yearStr, dte.timeStr)
	dirName := fmt.Sprintf("%v-%v_%v", dtStart, dtEnd, npfp.TaskID)

	filePath := path.Join(npfp.PathRoot, "/", sourceName, "/", dts.yearStr, "/", dts.mothStr, "/", dts.dayStr, "/", dirName)

	if err := os.MkdirAll(filePath, 0766); err != nil {
		return "", err
	}

	return filePath, nil
}

func CreateFileReadme(pathStorage string, npfp *NecessaryParametersFiltrationProblem) error {
	type FiltrationControlIPorNetorPortParameters struct {
		Any []string `xml:"any>value"`
		Src []string `xml:"src>value"`
		Dst []string `xml:"dst>value"`
	}

	type FilterSettings struct {
		Protocol      string                                   `xml:"filters>protocol"`
		DateTimeStart string                                   `xml:"filters>date_time_start"`
		DateTimeEnd   string                                   `xml:"filters>date_time_end"`
		IP            FiltrationControlIPorNetorPortParameters `xml:"filters>ip"`
		Port          FiltrationControlIPorNetorPortParameters `xml:"filters>port"`
		Network       FiltrationControlIPorNetorPortParameters `xml:"filters>network"`
	}

	type Information struct {
		XMLName            xml.Name `xml:"information"`
		DateTimeCreateTask string   `xml:"date_time_create_task"`
		SourceID           int      `xml:"source_id"`
		SourceShortName    string   `xml:"source_short_name"`
		TaskID             string   `xml:"task_id"`
		PathRoot           string   `xml:"directory_for_long-term_storage_of_files"`
		FilterSettings
		UseIndex                        bool  `xml:"use_index"`
		NumberFilesMeetFilterParameters int   `xml:"number_files_meet_filter_parameters"`
		NumberProcessedFiles            int   `xml:"number_processed_files"`
		NumberFilesFoundResultFiltering int   `xml:"number_files_found_result_filtering"`
		NumberDirectoryFiltartion       int   `xml:"number_directory_filtartion"`
		NumberErrorProcessedFiles       int   `xml:"number_error_processed_files"`
		SizeFilesMeetFilterParameters   int64 `xml:"size_files_meet_filter_parameters"`
		SizeFilesFoundResultFiltering   int64 `xml:"size_files_found_result_filtering"`
	}

	i := Information{
		UseIndex:                        npfp.UseIndex,
		DateTimeCreateTask:              time.Now().String(),
		SourceID:                        npfp.SourceID,
		SourceShortName:                 npfp.SourceShortName,
		TaskID:                          npfp.TaskID,
		NumberFilesMeetFilterParameters: npfp.NumberFilesMeetFilterParameters,
		NumberProcessedFiles:            npfp.NumberProcessedFiles,
		NumberErrorProcessedFiles:       npfp.NumberErrorProcessedFiles,
		NumberFilesFoundResultFiltering: npfp.NumberFilesFoundResultFiltering,
		NumberDirectoryFiltartion:       npfp.NumberDirectoryFiltartion,
		SizeFilesFoundResultFiltering:   npfp.SizeFilesFoundResultFiltering,
		SizeFilesMeetFilterParameters:   npfp.SizeFilesMeetFilterParameters,
		FilterSettings: FilterSettings{
			Protocol:      npfp.FiltrationOption.Protocol,
			DateTimeStart: time.Unix(npfp.FiltrationOption.DateTime.Start, 0).UTC().String(),
			DateTimeEnd:   time.Unix(npfp.FiltrationOption.DateTime.End, 0).UTC().String(),
			IP: FiltrationControlIPorNetorPortParameters{
				Any: npfp.FiltrationOption.Filters.IP.Any,
				Src: npfp.FiltrationOption.Filters.IP.Src,
				Dst: npfp.FiltrationOption.Filters.IP.Dst,
			},
			Port: FiltrationControlIPorNetorPortParameters{
				Any: npfp.FiltrationOption.Filters.Port.Any,
				Src: npfp.FiltrationOption.Filters.Port.Src,
				Dst: npfp.FiltrationOption.Filters.Port.Dst,
			},
			Network: FiltrationControlIPorNetorPortParameters{
				Any: npfp.FiltrationOption.Filters.Network.Any,
				Src: npfp.FiltrationOption.Filters.Network.Src,
				Dst: npfp.FiltrationOption.Filters.Network.Dst,
			},
		},
		PathRoot: pathStorage,
	}

	output, err := xml.MarshalIndent(i, "  ", "    ")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path.Join(pathStorage, "README.xml"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(output); err != nil {
		return err
	}

	return nil

}

var _ = Describe("Mytestpackages/CreatePathDownloadFiles", func() {
	validePath := "/home/ISEMS_NIH_master/ISEMS_NIH_master_RAW/313-OBU_ITC_Lipetsk/2019/August/11/11.08.2019T15:45-12.08.2019T07:23_hfeh8e83h38gh88485hg48"

	npfp := NecessaryParametersFiltrationProblem{
		SourceID:        313,
		SourceShortName: "OBU ITC Lipetsk",
		TaskID:          "hfeh8e83h38gh88485hg48",
		PathRoot:        "/home/ISEMS_NIH_master/ISEMS_NIH_master_RAW/",
		FiltrationOption: configure.FilteringOption{
			DateTime: configure.TimeInterval{
				Start: 1565538300, // 11.08.2019 15:45:00
				End:   1565594616, // 12.08.2019 07:23:36
			},
			Protocol: "tcp",
			Filters: configure.FilteringExpressions{
				IP: configure.FilteringNetworkParameters{
					Any: []string{"45.62.3.9", "78.3.6.4"},
					Src: []string{"78.100.2.3"},
				},
				Port: configure.FilteringNetworkParameters{
					Dst: []string{"22", "23", "25", "80", "443"},
				},
			},
		},
		NumberFilesMeetFilterParameters: 93,
		NumberProcessedFiles:            93,
		NumberFilesFoundResultFiltering: 4,
		NumberDirectoryFiltartion:       3,
		SizeFilesMeetFilterParameters:   49359594556,
		SizeFilesFoundResultFiltering:   2133445,
	}

	checkDirExist := func(path, name string) bool {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return false
		}

		for _, f := range files {
			if f.Name() == name {
				return true
			}
		}

		return false
	}

	Context("Тест 1: Создание директорий для формирования пути сохранения файлов при скачивании", func() {
		It("Должен быть сформирован путь директорий", func() {
			pathDir, err := CreatePathDirectory(&npfp)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(pathDir).Should(Equal(validePath))
		})
	})

	Context("Тест 2: Наличие директории для хранения файлов", func() {
		It("Должна быть создана директория для хранения файлов", func() {

			dirIsExist := checkDirExist("/home/ISEMS_NIH_master/ISEMS_NIH_master_RAW/313-OBU_ITC_Lipetsk/2019/August/11/", "11.08.2019T15:45-12.08.2019T07:23_hfeh8e83h38gh88485hg48")

			Expect(dirIsExist).Should(BeTrue())
		})
	})

	Context("Тест 3: Создание файла типа README в формате XML", func() {
		It("Должен быть создан файл с описанием в формате XML", func() {

			err := CreateFileReadme(validePath, &npfp)

			Expect(err).ShouldNot(HaveOccurred())

			fileIsExist := checkDirExist(validePath, "README.xml")
			Expect(fileIsExist).Should(BeTrue())
		})
	})

	Context("Тест 4: Создание многоуровневого отображения", func() {
		type testType struct {
			year, month, day int
		}
		testMap := map[int]map[string]*testType{}

		fmt.Printf("11111 Before:%v\n", testMap)

		if len(testMap[1]) == 0 {
			fmt.Println("111 - FEW 0")

			testMap[1] = map[string]*testType{}
		}

		testMap[1]["one"] = &testType{
			year:  2019,
			month: 1,
			day:   10,
		}

		fmt.Printf("11111 After:%v\n", testMap)

		if len(testMap[1]) > 0 {
			fmt.Println("222 - MORE 0")
		}

		testMap[1]["two"] = &testType{
			year:  2019,
			month: 2,
			day:   20,
		}

		fmt.Printf("22222 After:%v\n", testMap)

		It("Должон быть добавленно несколько параметров в одно отображение", func() {
			Expect(testMap[1]["one"].day).Should(Equal(10))
		})
	})
})
