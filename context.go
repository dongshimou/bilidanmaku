package bilidanmaku

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bitly/go-simplejson"
)

const (
	// CmdAll 订阅所有cmd事件时使用
	CmdAll CmdType = ""
	// CmdLive 直播开始
	CmdLive CmdType = "LIVE"
	// CmdPreparing 直播准备中
	CmdPreparing CmdType = "PREPARING"

	// CmdDanmuMsg 弹幕消息
	// {"cmd":"DANMU_MSG","info":[[0,1,25,16777215,1591004454028,-1464052061,0,"034722a0",0,0,0],"我改一下",[23624052,"星云之音喵",0,0,0,10000,1,""],[1,"谜酥","谜之声",5082,6406234,"",0],[12,0,6406234,"\u003e50000"],["",""],0,0,null,{"ts":1591004454,"},0,0,null,null,0]}
	CmdDanmuMsg CmdType = "DANMU_MSG"

	// CmdWelcomeGuard 管理进房
	// {"cmd":"WELCOME_GUARD","data":{"uid":2721583,"username":"超高校级的无节操","gul":3,"mock_effect":0}}
	CmdWelcomeGuard CmdType = "WELCOME_GUARD"

	// CmdWelcome 群众进房
	CmdWelcome CmdType = "WELCOME"
	// CmdSendGift 赠送礼物
	CmdSendGift CmdType = "SEND_GIFT"
	// CmdNoticeMsg 系统消息通知
	CmdNoticeMsg CmdType = "NOTICE_MSG"
	// CmdOnlineChange 在线人数变动,这不是一个标准cmd类型,仅为了统一handler接口而加入
	CmdOnlineChange CmdType = "ONLINE_CHANGE"

	// ROOM_REAL_TIME_MESSAGE_UPDATE 房间消息变动
	// {"cmd":"ROOM_REAL_TIME_MESSAGE_UPDATE","data":{"roomid":5279,"fans":747474,"red_notice":-1}}
	// VOICE_JOIN_ROOM_COUNT_INFO

	// VOICE_JOIN_LIST

	// {"cmd":"ROOM_RANK","data":{"roomid":79558,"rank_desc":"\u7f51\u6e38\u5c0f\u65f6\u699c 30","color":"#FB7299","h5_url":"https:\/\/live.bilibili.com\/p\/html\/live-app-rankcurrent\/index.html?is_live_half_webview=1&hybrid_half_ui=1,5,85p,70p,FFE293,0,30,100,10;2,2,320,100p,FFE293,0,30,100,0;4,2,320,100p,FFE293,0,30,100,0;6,5,65p,60p,FFE293,0,30,100,10;5,5,55p,60p,FFE293,0,30,100,10;3,5,85p,70p,FFE293,0,30,100,10;7,5,65p,60p,FFE293,0,30,100,10;&anchor_uid=6810019&rank_type=master_realtime_area_hour&area_hour=1&area_v2_id=252&area_v2_parent_id=2","web_url":"https:\/\/live.bilibili.com\/blackboard\/room-current-rank.html?rank_type=master_realtime_area_hour&area_hour=1&area_v2_id=252&area_v2_parent_id=2","timestamp":1590993900}}

	// COMBO_SEND

	// {"cmd":"ACTIVITY_BANNER_UPDATE_V2","data":{"id":378,"title":"\u7b2c44\u540d","cover":"","background":"https:\/\/i0.hdslb.com\/bfs\/activity-plat\/static\/20190904\/b5e210ef68e55c042f407870de28894b\/V3oI34frU-.png","jump_url":"https:\/\/live.bilibili.com\/p\/html\/live-app-rankcurrent\/index.html?is_live_half_webview=1&hybrid_half_ui=1,5,85p,70p,FFE293,0,30,100,10;2,2,320,100p,FFE293,0,30,100,0;4,2,320,100p,FFE293,0,30,100,0;6,5,65p,60p,FFE293,0,30,100,10;5,5,55p,60p,FFE293,0,30,100,10;3,5,85p,70p,FFE293,0,30,100,10;7,5,65p,60p,FFE293,0,30,100,10;&anchor_uid=7560829&is_new_rank_container=1&area_v2_id=89&area_v2_parent_id=2&rank_type=master_realtime_area_hour&area_hour=1","title_color":"#8B5817","closeable":1,"banner_type":4,"weight":18,"add_banner":0}}

	// {"cmd":"ENTRY_EFFECT","data":{"id":4,"uid":2721583,"target_id":478735811,"mock_effect":0,"face":"https://i0.hdslb.com/bfs/face/fc6ad83bae2a971125b0cb64c1ea62cb735ddfef.jpg","privilege_type":3,"copy_writing":"欢迎舰长 \u003c%超高校级的无...%\u0,"copy_color":"","highlight_color":"#E6FF00","priority":70,"basemap_url":"https://i0.hdslb.com/bfs/live/1fa3cc06258e16c0ac4c209e2645fda3c2791894.png","show_avatar":1,"effective_time":2,"web_basemap_url":"","web_effective_time":0,"web_effect_close":0,"web_close_time":0}}

	CmdSuperChatMsg CmdType = "SUPER_CHAT_MESSAGE" // 付费留言
	// {"cmd":"SUPER_CHAT_MESSAGE","data":{"id":"352006","uid":14349165,"price":30,"rate":1000,"message":"\u82ad\u5a1c\u5a1c\u4f60\u597d\u554a","trans_mark":0,"is_ranked":1,"message_trans":"","background_image":"http:\/\/i0.hdslb.com\/bfs\/live\/1aee2d5e9e8f03eed462a7b4bbfd0a7128bbc8b1.png","background_color":"#EDF5FF","background_icon":"","background_price_color":"#7497CD","background_bottom_color":"#2A60B2","ts":1591004473,"token":"C932D837","medal_info":null,"user_info":{"uname":"shine2015","face":"http:\/\/i1.hdslb.com\/bfs\/face\/3073ec94bb43cb0d9eecb86d932410fe8ac3736c.jpg","face_frame":"","guard_level":0,"user_level":1,"level_color":"#969696","is_vip":0,"is_svip":0,"is_main_vip":1,"title":"0","manager":0},"time":60,"start_time":1591004473,"end_time":1591004533,"gift":{"num":1,"gift_id":12000,"gift_name":"\u9192\u76ee\u7559\u8a00"}}}

	// {"cmd":"SUPER_CHAT_MESSAGE_JPN","data":{"id":"352006","uid":"14349165","price":30,"rate":1000,"message":"\u82ad\u5a1c\u5a1c\u4f60\u597d\u554a","message_jpn":"\u30d0\u30ca\u30ca\u3055\u3093\u3001\u3053\u3093\u306b\u3061\u306f","is_ranked":1,"background_image":"http:\/\/i0.hdslb.com\/bfs\/live\/1aee2d5e9e8f03eed462a7b4bbfd0a7128bbc8b1.png","background_color":"#EDF5FF","background_icon":"","background_price_color":"#7497CD","background_bottom_color":"#2A60B2","ts":1591004473,"token":"C932D837","medal_info":null,"user_info":{"uname":"shine2015","face":"http:\/\/i1.hdslb.com\/bfs\/face\/3073ec94bb43cb0d9eecb86d932410fe8ac3736c.jpg","face_frame":"","guard_level":0,"user_level":1,"level_color":"#969696","is_vip":0,"is_svip":0,"is_main_vip":1,"title":"0","manager":0},"time":60,"start_time":1591004473,"end_time":1591004533,"gift":{"num":1,"gift_id":12000,"gift_name":"\u9192\u76ee\u7559\u8a00"}}}
)

