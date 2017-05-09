package eStreamerEvent

import (
	"wssAPI"
)

const (
	AddSink = "AddSink"
	DelSink = "DelSink"
)

type EveAddSink struct {
	StreamName string
	SinkId     string
	Sinker     wssAPI.Obj
}

func (this *EveAddSink) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveAddSink) Type() string {
	return AddSink
}

type EveDelSink struct {
	StreamName string
	SinkId     string
}

func (this *EveDelSink) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveDelSink) Type() string {
	return DelSink
}
