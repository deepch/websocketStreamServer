package eLiveListCtrl

import (
	"container/list"
	"wssAPI"
)

type EveGetLiveList struct {
	Lives *list.List
}

func (this *EveGetLiveList) Receiver() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *EveGetLiveList) Type() string {
	return GetLiveList
}
