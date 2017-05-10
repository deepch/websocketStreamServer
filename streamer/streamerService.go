package streamer

import (
	"errors"
	"events/eLiveListCtrl"
	"events/eStreamerEvent"
	"fmt"
	"logger"
	"strings"
	"sync"
	"wssAPI"
)

const (
	streamTypeSource = "streamSource"
	streamTypeSink   = "streamSink"
)

func init() {
	logger.LOGT("streamer init^^^^^^")
}

type StreamerService struct {
	parent         wssAPI.Obj
	mutexSources   sync.RWMutex
	sources        map[string]*streamSource
	mutexBlackList sync.RWMutex
	blacks         map[string]string
	mutexWhiteList sync.RWMutex
	whites         map[string]string
	blackOn        bool
	whiteOn        bool
	mutexUpStream  sync.RWMutex
	upApps         map[string]*eLiveListCtrl.EveSetUpStreamApp
}

var service *StreamerService

func (this *StreamerService) Init(msg *wssAPI.Msg) (err error) {
	this.sources = make(map[string]*streamSource)
	this.blacks = make(map[string]string)
	this.whites = make(map[string]string)
	this.upApps = make(map[string]*eLiveListCtrl.EveSetUpStreamApp)
	service = this
	this.blackOn = false
	this.whiteOn = false
	return
}

func (this *StreamerService) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *StreamerService) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *StreamerService) GetType() string {
	return wssAPI.OBJ_StreamerServer
}

func (this *StreamerService) HandleTask(task wssAPI.Task) (err error) {

	if task == nil || task.Receiver() != this.GetType() {
		logger.LOGE("bad stask")
		return errors.New("invalid task")
	}
	switch task.Type() {
	case eStreamerEvent.AddSource:
		taskAddsrc, ok := task.(*eStreamerEvent.EveAddSource)
		if false == ok {
			return errors.New("invalid param")
		}
		taskAddsrc.SrcObj, err = this.addsource(taskAddsrc.StreamName)
		return
	case eStreamerEvent.DelSource:
		taskDelSrc, ok := task.(*eStreamerEvent.EveDelSource)
		if false == ok {
			return errors.New("invalid param")
		}
		taskDelSrc.StreamName = taskDelSrc.StreamName
		err = this.delSource(taskDelSrc.StreamName)
		return
	case eStreamerEvent.AddSink:
		taskAddSink, ok := task.(*eStreamerEvent.EveAddSink)
		if false == ok {
			return errors.New("invalid param")
		}
		err = this.addSink(taskAddSink.StreamName, taskAddSink.SinkId, taskAddSink.Sinker)
		return
	case eStreamerEvent.DelSink:
		taskDelSink, ok := task.(*eStreamerEvent.EveDelSink)
		if false == ok {
			return errors.New("invalid param")
		}
		err = this.delSink(taskDelSink.StreamName, taskDelSink.SinkId)
		return
	case eLiveListCtrl.EnableBlackList:
		taskEnableBlack, ok := task.(*eLiveListCtrl.EveEnableBlackList)
		if false == ok {
			return errors.New("invalid param")
		}
		err = enableBlackList(taskEnableBlack.Enable)
		return
	case eLiveListCtrl.EnableWhiteList:
		taskEnableWhite, ok := task.(*eLiveListCtrl.EveEnableWhiteList)
		if false == ok {
			return errors.New("invalid param")
		}
		err = enableWhiteList(taskEnableWhite.Enable)
	case eLiveListCtrl.SetBlackList:
		taskSetBlackList, ok := task.(*eLiveListCtrl.EveSetBlackList)
		if false == ok {
			return errors.New("invalid param")
		}
		if taskSetBlackList.Add == true {
			err = addBlackList(taskSetBlackList.Names)
		} else {
			err = delBlackList(taskSetBlackList.Names)
		}
		return
	case eLiveListCtrl.SetWhiteList:
		taskSetWhite, ok := task.(*eLiveListCtrl.EveSetWhiteList)
		if false == ok {
			return errors.New("invalid param")
		}
		if taskSetWhite.Add {
			err = addWhiteList(taskSetWhite.Names)
		} else {
			err = delWhiteList(taskSetWhite.Names)
		}
		return
	case eLiveListCtrl.GetLiveList:
		taskGetLiveList, ok := task.(*eLiveListCtrl.EveGetLiveList)
		if false == ok {
			return errors.New("invalid param")
		}
		taskGetLiveList.Lives, err = getLiveList()
		return
	case eLiveListCtrl.GetLivePlayerCount:
		taskGetLivePlayerCount, ok := task.(*eLiveListCtrl.EveGetLivePlayerCount)
		if false == ok {
			return errors.New("invalid param")
		}
		taskGetLivePlayerCount.Count, err = getPlayerCount(taskGetLivePlayerCount.LiveName)
		return
	case eLiveListCtrl.SetUpStreamApp:
		taskSetUpStream, ok := task.(*eLiveListCtrl.EveSetUpStreamApp)
		if false == ok {
			return errors.New("invalid param set upstream")
		}
		if taskSetUpStream.Add {
			err = this.addUpstream(taskSetUpStream)
		} else {
			err = this.delUpstream(taskSetUpStream)
		}
		return
	default:
		return errors.New("invalid task type:" + task.Type())
	}
	return
}

