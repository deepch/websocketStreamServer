package HLSService

import (
	"encoding/json"
	"errors"
	"logger"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"wssAPI"
)

type HLSService struct {
	sources   map[string]*HLSSource
	muxSource sync.RWMutex
	icoData   []byte
}

type HLSConfig struct {
	Port        int    `json:"Port"`
	StreamRoute string `json:"streamRoute"`
	ICO         string `json:"ico"`
}

var service *HLSService
var serviceConfig HLSConfig

func (this *HLSService) Init(msg *wssAPI.Msg) (err error) {
	defer func() {
		if nil != err {
			logger.LOGE(err.Error())
		}
	}()
	if nil == msg || nil == msg.Param1 {
		err = errors.New("invalid param")
		return
	}
	this.sources = make(map[string]*HLSSource)
	fileName := msg.Param1.(string)
	err = this.loadConfigFile(fileName)
	if err != nil {
		return
	}
	service = this
	if len(serviceConfig.ICO) > 0 {
		this.icoData, err = wssAPI.ReadFileAll(serviceConfig.ICO)
		if err != nil {
			return
		}
	}
	return
}

func (this *HLSService) loadConfigFile(fileName string) (err error) {
	buf, err := wssAPI.ReadFileAll(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, &serviceConfig)
	if err != nil {
		return err
	}
	return
}

func (this *HLSService) Start(msg *wssAPI.Msg) (err error) {

	go func() {
		strPort := ":" + strconv.Itoa(serviceConfig.Port)
		mux := http.NewServeMux()
		mux.Handle("/", this)
		err = http.ListenAndServe(strPort, mux)
		if err != nil {
			logger.LOGE("start websocket failed:" + err.Error())
		}
	}()
	return
}

func (this *HLSService) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *HLSService) GetType() string {
	return wssAPI.OBJ_HLSServer
}

func (this *HLSService) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *HLSService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *HLSService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	url := req.URL.Path
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	logger.LOGD(url)
	if strings.HasPrefix(url, serviceConfig.StreamRoute) {
		url = strings.TrimPrefix(url, serviceConfig.StreamRoute)
		if strings.HasPrefix(url, "/") {
			streamName := strings.TrimPrefix(url, "/")
			if strings.HasSuffix(url, ".ts") {
				subs := strings.Split(url, "/")
				streamName = strings.TrimSuffix(streamName, subs[len(subs)-1])
			}
			//
			this.muxSource.RLock()
			source, exist := this.sources[streamName]
			this.muxSource.RUnlock()
			if exist == false {
				source = this.createSource(streamName)
				if wssAPI.InterfaceIsNil(source) {

				} else {

				}
			}
			source.ServeHTTP(w, req)
		} else {
			w.WriteHeader(404)
			return
		}
	} else {
		//ico or invalid
		if "favicon.ico" == url {
			if len(this.icoData) > 0 {
				contentType := http.DetectContentType(this.icoData)
				w.Header().Set("Content-type", contentType)
				w.Write(this.icoData)
			}
		} else {
			w.WriteHeader(404)
		}
	}
}

func (this *HLSService) Add(key string, v *HLSSource) (err error) {
	this.muxSource.Lock()
	defer this.muxSource.Unlock()
	_, ok := this.sources[key]
	if true == ok {
		err = errors.New("source existed")
		return
	}
	this.sources[key] = v
	return
}

func (this *HLSService) DelSource(key, id string) {
	this.muxSource.Lock()
	defer this.muxSource.Unlock()
	src, exist := this.sources[key]
	if exist && src.clientId == id {
		delete(this.sources, key)
	}
}

func (this *HLSService) createSource(streamName string) (source *HLSSource) {
	chSvr := make(chan bool)
	msg := &wssAPI.Msg{
		Param1: streamName,
		Param2: chSvr}
	source = &HLSSource{}
	err := source.Init(msg)
	if err != nil {
		logger.LOGE(err.Error())
		return nil
	}
	select {
	case ret, ok := <-chSvr:
		if false == ok || false == ret {
			source.Stop(nil)
		} else {
			this.muxSource.Lock()
			defer this.muxSource.Unlock()
			old, exist := this.sources[streamName]
			if exist == true {
				//被抢先了
				source.Stop(nil)
				return old
			} else {
				this.sources[streamName] = source
				return source
			}
		}
	case <-time.After(time.Minute):
		source.Stop(nil)
		return nil
	}
	return
}
