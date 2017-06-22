package eLiveListCtrl

import (
	"container/list"
	"wssAPI"
)

type LiveInfo struct {
	StreamName  string
	PlayerCount int
}

type EveGetLiveList struct {
	Lives *list.List //value =*LiveInfo
}

func (this *EveGetLiveList) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveGetLiveList) Type() string {
	return GetLiveList
}
