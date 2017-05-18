package webSocketService

import (
	"container/list"
	"errors"
	"events/eStreamerEvent"
	"fmt"
	"logger"
	"mediaTypes/amf"
	"mediaTypes/flv"
	"mediaTypes/mp4"
	"sync"
	"time"
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
	lastCmd      int
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
	this.lastCmd = WSC_close
	return
}

func (this *websocketHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *websocketHandler) Stop(msg *wssAPI.Msg) (err error) {
	this.doClose()
	return
}

func (this *websocketHandler) GetType() string {
	return wsHandler
}

func (this *websocketHandler) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *websocketHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		this.appendFlvTag(tag)
	case wssAPI.MSG_PLAY_START:
		this.startPlay()
	case wssAPI.MSG_PLAY_STOP:
		this.stopPlay()
	case wssAPI.MSG_PUBLISH_START:
	case wssAPI.MSG_PUBLISH_STOP:
	}
	return
}

func (this *websocketHandler) appendFlvTag(tag *flv.FlvTag) {
	tag = tag.Copy()
	if this.stPlay.beginTime == 0 && tag.Timestamp > 0 {
		this.stPlay.beginTime = tag.Timestamp
	}
	tag.Timestamp -= this.stPlay.beginTime
	if false == this.stPlay.keyFrameWrited && tag.TagType == flv.FLV_TAG_Video {
		if this.stPlay.videoHeader == nil {
			this.stPlay.videoHeader = tag
		} else {
			if (tag.Data[0] >> 4) == 1 {
				this.stPlay.keyFrameWrited = true
			} else {
				return
			}
		}

	}
	this.stPlay.mutexCache.Lock()
	defer this.stPlay.mutexCache.Unlock()
	this.stPlay.cache.PushBack(tag)
}

func (this *websocketHandler) processWSMessage(data []byte) (err error) {
	if nil == data || len(data) < 4 {
		this.Stop(nil)
		return
	}
	msgType := int(data[0])
	switch msgType {
	case WS_pkt_audio:
	case WS_pkt_video:
	case WS_pkt_control:
		logger.LOGT(data)
		return this.controlMsg(data[1:])
	default:
		err = errors.New(fmt.Sprintf("msg type %d not supported", msgType))
		logger.LOGW("invalid binary data")
		return
	}
	return
}

func (this *websocketHandler) controlMsg(data []byte) (err error) {
	if nil == data || len(data) < 4 {
		return errors.New("invalid msg")
	}
	ctrlType, err := amf.AMF0DecodeInt24(data)
	if err != nil {
		logger.LOGE("get ctrl type failed")
		return
	}
	logger.LOGT(ctrlType)
	switch ctrlType {
	case WSC_play:
		return this.ctrlPlay(data[3:])
	case WSC_play2:
		return this.ctrlPlay2(data[3:])
	case WSC_resume:
		return this.ctrlResume(data[3:])
	case WSC_pause:
		return this.ctrlPause(data[3:])
	case WSC_seek:
		return this.ctrlSeek(data[3:])
	case WSC_close:
		return this.ctrlClose(data[3:])
	case WSC_dispose:
		return this.ctrlDispose(data[3:])
	case WSC_publish:
		return this.ctrlPublish(data[3:])
	case WSC_onMetaData:
		return this.ctrlOnMetadata(data[3:])
	default:
		logger.LOGE("unknowd websocket control type")
		return errors.New("invalid ctrl msg type")
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
		logger.LOGE(fmt.Sprintf("add sink %s %s failed :%s", streamName, clientId, err.Error()))
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

func (this *websocketHandler) startPlay() {
	this.stPlay.reset()
	go this.threadPlay()
}

func (this *websocketHandler) threadPlay() {
	this.isPlaying = true
	this.waitPlaying.Add(1)
	defer func() {
		this.waitPlaying.Done()
		this.stPlay.reset()
	}()
	fmp4Creater := &mp4.FMP4Creater{}
	for true == this.isPlaying {
		this.stPlay.mutexCache.Lock()
		if this.stPlay.cache == nil || this.stPlay.cache.Len() == 0 {
			this.stPlay.mutexCache.Unlock()
			time.Sleep(30 * time.Millisecond)
			continue
		}
		tag := this.stPlay.cache.Front().Value.(*flv.FlvTag)
		this.stPlay.cache.Remove(this.stPlay.cache.Front())
		this.stPlay.mutexCache.Unlock()
		if WSC_pause == this.lastCmd {
			continue
		}
		if tag.TagType == flv.FLV_TAG_ScriptData {
			err := SendWsControl(this.conn, WSC_onMetaData, tag.Data)
			if err != nil {
				logger.LOGE(err.Error())
				this.isPlaying = false
			}
			continue
		}
		slice := fmp4Creater.AddFlvTag(tag)
		if slice != nil {
			err := this.sendFmp4Slice(slice)
			if err != nil {
				logger.LOGE(err.Error())
				this.isPlaying = false
			}
		}
	}
}

func (this *websocketHandler) sendFmp4Slice(slice *mp4.FMP4Slice) (err error) {
	dataSend := make([]byte, len(slice.Data)+1)
	dataSend[0] = byte(slice.Type)
	copy(dataSend[1:], slice.Data)
	err = this.conn.WriteMessage(websocket.BinaryMessage, dataSend)
	return
}

func (this *websocketHandler) stopPlay() {
	this.isPlaying = false
	this.waitPlaying.Wait()
	this.delSink(this.streamName, this.clientId)
	this.stPlay.reset()
	SendWsStatus(this.conn, WS_status_status, NETSTREAM_PLAY_STOP, 0)
}

func (this *websocketHandler) stopPublish() {
	logger.LOGE("stop publish not code")
}
