package eRTMPEvent

import (
	"wssAPI"
)

const (
	PullRTMPStream = "PullRTMPStream"
)

type EvePullRTMPStream struct {
	Protocol   string //RTMP,RTMPS,RTMPS and so on
	App        string
	Address    string
	Port       int
	StreamName string
	Src        chan wssAPI.Obj
}

func (this *EvePullRTMPStream) Receiver() string {
	return wssAPI.OBJ_RTMPServer
}

func (this *EvePullRTMPStream) Type() string {
	return PullRTMPStream
}

func (this *EvePullRTMPStream) Init(protocol, app, addr, streamName string, port int) {
	this.Protocol = protocol
	this.App = app
	this.Address = addr
	this.Port = port
	this.StreamName = streamName
	this.Src = make(chan wssAPI.Obj)
}

func (this *EvePullRTMPStream) Copy() (out *EvePullRTMPStream) {
	out = &EvePullRTMPStream{}
	out.Protocol = this.Protocol
	out.App = this.App
	out.Address = this.Address
	out.Port = this.Port
	out.StreamName = this.StreamName
	out.Src = this.Src
	return
}
