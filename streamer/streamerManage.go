package streamer

import (
	"container/list"
	"errors"
	"events/eLiveListCtrl"
	"logger"
	"math/rand"
	"strings"
	"wssAPI"
)

func enableBlackList(enable bool) (err error) {

	service.mutexBlackList.Lock()
	defer service.mutexBlackList.Unlock()
	service.blackOn = enable
	return
}

func addBlackList(blackList *list.List) (err error) {

	service.mutexBlackList.Lock()
	defer service.mutexBlackList.Unlock()
	errs := ""
	for e := blackList.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if false == ok {
			logger.LOGE("add blackList itor not string")
			errs += " add blackList itor not string \n"
			continue
		}
		service.blacks[name] = name
		if service.blackOn {
			service.delSource(name, 0xffffffff)
		}
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func delBlackList(blackList *list.List) (err error) {
	service.mutexBlackList.Lock()
	defer service.mutexBlackList.Unlock()
	errs := ""
	for e := blackList.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("del blackList itor not string")
			errs += " del blackList itor not string \n"
			continue
		}
		delete(service.blacks, name)
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func enableWhiteList(enable bool) (err error) {

	service.mutexWhiteList.Lock()
	defer service.mutexWhiteList.Unlock()
	service.whiteOn = enable
	return
}

func addWhiteList(whiteList *list.List) (err error) {

	service.mutexWhiteList.Lock()
	defer service.mutexWhiteList.Unlock()
	errs := ""
	for e := whiteList.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("add whiteList itor not string")
			errs += " add blackList itor not string \n"
			continue
		}
		service.whites[name] = name
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func delWhiteList(whiteList *list.List) (err error) {

	service.mutexWhiteList.Lock()
	defer service.mutexWhiteList.Unlock()
	errs := ""
	for e := whiteList.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("del whiteList itor not string")
			errs += " del blackList itor not string \n"
			continue
		}
		delete(service.whites, name)
		if service.whiteOn {
			service.delSource(name, 0xffffffff)
		}
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func getLiveCount() (count int, err error) {
	service.mutexSources.RLock()
	defer service.mutexSources.RUnlock()
	count = len(service.sources)
	return
}

func getLiveList() (liveList *list.List, err error) {
	service.mutexSources.RLock()
	defer service.mutexSources.RUnlock()
	liveList = list.New()
	for k, v := range service.sources {
		info := &eLiveListCtrl.LiveInfo{}
		info.StreamName = k
		v.mutexSink.RLock()
		info.PlayerCount = len(v.sinks)
		v.mutexSink.RUnlock()
		liveList.PushBack(v)
	}
	return
}

func getPlayerCount(name string) (count int, err error) {

	service.mutexSources.RLock()
	defer service.mutexSources.RUnlock()
	src, exist := service.sources[name]
	if exist == false {
		count = 0
	} else {
		count = len(src.sinks)
	}

	return
}

func (this *StreamerService) checkStreamAddAble(appStreamname string) bool {
	tmp := strings.Split(appStreamname, "/")
	var name string
	if len(tmp) > 1 {
		name = tmp[1]
	} else {
		name = appStreamname
	}
	this.mutexBlackList.RLock()
	defer this.mutexBlackList.RUnlock()
	if this.blackOn {
		for k, _ := range this.blacks {
			if name == k {
				return false
			}
		}
	}
	this.mutexWhiteList.RLock()
	defer this.mutexWhiteList.RUnlock()
	if this.whiteOn {
		for k, _ := range this.whites {
			if name == k {
				return true
			}
		}
		return false
	}
	return true
}

func (this *StreamerService) addUpstream(app *eLiveListCtrl.EveSetUpStreamApp) (err error) {
	this.mutexUpStream.Lock()
	defer this.mutexUpStream.Unlock()
	exist := false
	if app.Weight < 1 {
		app.Weight = 1
	}
	for e := this.upApps.Front(); e != nil; e = e.Next() {
		v := e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
		if v.Equal(app) {
			exist = true
			break
		}
	}

	if exist {
		return errors.New("add up app:" + app.SinkApp + " existed")
	} else {
		this.upApps.PushBack(app.Copy())
	}

	return
}

func (this *StreamerService) delUpstream(app *eLiveListCtrl.EveSetUpStreamApp) (err error) {
	this.mutexUpStream.Lock()
	defer this.mutexUpStream.Unlock()
	for e := this.upApps.Front(); e != nil; e = e.Next() {
		v := e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
		if v.Equal(app) {
			this.upApps.Remove(e)
			return
		}
	}
	return errors.New("del up app: " + app.SinkApp + " not existed")
}

func (this *StreamerService) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}

func (this *StreamerService) badIni() {
	logger.LOGW("some bad init here!!!")
	//taskAddUp := eLiveListCtrl.NewSetUpStreamApp(true, "live", "rtmp", "live.hkstv.hk.lxdns.com", 1935)
	//	taskAddUp := eLiveListCtrl.NewSetUpStreamApp(true, "live", "rtmp", "127.0.0.1", 1935)
	//	this.HandleTask(taskAddUp)
}

func (this *StreamerService) InitUpstream(up eLiveListCtrl.EveSetUpStreamApp) {
	logger.LOGD(up)
	up.Add = true
	this.HandleTask(&up)
}

func (this *StreamerService) getUpAddrAuto() (addr *eLiveListCtrl.EveSetUpStreamApp) {
	this.mutexUpStream.RLock()
	defer this.mutexUpStream.RUnlock()
	size := this.upApps.Len()
	if size > 0 {
		totalWeight := 0
		for e := this.upApps.Front(); e != nil; e = e.Next() {
			v := e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
			totalWeight += v.Weight
		}
		if totalWeight == 0 {
			return
		}
		idx := rand.Intn(totalWeight) + 1
		cur := 0
		for e := this.upApps.Front(); e != nil; e = e.Next() {
			v := e.Value.(*eLiveListCtrl.EveSetUpStreamApp)
			cur += v.Weight
			if cur >= idx {
				return v
			}
		}
	}
	return
}
