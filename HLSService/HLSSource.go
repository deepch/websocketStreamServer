package HLSService

import (
	"errors"
	"events/eStreamerEvent"
	"logger"
	"mediaTypes/flv"
	"mediaTypes/ts"
	"net/http"
	"wssAPI"
	"container/list"
	"sync"
	"strings"
	"strconv"
)

type hlsTsData struct {
	buf []byte
	durationMs int
	idx int
}

const TsCacheLength=100

type HLSSource struct {
	sinkAdded   bool
	inSvrMap    bool
	chValid     bool
	chSvr       chan bool
	streamName  string
	urlPref string
	clientId    string
	audioHeader *flv.FlvTag
	videoHeader *flv.FlvTag
	segIdx      int64
	tsCur       *ts.TsCreater
	tsCache *list.List
	muxCache sync.RWMutex
	beginTime uint32
}

func (this *HLSSource) Init(msg *wssAPI.Msg) (err error) {
	this.sinkAdded = false
	this.inSvrMap = false
	this.chValid = false
	this.tsCache=list.New()
	this.segIdx=1
	this.beginTime=0
	var ok bool
	this.streamName, ok = msg.Param1.(string)
	if false == ok {
		return errors.New("invalid param init hls source")
	}
	this.chSvr, ok = msg.Param2.(chan bool)
	if false == ok {
		return errors.New("invalid param init hls source")
	}
	this.chValid = true

	//create source
	this.clientId = wssAPI.GenerateGUID()
	taskAddSink := &eStreamerEvent.EveAddSink{
		StreamName: this.streamName,
		SinkId:     this.clientId,
		Sinker:     this}
	wssAPI.HandleTask(taskAddSink)

	if strings.Contains(this.streamName,"/"){
		//subs:=strings.Split(this.streamName,"/")
		//this.urlPref=strings.TrimPrefix(this.streamName,subs[0])
		this.urlPref="/"+serviceConfig.StreamRoute+"/"+this.streamName
	}
	return
}

func (this *HLSSource) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *HLSSource) Stop(msg *wssAPI.Msg) (err error) {
	defer func() {
		if err := recover(); err != nil {
			logger.LOGD(err)
		}
	}()
	//从源移除
	if this.sinkAdded {
		taskDelSink := &eStreamerEvent.EveDelSink{}
		taskDelSink.StreamName = this.streamName
		taskDelSink.SinkId = this.clientId
		go wssAPI.HandleTask(taskDelSink)
		this.sinkAdded = false
		logger.LOGT("del sinker:" + this.clientId)
	}
	//从service移除
	if this.inSvrMap {
		this.inSvrMap = false
		service.DelSource(this.streamName, this.clientId)
	}
	//清理数据
	if this.chValid {
		close(this.chSvr)
		this.chValid = false
	}
	return
}

func (this *HLSSource) GetType() string {
	return ""
}

func (this *HLSSource) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *HLSSource) ProcessMessage(msg *wssAPI.Msg) (err error) {
	switch msg.Type {
	case wssAPI.MSG_GetSource_NOTIFY:
		if this.chValid {
			this.chSvr <- true
			this.inSvrMap = true
		}
	case wssAPI.MSG_GetSource_Failed:
		this.Stop(nil)
	case wssAPI.MSG_PLAY_START:
	case wssAPI.MSG_PLAY_STOP:
		//hls 停止就结束移除，不像RTMP等待
		this.Stop(nil)
	case wssAPI.MSG_FLV_TAG:
		tag := msg.Param1.(*flv.FlvTag)
		this.AddFlvTag(tag)
	default:
		logger.LOGT(msg.Type)
	}
	return
}

func (this *HLSSource) ServeHTTP(w http.ResponseWriter, req *http.Request,param string) {
	if strings.HasSuffix(param,".ts"){
		//get ts file
		this.serveTs(w,req,param)
	}else{
		//get m3u8 file
		this.serveM3u8(w,req,param)
	}
}

func (this *HLSSource)serveTs(w http.ResponseWriter,req *http.Request,param string)  {
	subs:=strings.Split(param,"/")
	strIdx:=strings.TrimSuffix(subs[len(subs)-1],".ts")
	idx,_:=strconv.Atoi(strIdx)
	this.muxCache.RLock()
	defer this.muxCache.RUnlock()
	for e:=this.tsCache.Front();e!=nil;e=e.Next(){
		tsData:=e.Value.(*hlsTsData)
		if tsData.idx==idx{
			w.Write(tsData.buf)
			return
		}
	}

}

