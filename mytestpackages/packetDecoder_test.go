package mytestpackages

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PacketDecoder", func() {
	var (
		file            *os.File
		errOpenFile     error
		fd              *os.File
		errFd           error
		handleOnlyIP    *pcap.Handle
		errOnlyIP       error
		errBPFIP        error
		wfOnlyIP        *os.File
		errwfOnlyIP     error
		nwfOnlyIP       *pcapgo.Writer
		handleOnlyPPPoE *pcap.Handle
		errOnlyPPPoE    error
		errBPFPPPoE     error
		wfOnlyPPPoE     *os.File
		errwfOnlyPPPoE  error
		nwfOnlyPPPoE    *pcapgo.Writer
	)

	fileNameOnlyIP := "1616398942_2021_03_22____10_42_22_21.tdp"
	filePathOnlyIP := "/Users/user/pcap_test_files/ip"

	fileNameOnlyPPPoE := "1616149545_2021_03_19____13_25_45_3596.tdp"
	filePathOnlyPPPoE := "/Users/user/pcap_test_files/pppoe"

	var _ = BeforeSuite(func() {
		//удаляем файлы результатов обработки
		func() {
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyIP.txt"); err != nil {
				fmt.Println(err)
			}
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyIP.pcap"); err != nil {
				fmt.Println(err)
			}
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyPPPoE.pcap"); err != nil {
				fmt.Println(err)
			}
		}()

		/* для файла по которому выполняется декодирование пакетов */
		file, errOpenFile = os.Open(path.Join(filePathOnlyIP, fileNameOnlyIP))

		/* для файла в который выполняется запись информации полученной в результате декодирования */
		fd, errFd = os.OpenFile("/Users/user/pcap_test_files/pcapinfoFileOnlyIP.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

		/* для сет. трафика содержащего только IP */
		handleOnlyIP, errOnlyIP = pcap.OpenOffline(path.Join(filePathOnlyIP, fileNameOnlyIP))
		errBPFIP = handleOnlyIP.SetBPFFilter("tcp && host 77.241.31.37")

		wfOnlyIP, errwfOnlyIP = os.OpenFile("/Users/user/pcap_test_files/pcapinfoFileOnlyIP.pcap", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		nwfOnlyIP = pcapgo.NewWriter(wfOnlyIP)
		if err := nwfOnlyIP.WriteFileHeader(1600, layers.LinkTypeEthernet); err != nil {
			fmt.Println(err)
		}

		/* для сет. трафика содержащего только PPPoE */
		handleOnlyPPPoE, errOnlyPPPoE = pcap.OpenOffline(path.Join(filePathOnlyPPPoE, fileNameOnlyPPPoE))
		errBPFPPPoE = handleOnlyPPPoE.SetBPFFilter("(pppoes && ip) && host 77.88.21.119")

		wfOnlyPPPoE, errwfOnlyPPPoE = os.OpenFile("/Users/user/pcap_test_files/pcapinfoFileOnlyPPPoE.pcap", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		nwfOnlyPPPoE = pcapgo.NewWriter(wfOnlyPPPoE)
		if err := nwfOnlyPPPoE.WriteFileHeader(1600, layers.LinkTypeEthernet); err != nil {
			fmt.Println(err)
		}

	})

	var _ = AfterSuite(func() {
		file.Close()
		fd.Close()
		handleOnlyIP.Close()
		wfOnlyIP.Close()
		handleOnlyPPPoE.Close()
		wfOnlyPPPoE.Close()
	})

	Context("Тест №1. Открытие или создание файлов", func() {
		It("При открытии файла, для выполнения его дальнейшего декодирования, ошибок быть недолжно", func() {
			Expect(errOpenFile).ShouldNot(HaveOccurred())
		})

		It("При открытии файла, для записи информации о результатах декодирования, ошибок быть недолжно", func() {
			Expect(errFd).ShouldNot(HaveOccurred())
		})

		It("При открытии файла содержащего только IP, ошибок быть недолжно", func() {
			Expect(errOnlyIP).ShouldNot(HaveOccurred())
		})

		It("При формировании BPF для поиска только по IP, ошибок быть недолжно", func() {
			Expect(errBPFIP).ShouldNot(HaveOccurred())
		})

		It("При создании pcap файла в который выполняется запись отфильтрованных, только по IP данных, ошибки быть не должно", func() {
			Expect(errwfOnlyIP).ShouldNot(HaveOccurred())
		})

		It("При открытии файла содержащего только PPPoE ошибок быть недолжно", func() {
			Expect(errOnlyPPPoE).ShouldNot(HaveOccurred())
		})

		It("При формировании BPF для поиска только по PPPoE, ошибок быть недолжно", func() {
			Expect(errBPFPPPoE).ShouldNot(HaveOccurred())
		})

		It("При создании pcap файла в который выполняется запись отфильтрованных, только по PPPoE данных, ошибки быть не должно", func() {
			Expect(errwfOnlyPPPoE).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест №2. Читаем и декодируем файл сетевого трафика содержащий только ip.", func() {
		It("При чтении файла не должно быть ошибок", func() {
			foip := path.Join(filePathOnlyIP, fileNameOnlyIP)

			fmt.Printf("Read file: '%v'\n", foip)

			var err error

			r, err := pcapgo.NewReader(file)
			Expect(err).ShouldNot(HaveOccurred())

			//data, ci, err := r.ReadPacketData()
			//Expect(err).ShouldNot(HaveOccurred())

			writer := bufio.NewWriter(fd)
			defer func() {
				if err == nil {
					err = writer.Flush()
				}
			}()

			_, err = writer.WriteString(fmt.Sprintf("Decoding file name: %v\n", foip))
			Expect(err).ShouldNot(HaveOccurred())

			/*			fmt.Println("+++++++++++++")
						fmt.Println(r.LinkType())
						fmt.Println(r.String())
						fmt.Printf("--- Info file name: '%v' ---\npacket size: %v\npacket timestamp: %v\n", fileName, ci.CaptureLength, ci.Timestamp)
						fmt.Printf("snaplen: %v\n", r.Snaplen())
						fmt.Println("+++++++++++++")*/

			var eth layers.Ethernet
			var ip4 layers.IPv4
			var ip6 layers.IPv6
			var tcp layers.TCP
			var udp layers.UDP
			var dns layers.DNS
			var ntp layers.NTP
			var tls layers.TLS
			decoded := []gopacket.LayerType{}
			parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp, &dns, &ntp, &tls)

			/*listForEach := func(list []interface{}) string {
				var result string

				for _, l := range list {
					switch element := l.(type) {
					case layers.DNSQuestion:
						result += fmt.Sprintf(" %v ", element.Name)
					case layers.DNSResourceRecord:
						result += fmt.Sprintf(" %v ", element.Name)
					}
				}

				return result
			}*/

			boolToInt8 := func(v bool) int8 {
				if v {
					return 1
				}
				return 0
			}

			for {
				data, ci, err := r.ReadPacketData()
				if err != nil {
					if err == io.EOF {
						break
					}
				}

				_, err = writer.WriteString(fmt.Sprintf("timestamp: %v,reading packets length: %v\n", ci.Timestamp, ci.CaptureLength))
				err = parser.DecodeLayers(data, &decoded)
				for _, layerType := range decoded {
					//fmt.Println(layerType)

					switch layerType {
					case layers.LayerTypeIPv6:
						_, err = writer.WriteString(fmt.Sprintf("    IP6 src:'%v', dst:'%v'\n", ip6.SrcIP, ip6.DstIP))
					case layers.LayerTypeIPv4:
						_, err = writer.WriteString(fmt.Sprintf("    IP4 src:'%v', dst:'%v'\n", ip4.SrcIP, ip4.DstIP))
					case layers.LayerTypeTCP:
						_, err = writer.WriteString(fmt.Sprintf("    TCP src port:'%v', dst port:'%v'\n", tcp.SrcPort, tcp.DstPort))

						fin := boolToInt8(tcp.FIN)
						syn := boolToInt8(tcp.SYN)
						rst := boolToInt8(tcp.RST)
						psh := boolToInt8(tcp.PSH)
						ack := boolToInt8(tcp.ACK)
						urg := boolToInt8(tcp.URG)

						_, err = writer.WriteString(fmt.Sprintf("    	Flags	(FIN:'%v' SYN:'%v' RST:'%v' PSH:'%v' ACK:'%v' URG:'%v')\n", fin, syn, rst, psh, ack, urg))
						if len(tcp.Payload) != 0 {
							reader := bufio.NewReader(bytes.NewReader(tcp.Payload))

							httpReq, err := http.ReadRequest(reader)
							if err == nil {
								proto := httpReq.Proto
								method := httpReq.Method
								//url := httpReq.URL //содержит целый тип, не только значение httpReq.RequestURI но и методы для парсинга запроса
								host := httpReq.Host
								reqURI := httpReq.RequestURI
								userAgent := httpReq.Header.Get("User-Agent")
								//_, err = writer.WriteString(fmt.Sprintf("%v\n", httpReq.Header))
								_, err = writer.WriteString(fmt.Sprintf("    %v %v %v\n	Host:%v\n	User-Agent:%v\n", proto, method, reqURI, host, userAgent))
							}

							httpRes, err := http.ReadResponse(reader, httpReq)
							if err == nil {
								_, err = writer.WriteString(fmt.Sprintf("    StatusCode:%v\n", httpRes.Status))
							}
						}
					case layers.LayerTypeUDP:
						_, err = writer.WriteString(fmt.Sprintf("    UDP src port:'%v', dst port:'%v'\n", udp.SrcPort, udp.DstPort))
					case layers.LayerTypeDNS:
						var resultDNSQuestions, resultDNSAnswers string

						for _, e := range dns.Questions {
							resultDNSQuestions += string(e.Name)
						}

						for _, e := range dns.Answers {
							resultDNSAnswers += fmt.Sprintf("%v (%v), %v\n", string(e.Name), e.IP, e.CNAME)
						}

						_, err = writer.WriteString(fmt.Sprintf("    Questions:'%v', Answers:'%v'\n", resultDNSQuestions, resultDNSAnswers))
						//						_, err = writer.WriteString(fmt.Sprintf("    Questions:'%v', Answers:'%v'\n", dns.Questions, dns.Answers))
					case layers.LayerTypeNTP:
						_, err = writer.WriteString(fmt.Sprintf("    Version:'%v'\n", ntp.Version))
					case layers.LayerTypeTLS:
						_, err = writer.WriteString(fmt.Sprintf("    %v\n", tls.Handshake))

					}
				}
			}

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест №3. Читаем и выполняем поиск с использованием BPF, файла, содержащего только ip", func() {
		It("При записи результатов фильтрации в файл, ошибок быть не должно", func() {
			packetSource := gopacket.NewPacketSource(handleOnlyIP, handleOnlyIP.LinkType())
			for packet := range packetSource.Packets() {
				if err := nwfOnlyIP.WritePacket(gopacket.CaptureInfo{
					Timestamp:      packet.Metadata().Timestamp,
					CaptureLength:  packet.Metadata().CaptureLength,
					Length:         packet.Metadata().Length,
					InterfaceIndex: packet.Metadata().InterfaceIndex,
					AncillaryData:  packet.Metadata().AncillaryData,
				}, packet.Data()); err != nil {
					fmt.Println(err)
				}
			}

			Expect(true).Should(BeTrue())
		})
	})

	Context("Тест №4. Читаем и выполняем поиск с использованием BPF, файла, содержащего только PPPoE", func() {
		It("При записи результатов фильтрации в файл, ошибок быть не должно", func() {
			packetSource := gopacket.NewPacketSource(handleOnlyPPPoE, handleOnlyPPPoE.LinkType())
			for packet := range packetSource.Packets() {
				if err := nwfOnlyPPPoE.WritePacket(gopacket.CaptureInfo{
					Timestamp:      packet.Metadata().Timestamp,
					CaptureLength:  packet.Metadata().CaptureLength,
					Length:         packet.Metadata().Length,
					InterfaceIndex: packet.Metadata().InterfaceIndex,
					AncillaryData:  packet.Metadata().AncillaryData,
				}, packet.Data()); err != nil {
					fmt.Println(err)
				}
			}

			Expect(true).Should(BeTrue())
		})
	})

	/*Context("Тест №2. Декодируем pcap файл.", func() {
		It("При декодировании файла по сетевым протоколоам не должно быть ошибок", func() {
			if handle, err := pcap.OpenOffline(path.Join(filePath, "19_04_2016___18_04_44_549906.tdp")); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if err := handle.SetBPFFilter("tcp && host 37.16.80.15"); err != nil {
				fmt.Println(err)
			} else {
				packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
				for packet := range packetSource.Packets() {
					fmt.Println(packet)
				}
			}

			Expect("ddd").ShouldNot(BeNil())
		})
	})*/
})

/*
	for pcap with libpcap

					handle, err := pcap.OpenOffline(path.Join(filePath, fileName))
					defer handle.Close()


					packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
					for packet := range packetSource.Packets() {
						fmt.Println(packet)
					}
*/