// Context 消息上下文环境,提供快捷提取消息数据的功能
type Context struct {
	Msg     *simplejson.Json
	RoomID  int
	Cmd     CmdType
	isAbort bool
}

func (c *Context) Abort() {
	c.isAbort = true
}
func (c *Context) IsAbort() bool {
	return c.isAbort
}

// DanmuInfo 弹幕信息
type DanmuInfo struct {
	UID         int    `json:"uid"`          //用户ID
	Uname       string `json:"uname"`        //用户名称
	Rank        int    `json:"rank"`         //用户排名
	Level       int    `json:"level"`        //用户等级
	Text        string `json:"text"`         //说的话
	MedalLevel  int    `json:"medal_level"`  //勋章等级
	MedalName   string `json:"medal_name"`   //勋章名称
	MedalAnchor string `json:"medal_anchor"` //勋章所属主播
}

// GetDanmuInfo 在Handler中调用，从simplejson.Json中提取弹幕信息
func (p *Context) GetDanmuInfo() (dInfo DanmuInfo) {
	dInfo.Text, _ = p.Msg.Get("info").GetIndex(1).String()
	dInfo.Uname, _ = p.Msg.Get("info").GetIndex(2).GetIndex(1).String()
	dInfo.UID, _ = p.Msg.Get("info").GetIndex(2).GetIndex(0).Int()
	dInfo.MedalLevel, _ = p.Msg.Get("info").GetIndex(3).GetIndex(0).Int()
	dInfo.MedalName, _ = p.Msg.Get("info").GetIndex(3).GetIndex(1).String()
	dInfo.MedalAnchor, _ = p.Msg.Get("info").GetIndex(3).GetIndex(2).String()
	dInfo.Level, _ = p.Msg.Get("info").GetIndex(4).GetIndex(0).Int()
	dInfo.Rank, _ = p.Msg.Get("info").GetIndex(4).GetIndex(2).Int()
	return
}

