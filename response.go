package bilidanmaku

type ResponseDanmuConf struct {
	Code    int                   `json:"code"`
	Msg     string                `json:"msg"`
	Message string                `json:"message"`
	Data    ResponseDanmuConfData `json:"data"`
}

type ResponseDanmuConfData struct {
	RefreshRowFactor float64                       `json:"refresh_row_factor"`
	RefreshRate      int                           `json:"refresh_rate"`
	MaxDelay         int                           `json:"max_delay"`
	Port             int                           `json:"port"`
	Host             string                        `json:"host"`
	HostServerList   []ResponseDanmuConfDataServer `json:"host_server_list"`
	ServerList       []ResponseDanmuConfDataServer `json:"server_list"`
	Token            string                        `json:"token"`
}

type ResponseDanmuConfDataServer struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	WssPort int    `json:"wss_port,omitempty"`
	WsPort  int    `json:"ws_port,omitempty"`
}

type roomInitResult struct {
	Code int `json:"code"`
	Data struct {
		Encrypted   bool `json:"encrypted"`
		HiddenTill  int  `json:"hidden_till"`
		IsHidden    bool `json:"is_hidden"`
		IsLocked    bool `json:"is_locked"`
		LockTill    int  `json:"lock_till"`
		NeedP2p     int  `json:"need_p2p"`
		PwdVerified bool `json:"pwd_verified"`
		RoomID      int  `json:"room_id"`
		ShortID     int  `json:"short_id"`
		UID         int  `json:"uid"`
	} `json:"data",omitempty`
	Message string `json:"message"`
	Msg     string `json:"msg"`
}
