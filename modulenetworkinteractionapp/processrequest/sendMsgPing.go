package processrequest

import (
	"encoding/json"

	"ISEMS-NIH_master/configure"
)

//SendMsgPing отправить сообщение типа Ping
func SendMsgPing(ss *configure.SourceSetting) ([]byte, error) {
	msgPing := configure.MsgTypePingPong{
		MsgType: "ping",
		Info: configure.DetailInfoMsgPingPong{
			EnableTelemetry: ss.Settings.EnableTelemetry,
			StorageFolders:  ss.Settings.StorageFolders,
		},
	}

	formatJSON, err := json.Marshal(msgPing)
	if err != nil {
		return nil, err
	}

	return formatJSON, nil
}
