package streamer

import (
	"container/list"
	"errors"
	"events/eLiveListCtrl"
	"logger"
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
			service.delSource(name)
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
			service.delSource(name)
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
	for k, _ := range service.sources {
		liveList.PushBack(k)
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
	_, exist := this.upApps[app.App]
	if true == exist {
		return errors.New("add up app:" + app.App + " existed")
	}
	this.upApps[app.App] = app.Copy()
	return
}

func (this *StreamerService) delUpstream(app *eLiveListCtrl.EveSetUpStreamApp) (err error) {
	this.mutexUpStream.Lock()
	defer this.mutexUpStream.Unlock()
	_, exist := this.upApps[app.App]
	if false == exist {
		return errors.New("del up app: " + app.App + " not existed")
	}
	delete(this.upApps, app.App)
	return
}

func (this *StreamerService) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}

func (this *StreamerService) badIni() {
	logger.LOGW("some bad init here!!!")
	taskAddUp := eLiveListCtrl.NewSetUpStreamApp(true, "live", "rtmp", "127.0.0.1", 1935)
	this.HandleTask(taskAddUp)
}
