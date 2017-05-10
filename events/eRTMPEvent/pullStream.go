package eRTMPEvent

import (
	"wssAPI"
)

const (
	PullRTMPStream = "PullRTMPStream"
)

type EvePullRTMPStream struct {
	Protocol string //RTMP,RTMPS,RTMPS and so on
	App      string
	Address  string
	Port     int
	Src      chan wssAPI.Obj
}

func (this *EvePullRTMPStream) Receiver() string {
	return wssAPI.OBJ_RTMPServer
}

func (this *EvePullRTMPStream) Type() string {
	return PullRTMPStream
}

func (this *EvePullRTMPStream) Init(protocol, app, addr string, port int) {
	this.Protocol = protocol
	this.App = app
	this.Address = addr
	this.Port = port
	this.Src = make(chan wssAPI.Obj)
}
