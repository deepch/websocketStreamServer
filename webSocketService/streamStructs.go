package webSocketService

import (
	"github.com/gorilla/websocket"
)

const (
	WS_status_ok       = 200
	WS_status_notfound = 404
)

//1byte type,
const (
	WS_pkt_audio   = 8
	WS_pkt_video   = 9
	WS_pkt_control = 18
)

const (
	WSC_play = 1 + iota
	WSC_play2
	WSC_resume
	WSC_pause
	WSC_seek
	WSC_close
	WSC_dispose
	WSC_publish
	WSC_onMetaData = "onMetaData"
)

func SendWsControl(conn *websocket.Conn, ctrlType int, data []byte) (err error) {
	dataSend := make([]byte, len(data)+2)
	dataSend[0] = WS_pkt_control
	dataSend[1] = byte(ctrlType)
	copy(dataSend[2:], data)
	return conn.WriteMessage(websocket.BinaryMessage, dataSend)
}
