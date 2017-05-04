package streamer

import (
	"errors"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"sync"
	"wssAPI"
)

type streamSource struct {
	bProducer    bool
	mutexSink    sync.RWMutex
	sinks        map[string]*streamSink
	streamName   string
	metadata     *flv.FlvTag
	audioHeader  *flv.FlvTag
	videoHeader  *flv.FlvTag
	lastKeyFrame *flv.FlvTag
}

func (this *streamSource) Init(msg *wssAPI.Msg) (err error) {
	this.sinks = make(map[string]*streamSink)
	this.streamName = msg.Param1.(string)
	return
}

func (this *streamSource) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *streamSource) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *streamSource) GetType() string {
	return streamTypeSource
}

func (this *streamSource) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *streamSource) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		switch tag.TagType {
		case flv.FLV_TAG_Audio:
			if this.audioHeader == nil {
				this.audioHeader = tag.Copy()
				this.audioHeader.Timestamp = 0
			}
		case flv.FLV_TAG_Video:
			if this.videoHeader == nil {
				this.videoHeader = tag.Copy()
				this.videoHeader.Timestamp = 0
			}
			if (tag.Data[0] >> 4) == 1 {
				this.lastKeyFrame = tag.Copy()
			}

		case flv.FLV_TAG_ScriptData:
			if this.metadata == nil {
				this.metadata = tag.Copy()
			}
		}
		this.mutexSink.RLock()
		defer this.mutexSink.RUnlock()
		for _, v := range this.sinks {
			v.ProcessMessage(msg)
		}
		return
	default:
		logger.LOGW(fmt.Sprintf("msg type %d not processed", msg.Type))
	}
	return
}

func (this *streamSource) HasProducer() bool {
	return this.bProducer
}

func (this *streamSource) SetProducer(status bool) (remove bool) {
	if status == this.bProducer {
		return
	}
	this.bProducer = status
	if this.bProducer == false {
		//clear cache
		this.clearCache()
		//notify sinks stop
		if 0 == len(this.sinks) {
			return true
		}
		this.mutexSink.RLock()
		defer this.mutexSink.RUnlock()
		for _, v := range this.sinks {
			v.Stop(nil)
		}
		return
	} else {
		//notify sinks start
		this.mutexSink.RLock()
		defer this.mutexSink.RUnlock()
		for _, v := range this.sinks {
			v.Start(nil)
		}
		return
	}
}

func (this *streamSource) AddSink(id string, sinker wssAPI.Obj) (err error) {
	this.mutexSink.Lock()
	defer this.mutexSink.Unlock()
	logger.LOGT(this.streamName + " add sink:" + id)
	_, exist := this.sinks[id]
	if true == exist {
		return errors.New("sink " + id + " exist")
	}
	sink := &streamSink{}
	msg := &wssAPI.Msg{}
	msg.Param1 = id
	msg.Param2 = sinker
	err = sink.Init(msg)
	if err != nil {
		logger.LOGE("sink init failed")
		return
	}

	this.sinks[id] = sink
	if this.bProducer {
		err = sink.Start(nil)
		if this.audioHeader != nil {
			msg.Param1 = this.audioHeader
			msg.Type = wssAPI.MSG_FLV_TAG
			sink.ProcessMessage(msg)
		}
		if this.videoHeader != nil {
			msg.Param1 = this.videoHeader
			msg.Type = wssAPI.MSG_FLV_TAG
			sink.ProcessMessage(msg)
		}
		if this.metadata != nil {
			msg.Param1 = this.metadata
			msg.Type = wssAPI.MSG_FLV_TAG
			sink.ProcessMessage(msg)
		}
		if this.lastKeyFrame != nil {
			msg.Param1 = this.lastKeyFrame
			msg.Type = wssAPI.MSG_FLV_TAG
			sink.ProcessMessage(msg)
		}
	}
	return
}

func (this *streamSource) clearCache() {
	logger.LOGT("clear cache")
	this.metadata = nil
	this.audioHeader = nil
	this.videoHeader = nil
	this.lastKeyFrame = nil
}
