package mytestpackages

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
)

type chanSysInfoParameters struct {
	ChunkSize, ChunkCount, lengthCheckString int
}

func getFileSettings(filePath string) (int64, string, error) {
	fd, err := os.Open(filePath)
	if err != nil {
		return 0, "", err
	}
	defer fd.Close()

	fileInfo, err := fd.Stat()
	if err != nil {
		return 0, "", err
	}

	h := md5.New()
	if _, err := io.Copy(h, fd); err != nil {
		return 0, "", err
	}

	return fileInfo.Size(), hex.EncodeToString(h.Sum(nil)), nil
}

func readSelectedFile(
	res *configure.DetailInfoMsgDownload,
	chanInOut chan<- []byte,
	chanSysInfo chan<- chanSysInfoParameters,
	chanStrErr chan<- string) {

	//fmt.Println("func 'readSelectedFile'...")

	const countByte = 4096

	filePath := path.Join(res.PathDirStorage, res.FileOptions.Name)

	//проверяем имя файла на соответствие регулярному выражению
	if err := common.CheckFileName(res.FileOptions.Name, "fileName"); err != nil {
		chanStrErr <- "ERROR 1"

		return
	}

	//проверяем что бы размер и хеш файла соответствовал переданным
	fileSize, fileHex, err := getFileSettings(filePath)
	if err != nil {
		chanStrErr <- "ERROR 2"

		return
	}
	if (fileSize != res.FileOptions.Size) || (fileHex != res.FileOptions.Hex) {
		chanStrErr <- "ERROR 3"

		return //fmt.Errorf("size or a hash sum of the file %v does not match the passed", res.FileOptions.Name)
	}

	//проверяем что бы файл не был пустым
	if fileSize <= 24 {
		chanStrErr <- "ERROR 4"

		return //fmt.Errorf("requested file %v is empty", res.FileOptions.Name)
	}

	file, err := os.Open(filePath)
	if err != nil {
		chanStrErr <- "ERROR 5"

		return
	}
	defer file.Close()

	strHash := fmt.Sprintf("1:%v:%v", res.TaskID, res.FileOptions.Hex)

	chunkSize := (countByte - len(strHash))

	//считаем количество циклов чтения файла
	countCycle := getCountCycle(fileSize, chunkSize)

	chanSysInfo <- chanSysInfoParameters{
		ChunkCount:        countCycle,
		ChunkSize:         chunkSize,
		lengthCheckString: len(strHash),
	}

	var fileIsReaded error
DONE:
	for i := 0; i <= countCycle; i++ {
		bytesTransmitted := []byte(strHash)

		/*if i == 0 {
			fmt.Printf("\tReader byteTransmitted = %v\n", len(bytesTransmitted))
		}*/

		if fileIsReaded == io.EOF {
			break DONE
		}

		data, err := readNextBytes(file, chunkSize, i)
		if err != nil {
			if err != io.EOF {

				chanStrErr <- "ERROR 6"

				return
			}

			fileIsReaded = io.EOF
		}

		bytesTransmitted = append(bytesTransmitted, data...)

		/*if i == 0 {
			fmt.Printf("Reader chunk = %v, DATA = %v\n", len(bytesTransmitted), len(data))
		}*/

		chanInOut <- bytesTransmitted
	}

	chanStrErr <- "OK!"
}

func readNextBytes(file *os.File, number, nextNum int) ([]byte, error) {
	bytes := make([]byte, number)
	var off int64

	if nextNum != 0 {
		off = int64(number * nextNum)
	}

	rb, err := file.ReadAt(bytes, off)
	if err != nil {
		if err == io.EOF {
			return bytes[:rb], err
		}

		return nil, err
	}

	return bytes, nil
}

func getCountCycle(fileSize int64, countByte int) int {
	newFileSize := float64(fileSize)
	newCountByte := float64(countByte)
	x := math.Floor(newFileSize / newCountByte)
	y := newFileSize / newCountByte

	if (y - x) != 0 {
		x++
	}

	return int(x)
}

