package webSocketService

import (
	"container/list"
	"errors"
	"events/eStreamerEvent"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"mediaTypes/mp4"
	"sync"
	"wssAPI"

	"github.com/gorilla/websocket"
)

const (
	wsHandler = "websocketHandler"
)

type websocketHandler struct {
	parent       wssAPI.Obj
	conn         *websocket.Conn
	app          string
	streamName   string
	playName     string
	pubName      string
	clientId     string
	isPlaying    bool
	mutexPlaying sync.RWMutex
	waitPlaying  *sync.WaitGroup
	stPlay       playInfo
	isPublish    bool
	mutexPublish sync.RWMutex
	hasSink      bool
	mutexbSink   sync.RWMutex
	hasSource    bool
	mutexbSource sync.RWMutex
	source       wssAPI.Obj
}

type playInfo struct {
	cache          *list.List
	mutexCache     sync.RWMutex
	audioHeader    *flv.FlvTag
	videoHeader    *flv.FlvTag
	metadata       *flv.FlvTag
	keyFrameWrited bool
	beginTime      uint32
}

func (this *websocketHandler) Init(msg *wssAPI.Msg) (err error) {
	this.conn = msg.Param1.(*websocket.Conn)
	this.waitPlaying = new(sync.WaitGroup)
	return
}

func (this *websocketHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *websocketHandler) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *websocketHandler) GetType() string {
	return wsHandler
}

func (this *websocketHandler) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *websocketHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		switch tag.TagType {
		case flv.FLV_TAG_Audio:
			if this.stPlay.audioHeader == nil {
				this.stPlay.audioHeader = tag
				this.stPlay.audioHeader.Timestamp = 0
				return
			}
		case flv.FLV_TAG_Video:
			if this.stPlay.videoHeader == nil {
				this.stPlay.videoHeader = tag
				this.stPlay.videoHeader.Timestamp = 0
				return
			}
			if false == this.stPlay.keyFrameWrited {
				if (tag.Data[0] >> 4) == 1 {
					this.stPlay.keyFrameWrited = true
					this.stPlay.beginTime = tag.Timestamp
					this.stPlay.addInitPkts()
				} else {
					return
				}
			}

		case flv.FLV_TAG_ScriptData:
			if this.stPlay.metadata == nil {
				this.stPlay.metadata = tag

				return nil
			}
		}
		if false == this.stPlay.keyFrameWrited {
			return
		}
		tag.Timestamp -= this.stPlay.beginTime
		this.stPlay.mutexCache.Lock()
		this.stPlay.cache.PushBack(tag)
		this.stPlay.mutexCache.Unlock()
	case wssAPI.MSG_PLAY_START:
	case wssAPI.MSG_PLAY_STOP:
	case wssAPI.MSG_PUBLISH_START:
	case wssAPI.MSG_PUBLISH_STOP:
	}
	return
}

func (this *websocketHandler) processWSMessage(data []byte) (err error) {
	if nil == data {
		this.Stop(nil)
		return
	}
	msgType := int(data[0])
	switch msgType {
	case WS_pkt_audio:
	case WS_pkt_video:
	case WS_pkt_control:
		logger.LOGT(data)

	default:
		err = errors.New(fmt.Sprintf("msg type %d not supported", msgType))
		logger.LOGW("invalid binary data")
		return
	}
	return
}

func (this *websocketHandler) sendSlice(slice *mp4.FMP4Slice) (err error) {
	dataSend := make([]byte, len(slice.Data)+1)
	dataSend[0] = byte(slice.Type)
	copy(dataSend[1:], slice.Data)
	return this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
}

func (this *websocketHandler) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}

func (this *websocketHandler) addSource(streamName string) (id int, src wssAPI.Obj, err error) {
	taskAddSrc := &eStreamerEvent.EveAddSource{StreamName: streamName}
	err = wssAPI.HandleTask(taskAddSrc)
	if err != nil {
		logger.LOGE("add source " + streamName + " failed")
		return
	}
	return
}

func (this *websocketHandler) delSource(streamName string, id int) (err error) {
	taskDelSrc := &eStreamerEvent.EveDelSource{StreamName: streamName, Id: int64(id)}
	err = wssAPI.HandleTask(taskDelSrc)
	if err != nil {
		logger.LOGE("del source " + streamName + " failed:" + err.Error())
		return
	}
	return
}

func (this *websocketHandler) addSink(streamName, clientId string, sinker wssAPI.Obj) (err error) {
	taskAddsink := &eStreamerEvent.EveAddSink{StreamName: streamName, SinkId: clientId, Sinker: sinker}
	err = wssAPI.HandleTask(taskAddsink)
	if err != nil {
		logger.LOGE(fmt.Sprintf("add sink %s %s failed :\n%s", streamName, clientId, err.Error()))
		return
	}
	return
}

func (this *websocketHandler) delSink(streamName, clientId string) (err error) {
	taskDelSink := &eStreamerEvent.EveDelSink{StreamName: streamName, SinkId: clientId}
	err = wssAPI.HandleTask(taskDelSink)
	if err != nil {
		logger.LOGE(fmt.Sprintf("del sink %s %s failed:\n%s", streamName, clientId, err.Error()))
	}
	return
}

func (this *playInfo) reset() {
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()
	this.cache = list.New()
	this.audioHeader = nil
	this.videoHeader = nil
	this.metadata = nil
	this.keyFrameWrited = false
	this.beginTime = 0
}

func (this *playInfo) addInitPkts() {
	this.mutexCache.Lock()
	defer this.mutexCache.Unlock()
	if this.audioHeader != nil {
		this.cache.PushBack(this.audioHeader)
	}
	if this.videoHeader != nil {
		this.cache.PushBack(this.videoHeader)
	}
	if this.metadata != nil {
		this.cache.PushBack(this.metadata)
	}
}
