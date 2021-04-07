package mytestpackages

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"ISEMS-NIH_master/common"
	"ISEMS-NIH_master/configure"
	"ISEMS-NIH_master/savemessageapp"
	//. "ISEMS-NIH_master/mytestpackages"
)

var _ = Describe("QueueTaskStorage", func() {
	saveMessageApp, _ := savemessageapp.New()
	qts := configure.NewRepositoryQTS(saveMessageApp)

	clientID := "mifw77g6f63g"
	var clientTaskID string
	var isSearchTaskID string

	for _, sID := range []int{10, 11, 12} {
		for key, tID := range []string{"taskID_1", "taskID_2"} {
			taskID := common.GetUniqIDFormatMD5(fmt.Sprintf("md5%v%v", tID, string(sID)))
			ctID := common.GetUniqIDFormatMD5(fmt.Sprintf("clientmd5%v%v", tID, string(sID)))

			//			fmt.Printf("SID:%v, TID:%v, key:%v\n", sID, tID, key)

			if (sID == 11) && (key == 1) {
				clientTaskID = ctID
				isSearchTaskID = taskID
			}

			qts.AddQueueTaskStorage(
				taskID,
				sID,
				configure.CommonTaskInfo{
					IDClientAPI:     clientID,
					TaskIDClientAPI: ctID,
					TaskType:        "filtration control",
				},
				&configure.DescriptionParametersReceivedFromUser{
					FilterationParameters: configure.FilteringOption{
						DateTime: configure.TimeInterval{
							Start: 123456789,
							End:   123456799,
						},
					},
				})
		}
	}

	Context("Тест 1: Добавление в очередь новых задач", func() {
		It("Должен вернуться список добавленных задач", func() {

			lt := qts.GetAllSourcesQueueTaskStorage()

			fmt.Println(lt)

			Expect(len(lt)).ShouldNot(Equal(0))
		})
	})

	Context("Тест 2: Поиск информации о задаче по ID задачи присвоенному клиентом API", func() {
		It("Должен быть найден номер источника и ID задачи по ID задачи полученному от клиента", func() {
			fmt.Printf("Client task ID = %v\n", clientTaskID)

			sID, tID, err := qts.SearchTaskForClientIDQueueTaskStorage(clientTaskID)

			fmt.Printf("sID:%v, tID:%v\n", sID, tID)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(tID).Should(Equal(isSearchTaskID))
			Expect(sID).Should(Equal(11))
		})
	})
})
