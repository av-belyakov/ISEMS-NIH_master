package mytestpackages_test

import (
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"

	"github.com/google/gopacket/pcap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//"ISEMS-NIH_master/mytestpackages"
)

type HandlerType struct {
	Handler *pcap.Handle
}

func (ht *HandlerType) getHandler() (*pcap.Handle, error) {
	if ht.Handler != nil {
		return ht.Handler, nil
	}

	handle, err := pcap.OpenOffline(path.Join("/Users/user/pcap_test_files/vlan_ip", "1617693945_2021_04_06____11_25_45_538.tdp"))
	ht.Handler = handle

	return ht.Handler, err
}

func (ht *HandlerType) closeHandler() {
	ht.Handler.Close()
}

var _ = Describe("CreatePatternScript", func() {
	ht := HandlerType{}

	var _ = AfterSuite(func() {
		ht.closeHandler()
	})

	Context("Тест 1. Проверка формирования шаблона скрипта содержащего BPF фильтр ТОЛЬКО с одними IP адресами", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
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
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP) Тест 1. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect((len(pattern) > 0)).Should(BeTrue())

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 1.1 Проверка формирования шаблона скрипта содержащего BPF фильтр ТОЛЬКО с одними IP адресами", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "udp",
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
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP and UDP) Тест 1.1. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect((len(pattern) > 0)).Should(BeTrue())

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 1.2. Проверка формирования шаблона скрипта содержащего BPF фильтр ТОЛЬКО с одними IP адресами ANY, SRC и DST", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{"194.50.141.29", "37.9.96.23"},
						Src: []string{"23.1.34.5"},
						Dst: []string{"90.1.2.3"},
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
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP any, src, dst and proto) Тест 1.2. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect((len(pattern) > 0)).Should(BeTrue())

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 1.3. Проверка формирования шаблона скрипта содержащего BPF фильтр ТОЛЬКО с одними IP адресами SRC и DST", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{"23.1.34.5", "194.50.141.29"},
						Dst: []string{"90.1.2.3", "194.50.141.29"},
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
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP src, dst and proto) Тест 1.3. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect((len(pattern) > 0)).Should(BeTrue())

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 1.4. Проверка формирования шаблона скрипта содержащего BPF фильтр ТОЛЬКО с одними IP адресами ANY и SRC", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				//Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{"194.50.141.29", "37.9.96.23"},
						Src: []string{"23.1.34.5", "90.1.2.3"},
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
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP any и src and proto) Тест 1.4. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect((len(pattern) > 0)).Should(BeTrue())

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 2. Проверка формирования шаблона скрипта содержащего BPF фильтр ТОЛЬКО с IP адресами", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{"194.50.141.29", "37.9.96.23"},
						Src: []string{},
						Dst: []string{"45.2.13.4", "87.100.0.34"},
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
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP any, dst) Тест 1.5. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()

			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect((len(pattern) > 0)).Should(BeTrue())

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 2. Проверка формирования шаблона скрипта содержащего BPF фильтр ТОЛЬКО с IP адресами и подсетями", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
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
						Any: []string{"34.2.3.4/32"},
						Src: []string{},
						Dst: []string{"45.2.13.4/32", "87.100.0.34/25"},
					},
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP any, NETWORK any, dst) Тест 2. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()

			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect((len(pattern) > 0)).Should(BeTrue())

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 3. Проверка формирования шаблона скрипта содержащего BPF фильтр с IP адресами, портами и подсетями", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{"194.50.141.29", "37.9.96.23"},
						Src: []string{},
						Dst: []string{},
					},
					Port: FiltrationControlIPorNetorPortParameters{
						Any: []string{"80", "443"},
						Src: []string{},
						Dst: []string{},
					},
					Network: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{"34.2.3.4/32", "67.100.0.2/24"},
						Dst: []string{},
					},
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP any, PORT any, NETWORK src)Тест 3. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect((len(pattern) > 0)).Should(BeTrue())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 3.1. Проверка формирования шаблона скрипта содержащего BPF фильтр с IP адресами, портами и подсетями", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				//Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{"194.50.141.29", "37.9.96.23"},
						Src: []string{},
						Dst: []string{},
					},
					Port: FiltrationControlIPorNetorPortParameters{
						Any: []string{"80", "443"},
						Src: []string{},
						Dst: []string{"22", "23"},
					},
					Network: FiltrationControlIPorNetorPortParameters{
						Any: []string{"56.100.2.3/26", "10.23.2.4/27"},
						Src: []string{"34.2.3.4/32", "67.100.0.2/24"},
						Dst: []string{},
					},
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP any, PORT any, dst, NETWORK any, src) Тест 3.1. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect((len(pattern) > 0)).Should(BeTrue())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	//------------ начало новых тестов ----
	Context("Тест 4.1. Проверка формирования шаблона скрипта содержащего BPF фильтр с IP адресами и портами", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{"194.50.141.29"},
						Src: []string{},
						Dst: []string{},
					},
					Port: FiltrationControlIPorNetorPortParameters{
						Any: []string{"80", "443"},
						Src: []string{},
						Dst: []string{},
					},
					Network: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{},
						Dst: []string{},
					},
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP any, PORT any) Тест 4.1. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect((len(pattern) > 0)).Should(BeTrue())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 4.2. Проверка формирования шаблона скрипта содержащего BPF фильтр с IP адресами и портами", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{"194.50.141.29"},
						Dst: []string{"67.3.2.45"},
					},
					Port: FiltrationControlIPorNetorPortParameters{
						Any: []string{"443"},
						Src: []string{},
						Dst: []string{},
					},
					Network: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{},
						Dst: []string{},
					},
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP src, dst, PORT any) Тест 4.2. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect((len(pattern) > 0)).Should(BeTrue())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 4.3. Проверка формирования шаблона скрипта содержащего BPF фильтр с IP адресами и портами", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{"67.3.2.45", "194.50.141.29"},
						Dst: []string{},
					},
					Port: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{"443"},
						Dst: []string{},
					},
					Network: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{},
						Dst: []string{},
					},
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (IP src, PORT src) Тест 4.3. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect((len(pattern) > 0)).Should(BeTrue())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 5.1. Проверка формирования шаблона скрипта содержащего BPF фильтр с сетями и портами", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{},
						Dst: []string{},
					},
					Port: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{"443"},
						Dst: []string{"3223"},
					},
					Network: FiltrationControlIPorNetorPortParameters{
						Any: []string{"67.3.2.45/25", "194.50.141.29/32"},
						Src: []string{},
						Dst: []string{},
					},
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (Network any, PORT src, dst) Тест 5.1. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect((len(pattern) > 0)).Should(BeTrue())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})

	Context("Тест 5.2. Проверка формирования шаблона скрипта содержащего BPF фильтр с сетями и портами", func() {
		It("При формировании шаблона BPF фильтра не должно быть ошибок", func() {
			pattern := createPatternScript(FiltrationTasks{
				Protocol: "tcp",
				Filters: FiltrationControlParametersNetworkFilters{
					IP: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{},
						Dst: []string{},
					},
					Port: FiltrationControlIPorNetorPortParameters{
						Any: []string{"22", "3876"},
						Src: []string{"443"},
						Dst: []string{"3223"},
					},
					Network: FiltrationControlIPorNetorPortParameters{
						Any: []string{},
						Src: []string{"67.3.2.45/25", "194.50.141.29/32"},
						Dst: []string{"56.3.22.4/27"},
					},
				}}, "pppoe")

			fmt.Printf("\n---=== BPF filter (Network any, PORT src, dst) Тест 5.2. ===---\n'%v'\n", pattern)

			handle, errCreate := ht.getHandler()
			errBPFFilter := handle.SetBPFFilter(pattern)

			Expect(errCreate).ShouldNot(HaveOccurred())
			Expect((len(pattern) > 0)).Should(BeTrue())
			Expect(errBPFFilter).ShouldNot(HaveOccurred())
		})
	})
})

