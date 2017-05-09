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

const (
	Streamer_OP_set_blackList      = iota //param2 bool
	Streamer_OP_addBlackList              //param2 blackList list
	Streamer_OP_delBlackList              //param2 blackList list
	Streamer_OP_set_whiteList             //param2 bool
	Streamer_OP_addWhiteList              //param2 whiteList list
	Streamer_OP_delWhiteList              //param2 whiteList list
	Streamer_OP_getLiveCount              //return param2
	Streamer_OP_getLiveList               //return params
	Streamer_OP_getLivePlayerCount        //param2 streamName return param2
	Streamer_OP_AddUpStreamAddress        //param2 *UpStreamAddr
	Streamer_OP_DelUpStreamAddress        //param2 appName
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
