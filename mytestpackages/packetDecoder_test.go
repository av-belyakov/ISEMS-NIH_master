package mytestpackages

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PacketDecoder", func() {
	var file *os.File
	fileName := "test.pcap"
	filePath := "/home/miastr/tmp"

	file, err := os.Open(path.Join(filePath, fileName))
	if err != nil {
		log.Println(fmt.Sprintln(err))

		return
	}
	//	defer file.Close()

	Context("Тест №1. Читаем pcap файл.", func() {
		It("При чтении файла не должно быть ошибок", func() {

			fmt.Printf("Read file: '%v'\n", fileName)

			/*
				for pcap with libpcap

								handle, err := pcap.OpenOffline(path.Join(filePath, fileName))
								defer handle.Close()


								packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
								for packet := range packetSource.Packets() {
									fmt.Println(packet)
								}
			*/

			var err error

			fd, err := os.OpenFile("/home/miastr/tmp/pcapinfo.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer fd.Close()
			Expect(err).ShouldNot(HaveOccurred())

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

			_, err = writer.WriteString(fmt.Sprintf("Decoding file name: %v\n", fileName))
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

			Expect("ddd").ShouldNot(BeNil())
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
