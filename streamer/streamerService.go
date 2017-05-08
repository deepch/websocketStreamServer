package streamer

import (
	"errors"
	"logger"
	"sync"
	"wssAPI"
)

const (
	streamTypeSource = "streamSource"
	streamTypeSink   = "streamSink"
)

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
	upApps         map[string]*wssAPI.UpStreamAddr
}

var service *StreamerService

func (this *StreamerService) Init(msg *wssAPI.Msg) (err error) {
	this.sources = make(map[string]*streamSource)
	this.blacks = make(map[string]string)
	this.whites = make(map[string]string)
	this.upApps = make(map[string]*wssAPI.UpStreamAddr)
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

func (this *StreamerService) HandleTask(task *wssAPI.Task) (err error) {
	switch task.Type {
	case wssAPI.TASK_StreamerManage:
		return this.manageStreams(task)
	case wssAPI.TASK_StreamerUSC:
		return this.streamerUSC(task)
	default:
		logger.LOGW(task.Type + " should not handle here")
	}
	task = &wssAPI.Task{}
	task.Reciver = wssAPI.OBJ_StreamerServer
	task.Param1 = wssAPI.Streamer_OP_addBlackList
	//task.Params //一个list

	return
}

func (this *StreamerService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *StreamerService) manageStreams(task *wssAPI.Task) (err error) {
	op, ok := task.Param1.(int)
	if false == ok {
		return errors.New("invalid task params")
	}
	switch op {
	case wssAPI.Streamer_OP_set_blackList:
		return this.setblackList(task)
	case wssAPI.Streamer_OP_addBlackList:
		return this.addBlackList(task)
	case wssAPI.Streamer_OP_delBlackList:
		return this.delBlackList(task)
	case wssAPI.Streamer_OP_set_whiteList:
		return this.setwhiteList(task)
	case wssAPI.Streamer_OP_addWhiteList:
		return this.addWhiteList(task)
	case wssAPI.Streamer_OP_delWhiteList:
		return this.delWhiteList(task)
	case wssAPI.Streamer_OP_getLiveCount:
		return this.getLiveCount(task)
	case wssAPI.Streamer_OP_getLiveList:
		return this.getLiveList(task)
	case wssAPI.Streamer_OP_getLivePlayerCount:
		return this.getPlayerCount(task)
	default:
		return errors.New("unknown op")
	}
	return
}

func (this *StreamerService) streamerUSC(task *wssAPI.Task) (err error) {
	op, ok := task.Param1.(int)
	if false == ok {
		return errors.New("invalid param streamUSC")
	}
	switch op {
	case wssAPI.Streamer_OP_AddUpStreamAddress:
		addr, ok := task.Param2.(*wssAPI.UpStreamAddr)
		if false == ok {
			return errors.New("bad param for add USC")
		}
		this.mutexUpStream.Lock()
		defer this.mutexUpStream.Unlock()
		_, exist := this.upApps[addr.App]
		if true == exist {
			return errors.New("app " + addr.App + " existed")
		}
		this.upApps[addr.App] = addr.Copy()
		return
	case wssAPI.Streamer_OP_DelUpStreamAddress:

		app, ok := task.Param2.(string)
		if false == ok {
			return errors.New("bad app for del USC")
		}
		this.mutexUpStream.Lock()
		defer this.mutexUpStream.Unlock()
		delete(this.upApps, app)
		return
	default:
		return errors.New("unknown op")
	}
	return
}

//src control sink
//add source:not start src,start sinks
//del source:not stop src,stop sinks
func Addsource(path string) (src wssAPI.Obj, err error) {
	if service == nil {
		logger.LOGE("streamer service null")
		err = errors.New("streamer invalid")
		return
	}
	if false == service.checkBlack(path) || false == service.checkWhite(path) {
		return nil, errors.New("bad name")
	}
	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
	logger.LOGT("add source:" + path)
	oldSrc, exist := service.sources[path]
	if exist == false {
		oldSrc = &streamSource{}
		msg := &wssAPI.Msg{}
		msg.Param1 = path
		oldSrc.Init(msg)
		oldSrc.SetProducer(true)
		service.sources[path] = oldSrc
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

func DelSource(path string) (err error) {
	if service == nil {
		return errors.New("streamer invalid")
	}
	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
	logger.LOGT("del source:" + path)
	oldSrc, exist := service.sources[path]
	if exist == false {
		return errors.New(path + " not found")
	} else {
		/*remove := */ oldSrc.SetProducer(false)
		//if remove == true {
		if 0 == len(oldSrc.sinks) {
			delete(service.sources, path)
		}
		//}
		return
	}
	return
}

//add sink:auto start sink by src
//del sink:not stop sink,stop by sink itself
func AddSink(path, sinkId string, sinker wssAPI.Obj) (err error) {
	if service == nil {
		return errors.New("streamer invalid")
	}
	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
	src, exist := service.sources[path]
	if false == exist {
		err = errors.New("source not found in add sink")
		return
	} else {
		return src.AddSink(sinkId, sinker)
	}
	return
}

func DelSink(path, sinkId string) (err error) {
	if service == nil {
		return errors.New("streamer invalid")
	}
	service.mutexSources.Lock()
	defer service.mutexSources.Unlock()
	src, exist := service.sources[path]
	if false == exist {
		return errors.New("source not found in del sink")
	} else {

		src.mutexSink.Lock()
		defer src.mutexSink.Unlock()
		delete(src.sinks, sinkId)
		if 0 == len(src.sinks) && src.bProducer == false {
			delete(service.sources, path)
		}
	}
	return
}
