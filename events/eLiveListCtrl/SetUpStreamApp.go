package eLiveListCtrl

import (
	"wssAPI"
)

type EveSetUpStreamApp struct {
	SinkApp  string `json:"sinkApp"`
	Add      bool
	App      string `json:"app"`
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Addr     string `json:"addr"`
}

func (this *EveSetUpStreamApp) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveSetUpStreamApp) Type() string {
	return SetUpStreamApp
}

func NewSetUpStreamApp(add bool, app, protocol, addr, name string, port int) (out *EveSetUpStreamApp) {
	out = &EveSetUpStreamApp{}
	out.Add = add
	out.App = app
	out.Protocol = protocol
	out.Addr = addr
	out.Port = port
	out.SinkApp = name
	return
}

func (this *EveSetUpStreamApp) Copy() (out *EveSetUpStreamApp) {
	out = &EveSetUpStreamApp{}
	out.SinkApp = this.SinkApp
	out.Add = this.Add
	out.App = this.App
	out.Protocol = this.Protocol
	out.Addr = this.Addr
	out.Port = this.Port
	return
}
