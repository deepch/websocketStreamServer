package backend

import (
	"encoding/json"
	"errors"
	"logger"
	"net/http"
	"strconv"
	"wssAPI"
)

type BackendService struct {
	parent wssAPI.Obj
}

type BackendConfig struct {
	Port        int    `json:"Port"`
	RootName    string `json:"Usr"`
	RootPwd     string `json:"Pwd"`
	ActionToken string
}

var service *BackendService
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
	service = this

	go func() {
		strPort := ":" + strconv.Itoa(serviceConfig.Port)
		//http.Handle("/admin",http.StripPrefix("/admin",this))
		http.Handle("/admin/login", http.StripPrefix("/admin/login", this))
		err = http.ListenAndServe(strPort, nil)
		if err != nil {
			logger.LOGE("start backend serve failed")
		}
	}()
	return
}

func (this *BackendService) SetParent(parent wssAPI.Obj) {
	this.parent = parent
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

//handle
func (this *BackendService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == "/admin/login" {
		this.HandleLoginRequest(w, req)
	}
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

func (this *BackendService) HandleTask(task *wssAPI.Task) (err error) {
	if task.Reciver == this.GetType() {
		return
	}
	if nil != this.parent {
		return this.parent.HandleTask(task)
	}
	return
}

func (this *BackendService) ProcessMessage(msg *wssAPI.Msg) (err error) {
	return
}

func (this *BackendService) HandleLoginRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		result, err := badRequest(1, "bad request")
		SendResponse(result, err, w)
	} else {
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")
		if len(username) > 0 && len(password) > 0 {
			ispass, authToken := compAuth(username, password)
			if ispass {
				serviceConfig.ActionToken = authToken
				responseData, err := passAuthResponseData(authToken)
				SendResponse(responseData, err, w)
			} else {
				responseData, err := badRequest(2, "login auth error")
				SendResponse(responseData, err, w)
			}
		} else {
			responseData, err := badRequest(2, "login auth error")
			SendResponse(responseData, err, w)
		}
	}
}

func badRequest(code int, msg string) ([]byte, error) {
	result := &ResponseData{}
	result.Code = code
	result.Msg = msg
	bytes, err := json.Marshal(result)
	return bytes, err
}
