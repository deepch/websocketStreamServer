package streamer

import (
	"container/list"
	"errors"
	"logger"
	"wssAPI"
)

func (this *StreamerService) setblackList(task *wssAPI.Task) (err error) {
	enable, ok := task.Param2.(bool)
	if false == ok {
		return errors.New("set blackList invalid param")
	}
	this.mutexBlackList.Lock()
	defer this.mutexBlackList.Unlock()
	this.blackOn = enable
	return
}

func (this *StreamerService) addBlackList(task *wssAPI.Task) (err error) {
	if task.Params == nil {
		return errors.New("add blackList invalid params")
	}
	this.mutexBlackList.Lock()
	defer this.mutexBlackList.Unlock()
	errs := ""
	for e := task.Params.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if false == ok {
			logger.LOGE("add blackList itor not string")
			errs += " add blackList itor not string \n"
			continue
		}
		this.blacks[name] = name
		if this.blackOn {
			DelSource(name)
		}
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func (this *StreamerService) delBlackList(task *wssAPI.Task) (err error) {
	if task.Params == nil {
		return errors.New("del blackList invalid params")
	}
	this.mutexBlackList.Lock()
	defer this.mutexBlackList.Unlock()
	errs := ""
	for e := task.Params.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("del blackList itor not string")
			errs += " del blackList itor not string \n"
			continue
		}
		delete(this.blacks, name)
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func (this *StreamerService) setwhiteList(task *wssAPI.Task) (err error) {
	enable, ok := task.Param2.(bool)
	if false == ok {
		return errors.New("set whiteList invalid param ")
	}
	this.mutexWhiteList.Lock()
	defer this.mutexWhiteList.Unlock()
	this.whiteOn = enable
	return
}

func (this *StreamerService) addWhiteList(task *wssAPI.Task) (err error) {
	if task.Params == nil {
		return errors.New("add whiteList invalid params")
	}
	this.mutexWhiteList.Lock()
	defer this.mutexWhiteList.Unlock()
	errs := ""
	for e := task.Params.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("add whiteList itor not string")
			errs += " add blackList itor not string \n"
			continue
		}
		this.whites[name] = name
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func (this *StreamerService) delWhiteList(task *wssAPI.Task) (err error) {
	if task.Params == nil {
		return errors.New("del whiteList invalid params")
	}
	this.mutexWhiteList.Lock()
	defer this.mutexWhiteList.Unlock()
	errs := ""
	for e := task.Params.Front(); e != nil; e = e.Next() {
		name, ok := e.Value.(string)
		if ok == false {
			logger.LOGE("del whiteList itor not string")
			errs += " del blackList itor not string \n"
			continue
		}
		delete(this.whites, name)
		if this.whiteOn {
			DelSource(name)
		}
	}
	if len(errs) > 0 {
		err = errors.New(errs)
	}
	return
}

func (this *StreamerService) getLiveCount(task *wssAPI.Task) (err error) {
	this.mutexSources.RLock()
	defer this.mutexSources.RUnlock()
	task.Param2 = len(this.sources)
	return
}

func (this *StreamerService) getLiveList(task *wssAPI.Task) (err error) {
	this.mutexSources.RLock()
	defer this.mutexSources.RUnlock()
	task.Params = list.New()
	for k, _ := range this.sources {
		task.Params.PushBack(k)
	}
	return
}

func (this *StreamerService) getPlayerCount(task *wssAPI.Task) (err error) {
	name, ok := task.Param2.(string)
	if false == ok {
		return errors.New("getLiveplayerCount invalid param")
	}
	this.mutexSources.RLock()
	defer this.mutexSources.RUnlock()
	src, exist := this.sources[name]
	if exist == false {
		task.Param2 = 0
	} else {
		task.Param2 = len(src.sinks)
	}
	return
}

func (this *StreamerService) checkBlack(path string) bool {
	this.mutexBlackList.RLock()
	defer this.mutexBlackList.RUnlock()
	if false == this.blackOn {
		return true
	}
	for k, _ := range this.blacks {
		if k == path {
			return false
		}
	}
	return true
}

func (this *StreamerService) checkWhite(path string) bool {
	this.mutexWhiteList.RLock()
	defer this.mutexWhiteList.RUnlock()
	if false == this.whiteOn {
		return true
	}
	for k, _ := range this.whites {
		if k == path {
			return true
		}
	}
	return false
}
