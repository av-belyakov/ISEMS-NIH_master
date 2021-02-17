package mytestpackages

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
	fileName := "1502870565_2017_08_16____11_02_45_283.tdp"
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
			decoded := []gopacket.LayerType{}
			parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp, &dns)
			/*err = parser.DecodeLayers(data, &decoded)
			Expect(err).ShouldNot(HaveOccurred())

			for _, layerType := range decoded {
				fmt.Println(layerType)

				switch layerType {
				case layers.LayerTypeIPv6:
					fmt.Printf("    IP6 src:'%v', dst:'%v'\n", ip6.SrcIP, ip6.DstIP)
				case layers.LayerTypeIPv4:
					fmt.Printf("    IP4 src:'%v', dst:'%v'\n", ip4.SrcIP, ip4.DstIP)
				case layers.LayerTypeTCP:
					fmt.Printf("    TCP src port:'%v', dst port:'%v'\n", tcp.SrcPort, tcp.DstPort)
					fmt.Println("    Flags:")
					fmt.Printf("    	FIN: '%v'\n", tcp.FIN)
					fmt.Printf("    	SYN: '%v'\n", tcp.SYN)
					fmt.Printf("    	RST: '%v'\n", tcp.RST)
					fmt.Printf("    	PSH: '%v'\n", tcp.PSH)
					fmt.Printf("    	ACK: '%v'\n", tcp.ACK)
					fmt.Printf("    	URG: '%v'\n", tcp.URG)
				}
			}*/

			for {
				data, ci, err := r.ReadPacketData()
				if err != nil {
					if err == io.EOF {
						break
					}
				}

				_, err = writer.WriteString(fmt.Sprintf("reading packets length: %v\n", ci.CaptureLength))
				err = parser.DecodeLayers(data, &decoded)
				for _, layerType := range decoded {
					fmt.Println(layerType)

					switch layerType {
					case layers.LayerTypeIPv6:
						//fmt.Printf("    IP6 src:'%v', dst:'%v'\n", ip6.SrcIP, ip6.DstIP)
						_, err = writer.WriteString(fmt.Sprintf("    IP6 src:'%v', dst:'%v'\n", ip6.SrcIP, ip6.DstIP))
					case layers.LayerTypeIPv4:
						//fmt.Printf("    IP4 src:'%v', dst:'%v'\n", ip4.SrcIP, ip4.DstIP)
						_, err = writer.WriteString(fmt.Sprintf("    IP4 src:'%v', dst:'%v'\n", ip4.SrcIP, ip4.DstIP))
					case layers.LayerTypeTCP:
						/*fmt.Printf("    TCP src port:'%v', dst port:'%v'\n", tcp.SrcPort, tcp.DstPort)
						fmt.Println("    Flags:")
						fmt.Printf("    	FIN: '%v'\n", tcp.FIN)
						fmt.Printf("    	SYN: '%v'\n", tcp.SYN)
						fmt.Printf("    	RST: '%v'\n", tcp.RST)
						fmt.Printf("    	PSH: '%v'\n", tcp.PSH)
						fmt.Printf("    	ACK: '%v'\n", tcp.ACK)
						fmt.Printf("    	URG: '%v'\n", tcp.URG)*/
						_, err = writer.WriteString(fmt.Sprintf("    TCP src port:'%v', dst port:'%v'\n", tcp.SrcPort, tcp.DstPort))
						_, err = writer.WriteString(fmt.Sprintf("    	Flags:\n	FIN: '%v'\n    	SYN: '%v'\n    	RST: '%v'\n    	PSH: '%v'\n    	ACK: '%v'\n    	URG: '%v'\n", tcp.FIN, tcp.SYN, tcp.RST, tcp.PSH, tcp.ACK, tcp.URG))
					case layers.LayerTypeUDP:
						_, err = writer.WriteString(fmt.Sprintf("    UDP src port:'%v', dst port:'%v'\n", udp.SrcPort, udp.DstPort))
					case layers.LayerTypeDNS:
						_, err = writer.WriteString(fmt.Sprintf("    Questions:'%v', Answers:'%v'\n", dns.Questions, dns.Answers))

					}
				}
			}

			/*
				packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
				//ip4 := packet.Layer(layers.LayerTypeIPv4)
				//fmt.Println(ip4)

				app := packet.ApplicationLayer()

				fmt.Println(app)

				/*for d, b := range layers {
					fmt.Printf("num: %v, byte: %v\n", d, b)
				}*/

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