type FiltrationTasks struct {
	DateTimeStart, DateTimeEnd      int64
	Protocol                        string
	Filters                         FiltrationControlParametersNetworkFilters
	Status                          string
	UseIndex                        bool
	NumberFilesMeetFilterParameters int
	NumberFilesFoundResultFiltering int
	NumberProcessedFiles            int
	NumberErrorProcessedFiles       int
	SizeFilesMeetFilterParameters   int64
	SizeFilesFoundResultFiltering   int64
	FileStorageDirectory            string
	ListChanStopFiltration          []chan struct{}
	ListFiles                       map[string][]string
}

//FiltrationControlParametersNetworkFilters параметры сетевых фильтров
type FiltrationControlParametersNetworkFilters struct {
	IP      FiltrationControlIPorNetorPortParameters `json:"ip"`
	Port    FiltrationControlIPorNetorPortParameters `json:"pt"`
	Network FiltrationControlIPorNetorPortParameters `json:"nw"`
}

//FiltrationControlIPorNetorPortParameters параметры для ip или network
type FiltrationControlIPorNetorPortParameters struct {
	Any []string `json:"any"`
	Src []string `json:"src"`
	Dst []string `json:"dst"`
}

//createPatternScript подготовка шаблона для фильтрации
// Логика работы: сетевой протокол '&&' (набор IP '||' набор Network) '&&' набор портов,
// (ANY '||' (SRC '&&' DST))
func createPatternScript(filtrationParameters FiltrationTasks, typeArea string) string {
	var pAnd, patterns string

	/*listTypeArea := map[string]string{
		"ip":         "",
		"pppoe":      "(pppoes && ip) && ",
		"vlan/pppoe": "(vlan && pppoes && ip) && ",
		"pppoe/vlan": "(pppoes && vlan && ip) && ",
	}*/

	listTypeArea := map[string]string{
		"ip":         "(vlan && ip) && ",
		"pppoe":      "(pppoes && ip) && ",
		"vlan/pppoe": "(vlan && pppoes && ip) && ",
	}

	rangeFunc := func(s []string, pattern, typeElem string) string {
		countAny := len(s)
		if countAny == 0 {
			return ""
		}

		num := 0
		for _, v := range s {
			pEnd := " ||"
			if num == countAny-1 {
				pEnd = ")"
			}

			pattern += " " + v + pEnd
			num++
		}

		return pattern
	}

	patternTypeProtocol := func(proto string) string {
		if proto == "any" || proto == "" {
			return ""
		}

		return fmt.Sprintf("%v &&", proto)
	}

	formingPatterns := func(p *FiltrationControlIPorNetorPortParameters, a, s, d, typeElem string, protoIsExist bool) string {
		var i string
		numAny := len(p.Any)
		numSrc := len(p.Src)
		numDst := len(p.Dst)

		if protoIsExist {
			i = ")"
		}

		if (numAny == 0) && (numSrc == 0) && (numDst == 0) {
			return ""
		}

		pAny := rangeFunc(p.Any, a, typeElem)
		pSrc := rangeFunc(p.Src, s, typeElem)
		pDst := rangeFunc(p.Dst, d, typeElem)

		if (numAny > 0) && ((numSrc == 0) && (numDst == 0)) {
			return fmt.Sprintf("%v%v", pAny, i)
		} else if (numAny > 0) && ((numSrc > 0) || (numDst > 0)) {
			if (numSrc == 0) && (numDst > 0) {
				return fmt.Sprintf("(%v%v || %v)%v", pAny, i, pDst, i)
			} else if (numSrc > 0) && (numDst == 0) {
				return fmt.Sprintf("(%v%v || %v)%v", pAny, i, pSrc, i)
			} else {
				return fmt.Sprintf("(%v%v || %v%v && %v)%v", pAny, i, pSrc, i, pDst, i)
			}
		} else if (numAny == 0) && ((numSrc > 0) || (numDst > 0)) {
			if (numSrc == 0) && (numDst > 0) {
				return fmt.Sprintf("%v%v", pDst, i)
			} else if (numSrc > 0) && (numDst == 0) {
				return fmt.Sprintf("%v%v", pSrc, i)
			} else {
				return fmt.Sprintf("(%v%v && %v)%v", pSrc, i, pDst, i)
			}
		}

		return ""
	}

	patternIPAddress := func(ip *FiltrationControlIPorNetorPortParameters, pProto string) string {
		var protoIsExist bool
		a := "(host"
		s := "(src"
		d := "(dst"

		if pProto != "" {
			protoIsExist = true

			a = fmt.Sprintf("(%v (host", pProto)
			s = fmt.Sprintf("(%v (src", pProto)
			d = fmt.Sprintf("(%v (dst", pProto)
		}

		return formingPatterns(ip, a, s, d, "ip", protoIsExist)
	}

	patternPort := func(port *FiltrationControlIPorNetorPortParameters) string {
		return formingPatterns(port, "(port", "(src port", "(dst port", "port", false)
	}

	patternNetwork := func(network *FiltrationControlIPorNetorPortParameters, pProto string) string {
		//приводим значение сетей к валидным сетевым маскам
		forEachFunc := func(list []string) []string {
			newList := make([]string, 0, len(list))

			for _, v := range list {
				t := strings.Split(v, "/")

				mask, _ := strconv.Atoi(t[1])

				ipv4Addr := net.ParseIP(t[0])
				ipv4Mask := net.CIDRMask(mask, 32)

				newList = append(newList, fmt.Sprintf("%v/%v", ipv4Addr.Mask(ipv4Mask).String(), t[1]))
			}

			return newList
		}

		var protoIsExist bool
		a := "(net"
		s := "(src net"
		d := "(dst net"

		if pProto != "" {
			protoIsExist = true

			a = fmt.Sprintf("(%v (net", pProto)
			s = fmt.Sprintf("(%v (src net", pProto)
			d = fmt.Sprintf("(%v (dst net", pProto)
		}

		return formingPatterns(&FiltrationControlIPorNetorPortParameters{
			Any: forEachFunc(network.Any),
			Src: forEachFunc(network.Src),
			Dst: forEachFunc(network.Dst),
		}, a, s, d, "network", protoIsExist)
	}

	//формируем шаблон для фильтрации по протоколам сетевого уровня
	pProto := patternTypeProtocol(filtrationParameters.Protocol)

	//формируем шаблон для фильтрации по ip адресам
	pIP := patternIPAddress(&filtrationParameters.Filters.IP, pProto)

	//формируем шаблон для фильтрации по сетевым портам
	pPort := patternPort(&filtrationParameters.Filters.Port)

	//формируем шаблон для фильтрации по сетям
	pNetwork := patternNetwork(&filtrationParameters.Filters.Network, pProto)

	if len(pPort) > 0 && (len(pIP) > 0 || len(pNetwork) > 0) {
		pAnd = " && "
	}

	if len(pIP) > 0 && len(pNetwork) > 0 {
		patterns = fmt.Sprintf(" (%v || %v)%v%v", pIP, pNetwork, pAnd, pPort)
	} else {
		patterns = fmt.Sprintf("%v%v%v%v", pIP, pNetwork, pAnd, pPort)
	}

	return fmt.Sprintf("%v || (%v%v)", patterns, listTypeArea[typeArea], patterns)
}

/*
	5.1
	(tcp && (net 67.3.2.0/25 || 194.50.141.29/32)) && ((src port 443) && (dst port 3223))

	5.2
	(
		(tcp && (src net 67.3.2.0/25 || 194.50.141.29/32))
		&&
		(tcp && (dst net 56.3.22.0/27))
	)
	&&
	(
		(port 22 || 3876) || (src port 443) && (dst port 3223)
	)
	||
	(
		(pppoes && ip)
		&&
		(
			(tcp && (src net 67.3.2.0/25 || 194.50.141.29/32)) && (tcp && (dst net 56.3.22.0/27))
		)
		&&
		(
			(port 22 || 3876)
			||
			(src port 443) && (dst port 3223)
		)
	)
*/
