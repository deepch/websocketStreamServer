package backend

type ResponseData struct {
	Code int `json:"code"`
	Data Data `json:"data,omitempty"`
	Msg string `json:"msg"`
}

type Data struct {
	UserData Usr `json:"usr_data"`
	Action Action `json:"action"`
}

type Usr struct {
	Usrname string `json:"usrname"`
	Token string `json:"token"`
}

type Action struct {
	ActionCode int `json:"action_code"`
	ActionToken string `json:"action_token"`
}

const (
	WS_SHOW_ALL_STREAM = iota
)