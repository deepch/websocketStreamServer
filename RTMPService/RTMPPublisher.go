package RTMPService

import (
	"wssAPI"
)

type rtmpPublisher struct {
}

func (this *rtmpPublisher) Init(msg *wssAPI.Msg) (err error) {
	return
}

func (this *rtmpPublisher) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *rtmpPublisher) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *rtmpPublisher) GetType() string {
	return ""
}

func (this *rtmpPublisher) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *rtmpPublisher) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}