// GetOnlineNumber 在Handler中调用，从simplejson.Json中提取房间在线人气值
func (p *Context) GetOnlineNumber() int {
	return p.Msg.Get("online").MustInt()
}

// WelcomeGuardInfo 管理进房信息
type WelcomeGuardInfo struct {
	GuardLevel string `json:"guard_level"`
	UID        int    `json:"uid"`
	Username   string `json:"username"`
}

// GetWelcomeGuardInfo 在Handler中调用，从一个simplejson.Json中提取管理进房信息
func (p *Context) GetWelcomeGuardInfo() (wInfo WelcomeGuardInfo) {
	wInfo.GuardLevel = p.Msg.Get("data").Get("guard_level").MustString()
	wInfo.UID = p.Msg.Get("data").Get("uid").MustInt()
	wInfo.Username = p.Msg.Get("data").Get("username").MustString()
	return
}

// WelcomeInfo 普通人员进房信息
type WelcomeInfo struct {
	IsAdmin bool   `json:"is_admin"`
	UID     int    `json:"uid"`
	Uname   string `json:"uname"`
	Vip     int    `json:"vip"`
	Svip    int    `json:"svip"`
}

// GetWelcomeInfo 在Handler中调用，从一个simplejson.Json中提取普通人员进房信息
func (p *Context) GetWelcomeInfo() (wInfo WelcomeInfo) {
	wInfo.IsAdmin = p.Msg.Get("data").Get("is_admin").MustBool() || p.Msg.Get("data").Get("isadmin").MustBool()
	wInfo.UID = p.Msg.Get("data").Get("uid").MustInt()
	wInfo.Uname = p.Msg.Get("data").Get("uname").MustString()
	wInfo.Vip = p.Msg.Get("data").Get("vip").MustInt()
	wInfo.Svip = p.Msg.Get("data").Get("svip").MustInt()
	return
}

