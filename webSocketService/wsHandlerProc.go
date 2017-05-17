package webSocketService

import (
	"encoding/json"
	"logger"
	"wssAPI"
)

func (this *websocketHandler) ctrlPlay(data []byte) (err error) {
	st := &stPlay{}
	defer func() {
		if err != nil {
			logger.LOGE("play failed")
			err = SendWsStatus(this.conn, WS_status_error, NETSTREAM_PLAY_FAILED, st.Req)
		} else {
			this.lastCmd = WSC_play
		}
	}()
	err = json.Unmarshal(data, st)
	if err != nil {
		return err
	}
	if false == supportNewCmd(this.lastCmd, WSC_play) {
		logger.LOGE("bad status")
		return
	}
	logger.LOGT("play")
	this.clientId = wssAPI.GenerateGUID()
	this.streamName = serviceConfig.PlayPath + "/" + st.Name
	err = this.addSink(this.streamName, this.clientId, this)
	if err != nil {
		logger.LOGE("add sink failed: " + err.Error())
		return
	}

	err = SendWsStatus(this.conn, WS_status_status, NETSTREAM_PLAY_START, st.Req)
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
