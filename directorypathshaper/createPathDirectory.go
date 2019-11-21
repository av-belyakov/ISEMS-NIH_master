package directorypathshaper

import (
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
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

//FileStorageDirectiry создает директорию для хранения файлов и формирует файл README.xml с кратким описание задачи
func FileStorageDirectiry(npfp *NecessaryParametersFiltrationProblem) (string, error) {
	pathStorage, err := CreatePathDirectory(npfp)
	if err != nil {
		return "", err
	}

	if err := CreateFileReadme(pathStorage, npfp); err != nil {
		return pathStorage, err
	}

	return pathStorage, nil
}

//CreatePathDirectory создает каскад директорий и возвращает путь к ним
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

//CreateFileReadme создает XML файл с кратким описанием задачи
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
