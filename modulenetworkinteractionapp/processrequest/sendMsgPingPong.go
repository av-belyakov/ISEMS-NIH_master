package processrequest

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
)

//SendMsgPingPong отправить сообщение типа Ping
func SendMsgPingPong(msgType string, countProsessFilter int8) ([]byte, error) {
	fmt.Println("для клиента, отправить запрос PING")

	msgPing := configure.MsgTypePingPong{
		MsgType: msgType,
		Info: configure.DetailInfoMsgPingPong{
			MaxCountProcessfiltration: countProsessFilter,
		},
	}

	formatJSON, err := json.Marshal(msgPing)
	if err != nil {
		return nil, err
	}

	return formatJSON, nil
}
