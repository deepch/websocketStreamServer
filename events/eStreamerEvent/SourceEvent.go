package eStreamerEvent

import (
	"wssAPI"
)

const (
	AddSource = "AddSource"
	DelSource = "DelSource"
)

type EveAddSource struct {
	StreamName string
	SrcObj     wssAPI.Obj
}

func (this *EveAddSource) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveAddSource) Type() string {
	return AddSource
}

type EveDelSource struct {
	StreamName string
}

func (this *EveDelSource) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveDelSource) Type() string {
	return DelSource
}
