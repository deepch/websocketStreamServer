package RTMPService

import (
	"container/list"
	"errors"
	"fmt"
	"logger"
	"mediaTypes/flv"
	"streamer"
	"sync"
	"wssAPI"
)

type RTMPHandler struct {
	parent       wssAPI.Obj
	mutexStatus  sync.RWMutex
	rtmpInstance *RTMP
	source       wssAPI.Obj
	sinke        wssAPI.Obj
	srcAdded     bool
	sinkAdded    bool
	streamName   string
	clientId     string
	playInfo     RTMPPlayInfo
	app          string
	player       rtmpPlayer
	publisher    rtmpPublisher
}
type RTMPPlayInfo struct {
	playReset      bool
	playing        bool //true for thread send playing data
	waitPlaying    *sync.WaitGroup
	mutexCache     sync.RWMutex
	cache          *list.List
	audioHeader    *flv.FlvTag
	videoHeader    *flv.FlvTag
	metadata       *flv.FlvTag
	keyFrameWrited bool
	beginTime      uint32
	startTime      float32
	duration       float32
	reset          bool
}

func (this *RTMPHandler) Init(msg *wssAPI.Msg) (err error) {
	this.rtmpInstance = msg.Param1.(*RTMP)
	msgInit := &wssAPI.Msg{}
	msgInit.Param1 = this.rtmpInstance
	this.player.Init(msgInit)
	this.publisher.Init(msgInit)
	return
}

func (this *RTMPHandler) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPHandler) Stop(msg *wssAPI.Msg) (err error) {
	if this.srcAdded {
		streamer.DelSource(this.streamName)
		logger.LOGT("del source:" + this.streamName)
		this.srcAdded = false
	}
	if this.sinkAdded {
		this.sinkAdded = false
		streamer.DelSink(this.streamName, this.clientId)
		logger.LOGT("del sinker:" + this.clientId)
	}
	this.player.Stop(msg)
	this.publisher.Stop(msg)
	return
}

func (this *RTMPHandler) GetType() string {
	return rtmpTypeHandler
}

func (this *RTMPHandler) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *RTMPHandler) ProcessMessage(msg *wssAPI.Msg) (err error) {

	if msg == nil {
		return errors.New("nil message")
	}
	switch msg.Type {
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		this.player.appendFlvTag(tag)

	case wssAPI.MSG_PLAY_START:
		this.player.startPlay()
		return
	case wssAPI.MSG_PLAY_STOP:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		this.sourceInvalid()
		return
	case wssAPI.MSG_PUBLISH_START:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		if err != nil {
			logger.LOGE("start publish failed")
			return
		}
		if false == this.publisher.startPublish() {
			logger.LOGE("start publish falied")
			if true == this.srcAdded {
				streamer.DelSource(this.streamName)
			}
		}
		return
	case wssAPI.MSG_PUBLISH_STOP:
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		if err != nil {
			logger.LOGE("stop publish failed")
			return
		}
		this.publisher.stopPublish()
		return
	default:
		logger.LOGW(fmt.Sprintf("msg type: %s not processed", msg.Type))
		return
	}
	return
}

func (this *RTMPHandler) sourceInvalid() {
	logger.LOGT("stop play,keep sink")
	this.player.stopPlay()
}