// GiftInfo 礼物信息
type GiftInfo struct {
	Action    string `json:"action"`
	AddFollow int    `json:"addFollow"`
	BeatID    string `json:"beatId"`
	BizSource string `json:"biz_source"`
	Capsule   struct {
		Colorful struct {
			Change   int `json:"change"`
			Coin     int `json:"coin"`
			Progress struct {
				Max int `json:"max"`
				Now int `json:"now"`
			} `json:"progress"`
		} `json:"colorful"`
		Normal struct {
			Change   int `json:"change"`
			Coin     int `json:"coin"`
			Progress struct {
				Max int `json:"max"`
				Now int `json:"now"`
			} `json:"progress"`
		} `json:"normal"`
	} `json:"capsule"`
	EventNum   int    `json:"eventNum"`
	EventScore int    `json:"eventScore"`
	GiftID     int    `json:"giftId"`
	GiftName   string `json:"giftName"`
	GiftType   int    `json:"giftType"`
	CoinType   string `json:"coin_type"` // 礼物类型 silver(银瓜子) gold(金瓜子)
	Gold       int    `json:"gold"`
	// Medal       interface{} `json:"medal"`
	Metadata string `json:"metadata"`
	NewMedal int    `json:"newMedal"`
	NewTitle int    `json:"newTitle"`
	// NoticeMsg   interface{} `json:"notice_msg"`
	Num    int    `json:"num"`
	Price  int    `json:"price"`
	Rcost  int    `json:"rcost"`
	Remain int    `json:"remain"`
	Rnd    string `json:"rnd"`
	Silver int    `json:"silver"`
	// SmalltvMsg  interface{} `json:"smalltv_msg"`
	// SpecialGift interface{} `json:"specialGift"`
	Super     int    `json:"super"`
	Timestamp int    `json:"timestamp"`
	Title     string `json:"title"`
	TopList   *[]struct {
		Face       string `json:"face"`
		GuardLevel int    `json:"guard_level"`
		IsSelf     int    `json:"isSelf"`
		Rank       int    `json:"rank"`
		Score      int    `json:"score"`
		UID        int    `json:"uid"`
		Uname      string `json:"uname"`
	} `json:"top_list"`
	UID   int    `json:"uid"`
	Uname string `json:"uname"`
}

// GetGiftInfo 获取礼物信息
func (p *Context) GetGiftInfo() *GiftInfo {
	gInfo := &GiftInfo{}
	jbytes, _ := p.Msg.Get("data").Encode()
	jbytes = bytes.Replace(jbytes, []byte(`"beatId":0,`), []byte(`"beatId":"0",`), -1)
	jbytes = bytes.Replace(jbytes, []byte(`"rnd":0,`), []byte(`"rnd":"0",`), -1)
	if err := json.Unmarshal(jbytes, gInfo); err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(jbytes))
		gInfo.Action = p.Msg.Get("data").Get("action").MustString()
		gInfo.AddFollow = p.Msg.Get("data").Get("addFollow").MustInt()
		gInfo.BeatID = p.Msg.Get("data").Get("beatId").MustString()
		gInfo.BizSource = p.Msg.Get("data").Get("biz_source").MustString()
		gInfo.EventNum = p.Msg.Get("data").Get("eventNum").MustInt()
		gInfo.EventScore = p.Msg.Get("data").Get("eventScore").MustInt()
		gInfo.GiftID = p.Msg.Get("data").Get("giftId").MustInt()
		gInfo.GiftName = p.Msg.Get("data").Get("giftName").MustString()
		gInfo.GiftType = p.Msg.Get("data").Get("giftType").MustInt()
		gInfo.Gold = p.Msg.Get("data").Get("gold").MustInt()
		// gInfo.Medal = p.Msg.Get("data").Get("medal")
		gInfo.Metadata = p.Msg.Get("data").Get("metadata").MustString()
		gInfo.NewMedal = p.Msg.Get("data").Get("newMedal").MustInt()
		gInfo.NewTitle = p.Msg.Get("data").Get("newTitle").MustInt()
		// gInfo.NoticeMsg = p.Msg.Get("data").Get("")
		gInfo.Num = p.Msg.Get("data").Get("num").MustInt()
		gInfo.Price = p.Msg.Get("data").Get("price").MustInt()
		gInfo.Rcost = p.Msg.Get("data").Get("rcost").MustInt()
		gInfo.Remain = p.Msg.Get("data").Get("remain").MustInt()
		gInfo.Rnd = p.Msg.Get("data").Get("rnd").MustString()
		gInfo.Silver = p.Msg.Get("data").Get("silver").MustInt()
		// gInfo.SmalltvMsg = p.Msg.Get("data").Get("")
		// gInfo.SpecialGift = p.Msg.Get("data").Get("")
		gInfo.Super = p.Msg.Get("data").Get("super").MustInt()
		gInfo.Timestamp = p.Msg.Get("data").Get("timestamp").MustInt()
		gInfo.Title = p.Msg.Get("data").Get("title").MustString()
	}
	return gInfo
}

