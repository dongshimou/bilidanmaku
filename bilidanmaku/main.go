package main

import (
	"flag"
	"time"

	"github.com/dongshimou/bilidanmaku"

	prefixed "github.com/dongshimou/logrus-prefixed-formatter"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	formatter := new(prefixed.TextFormatter)
	//formatter.CallerPrettyfier= func(frame *runtime.Frame) (function string, file string) {
	//	return frame.Function,fmt.Sprintf("%s:%d",frame.File,frame.Line)
	//}
	formatter.FullTimestamp = true
	formatter.TimestampFormat = "15:04:05"
	log.Formatter = formatter
	log.Level = logrus.InfoLevel
	log.SetReportCaller(true)
}

func main() {
	var roomid int
	flag.IntVar(&roomid, "room", 1, "直播房间号")
	flag.Parse()

	bili := bilidanmaku.NewBiliBiliClient()
	// bili.RegHandleFunc(bilidanmaku.CmdAll, bilidanmaku.DefaultHandler)
	// bili.RegHandleFunc(bilidanmaku.CmdAll, bilidanmaku.DebugHandler)
	bili.RegHandleFunc(bilidanmaku.CmdDanmuMsg, func(c *bilidanmaku.Context) {
		dinfo := c.GetDanmuInfo()
		if dinfo.Uname != "" {
			log.Infof("[%d]%s(%d) 说: %s", c.RoomID, dinfo.Uname, dinfo.UID, dinfo.Text)
		} else {
			log.Infof("[%d]%d 说: %s", c.RoomID, dinfo.UID, dinfo.Text)
		}
	})
	bili.RegHandleFunc(bilidanmaku.CmdWelcome, func(c *bilidanmaku.Context) {
		winfo := c.GetWelcomeInfo()
		if winfo.Uname != "" {
			log.Infof("[%d]%s(%d) 进入了房间", c.RoomID, winfo.Uname, winfo.UID)
		} else {
			log.Infof("[%d]%d 进入了房间", c.RoomID, winfo.UID)
		}
	})

	bili.RegHandleFunc(bilidanmaku.CmdSuperChatMsg, func(c *bilidanmaku.Context) {
		msg := c.GetSuperChatMsg()
		log.Debug(c.Msg)
		log.Infof("[%d]房间%s(%d) 超级留言了%s (价值%.3f)", c.RoomID, msg.UserInfo.Uname, msg.Uid, msg.Message, msg.Price)
	})

	bili.RegHandleFunc(bilidanmaku.CmdSendGift, func(c *bilidanmaku.Context) {
		gInfo := c.GetGiftInfo()
		log.Debug(c.Msg)
		log.Infof("[%d]%s(%d) %s 了 %s x %d (价值%.3f)", c.RoomID, gInfo.Uname, gInfo.UID, gInfo.Action, gInfo.GiftName, gInfo.Num, float32(gInfo.Price*gInfo.Num)/1000)
	})

	bili.RegHandleFunc(bilidanmaku.CmdOnlineChange, func(c *bilidanmaku.Context) {
		online := c.GetOnlineNumber()
		log.Infof("[%d]房间里当前人气值：%d", c.RoomID, online)
	})

	bili.RegHandleFunc(bilidanmaku.CmdNoticeMsg, func(c *bilidanmaku.Context) {
		nMsg := c.GetNoticeMsg()
		log.Infof("[%d]系统消息通知: %s", c.RoomID, nMsg.MsgCommon)
	})

	for {
		err := bili.ConnectServer(roomid)
		log.Warn("与弹幕服务器连接中断,3秒后重连。原因:", err)
		time.Sleep(time.Second * 3)
	}
}