func (this *HLSSource)serveM3u8(w http.ResponseWriter,req *http.Request,param string)()  {

		this.muxCache.RLock()
		tsCacheCopy:=list.New()
		for e:=this.tsCache.Front();e!=nil;e=e.Next(){
			tsCacheCopy.PushBack(e.Value)
		}
		this.muxCache.RUnlock()
		if tsCacheCopy.Len()>0{
			w.Header().Set("Content-Type","Application/vnd.apple.mpegurl")
			//max duration
			maxDuration:=0
			for e:=tsCacheCopy.Front();e!=nil;e=e.Next(){
				maxDuration=e.Value.(*hlsTsData).durationMs
			}
			//sequence
			sequence:=tsCacheCopy.Front().Value.(*hlsTsData).idx
			strOut:="#EXTM3U"+"\r\n"
			strOut+="#EXT-X-TARGETDURATION:"+strconv.Itoa(1+maxDuration/1000)+"\r\n"
			strOut+="#EXT-X-MEDIA-SEQUENCE:"+strconv.Itoa(sequence)+"\r\n"
			for e:=tsCacheCopy.Front();e!=nil ;e=e.Next() {
				tmp := e.Value.(*hlsTsData)
				strOut += "#EXTINF:" + strconv.Itoa(tmp.durationMs/1000+1) + "\r\n"
				strOut += this.urlPref+"/"+strconv.Itoa(tmp.idx) + ".ts" + "\r\n"
			}
			w.Write([]byte(strOut))
		}else{
			//wait for new
			logger.LOGE("no data now")
		}
}

func (this *HLSSource) AddFlvTag(tag *flv.FlvTag) {
	if this.audioHeader == nil && tag.TagType == flv.FLV_TAG_Audio {
		this.audioHeader = tag.Copy()
		return
	}
	if this.videoHeader == nil && tag.TagType == flv.FLV_TAG_Video {
		this.videoHeader = tag.Copy()
		return
	}

	//如果是关键帧，新建一个切片
	if tag.TagType == flv.FLV_TAG_Video && tag.Data[0] == 0x17 && tag.Data[1] == 1 {
		this.createNewTSSegment(tag)
	} else {
		this.appendTag(tag)
	}
}

func (this *HLSSource) createNewTSSegment(keyframe *flv.FlvTag) {
	//可能有多帧
	if this.tsCur == nil {
		this.tsCur = &ts.TsCreater{}
		if this.audioHeader != nil {
			this.tsCur.AddTag(this.audioHeader)
		}
		if this.videoHeader != nil {
			this.tsCur.AddTag(this.videoHeader)
		}
		this.tsCur.AddTag(keyframe)
	} else {
		//flush data
		if this.tsCur.GetDuration()<5000 {
			this.appendTag(keyframe)
			return
		}
		data := this.tsCur.FlushTsList()
		this.muxCache.Lock()
		defer this.muxCache.Unlock()
		if this.tsCache.Len()>TsCacheLength{
			this.tsCache.Remove(this.tsCache.Front())
		}
		tsdata:=&hlsTsData{}
		tsdata.durationMs=this.tsCur.GetDuration()
		tsdata.buf=make([]byte,ts.TS_length*data.Len())
		ptr:=0
		for e:=data.Front();e!=nil;e=e.Next(){
			copy(tsdata.buf[ptr:],e.Value.([]byte))
			ptr+=ts.TS_length
		}
		tsdata.idx=int(this.segIdx&0xffffffff)
		this.segIdx++
		this.tsCache.PushBack(tsdata)
		this.tsCur=&ts.TsCreater{}

		if this.audioHeader != nil {
			this.tsCur.AddTag(this.audioHeader)
		}
		if this.videoHeader != nil {
			this.tsCur.AddTag(this.videoHeader)
		}
		this.tsCur.AddTag(keyframe)
	}
}

func (this *HLSSource) appendTag(tag *flv.FlvTag) {
	if this.tsCur != nil {
		if this.beginTime==0&&tag.Timestamp>0{
			this.beginTime=tag.Timestamp
		}
		tagIn:=tag.Copy()
		//tagIn.Timestamp-=this.beginTime
		this.tsCur.AddTag(tagIn)
	}
}
