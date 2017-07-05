package HLSService

import (
	"errors"
	"events/eStreamerEvent"
	"logger"
	"mediaTypes/flv"
	"mediaTypes/ts"
	"net/http"
	"os"
	"wssAPI"
)

type HLSSource struct {
	sinkAdded   bool
	inSvrMap    bool
	chValid     bool
	chSvr       chan bool
	streamName  string
	clientId    string
	audioHeader *flv.FlvTag
	videoHeader *flv.FlvTag
	segIdx      int64
	tsCur       *ts.TsCreater
}

func (this *HLSSource) Init(msg *wssAPI.Msg) (err error) {
	this.sinkAdded = false
	this.inSvrMap = false
	this.chValid = false
	var ok bool
	this.streamName, ok = msg.Param1.(string)
	if false == ok {
		return errors.New("invalid param init hls source")
	}
	this.chSvr, ok = msg.Param2.(chan bool)
	if false == ok {
		return errors.New("invalid param init hls source")
	}
	this.chValid = true

	//create source
	this.clientId = wssAPI.GenerateGUID()
	taskAddSink := &eStreamerEvent.EveAddSink{
		StreamName: this.streamName,
		SinkId:     this.clientId,
		Sinker:     this}
	wssAPI.HandleTask(taskAddSink)

	return
}

func (this *HLSSource) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *HLSSource) Stop(msg *wssAPI.Msg) (err error) {
	defer func() {
		if err := recover(); err != nil {
			logger.LOGD(err)
		}
	}()
	//从源移除
	if this.sinkAdded {
		taskDelSink := &eStreamerEvent.EveDelSink{}
		taskDelSink.StreamName = this.streamName
		taskDelSink.SinkId = this.clientId
		go wssAPI.HandleTask(taskDelSink)
		this.sinkAdded = false
		logger.LOGT("del sinker:" + this.clientId)
	}
	//从service移除
	if this.inSvrMap {
		this.inSvrMap = false
		service.DelSource(this.streamName, this.clientId)
	}
	//清理数据
	if this.chValid {
		close(this.chSvr)
		this.chValid = false
	}
	return
}

func (this *HLSSource) GetType() string {
	return ""
}

func (this *HLSSource) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *HLSSource) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_GetSource_NOTIFY:
		if this.chValid {
			this.chSvr <- true
			this.inSvrMap = true
		}
	case wssAPI.MSG_GetSource_Failed:
		this.Stop(nil)
	case wssAPI.MSG_PLAY_START:
	case wssAPI.MSG_PLAY_STOP:
		//hls 停止就结束移除，不像RTMP等待
		this.Stop(nil)
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		this.AddFlvTag(tag)
	default:
		logger.LOGT(msg.Type)
	}
	return
}

func (this *HLSSource) ServeHTTP(w http.ResponseWriter, req *http.Request) {

}

func (this *HLSSource) AddFlvTag(tag *flv.FlvTag) {
	if this.audioHeader == nil && tag.TagType == flv.FLV_TAG_Audio {
		this.audioHeader = tag.Copy()
		return
	}
	if this.videoHeader == nil && tag.TagType == flv.FLV_TAG_Video {
		this.videoHeader = tag.Copy()
		return
	}

	//如果是关键帧，新建一个切片
	if tag.TagType == flv.FLV_TAG_Video && tag.Data[0] == 0x17 && tag.Data[1] == 1 {
		this.createNewTSSegment(tag)
	} else {
		this.appendTag(tag)
	}
}

func (this *HLSSource) createNewTSSegment(keyframe *flv.FlvTag) {
	//可能有多帧
	if this.tsCur == nil {
		this.tsCur = &ts.TsCreater{}
		if this.audioHeader != nil {
			this.tsCur.AddTag(this.audioHeader)
		}
		if this.videoHeader != nil {
			this.tsCur.AddTag(this.videoHeader)
		}
		this.tsCur.AddTag(keyframe)
	} else {
		//flush data
		if this.tsCur.GetDuration()<10000{
			return
		}
		data := this.tsCur.FlushTsList()
		wssAPI.CreateDirectory("audio")
		fp, err := os.Create("audio/init.ts")
		if err != nil {
			logger.LOGE(err.Error())
			return
		}
		for e := data.Front(); e != nil; e = e.Next() {
			fp.Write(e.Value.([]byte))
		}
		fp.Close()
		logger.LOGD(data.Len())
		logger.LOGF("one seg ok")
	}
}

func (this *HLSSource) appendTag(tag *flv.FlvTag) {
	if this.tsCur != nil {
		this.tsCur.AddTag(tag)
	}
}
