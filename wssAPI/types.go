package wssAPI

const (
	OBJ_ServerBus       = "ServerBus"
	OBJ_RTMPServer      = "RTMPServer"
	OBJ_WebSocketServer = "WebsocketServer"
	OBJ_BackendServer   = "BackendServer"
)

const (
	TASK_StreamerManage = "StreamerManage" //param:op
)

const (
	MSG_FLV_TAG       = "FLVTag"
	MSG_PUBLISH_START = "NetStream.Publish.Start"
	MSG_PUBLISH_STOP  = "NetStream.Publish.Stop"
	MSG_PLAY_START    = "NetStream.Play.Start"
	MSG_PLAY_STOP     = "NetStream.Play.Stop"
)

const (
	Streamer_OP_set_blackList      = iota //param2 bool
	Streamer_OP_addBlackList              //params blackList list
	Streamer_OP_delBlackList              //params blackList list
	Streamer_OP_set_whiteList             //param2 bool
	Streamer_OP_addWhiteList              //params whiteList list
	Streamer_OP_delWhiteList              //params whiteList list
	Streamer_OP_getLiveCount              //return param2
	Streamer_OP_getLiveList               //return params
	Streamer_OP_getLivePlayerCount        //param2 streamName return param2
)
