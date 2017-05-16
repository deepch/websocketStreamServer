package webSocketService

import (
	"encoding/json"
	"logger"
)

func (this *websocketHandler) ctrlPlay(data []byte) (err error) {
	st := &stPlay{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlPlay2(data []byte) (err error) {
	st := &stPlay2{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlResume(data []byte) (err error) {
	st := &stResume{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlPause(data []byte) (err error) {
	st := &stPause{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlSeek(data []byte) (err error) {
	st := &stSeek{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlClose(data []byte) (err error) {
	st := &stClose{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlDispose(data []byte) (err error) {
	st := &stDispose{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlPublish(data []byte) (err error) {
	st := &stPublish{}
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	return
}

func (this *websocketHandler) ctrlOnMetadata(data []byte) (err error) {
	logger.LOGT(string(data))
	logger.LOGW("on metadata not processed")
	return
}