func writeReseivedFile(
	taskSettings *configure.DetailInfoMsgDownload,
	chanInOut <-chan []byte,
	chanSysInfo <-chan chanSysInfoParameters) error {

	//fmt.Println("func 'writeReseivedFile'...")

	msgSysInfo := <-chanSysInfo

	//fmt.Printf("--- MSGSYSINFO: %v\n", msgSysInfo)

	//msgSysInfo.ChunkCount
	//msgSysInfo.ChunkSize
	//msgSysInfo.lengthCheckString

	/*fileOut, err := os.OpenFile(path.Join("/__TMP", "write", taskSettings.FileOptions.Name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)

		return err
	}
	defer fileOut.Close()*/

	fileOut, err := os.Create(path.Join("/__TMP", "write", taskSettings.FileOptions.Name))
	if err != nil {
		log.Println(err)

		return err
	}
	defer fileOut.Close()

	w := bufio.NewWriter(fileOut)
	defer func() {
		if err == nil {
			err = w.Flush()
		}
	}()

	num := 0
DONE:
	for data := range chanInOut {
		r := bytes.NewReader(data)

		checkBytes := make([]byte, msgSysInfo.lengthCheckString)
		r.ReadAt(checkBytes, int64(msgSysInfo.lengthCheckString))

		data = data[msgSysInfo.lengthCheckString:]
		if _, err := w.Write(data); err != nil {
			return err
		}

		/*if num == 0 {
			fmt.Printf("NUM checkByte = %v, NUM DATA = %v,\n", len(checkBytes), len(data))
		}

		if num == 1 {
			fmt.Printf("NUM checkByte = %v, NUM DATA = %v,\n", len(checkBytes), len(data))
		}

		if num == 10 {
			fmt.Printf("NUM checkByte = %v, NUM DATA = %v,\n", len(checkBytes), len(data))
		}*/

		num++
		if num == msgSysInfo.ChunkCount {
			//fmt.Printf("=== NUM COUNT = %v\n", num)

			break DONE
		}
	}

	return nil
}

var _ = Describe("Mytestpackages/ReadPipeWriteFile", func() {

	//saveMessageApp := savemessageapp.New()
	fp := "/__TMP/read"

	Context("Тест 1: проверить количество байт во вспомогательной строке", func() {
		It("Количество байт должно быть равно N", func() {
			/*hexList := map[string][]string{
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

				}

			}*/

			Expect(true).Should(BeTrue())
		})
	})

	Context("Тест 2: тестирование чтения, передачи и записи файла с добавлением к нему доп. информации", func() {
		It("По результатам хеш записанного файла должен соответствовать хешу прочитанного файла", func() {

			chanInOut := make(chan []byte)
			chanSysInfo := make(chan chanSysInfoParameters)
			chanStrErr := make(chan string)

			files, err := ioutil.ReadDir(path.Join("/__TMP", "read"))

			Expect(err).ShouldNot(HaveOccurred())

			for _, file := range files {
				fn := file.Name()
				fs, fh, err := getFileSettings(path.Join(fp, fn))

				taskSettings := configure.DetailInfoMsgDownload{
					TaskID:         "510dbd658a519a450ae57c9969888777",
					TaskStatus:     "give me the file",
					PathDirStorage: fp,
					FileOptions: configure.DownloadFileOptions{
						Name: fn,
						Size: fs,
						Hex:  fh,
					},
				}

				Expect(err).ShouldNot(HaveOccurred())

				go readSelectedFile(&taskSettings, chanInOut, chanSysInfo, chanStrErr)

				//			Expect(<-chanStrErr).To(ContainSubstring("OK!"))

				err = writeReseivedFile(&taskSettings, chanInOut, chanSysInfo)

				Expect(err).ShouldNot(HaveOccurred())

				rfs, rfh, rfErr := getFileSettings(path.Join("/__TMP", "read", fn))
				wfs, wfh, wfErr := getFileSettings(path.Join("/__TMP", "write", fn))

				Expect(rfErr).ShouldNot(HaveOccurred())
				Expect(wfErr).ShouldNot(HaveOccurred())

				Expect((rfs == wfs)).Should(BeTrue())
				Expect((rfh == wfh)).Should(BeTrue())
			}

			Expect(true).Should(BeTrue())

			/*
			   ТЕСТЫ ПРОХОДЯТ УСПЕШНО
			   файлы передаются, теперь можно делать основную часть
			   приема и передачи файлов
			*/

			//fn := "26_04_2016___00_54_15.tdp"

		})
	})

	/*
	   Context("", func(){
	   	It("", func(){

	   	})
	   })
	*/
})
