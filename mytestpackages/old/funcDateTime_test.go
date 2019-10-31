package mytestpackages_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//. "ISEMS-NIH_master"
)

func unixTimeConvert(unixTime int64) string {
	fmt.Printf("Current Unix time:%v\n", unixTime)

	return ""
}

//ListFileDescription хранит список файловых дескрипторов и канал для доступа к ним
type ListFileDescription struct {
	list    map[string]*os.File
	chanReq chan channelReqSettings
}

type channelReqSettings struct {
	command, fileHex, filePath string
	channelRes                 chan channelResSettings
}

type channelResSettings struct {
	fd  *os.File
	err error
}

//NewListFileDescription создание нового репозитория со списком дескрипторов файлов
func NewListFileDescription() *ListFileDescription {
	lfd := ListFileDescription{}
	lfd.list = map[string]*os.File{}
	lfd.chanReq = make(chan channelReqSettings)

	go func() {
		for msg := range lfd.chanReq {
			switch msg.command {
			case "add":
				crs := channelResSettings{}
				if _, ok := lfd.list[msg.fileHex]; !ok {
					//создаем дескриптор файла для последующей записи в него
					f, err := os.Create(msg.filePath)
					if err != nil {
						crs.err = err
					} else {
						lfd.list[msg.fileHex] = f
					}
				}

				msg.channelRes <- crs

				close(msg.channelRes)

			case "get":
				crs := channelResSettings{}
				fd, ok := lfd.list[msg.fileHex]
				if !ok {
					crs.err = fmt.Errorf("file descriptor not found")
				} else {
					crs.fd = fd
				}

				msg.channelRes <- crs

				close(msg.channelRes)

			case "del":
				delete(lfd.list, msg.fileHex)

				close(msg.channelRes)
			}
		}
	}()

	return &lfd
}

func (lfd *ListFileDescription) addFileDescription(fh, fp string) error {
	chanRes := make(chan channelResSettings)

	lfd.chanReq <- channelReqSettings{
		command:    "add",
		fileHex:    fh,
		filePath:   fp,
		channelRes: chanRes,
	}

	return (<-chanRes).err
}

func (lfd *ListFileDescription) getFileDescription(fh string) (*os.File, error) {
	chanRes := make(chan channelResSettings)

	lfd.chanReq <- channelReqSettings{
		command:    "get",
		fileHex:    fh,
		channelRes: chanRes,
	}

	res := <-chanRes

	return res.fd, res.err
}

func (lfd *ListFileDescription) delFileDescription(fh string) {
	chanRes := make(chan channelResSettings)

	lfd.chanReq <- channelReqSettings{
		command:    "del",
		fileHex:    fh,
		channelRes: chanRes,
	}

	<-chanRes
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

	Context("Тест 3. Тестируем управление дескриптором файла", func() {
		fileHex := "b73aaca054c920d13500a6ad9beb0c3b"
		filePath := "old/testFileDescription.txt"

		lfd := NewListFileDescription()

		It("Должен быть создан дескриптор файла без ошибки", func() {
			err := lfd.addFileDescription(fileHex, filePath)

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Должен быть получен дескриптор файла", func() {
			fd, err := lfd.getFileDescription(fileHex)

			fmt.Println(fd)

			Expect(err).ShouldNot(HaveOccurred())
		})

		It("После удаления болжна быть ошибка при поиске дескриптора файла", func() {
			lfd.delFileDescription(fileHex)

			os.Remove(filePath)

			_, err := lfd.getFileDescription(fileHex)

			Expect(err).Should(HaveOccurred())
		})
	})

})
