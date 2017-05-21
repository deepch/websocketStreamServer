package backend

import (
	"fmt"
	"net/http"
	"strconv"
	"events/eLiveListCtrl"
	"events/eStreamerEvent"
	"events/eRTMPEvent"
	"wssAPI"
	"errors"
)

type AdminStreamManageHandler struct {
	Route string
}

type StreamManageRequestData struct {
	Action Action
}


func (this *AdminStreamManageHandler) Init(data *wssAPI.Msg) (err error) {
	this.Route = "/admin/stream/manage"
	return
}

func (this *AdminStreamManageHandler) GetRoute() (route string) {
	return this.Route
}

func (this *AdminStreamManageHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == this.Route {
		this.handleStreamManageRequest(w, req)
	} else {
		badrequestResponse, err := BadRequest(WSS_SeverError, "server error in login")
		SendResponse(badrequestResponse, err, w)
	}
}

func (this *AdminStreamManageHandler) handleStreamManageRequest(w http.ResponseWriter, req *http.Request) {
	if !LoginHandler.isLogin {
		doManage(w)
		response, err := BadRequest(WSS_NotLogin, "please login")
		SendResponse(response, err, w)
		return
	} else {
		requestData := getRequestData(req)
		if requestData.Action.ActionToken != LoginHandler.AuthToken {
			response, err := BadRequest(WSS_UserAuthError, "Auth error")
			SendResponse(response, err, w)
			return
		} else { //do manage
			doManage(w)
		}
	}
}

func getRequestData(req *http.Request) StreamManageRequestData {
	data := StreamManageRequestData{}
	code := req.PostFormValue("action_code")
	codeInt, err := strconv.Atoi(code)
	if err != nil {
		data.Action.ActionCode = codeInt
	} else {
		data.Action.ActionCode = -1
	}
	data.Action.ActionToken = req.PostFormValue("action_token")
	return data
}

func doManage(w http.ResponseWriter) {
	eve := eLiveListCtrl.EveGetLiveList{}
	wssAPI.HandleTask(&eve)
	fmt.Println(eve)
}


func getReuqestActionEvent(req *http.Request)(wssAPI.Task, error){
	actionCode := req.PostFormValue("action_code")
	liveName := req.PostFormValue("live_name")
	if len(actionCode)==0{
		return nil,errors.New("no action code")
	}

	intCode,err :=strconv.Atoi(actionCode)

	var task wssAPI.Task
	if err != nil{
		return nil, errors.New("action code is error")
	}
	switch intCode{
		case WS_SHOW_ALL_STREAM:
			eve := &eLiveListCtrl.EveGetLiveList{}
			task = eve
		case WS_GET_LIVE_PLAYER_COUNT:
			eve := &eLiveListCtrl.EveGetLivePlayerCount{}
			eve.LiveName = liveName
			task = eve
		case WS_ENABLE_BLACK_LIST:
			eve := &eLiveListCtrl.EveEnableBlackList{}
			//eve.Enable = true
			task = eve
		case WS_SET_BLACK_LIST:
			eve := &eLiveListCtrl.EveSetBlackList{}
			//eve.Add = true
		case WS_ENABLE_WHITE_LIST:
			eve := &eLiveListCtrl.EveEnableWhiteList{}
			task = eve
		case WS_SET_WHITE_LIST:
			task = &eLiveListCtrl.EveSetWhiteList{}
		case WS_SET_UP_STREAM_APP:
			eve := &eLiveListCtrl.EveSetUpStreamApp{}
		case WS_PULL_RTMP_STREAM:
			task = &eRTMPEvent.EvePullRTMPStream{}
		case WS_ADD_SINK:
			task = &eStreamerEvent.EveAddSink{}
		case WS_DEL_SINK:
			task = &eStreamerEvent.EveDelSink{}
		case WS_ADD_SOURCE:
			task = &eStreamerEvent.EveAddSource{}
		case WS_DEL_SOURCE:
			task = &eStreamerEvent.EveDelSource{}
		case WS_GET_SOURCE:
			task = &eStreamerEvent.EveGetSource{}
		default:
			task = nil
	}
	if task == nil{
		return nil, errors.New("no function")
	}

	wssAPI.HandleTask(task)

	return task, nil
}