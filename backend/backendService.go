package backend

import (
	"encoding/json"
	"errors"
	"logger"
	"net/http"
	"strconv"
	"wssAPI"
)

type BackendHander interface {
	Init(msg *wssAPI.Msg) error
	GetRoute() string
}

type BackendService struct {
}

type BackendConfig struct {
	Port     int    `json:"Port"`
	RootName string `json:"Usr"`
	RootPwd  string `json:"Pwd"`
}

var serviceConfig BackendConfig

func (this *BackendService) Init(msg *wssAPI.Msg) (err error) {
	if msg == nil || msg.Param1 == nil {
		logger.LOGE("init backend service failed")
		return errors.New("invalid param!")
	}

	fileName := msg.Param1.(string)
	err = this.loadConfigFile(fileName)
	if err != nil {
		logger.LOGE(err.Error())
		return errors.New("load backend config failed")
	}

	go func() {
		strPort := ":" + strconv.Itoa(serviceConfig.Port)
		handlers := backendHandlerInit()
		for _, item := range handlers {
			backHandler := item.(BackendHander)
			http.Handle(backHandler.GetRoute(), http.StripPrefix(backHandler.GetRoute(), backHandler.(http.Handler)))
		}
		err = http.ListenAndServe(strPort, nil)
		if err != nil {
			logger.LOGE("start backend serve failed")
		}
	}()
	return
}

func (this *BackendService) loadConfigFile(fileName string) (err error) {
	data, err := wssAPI.ReadFileAll(fileName)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &serviceConfig)
	if err != nil {
		return
	}
	return
}

func (this *BackendService) Start(msg *wssAPI.Msg) (err error) {
	return
}

func (this *BackendService) Stop(msg *wssAPI.Msg) (err error) {
	return
}

func (this *BackendService) GetType() string {
	return wssAPI.OBJ_BackendServer
}

func (this *BackendService) HandleTask(task wssAPI.Task) (err error) {
	return
}

func (this *BackendService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *BackendService) SetParent(parent wssAPI.Obj) {
	return
}

func backendHandlerInit() []BackendHander {
	handers := make([]BackendHander, 0)

	adminLoginHandle := &AdminLoginHandler{}
	lgData := &wssAPI.Msg{}
	loginData := AdminLoginData{}
	loginData.password = serviceConfig.RootPwd
	loginData.username = serviceConfig.RootName
	lgData.Param1 = loginData
	err := adminLoginHandle.Init(lgData)
	if err == nil {
		handers = append(handers, adminLoginHandle)
	} else {
		if err != nil {
			logger.LOGE("add adminLoginHandle error!")
		}
	}

	streamManagerHandle := &AdminStreamManageHandler{}
	streamManagerHandle.Init(nil)
	handers = append(handers, streamManagerHandle)

	return handers
}
