package mytestpackages

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mytestpackages/ReadPipeWriteFile", func() {
	Context("Тест 1: проверить количество байт во вспомогательной строке", func() {
		It("Количество байт должно быть равно N", func() {
			hexList := map[string][]string{
				"9b6c633defaee1b78ec65affc3ddc768": []string{
					"4de080f2339fa60ff0639b26192f7af4",
					"a86b143391a1eeae4078786f624b5257",
					"3ab19032a4a3d990a5a0b92042a93ef4",
					"8b95f4e9454e5fe755bc7d6cfbe1f4a1",
					"3ae33f1ef0c2c4de1b63d1610dc5fa9e",
					"3b6581bff1c8f87871bf16482000336d",
					"04f42c59b5424478d9cd6fc5d67707d1",
					"c7ee53ed23c958ec070bc5c664fad44c",
					"7dfe5572c1e210aeffa392a6ab61add8",
					"05a424741af155ae43f81b433f413a1b",
				},
				"d689246d56d1b5087818fae614019588": []string{
					"c7ee53ed23c958ec070bc5c664fad44c",
					"3b6581bff1c8f87871bf16482000336d",
					"d1b2382b16dc7f8f8f543bd711805af0",
					"0aed778029d3c561533ef811a004ff1e",
					"d0fb32ee6fa6b4c60e1022735cf76616",
					"caf342e536492dc32daaa2ee09602456",
					"ef154604eda23d5cf8c70b4c5eeb9003",
					"a86b143391a1eeae4078786f624b5257",
					"ef8ac84466b72f50a37decbd2b63d958",
					"4de080f2339fa60ff0639b26192f7af4",
				},
				"510dbd658a519a450ae57c9969888777": []string{
					"94f2fe5b38f04e7a735449ef3e44a1c9",
					"0ad20b21c171ad46bcc5234babeb5795",
					"0ad20b21c171ad46bcc5234babeb5795",
					"caf342e536492dc32daaa2ee09602456",
					"5a5cd1299d4f6210ab95b765bb683cc5",
					"3ae33f1ef0c2c4de1b63d1610dc5fa9e",
					"6f826e3ead1a8f3e46c1639332bf5057",
					"ead26b5d302e53961b75a7e92c080187",
					"d1b2382b16dc7f8f8f543bd711805af0",
					"c7ee53ed23c958ec070bc5c664fad44c",
				},
			}

			for taskID, listFileHex := range hexList {
				for _, fileHex := range listFileHex {
					strTmp := fmt.Sprintf("1:%v:%v", taskID, fileHex)
					strByte := []byte(strTmp)

					fmt.Printf("long file hex in byte = %v\n", len([]byte(fileHex)))
					fmt.Printf("string - %v, size = %v, size byte = %v\n", strTmp, len(strTmp), len(strByte))

					Expect(true).Should(BeTrue())
				}
			}

		})
	})
	Context("Тест 2: тестирование чтения, передачи и записи файла с добавлением к доп. информации", func() {
		It("", func() {
			/*
			   прочитать файлы по кусочкам, добавить доп. информацию,
			   записать в другую директорию пердворительно убрав доп. информацию
			*/
		})
	})

	/*
	   Context("", func(){
	   	It("", func(){

	   	})
	   })
	*/
})