func (this *RTMPHandler) HandleRTMPPacket(packet *RTMPPacket) (err error) {
	if nil == packet {
		//this.updateStatus(rtmp_status_idle)
		this.Stop(nil)
		return
	}
	switch packet.MessageTypeId {
	case RTMP_PACKET_TYPE_CHUNK_SIZE:
		this.rtmpInstance.RecvChunkSize, err = AMF0DecodeInt32(packet.Body)
		logger.LOGT(fmt.Sprintf("chunk size:%d", this.rtmpInstance.RecvChunkSize))
	case RTMP_PACKET_TYPE_CONTROL:
		err = this.rtmpInstance.HandleControl(packet)
	case RTMP_PACKET_TYPE_BYTES_READ_REPORT:
	case RTMP_PACKET_TYPE_SERVER_BW:
		this.rtmpInstance.TargetBW, err = AMF0DecodeInt32(packet.Body)
	case RTMP_PACKET_TYPE_CLIENT_BW:
		this.rtmpInstance.SelfBW, err = AMF0DecodeInt32(packet.Body)
		this.rtmpInstance.LimitType = uint32(packet.Body[4])
	case RTMP_PACKET_TYPE_FLEX_MESSAGE:
		err = this.handleInvoke(packet)
	case RTMP_PACKET_TYPE_INVOKE:
		err = this.handleInvoke(packet)
	case RTMP_PACKET_TYPE_AUDIO:
		if this.publisher.isPublishing() && this.source != nil {
			msg := &wssAPI.Msg{}
			msg.Type = wssAPI.MSG_FLV_TAG
			msg.Param1 = packet.ToFLVTag()
			this.source.ProcessMessage(msg)
		} else {
			logger.LOGE("bad status")
			logger.LOGE(this.source)
		}
	case RTMP_PACKET_TYPE_VIDEO:
		if this.publisher.isPublishing() && this.source != nil {
			msg := &wssAPI.Msg{}
			msg.Type = wssAPI.MSG_FLV_TAG
			msg.Param1 = packet.ToFLVTag()
			this.source.ProcessMessage(msg)
		} else {
			logger.LOGE("bad status")
		}
	case RTMP_PACKET_TYPE_INFO:
		if this.publisher.isPublishing() && this.source != nil {
			msg := &wssAPI.Msg{}
			msg.Type = wssAPI.MSG_FLV_TAG
			//logger.LOGI(packet.ChunkStreamID)
			msg.Param1 = packet.ToFLVTag()
			this.source.ProcessMessage(msg)
		} else {
			logger.LOGE("bad status")
		}
	default:
		logger.LOGW(fmt.Sprintf("rtmp packet type %d not processed", packet.MessageTypeId))
	}
	return
}

