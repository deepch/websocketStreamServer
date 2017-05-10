package RTMPService

import (
	"errors"
	"events/eRTMPEvent"
	"events/eStreamerEvent"
	"logger"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
	"wssAPI"
)

type RTMPPuller struct {
	rtmp       RTMP
	parent     wssAPI.Obj
	src        wssAPI.Obj
	pullParams *eRTMPEvent.EvePullRTMPStream
	waitRead   *sync.WaitGroup
	reading    bool
}

func PullRTMPLive(task *eRTMPEvent.EvePullRTMPStream) {
	puller := &RTMPPuller{}
	msg := &wssAPI.Msg{}
	msg.Param1 = task
	puller.Init(msg)
	puller.Start(nil)
}

func (this *RTMPPuller) Init(msg *wssAPI.Msg) (err error) {
	this.pullParams = msg.Param1.(*eRTMPEvent.EvePullRTMPStream).Copy()
	this.initRTMPLink()
	this.waitRead = new(sync.WaitGroup)
	return
}

func (this *RTMPPuller) initRTMPLink() {
	this.rtmp.Link.Protocol = this.pullParams.Protocol
	this.rtmp.Link.App = this.pullParams.App
	this.rtmp.Link.Path = this.pullParams.StreamName
	this.rtmp.Link.TcUrl = this.pullParams.Protocol + "://" +
		this.pullParams.Address + ":" +
		strconv.Itoa(this.pullParams.Port) + "/" +
		this.pullParams.App + "/" + this.pullParams.StreamName
	logger.LOGT(this.rtmp.Link.TcUrl)
}

func (this *RTMPPuller) Start(msg *wssAPI.Msg) (err error) {
	defer func() {
		if err != nil {
			close(this.pullParams.Src)
			if nil != this.rtmp.Conn {
				this.rtmp.Conn.Close()
				this.rtmp.Conn = nil
			}
		}
	}()
	//start pull
	//connect
	addr := this.pullParams.Address + ":" + strconv.Itoa(this.pullParams.Port)
	this.rtmp.Conn, err = net.Dial("tcp", addr)
	if err != nil {
		logger.LOGE("connect failed:" + err.Error())
		return
	}
	//just simple handshake
	err = this.handleShake()
	if err != nil {
		logger.LOGE("handle shake failed")
		return
	}
	//start read thread
	go this.threadRead()
	//play
	return
}

func (this *RTMPPuller) Stop(msg *wssAPI.Msg) (err error) {
	//stop pull
	this.reading = false
	this.waitRead.Wait()

	if nil != this.rtmp.Conn {
		this.rtmp.Conn.Close()
		this.rtmp.Conn = nil
	}
	//del src
	if nil != this.src {
		taskDelSrc := &eStreamerEvent.EveDelSource{}
		taskDelSrc.StreamName = this.pullParams.App + "/" + this.pullParams.StreamName
		err = wssAPI.HandleTask(taskDelSrc)
		if err != nil {
			logger.LOGE(err.Error())
		}
	}

	return
}

func (this *RTMPPuller) handleShake() (err error) {
	randomSize := 1528
	//send c0
	conn := this.rtmp.Conn
	c0 := make([]byte, 1)
	c0[0] = 3
	_, err = wssAPI.TcpWrite(conn, c0)
	if err != nil {
		logger.LOGE("send c0 failed")
		return
	}
	//send c1
	c1 := make([]byte, randomSize+4+4)
	for idx := 8; idx < len(c1); idx++ {
		c1[idx] = byte(rand.Intn(255))
	}
	_, err = wssAPI.TcpWrite(conn, c1)
	if err != nil {
		logger.LOGE("send c1 failed")
		return
	}
	//read s0
	s0, err := wssAPI.TcpRead(conn, 1)
	if err != nil {
		logger.LOGE("read s0 failed")
		return
	}
	logger.LOGT(s0)
	//read s1
	s1, err := wssAPI.TcpRead(conn, randomSize+8)
	if err != nil {
		logger.LOGE("read s1 failed")
		return
	}
	//send c2
	_, err = wssAPI.TcpWrite(conn, s1)
	if err != nil {
		logger.LOGE("send c2 failed")
		return
	}
	//read s2
	s2, err := wssAPI.TcpRead(conn, randomSize+8)
	if err != nil {
		logger.LOGE("read s2 failed")
		return
	}
	for idx := 0; idx < len(s2); idx++ {
		if c1[idx] != s2[idx] {
			logger.LOGE("invalid s2")
			return errors.New("invalid s2")
		}
	}
	logger.LOGT("handleshake ok")
	return
}

func (this *RTMPPuller) GetType() string {
	return rtmpTypePuller
}

func (this *RTMPPuller) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *RTMPPuller) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *RTMPPuller) SetParent(parent wssAPI.Obj) {
	this.parent = parent
}

func (this *RTMPPuller) threadRead() {
	this.reading = true
	this.waitRead.Add(1)
	defer func() {
		this.waitRead.Done()
		this.rtmp.Conn.Close()
		this.rtmp.Conn = nil
		logger.LOGT("stop read,close conn")
	}()
	for this.reading {
		this.rtmp.ReadPacket()
	}
}

func (this *RTMPPuller) readRTMPPkt() (packet *RTMPPacket, err error) {
	err = this.rtmp.Conn.SetReadDeadline(time.Now().Add(time.Duration(serviceConfig.TimeoutSec) * time.Second))
	if err != nil {
		logger.LOGE(err.Error())
		return
	}
	defer this.rtmp.Conn.SetReadDeadline(time.Time{})
	packet, err = this.rtmp.ReadPacket()
	return
}
