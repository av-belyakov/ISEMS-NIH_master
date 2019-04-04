package processrequest

import (
	"encoding/json"
	"fmt"

	"ISEMS-NIH_master/configure"
)

//SendMsgPing отправить сообщение типа Ping
func SendMsgPing(ss *configure.SourceSetting) ([]byte, error) {
	fmt.Println("для клиента, отправить запрос PING")

	fmt.Println(ss.Settings)

	msgPing := configure.MsgTypePingPong{
		MsgType: "ping",
		Info: configure.DetailInfoMsgPingPong{
			MaxCountProcessFiltration: ss.Settings.MaxCountProcessFiltration,
			EnableTelemetry:           ss.Settings.EnableTelemetry,
			StorageFolders:            ss.Settings.StorageFolders,
		},
	}

	formatJSON, err := json.Marshal(msgPing)
	if err != nil {
		return nil, err
	}

	return formatJSON, nil
}
