package mytestpackages

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	//. "ISEMS-NIH_master"
)

//ParametersForPathDirectory содержит набор параметров для формирования директории
type ParametersForPathDirectory struct {
	SourceID                          int
	SourceShortName, TaskID, PathRoot string
	DateTimeStart, DateTimeEnd        int64
}

func CreatePathDirectory(pfpd *ParametersForPathDirectory) (string, error) {
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

		//fmt.Printf("((%v))\n", tm)

		//fmt.Printf("-(%v:%v)-\n", tm.Hour(), tm.Minute())

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

	dts := timeConv(time.Unix(pfpd.DateTimeStart, 0))
	dte := timeConv(time.Unix(pfpd.DateTimeEnd, 0))

	sourceName := fmt.Sprintf("%v-%v", pfpd.SourceID, strings.ReplaceAll(pfpd.SourceShortName, " ", "_"))
	dtStart := fmt.Sprintf("%v.%v.%v%v", dts.dayStr, dts.mothInt, dts.yearStr, dts.timeStr)
	dtEnd := fmt.Sprintf("%v.%v.%v%v", dte.dayStr, dte.mothInt, dte.yearStr, dte.timeStr)

	dirName := fmt.Sprintf("%v-%v_%v", dtStart, dtEnd, pfpd.TaskID)

	filePath := path.Join(pfpd.PathRoot, "/", sourceName, "/", dts.yearStr, "/", dts.mothStr, "/", dts.dayStr, "/", dirName)

	return filePath, nil
}

var _ = Describe("Mytestpackages/CreatePathDownloadFiles", func() {
	Context("Тест 1: Создание директорий для формирования пути сохранения файлов при скачивании", func() {
		It("Должен быть сформирован путь директорий", func() {
			pathDir, err := CreatePathDirectory(&ParametersForPathDirectory{
				SourceID:        313,
				SourceShortName: "AO Tambov",
				TaskID:          "hfeh8e83h38gh88485hg48",
				PathRoot:        "/home/ISEMS_NIH_master/ISEMS_NIH_master_OBJECT/",
				DateTimeStart:   1565538300, // 11.08.2019 15:45:00
				DateTimeEnd:     1565594616, // 12.08.2019 07:23:36
			})

			validePath := "/home/ISEMS_NIH_master/ISEMS_NIH_master_OBJECT/313-AO_Tambov/2019/August/11/11.08.2019T15:45-12.08.2019T07:23_hfeh8e83h38gh88485hg48"

			fmt.Println(pathDir)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(pathDir).Should(Equal(validePath))
		})
	})
})
