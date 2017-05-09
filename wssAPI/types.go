package wssAPI

const (
	OBJ_ServerBus       = "ServerBus"
	OBJ_RTMPServer      = "RTMPServer"
	OBJ_WebSocketServer = "WebsocketServer"
	OBJ_BackendServer   = "BackendServer"
	OBJ_StreamerServer  = "StreamerServer"
)

const (
	MSG_FLV_TAG       = "FLVTag"
	MSG_PUBLISH_START = "NetStream.Publish.Start"
	MSG_PUBLISH_STOP  = "NetStream.Publish.Stop"
	MSG_PLAY_START    = "NetStream.Play.Start"
	MSG_PLAY_STOP     = "NetStream.Play.Stop"
)

const (
	TASK_PullRTMPLive = "PullRTMPPlive" //param UpStreamAddr,result streamSrc
)

type UpStreamAddr struct {
	Protocol string
	App      string
	Address  string
	Port     int
}

func (this *UpStreamAddr) Copy() (out *UpStreamAddr) {
	out = &UpStreamAddr{}
	out.Address = this.Address
	out.App = this.App
	out.Protocol = this.Protocol
	out.Port = this.Port
	return
}
