package processrequest

import (
	"encoding/json"

	"ISEMS-NIH_master/configure"
)

//SendMsgPing отправить сообщение типа Ping
func SendMsgPing(ss *configure.SourceSetting) ([]byte, error) {
	msgPing := configure.MsgTypePing{
		MsgType: "ping",
		Info: configure.DetailInfoMsgPing{
			EnableTelemetry: ss.Settings.EnableTelemetry,
			StorageFolders:  ss.Settings.StorageFolders,
			TypeAreaNetwork: ss.Settings.TypeAreaNetwork,
		},
	}

	formatJSON, err := json.Marshal(msgPing)
	if err != nil {
		return nil, err
	}

	return formatJSON, nil
}
