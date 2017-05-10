package eLiveListCtrl

import (
	"wssAPI"
)

type EveSetUpStreamApp struct {
	Add      bool
	App      string
	Protocol string
	Port     int
	Addr     string
}

func (this *EveSetUpStreamApp) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveSetUpStreamApp) Type() string {
	return SetUpStreamApp
}

func NewSetUpStreamApp(add bool, app, protocol, addr string, port int) (out *EveSetUpStreamApp) {
	out = &EveSetUpStreamApp{}
	out.Add = add
	out.App = app
	out.Protocol = protocol
	out.Addr = addr
	out.Port = port
	return
}

func (this *EveSetUpStreamApp) Copy() (out *EveSetUpStreamApp) {
	out = &EveSetUpStreamApp{}
	out.Add = this.Add
	out.App = this.App
	out.Protocol = this.Protocol
	out.Addr = this.Addr
	out.Port = this.Port
	return
}