type NoticeMsg struct {
	MsgCommon string `json:msg_common`
}

// GetNoticeMsg 获取系统消息通知
func (p *Context) GetNoticeMsg() (nMsg NoticeMsg) {
	nMsg.MsgCommon = p.Msg.Get("msg_common").MustString()
	return
}

type SuperChatMsg struct {
	Id                    int64   `json:"id"`
	Uid                   int64   `json:"uid"`     // 用户id
	Price                 float64 `json:"price"`   // 价值
	Rate                  float64 `json:"rate"`    // (可能是金瓜子与rmb的比例)
	Message               string  `json:"message"` // 留言消息
	TransMark             int     `json:"trans_mark"`
	IsRanked              int     `json:"is_ranked"`
	MessageTrans          string  `json:"message_trans"`
	BackgroundImage       string  `json:"background_image"`
	BackgroundColor       string  `json:"background_color"`
	BackgroundIcon        string  `json:"background_icon"`
	BackgroundPriceColor  string  `json:"background_price_color"`
	BackgroundBottomColor string  `json:"background_bottom_color"`
	Ts                    int64   `json:"ts"`
	Token                 string  `json:"token"`
	// MedalInfo object
	UserInfo struct {
		Uname      string `json:"uname"`
		Face       string `json:"face"`
		FaceFrame  string `json:"face_frame"`
		GuardLevel int    `json:"guard_level"`
		UserLevel  int    `json:"user_level"`
		LevelColor string `json:"level_color"`
		IsVip      int    `json:"is_vip"`
		IsSvip     int    `json:"is_svip"`
		IsMainVip  int    `json:"is_main_vip"`
		Title      string `json:"title"`
		Manager    int    `json:"manager"`
	} `json:"user_info"` // 用户信息
	Time      int64 `json:"time"`       // 持续秒
	StartTime int64 `json:"start_time"` // 开始时间戳
	EndTime   int64 `json:"end_time"`   // 结束时间戳
	Gift      struct {
		Num      int    `json:"num"`
		GiftId   int    `json:"gift_id"`
		GiftName string `json:"gift_name"`
	} `json:"gift"`
}

// 获取付费留言
func (p *Context) GetSuperChatMsg() *SuperChatMsg {
	msg := &SuperChatMsg{}
	jbytes, _ := p.Msg.Get("data").Encode()
	if err := json.Unmarshal(jbytes, msg); err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(jbytes))
		msg.Uid = p.Msg.Get("data").Get("uid").MustInt64()
		msg.Price = p.Msg.Get("data").Get("price").MustFloat64()
		msg.Message = p.Msg.Get("data").Get("message").MustString()
		msg.UserInfo.Uname = p.Msg.Get("data").Get("user_info").Get("uname").MustString()
		msg.UserInfo.Face = p.Msg.Get("data").Get("user_info").Get("face").MustString()
		msg.Time = p.Msg.Get("data").Get("time").MustInt64()
		msg.StartTime = p.Msg.Get("data").Get("start_time").MustInt64()
		msg.EndTime = p.Msg.Get("data").Get("end_time").MustInt64()
	}
	return msg
}
