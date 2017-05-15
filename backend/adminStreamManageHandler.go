package backend

import (
	"fmt"
	"net/http"
	"strconv"
	//"streamer"
	"events/eLiveListCtrl"
	"wssAPI"
)

type AdminStreamManageHandler struct {
	Route string
}

type StreamManageRequestData struct {
	Action Action
}

//var managers streamer.StreamerService

func (this *AdminStreamManageHandler) Init(data *wssAPI.Msg) (err error) {
	this.Route = "/admin/stream/manage"
	//	managers := &streamer.StreamerService{}
	//	managers.Init(nil)
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