func (this *StreamerService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *StreamerService) createSrcFromUpstream(app string, chRet chan *streamSource) {
	this.mutexUpStream.RLock()
	addr, exist := this.upApps[app]
	if exist == false {
		logger.LOGE(fmt.Sprintf("%s upstream not found", app))
		this.mutexUpStream.RUnlock()
		return
	}
	this.mutexUpStream.RUnlock()
	switch addr.Protocol {
	case "RTMP":
		//		task := &wssAPI.Task{}
		//		task.Type = wssAPI.TASK_PullRTMPLive
		//		task.Reciver = wssAPI.OBJ_RTMPServer
		//		task.Param1 = addr.Copy()
		//		task.Param2 = chRet
		//		this.parent.HandleTask(task)
	default:
		logger.LOGE(fmt.Sprintf("%s not support now...", addr.Protocol))
		return
	}
	return
}

func (this *StreamerService) checkUpStreamCreated(path string) {
	//usr chan
}

//src control sink
//add source:not start src,start sinks
//del source:not stop src,stop sinks
func (this *StreamerService) addsource(path string) (src wssAPI.Obj, err error) {

	if false == this.checkStreamAddAble(path) {
		return nil, errors.New("bad name")
	}
	this.mutexSources.Lock()
	defer this.mutexSources.Unlock()
	logger.LOGT("add source:" + path)
	oldSrc, exist := this.sources[path]
	if exist == false {
		oldSrc = &streamSource{}
		msg := &wssAPI.Msg{}
		msg.Param1 = path
		oldSrc.Init(msg)
		oldSrc.SetProducer(true)
		this.sources[path] = oldSrc
		src = oldSrc
		return
	} else {
		if oldSrc.HasProducer() {
			err = errors.New("bad name")
			return
		} else {
			logger.LOGT("source:" + path + " is idle")
			oldSrc.SetProducer(true)
			src = oldSrc
			return
		}
	}
	return
}

func (this *StreamerService) delSource(path string) (err error) {
	this.mutexSources.Lock()
	defer this.mutexSources.Unlock()
	logger.LOGT("del source:" + path)
	oldSrc, exist := this.sources[path]
	if exist == false {
		return errors.New(path + " not found")
	} else {
		/*remove := */ oldSrc.SetProducer(false)
		//if remove == true {
		if 0 == len(oldSrc.sinks) {
			delete(this.sources, path)
		}
		//}
		return
	}
	return
}

//add sink:auto start sink by src
//del sink:not stop sink,stop by sink itself
func (this *StreamerService) addSink(path, sinkId string, sinker wssAPI.Obj) (err error) {
	this.mutexSources.Lock()
	defer this.mutexSources.Unlock()
	src, exist := this.sources[path]
	if false == exist {
		app := strings.Split(path, "/")[0]
		this.mutexSources.Unlock()
		chRet := make(chan *streamSource)
		this.createSrcFromUpstream(app, chRet)
		this.mutexSources.Lock()
		src, ok := <-chRet
		if false == ok {
			return
		}
		//!add to map
		if src == nil {
			err = errors.New("source not found in add sink")
		}
		close(chRet)
		return src.AddSink(sinkId, sinker)
	} else {
		return src.AddSink(sinkId, sinker)
	}
	return
}

func (this *StreamerService) delSink(path, sinkId string) (err error) {

	this.mutexSources.Lock()
	defer this.mutexSources.Unlock()
	src, exist := this.sources[path]
	if false == exist {
		return errors.New("source not found in del sink")
	} else {

		src.mutexSink.Lock()
		defer src.mutexSink.Unlock()
		delete(src.sinks, sinkId)
		if 0 == len(src.sinks) && src.bProducer == false {
			delete(this.sources, path)
		}
	}
	return
}
