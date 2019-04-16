package mytestpackages_test

import (
	. "ISEMS-NIH_master/mytestpackages"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TimeConvert", func() {
	Context("Тест конвертирования времени", func() {
		It("время должно быть приобразованно из человеческого вида в формат unix time", func() {
			strTime := "Thu, 26 Jan 2012 20:51:50 GMT"

			ut, err := TimeConvert(strTime)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(ut).Should(Equal(1533984300))
		})
	})
})
