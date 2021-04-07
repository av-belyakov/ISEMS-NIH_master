package mytestpackages_test

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
	var _ = BeforeSuite(func() {
		//удаляем файлы результатов обработки
		func() {
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyIP.txt"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyIP.pcap"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyPPPoE.pcap"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyVLAN.pcap"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyVLAN_IP.pcap"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
			if err := os.Remove("/Users/user/pcap_test_files/pcapinfoFileOnlyVLAN_PPPoE.pcap"); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
		}()
	})

	/*Context("Тест №1. Читаем и декодируем файл сетевого трафика содержащий только ip.", func() {
			It("При выполнении декодирования файла сет. трафика ошибок быть не должно", func() {
				err := networkTrafficDecoder(networkTrafficFileSettingsType{
					filePathIn:  "/Users/user/pcap_test_files/ip",
					fileNameIn:  "1616398942_2021_03_22____10_42_22_21.tdp",
					filePathOut: "/Users/user/pcap_test_files",
					fileNameOut: "pcapinfoFileOnlyIP.txt",
				})

	//временно закоментирован
	//ТЕСТ проходит успешно, для декодирования файлов сет. трафика содержащих ТОЛЬКО IP все выполняется хорошо
	// тесты на файлах которые содержат PPPoE или VLAN пока не проводил (НАДО ПОПРОБОВАТЬ!!!)

				Expect(err).ShouldNot(HaveOccurred())
			})
		})*/

	/*
		   For filteration only VLAN
		   				filePathIn: "/Users/user/pcap_test_files/only_vlan",
		   				fileNameIn: "1616670006_2021_03_25____14_00_06_23.tdp",
		   				"(( (host 46.42.4.164))) || (vlan && (( (host 46.42.4.164))))"

						   Для такого шаблона поиск по IP или же VLAN работает полностью
	*/

	Context("Тест №2. Читаем и выполняем поиск с использованием BPF, файла, содержащего только IP", func() {
		It("При выполнении фильтрации файла содержащего только IP, ошибок быть не должно", func() {
			err := networkTrafficFilter(networkTrafficFileSettingsType{
				filePathIn:  "/Users/user/pcap_test_files/ip",
				fileNameIn:  "1616398942_2021_03_22____10_42_22_21.tdp",
				filePathOut: "/Users/user/pcap_test_files",
				fileNameOut: "pcapinfoFileOnlyIP.pcap",
			}, /*"(( (host 77.241.31.37))) || (vlan && (( (host 77.241.31.37))))")*/
				createPatternScript(FiltrationTasks{
					Protocol: "tcp",
					Filters: FiltrationControlParametersNetworkFilters{
						IP: FiltrationControlIPorNetorPortParameters{
							Any: []string{"77.241.31.37"},
							Src: []string{},
							Dst: []string{},
						},
						Port: FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{},
							Dst: []string{},
						},
						Network: FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{},
							Dst: []string{},
						},
					}}, "ip"))

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	/*
		Шаблон "(pppoes && ip) && (( (host 77.88.21.119))) || (vlan && (( (host 77.88.21.119))))"" handle.SetBPFFilter
		компилировать отказывается, однако шаблон "(pppoes && ip) && (( (host 77.88.21.119)))" успешно проходит
		компиляцию и с ним находятся сетевые пакеты
	*/
	Context("Тест №3. Читаем и выполняем поиск с использованием BPF, файла, содержащего только PPPoE", func() {
		It("При выполнении фильтрации файла содержащего только PPPoE, ошибок быть не должно", func() {
			err := networkTrafficFilter(networkTrafficFileSettingsType{
				filePathIn:  "/Users/user/pcap_test_files/pppoe",
				fileNameIn:  "1616149545_2021_03_19____13_25_45_3596.tdp",
				filePathOut: "/Users/user/pcap_test_files",
				fileNameOut: "pcapinfoFileOnlyPPPoE.pcap",
			}, /*"(host 77.88.21.119 || 10.15.128.47) || ((pppoes && ip) && (( (host 77.88.21.119 || 10.15.128.47))))")*/
				createPatternScript(FiltrationTasks{
					Protocol: "tcp",
					Filters: FiltrationControlParametersNetworkFilters{
						IP: FiltrationControlIPorNetorPortParameters{
							Any: []string{"77.88.21.119", "10.15.128.47"},
							Src: []string{},
							Dst: []string{},
						},
						Port: FiltrationControlIPorNetorPortParameters{
							Any: []string{"443", "51832"},
							Src: []string{},
							Dst: []string{},
						},
						Network: FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{},
							Dst: []string{},
						},
					}}, "pppoe"))

			//   "(host 77.88.21.119 || 10.15.128.47) || ((pppoes && ip) && (( (host 77.88.21.119 || 10.15.128.47))))"
			//   с таким фильтром выполняется фильтрация как просто по IP (10.15.128.47), так и по PPPoE+IP (77.88.21.119)

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест №4. Читаем и выполняем поиск с использованием BPF, файла, содержащего только VLAN и IP", func() {
		It("При выполнении фильтрации файла содержащего только VLAN и IP, ошибок быть не должно", func() {
			err := networkTrafficFilter(networkTrafficFileSettingsType{
				filePathIn:  "/Users/user/pcap_test_files/vlan_ip",
				fileNameIn:  "1617693945_2021_04_06____11_25_45_538.tdp",
				filePathOut: "/Users/user/pcap_test_files",
				fileNameOut: "pcapinfoFileOnlyVLAN_IP.pcap",
			}, //"(host 194.50.141.29 || 37.9.96.23) || ((vlan && ip) && (( (host 194.50.141.29 || 37.9.96.23))))")
				//"(( (host 194.50.141.29 || 37.9.96.23))) || ((pppoes && ip) && (( (host 194.50.141.29 || 37.9.96.23))))")
				createPatternScript(FiltrationTasks{
					Protocol: "tcp",
					Filters: FiltrationControlParametersNetworkFilters{
						IP: FiltrationControlIPorNetorPortParameters{
							Any: []string{"194.50.141.29", "37.9.96.23"},
							Src: []string{},
							Dst: []string{},
						},
						Port: FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{},
							Dst: []string{},
						},
						Network: FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{},
							Dst: []string{},
						},
					}}, "pppoe"))

			// "(host 194.50.141.29 || 37.9.96.23) || ((vlan && ip) && (( (host 194.50.141.29 || 37.9.96.23))))"
			// с таким фильтром выполняется фильтрация как просто по IP (194.50.141.29), так и по VLAN+IP (37.9.96.23)

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест №5. Читаем и выполняем поиск с использованием BPF, файла, содержащего только VLAN, PPPoE и IP", func() {
		It("При выполнении фильтрации файла содержащего только VLAN, PPPoE и IP, ошибок быть не должно", func() {
			err := networkTrafficFilter(networkTrafficFileSettingsType{
				filePathIn:  "/Users/user/pcap_test_files/vlan_and_pppoe",
				fileNameIn:  "1616398304_2021_03_22____10_31_44_3872.tdp",
				filePathOut: "/Users/user/pcap_test_files",
				fileNameOut: "pcapinfoFileOnlyVLAN_PPPoE.pcap",
			}, //"(host 213.155.193.7 || 37.9.96.23) || ((vlan && pppoes && ip) && (( (host 213.155.193.7 || 37.9.96.23))))")
				createPatternScript(FiltrationTasks{
					Protocol: "tcp",
					Filters: FiltrationControlParametersNetworkFilters{
						IP: FiltrationControlIPorNetorPortParameters{
							Any: []string{"213.155.193.7", "37.9.96.23"},
							Src: []string{},
							Dst: []string{},
						},
						Port: FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{},
							Dst: []string{},
						},
						Network: FiltrationControlIPorNetorPortParameters{
							Any: []string{},
							Src: []string{},
							Dst: []string{},
						},
					}}, "vlan/pppoe"))

			// "(host 213.155.193.7 || 37.9.96.23) || ((vlan && pppoes && ip) && (( (host 213.155.193.7 || 37.9.96.23))))"
			// с таким фильтром выполняется фильтрация как просто по IP (194.50.141.29), так и по VLAN+IP (37.9.96.23)

			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	/*Context("Тест №6. Читаем и выполняем поиск с использованием BPF, файла, содержащего первую связку IP, а за ней вторую связку IP", func() {
		It("При выполнении фильтрации файла содержащего первую связку IP, а за ней вторую связку IP, ошибок быть не должно", func() {
			err := networkTrafficFilter(networkTrafficFileSettingsType{
				filePathIn:  "/Users/user/pcap_test_files/ip_and_ip",
				fileNameIn:  "1616155490_2021_03_19____15_04_50_149.tdp",
				filePathOut: "/Users/user/pcap_test_files",
				fileNameOut: "pcapinfoFileOnlyIP_IP.pcap",
			}, "(ether host 87.250.251.15 || 37.9.96.23)")

			// с двумя парами IP адресов в одном ethernet пакете пока не понятно

			Expect(err).ShouldNot(HaveOccurred())
		})
	})*/

	/*
	   '(pppoes && ip) && (( (host 87.250.250.192))) || (vlan && (( (host 87.250.250.192))))'
	   '(pppoes && vlan && ip) && (( (host 87.250.250.192))) || (vlan && (( (host 87.250.250.192))))'
	*/

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

//networkTrafficFileSettingsType
// filePathIn - директория в которой находится файл подлежащий декодированию
// fileNameIn - имя файла подлежащего декодированию
// filePathOut - директория в которою сохраняются результаты декодированию
// fileNameOut - имя файла в который сохраняются результаты дикодирования
type networkTrafficFileSettingsType struct {
	filePathIn  string
	fileNameIn  string
	filePathOut string
	fileNameOut string
}

//networkTrafficDecoder декодировщик сетевого трафика
func networkTrafficDecoder(ntfs networkTrafficFileSettingsType) error {
	inputFile := path.Join(ntfs.filePathIn, ntfs.fileNameIn)

	fmt.Printf("Read file: '%v'\n", inputFile)

	// для файла по которому выполняется декодирование пакетов
	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// для файла в который выполняется запись информации полученной в результате декодирования
	fd, err := os.OpenFile(path.Join(ntfs.filePathOut, ntfs.fileNameOut), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer fd.Close()

	r, err := pcapgo.NewReader(file)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(fd)
	defer func() {
		if err == nil {
			err = writer.Flush()
		}
	}()

	_, writeErr := writer.WriteString(fmt.Sprintf("Decoding file name: %v\n", inputFile))
	if writeErr != nil {
		return writeErr
	}

	var (
		eth layers.Ethernet
		ip4 layers.IPv4
		ip6 layers.IPv6
		tcp layers.TCP
		udp layers.UDP
		dns layers.DNS
		ntp layers.NTP
		tls layers.TLS
	)
	decoded := []gopacket.LayerType{}
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp, &dns, &ntp, &tls)

	boolToInt8 := func(v bool) int8 {
		if v {
			return 1
		}
		return 0
	}

	for {
		data, ci, e := r.ReadPacketData()
		if e != nil {
			if e == io.EOF {
				break
			}
		}

		_, writeErr := writer.WriteString(fmt.Sprintf("timestamp: %v,reading packets length: %v\n", ci.Timestamp, ci.CaptureLength))
		if writeErr != nil {
			break
		}

		e = parser.DecodeLayers(data, &decoded)
		if e != nil {
			continue
		}

		for _, layerType := range decoded {
			switch layerType {
			case layers.LayerTypeIPv6:
				_, writeErr = writer.WriteString(fmt.Sprintf("    IP6 src:'%v', dst:'%v'\n", ip6.SrcIP, ip6.DstIP))
			case layers.LayerTypeIPv4:
				_, writeErr = writer.WriteString(fmt.Sprintf("    IP4 src:'%v', dst:'%v'\n", ip4.SrcIP, ip4.DstIP))
			case layers.LayerTypeTCP:
				_, writeErr = writer.WriteString(fmt.Sprintf("    TCP src port:'%v', dst port:'%v'\n", tcp.SrcPort, tcp.DstPort))

				fin := boolToInt8(tcp.FIN)
				syn := boolToInt8(tcp.SYN)
				rst := boolToInt8(tcp.RST)
				psh := boolToInt8(tcp.PSH)
				ack := boolToInt8(tcp.ACK)
				urg := boolToInt8(tcp.URG)

				_, writeErr = writer.WriteString(fmt.Sprintf("    	Flags	(FIN:'%v' SYN:'%v' RST:'%v' PSH:'%v' ACK:'%v' URG:'%v')\n", fin, syn, rst, psh, ack, urg))
				if len(tcp.Payload) != 0 {
					reader := bufio.NewReader(bytes.NewReader(tcp.Payload))

					httpReq, errHTTP := http.ReadRequest(reader)
					if errHTTP == nil {
						proto := httpReq.Proto
						method := httpReq.Method
						//url := httpReq.URL //содержит целый тип, не только значение httpReq.RequestURI но и методы для парсинга запроса
						host := httpReq.Host
						reqURI := httpReq.RequestURI
						userAgent := httpReq.Header.Get("User-Agent")
						//_, writeErr = writer.WriteString(fmt.Sprintf("%v\n", httpReq.Header))
						_, writeErr = writer.WriteString(fmt.Sprintf("    %v %v %v\n	Host:%v\n	User-Agent:%v\n", proto, method, reqURI, host, userAgent))
					}

					httpRes, errHTTP := http.ReadResponse(reader, httpReq)
					if errHTTP == nil {
						_, writeErr = writer.WriteString(fmt.Sprintf("    StatusCode:%v\n", httpRes.Status))
					}
				}
			case layers.LayerTypeUDP:
				_, writeErr = writer.WriteString(fmt.Sprintf("    UDP src port:'%v', dst port:'%v'\n", udp.SrcPort, udp.DstPort))
			case layers.LayerTypeDNS:
				var resultDNSQuestions, resultDNSAnswers string

				for _, dnsQ := range dns.Questions {
					resultDNSQuestions += string(dnsQ.Name)
				}

				for _, dnsA := range dns.Answers {
					resultDNSAnswers += fmt.Sprintf("%v (%v), %v\n", string(dnsA.Name), dnsA.IP, dnsA.CNAME)
				}

				_, writeErr = writer.WriteString(fmt.Sprintf("    Questions:'%v', Answers:'%v'\n", resultDNSQuestions, resultDNSAnswers))
				//						_, err = writer.WriteString(fmt.Sprintf("    Questions:'%v', Answers:'%v'\n", dns.Questions, dns.Answers))
			case layers.LayerTypeNTP:
				_, writeErr = writer.WriteString(fmt.Sprintf("    Version:'%v'\n", ntp.Version))
			case layers.LayerTypeTLS:
				_, writeErr = writer.WriteString(fmt.Sprintf("    %v\n", tls.Handshake))

			}
		}
	}

	return err
}

//networkTrafficFilter
func networkTrafficFilter(ntfs networkTrafficFileSettingsType, pattern string) error {
	var (
		handle *pcap.Handle
		wf     *os.File
		nwf    *pcapgo.Writer
		err    error
	)

	handle, err = pcap.OpenOffline(path.Join(ntfs.filePathIn, ntfs.fileNameIn))
	if err != nil {
		return err
	}
	defer handle.Close()

	err = handle.SetBPFFilter(pattern)
	if err != nil {
		return err
	}

	wf, err = os.OpenFile(path.Join(ntfs.filePathOut, ntfs.fileNameOut), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer wf.Close()

	nwf = pcapgo.NewWriter(wf)
	if err := nwf.WriteFileHeader(1600, layers.LinkTypeEthernet); err != nil {
		return err
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if err := nwf.WritePacket(gopacket.CaptureInfo{
			Timestamp:      packet.Metadata().Timestamp,
			CaptureLength:  packet.Metadata().CaptureLength,
			Length:         packet.Metadata().Length,
			InterfaceIndex: packet.Metadata().InterfaceIndex,
			AncillaryData:  packet.Metadata().AncillaryData,
		}, packet.Data()); err != nil {
			break
		}
	}

	return err
}
