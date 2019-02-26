package processrequest

import (
	"encoding/json"
	"fmt"
)

//DetailInformation подробная информация
type DetailInformation struct {
	MaxCountProcessFiltering int `json:"maxCountProcessFiltering"`
}

//MsgTypePing сообщение типа ping
type MsgTypePing struct {
	MsgType string            `json:"messageType"`
	Info    DetailInformation `json:"info"`
}

//SendMsgPing отправить сообщение типа Ping
func SendMsgPing(countProsessFilter int) ([]byte, error) {
	fmt.Println("для клиента, отправить запрос PING")

	msgPing := MsgTypePing{
		MsgType: "ping",
		Info: DetailInformation{
			MaxCountProcessFiltering: countProsessFilter,
		},
	}

	formatJSON, err := json.Marshal(msgPing)
	if err != nil {
		return nil, err
	}

	return formatJSON, nil
}
