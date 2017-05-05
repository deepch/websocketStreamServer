package backend

import (
	"wssAPI"
	"logger"
	"errors"
	"encoding/json"
	"strconv"
	"net/http"
)

type  BackendService struct {
}

type BackendConfig struct {
	Port	int	`json:"Port"`
	RootName string `json:"Usr"`
	RootPwd string `json:"Pwd"`
}

var service *BackendService
var serviceConfig BackendConfig

func (this *BackendService) Init(msg *wssAPI.Msg) (err error){
	if msg == nil || msg.Param1 == nil{
		logger.LOGE("init backend service failed")
		return errors.New("invalid param!")
	}

	fileName := msg.Param1.(string)
	err = this.loadConfigFile(fileName)
	if err != nil {
		logger.LOGE(err.Error())
		return errors.New("load backend config failed")
	}
	service = this

	go func() {
		strPort := ":" + strconv.Itoa(serviceConfig.Port)
		http.Handle("/admin",http.StripPrefix("/admin",this))
		err = http.ListenAndServe(strPort,nil)
		if err != nil {
			logger.LOGE("start backend serve failed")
		}
		}()
	return
}

func (this *BackendService) loadConfigFile(fileName string) (err error)  {
	data, err := wssAPI.ReadFileAll(fileName)
	if err !=nil {
		return
	}

	err = json.Unmarshal(data,&serviceConfig)
	if err !=nil {
		return
	}
	return
}

//handle
func (this *BackendService) ServeHTTP(w http.ResponseWriter,req *http.Request){
	w.Write([]byte("error!\n"))
}

func (this *BackendService)Start(msg *wssAPI.Msg) (err error){
	return
}

func (this *BackendService) Stop (msg *wssAPI.Msg) (err error){
	return ;
}


func (this *BackendService) GetType() string{
	return wssAPI.OBJ_BackendServer
}

func (this *BackendService) HandleTask(task *wssAPI.Task) (err error) {
	return
}

func (this *BackendService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}
