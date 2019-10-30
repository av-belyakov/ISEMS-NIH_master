package mytestpackages_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//. "ISEMS-NIH_master"
)

func unixTimeConvert(unixTime int64) string {
	fmt.Printf("Current Unix time:%v\n", unixTime)

	return ""
}

var _ = Describe("Mytestpackages/FuncDateTime", func() {
	Context("Тест 1. Тестирование функции конвертирования даты из Unix формата в человеческий", func() {
		It("Успех", func() {

			unixTimeConvert(time.Now().Unix())

			fmt.Printf("Time Now: %v\n", time.Now())

			Expect(true).Should(BeTrue())
		})
	})

	Context("Тест 2. Преобразование строки в байты и обратно", func() {
		It("Строка должна быть преобразована в срез байт и потом обратно в строку", func() {
			type TestStruct struct {
				Msg *[]byte
			}

			str := "test convert string"
			s := []byte(str)
			ts := TestStruct{
				Msg: &s,
			}

			newStr := string(*ts.Msg)

			Expect(newStr).Should(Equal(str))
		})
	})

})
