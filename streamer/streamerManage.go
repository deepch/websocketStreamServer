package streamer

import (
	"container/list"
	"errors"
	"logger"
	"wssAPI"
)

func SetblackList(enable bool) (err error) {

	service.mutexBlackList.Lock()
	defer service.mutexBlackList.Unlock()
	service.blackOn = enable
	return
}

func AddBlackList(blackList *list.List) (err error) {

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
			DelSource(name)
		}
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func DelBlackList(blackList *list.List) (err error) {
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

func SetwhiteList(enable bool) (err error) {

	service.mutexWhiteList.Lock()
	defer service.mutexWhiteList.Unlock()
	service.whiteOn = enable
	return
}

func AddWhiteList(whiteList *list.List) (err error) {

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

func DelWhiteList(whiteList *list.List) (err error) {

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
			DelSource(name)
		}
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func GetLiveCount() (count int, err error) {
	service.mutexSources.RLock()
	defer service.mutexSources.RUnlock()
	count = len(service.sources)
	return
}

func GetLiveList() (liveList *list.List, err error) {
	service.mutexSources.RLock()
	defer service.mutexSources.RUnlock()
	liveList = list.New()
	for k, _ := range service.sources {
		liveList.PushBack(k)
	}
	return
}

func GetPlayerCount(name string) (count int, err error) {

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

func CheckBlack(path string) bool {
	service.mutexBlackList.RLock()
	defer service.mutexBlackList.RUnlock()
	if false == service.blackOn {
		return true
	}
	for k, _ := range service.blacks {
		if k == path {
			return false
		}
	}
	return true
}

func CheckWhite(path string) bool {
	service.mutexWhiteList.RLock()
	defer service.mutexWhiteList.RUnlock()
	if false == service.whiteOn {
		return true
	}
	for k, _ := range service.whites {
		if k == path {
			return true
		}
	}
	return false
}

func AddUpStreamApp(addr *wssAPI.UpStreamAddr) (err error) {

	service.mutexUpStream.Lock()
	defer service.mutexUpStream.Unlock()
	_, exist := service.upApps[addr.App]
	if true == exist {
		return errors.New("app " + addr.App + " existed")
	}
	service.upApps[addr.App] = addr.Copy()
	return
}

func DelUpStreamApp(app string) (err error) {

	service.mutexUpStream.Lock()
	defer service.mutexUpStream.Unlock()
	_, exist := service.upApps[app]
	if exist == false {
		return errors.New(app + " not found")
	}
	delete(service.upApps, app)
	return
}

func (this *StreamerService) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}