func (this *RTMPHandler) handleInvoke(packet *RTMPPacket) (err error) {
	var amfobj *AMF0Object
	if RTMP_PACKET_TYPE_FLEX_MESSAGE == packet.MessageTypeId {
		amfobj, err = AMF0DecodeObj(packet.Body[1:])
	} else {
		amfobj, err = AMF0DecodeObj(packet.Body)
	}
	if err != nil {
		logger.LOGE("recved invalid amf0 object")
		return
	}
	if amfobj.Props.Len() == 0 {
		logger.LOGT(packet.Body)
		logger.LOGT(string(packet.Body))
		return
	}

	method := amfobj.Props.Front().Value.(*AMF0Property)

	switch method.Value.StrValue {
	case "connect":
		cmdObj := amfobj.AMF0GetPropByIndex(2)
		if cmdObj != nil {
			this.app = cmdObj.Value.ObjValue.AMF0GetPropByName("app").Value.StrValue
		}
		if this.app != serviceConfig.LivePath {
			logger.LOGW("path wrong")
		}
		err = this.rtmpInstance.AcknowledgementBW()
		if err != nil {
			return
		}
		err = this.rtmpInstance.SetPeerBW()
		if err != nil {
			return
		}
		//err = this.rtmpInstance.SetChunkSize(RTMP_better_chunk_size)
		//		if err != nil {
		//			return
		//		}
		err = this.rtmpInstance.OnBWDone()
		if err != nil {
			return
		}
		err = this.rtmpInstance.ConnectResult(amfobj)
		if err != nil {
			return
		}
	case "_checkbw":
		err = this.rtmpInstance.OnBWCheck()
	case "_result":
		this.handle_result(amfobj)
	case "releaseStream":
		//		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		//		err = this.rtmpInstance.CmdError("error", "NetConnection.Call.Failed",
		//			fmt.Sprintf("Method not found (%s).", "releaseStream"), idx)
	case "FCPublish":
		//		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		//		err = this.rtmpInstance.CmdError("error", "NetConnection.Call.Failed",
		//			fmt.Sprintf("Method not found (%s).", "FCPublish"), idx)
	case "createStream":
		idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
		err = this.rtmpInstance.CmdNumberResult(idx, 1.0)
	case "publish":
		//check prop
		if amfobj.Props.Len() < 4 {
			logger.LOGE("invalid props length")
			err = errors.New("invalid amf obj for publish")
			return
		}

		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
		//check status
		if true == this.publisher.isPublishing() {
			logger.LOGE("publish on bad status ")
			idx := amfobj.AMF0GetPropByIndex(1).Value.NumValue
			err = this.rtmpInstance.CmdError("error", "NetStream.Publish.Denied",
				fmt.Sprintf("can not publish (%s).", "publish"), idx)
			return
		}
		//add to source
		this.streamName = this.app + "/" + amfobj.AMF0GetPropByIndex(3).Value.StrValue
		this.source, err = streamer.Addsource(this.streamName)
		if err != nil {
			logger.LOGE("add source failed:" + err.Error())
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Publish.BadName",
				fmt.Sprintf("publish %s.", this.streamName), "", 0, RTMP_channel_Invoke)
			this.streamName = ""
			return errors.New("bad name")
		}
		this.srcAdded = true
		this.rtmpInstance.Link.Path = amfobj.AMF0GetPropByIndex(2).Value.StrValue
		if false == this.publisher.startPublish() {
			logger.LOGE("start publish failed:" + this.streamName)
			streamer.DelSource(this.streamName)
			return
		}
	case "FCUnpublish":
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
	case "deleteStream":
		this.mutexStatus.Lock()
		defer this.mutexStatus.Unlock()
	//do nothing now
	case "play":
		this.streamName = this.app + "/" + amfobj.AMF0GetPropByIndex(3).Value.StrValue
		this.rtmpInstance.Link.Path = this.streamName
		startTime := -2
		duration := -1
		reset := false
		this.playInfo.startTime = -2
		this.playInfo.duration = -1
		this.playInfo.reset = false
		if amfobj.Props.Len() >= 5 {
			this.playInfo.startTime = float32(amfobj.AMF0GetPropByIndex(4).Value.NumValue)
		}
		if amfobj.Props.Len() >= 6 {
			this.playInfo.duration = float32(amfobj.AMF0GetPropByIndex(5).Value.NumValue)
			if this.playInfo.duration < 0 {
				this.playInfo.duration = -1
			}
		}
		if amfobj.Props.Len() >= 7 {
			this.playInfo.reset = amfobj.AMF0GetPropByIndex(6).Value.BoolValue
		}

		//check player status,if playing,error
		if false == this.player.setPlayParams(this.streamName, startTime, duration, reset) {
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Play.Failed",
				"paly failed", this.streamName, 0, RTMP_channel_Invoke)

			return nil
		}
		err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}

		if true == this.playInfo.playReset {
			err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Reset",
				fmt.Sprintf("Playing and resetting %s", this.rtmpInstance.Link.Path),
				this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
			if err != nil {
				logger.LOGE(err.Error())
				return
			}
		}

		err = this.rtmpInstance.CmdStatus("status", "NetStream.Play.Start",
			fmt.Sprintf("Started playing %s", this.rtmpInstance.Link.Path), this.rtmpInstance.Link.Path, 0, RTMP_channel_Invoke)
		if err != nil {
			logger.LOGE(err.Error())
			return
		}

		this.clientId = wssAPI.GenerateGUID()
		err = streamer.AddSink(this.streamName, this.clientId, this)
		if err != nil {
			//404
			err = this.rtmpInstance.CmdStatus("error", "NetStream.Play.StreamNotFound",
				"paly failed", this.streamName, 0, RTMP_channel_Invoke)
			return nil
		}
		this.sinkAdded = true
	case "_error":
		amfobj.Dump()
	case "closeStream":
		amfobj.Dump()
	default:
		logger.LOGW(fmt.Sprintf("rtmp method <%s> not processed", method.Value.StrValue))
	}
	return
}

func (this *RTMPHandler) handle_result(amfobj *AMF0Object) {
	transactionId := int32(amfobj.AMF0GetPropByIndex(1).Value.NumValue)
	resultMethod := this.rtmpInstance.methodCache[transactionId]
	switch resultMethod {
	case "_onbwcheck":
	default:
		logger.LOGW("result of " + resultMethod + " not processed")
	}
}

func (this *RTMPHandler) startPublishing() (err error) {
	err = this.rtmpInstance.SendCtrl(RTMP_CTRL_streamBegin, 1, 0)
	if err != nil {
		logger.LOGE(err.Error())
		return nil
	}
	err = this.rtmpInstance.CmdStatus("status", "NetStream.Publish.Start",
		fmt.Sprintf("publish %s", this.rtmpInstance.Link.Path), "", 0, RTMP_channel_Invoke)
	if err != nil {
		logger.LOGE(err.Error())
		return nil
	}
	this.publisher.startPublish()
	return
}

func (this *RTMPHandler) isPlaying() bool {
	return this.player.IsPlaying()
}

func (this *RTMPHandler) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}
