package bilidanmaku

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	prefixed "github.com/dongshimou/logrus-prefixed-formatter"
	"github.com/sirupsen/logrus"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
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

const (
	min        = 10000000
	max        = 2000000000
	defaultCap = 8192
)

// CmdType cmd const,Cmd开头的一系列常量
type CmdType string

type Handler interface {
	HandleFunc(c *Context)
}

type HandleFunc func(c *Context)

func (f HandleFunc) HandleFunc(context *Context) { f(context) }

// 获取直播间长号
func getRealRoomID(rid int) (realID int, err error) {
	resp, err := http.Get(fmt.Sprintf("http://api.live.bilibili.com/room/v1/Room/room_init?id=%d", rid))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var res roomInitResult
	jbytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(jbytes, &res); err != nil {
		return
	}
	if res.Code == 0 {
		return res.Data.RoomID, nil
	}
	return 0, fmt.Errorf(res.Message)
}

// 连接类型
type ConnType string

const (
	WebSocket = ConnType("websocket")
	Tcp       = ConnType("tcp")
)

type BiliLiveClient struct {
	ctx             context.Context
	cal             context.CancelFunc
	roomID          int
	ChatPort        int
	protocolVersion uint16
	ChatHost        string
	tcpConn         net.Conn
	uid             int
	handlerMap      map[CmdType]([]Handler)
	handlerMutex    sync.RWMutex
	connected       bool
	eventChan       chan Context
	connectType     ConnType
	conf            *ResponseDanmuConf
	wsConn          *websocket.Conn
	oldOnline       uint32
}

// 新客户端
func NewBiliBiliClient() *BiliLiveClient {
	bili := new(BiliLiveClient)
	bili.ChatHost = "livecmt-1.bilibili.com"
	bili.ChatPort = 788
	// version=2 deflate 新增的flate压缩 2020/04/20
	bili.protocolVersion = 2
	bili.handlerMap = make(map[CmdType]([]Handler))
	bili.eventChan = make(chan Context, defaultCap)
	bili.handlerMutex = sync.RWMutex{}
	go bili.Run()
	return bili
}

// 获取事件function
func (bili *BiliLiveClient) getCmdFunc(cmd CmdType) []Handler {
	bili.handlerMutex.RLock()
	defer bili.handlerMutex.RUnlock()
	return bili.handlerMap[cmd]
}

// 事件处理
func (bili *BiliLiveClient) Run() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case e, ok := <-bili.eventChan:
			if !ok {
				return
			}
			// 特定的
			for _, f := range bili.getCmdFunc(e.Cmd) {
				f.HandleFunc(&e)
				if e.IsAbort() {
					break
				}
			}
			// 注册了全部的
			for _, f := range bili.getCmdFunc(CmdAll) {
				f.HandleFunc(&e)
				if e.IsAbort() {
					break
				}
			}
		case <-ticker.C:
			// log.Debug("client is running!")
		}
	}
}

func (bili *BiliLiveClient) RegHandler(cmd CmdType, handler Handler) {
	bili.handlerMutex.Lock()
	defer bili.handlerMutex.Unlock()
	funcAddr := append(bili.handlerMap[cmd], handler)
	bili.handlerMap[cmd] = funcAddr
}

func (bili *BiliLiveClient) RegHandleFunc(cmd CmdType, hfunc HandleFunc) {
	bili.RegHandler(cmd, hfunc)
}

func (bili *BiliLiveClient) websockConnet(roomId int) error {
	// 获取ws服务器地址
	res, err := http.Get(fmt.Sprintf("https://api.live.bilibili.com/room/v1/Danmu/getConf?room_id=%d&platform=pc&player=web", roomId))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	conf := &ResponseDanmuConf{}
	if err := json.Unmarshal(data, conf); err != nil {
		return err
	}
	bili.conf = conf
	u := url.URL{Scheme: "wss", Host: bili.conf.Data.Host, Path: "/sub"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return err
	}
	defer c.Close()
	bili.wsConn = c
	log.Info("弹幕链接中。。。")
	if err := bili.SendJoinChannel(roomId); err != nil {
		log.Info(err)
		return err
	}
	bili.connected = true
	log.Info("开始接收消息...")
	go bili.heartbeatLoop()
	return bili.receiveMessageLoop()
}

func (bili *BiliLiveClient) Write(data []byte) error {
	switch bili.connectType {
	case WebSocket:
		{
			return bili.wsConn.WriteMessage(websocket.TextMessage, data)
		}

	case Tcp:
		{
			_, err := bili.tcpConn.Write(data)
			return err
		}
	}
	return nil
}
func (bili *BiliLiveClient) recv(buf []byte, l int) error {
	if _, err := io.ReadAtLeast(bili.tcpConn, buf, l); err != nil {
		return err
	}
	return nil
}

func (bili *BiliLiveClient) tcpConnet(roomId int) error {
	dstAddr := fmt.Sprintf("%s:%d", bili.ChatHost, bili.ChatPort)
	dstConn, err := net.Dial("tcp", dstAddr)
	if err != nil {
		return err
	}
	defer dstConn.Close()
	bili.tcpConn = dstConn
	log.Info("弹幕链接中。。。")
	if err := bili.SendJoinChannel(roomId); err != nil {
		log.Info(err)
		return err
	}
	bili.connected = true
	log.Info("开始接收消息...")
	go bili.heartbeatLoop()
	return bili.receiveMessageLoop()
}

// ConnectServer define
func (bili *BiliLiveClient) ConnectServer(roomID int) error {
	log.Infof("%d 获取真实房间id(长号)...", roomID)
	roomId, err := getRealRoomID(roomID)
	bili.roomID = roomId
	if err != nil {
		return err
	}
	log.Info("开始进入房间...", roomId)
	// 默认方式为websocket (原tcp来源于官方弹幕姬,貌似协议有改动)
	bili.connectType = WebSocket
	bili.ctx, bili.cal = context.WithCancel(context.Background())
	switch bili.connectType {
	case WebSocket:
		return bili.websockConnet(roomId)
	case Tcp:
		return bili.tcpConnet(roomId)
	}
	return nil
}

// heartbeatLoop keep heartbeat and get online
func (bili *BiliLiveClient) heartbeatLoop() {
	log.Warn("开始心跳循环...")
	var (
		ticker    *time.Ticker
		heartData string
	)
	switch bili.connectType {
	case WebSocket:
		ticker = time.NewTicker(time.Second * 30)
		heartData = WsHeart
	case Tcp:
		ticker = time.NewTicker(time.Second * 5)
		heartData = ""
	}
	for {
		select {
		case <-ticker.C:
			{
				if bili.connected {
					err := bili.sendSocketData(0, 16, bili.protocolVersion, 2, 1, heartData)
					if err != nil {
						bili.connected = false
						log.Warn("心跳包错误:", err)
						return
					}
				}
			}
		case <-bili.ctx.Done():
			{
				log.Warn("心跳协程结束.")
				return
			}
		}
	}
}

//GetRoomID Get the current room ID
func (bili *BiliLiveClient) GetRoomID() int { return bili.roomID }

// SendJoinChannel define
func (bili *BiliLiveClient) SendJoinChannel(channelID int) error {
	bili.uid = rand.Intn(max) + min
	body := fmt.Sprintf("{\"roomid\":%d,\"uid\":%d}", channelID, bili.uid)
	// var packetModel = new {roomid = channelId, uid = 0, protover = 2, token=token, platform="danmuji"}; //c#

	// 0000 0101 0010 0001 0000 0007 0000 0001
	return bili.sendSocketData(0, 16, bili.protocolVersion, 7, 1, body)
}

// sendSocketData define
func (bili *BiliLiveClient) sendSocketData(packetlength uint32, magic uint16, ver uint16, action uint32, param uint32, body string) error {
	bodyBytes := []byte(body)
	if packetlength == 0 {
		packetlength = uint32(len(bodyBytes) + 16)
	}
	headerBytes := new(bytes.Buffer)
	var data = []interface{}{
		packetlength,
		magic,
		ver,
		action,
		param,
	}
	for _, v := range data {
		err := binary.Write(headerBytes, binary.BigEndian, v)
		if err != nil {
			return err
		}
	}
	socketData := append(headerBytes.Bytes(), bodyBytes...)
	return bili.Write(socketData)
}

func (bili *BiliLiveClient) copy(src []byte) []byte {
	dest := make([]byte, len(src))
	copy(dest, src)
	return dest
}

func (bili *BiliLiveClient) b4int(src []byte) int {
	return int(binary.BigEndian.Uint32(src))
}
func (bili *BiliLiveClient) b2int(src []byte) int {
	raw := make([]byte, 2)
	raw = append(raw, src...)
	return int(binary.BigEndian.Uint32(raw))
}

func (bili *BiliLiveClient) onlineChangeEvent(total uint32) {
	sj := simplejson.New()
	sj.Set("cmd", CmdOnlineChange)
	sj.Set("online", total)
	bili.callCmdHandlerChain(&Context{RoomID: bili.roomID, Msg: sj, Cmd: CmdOnlineChange})
}

func (bili *BiliLiveClient) wsRecv() error {
	_, buf, err := bili.wsConn.ReadMessage()
	if err != nil {
		log.Error(err)
		return err
	}
	msg := parseHead(buf)
	if msg.Blen <= 0 {
		return nil
	}
	msg.Data = buf[msg.Hlen:]
	// log.Debug("websocket head==>",msg.Head," data len:",len(msg.Data)," data :",string(msg.Data),"<==")
	switch msg.Action {
	case 1, 2, 3:
		// heart beat response
		// log.Debug("1,2,3:",msg.Data)
		newOnline := uint32(b4int(msg.Data))
		if bili.oldOnline != newOnline {
			bili.oldOnline = newOnline
			bili.onlineChangeEvent(newOnline)
		}
	case 4, 5:
		if err := bili.parseDanMu(msg.Data, msg.isZlib()); err != nil {
			log.Info(err)
			return err
		}
	case 6, 7, 8:
		// ok response
		log.Debug("6,7,8:", string(msg.Data))
	case 17:
		log.Debug("17:", string(msg.Data))
	default:
		log.Debug("default:", msg.Data)
	}
	return nil
}

func (bili *BiliLiveClient) tcpRecv() error {
	buf := make([]byte, HeadLength)
	if err := bili.recv(buf, HeadLength); err != nil {
		log.Error(err)
		return err
	}
	msg := parseHead(buf)
	bLen := int(msg.Blen - HeadLength)
	if bLen <= 0 {
		return nil
	}
	log.Debug("tcp head==>", buf)
	switch msg.Action {
	case 1, 2, 3: // 3 心跳
		buf = make([]byte, HeartLength)
		if err := bili.recv(buf, HeartLength); err != nil {
			log.Error(err)
			return err
		}
		newOnline := binary.BigEndian.Uint32(buf)
		if bili.oldOnline != newOnline {
			bili.oldOnline = newOnline
			bili.onlineChangeEvent(newOnline)
		}
	case 4, 5: // 5 消息
		buf = make([]byte, bLen)
		if err := bili.recv(buf, bLen); err != nil {
			log.Error(err)
			return err
		}
		if err := bili.parseDanMu(buf, msg.isZlib()); err != nil {
			log.Info(err)
			return err
		}
	case 6, 7, 8:
		buf = make([]byte, bLen)
		if err := bili.recv(buf, bLen); err != nil {
			log.Error(err)
			return err
		}
	case 17:
	default:
		buf = make([]byte, bLen)
		if err := bili.recv(buf, bLen); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func (bili *BiliLiveClient) receiveMessageLoop() (err error) {
	defer CatchThrowHandle(func(e error) {
		bili.connected = false
		bili.cal()
		err = e
	})
	for bili.connected {
		if err := func() error {
			switch bili.connectType {
			case Tcp:
				return bili.tcpRecv()
			case WebSocket:
				return bili.wsRecv()
			}
			return nil
		}(); err != nil {
			if err == io.EOF {
				log.Info(err)
			} else {
				log.Error(err)
			}
			// websocket: close 1006 (abnormal closure): unexpected EOF. // 心跳包错误
			panic(err)
		}
	}
	return nil
}

// 使用flate解压
func (bili *BiliLiveClient) flate(message []byte) []byte {
	buff, err := ioutil.ReadAll(flate.NewReader(bytes.NewReader(message[2:])))
	if err != nil {
		log.Info(err)
		return nil
	}
	// log.Info("flate===>",string(buff))
	// log.Info("================")
	return buff
}

// 使用zlib解压
func (bili *BiliLiveClient) zlib(message []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(message))
	if err != nil {
		return nil, err
	}
	enflated, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	// log.Info("zlib===>",string(enflated))
	// log.Info("================")
	return enflated, err
}

// 全站消息
// {"full":{"head_icon":"http:\/\/i0.hdslb.com\/bfs\/live\/b049ac07021f3e4269d22a79ca53e6e7815af9ba.png","tail_icon":"http:\/\/i0.hdslb.com\/bfs\/live\/822da481fdaba986d738db5d8fd469ffa95a8fa1.webp","head_icon_fa":"http:\/\/i0.hdslb.com\/bfs\/live\/b049ac07021f3e4269d22a79ca53e6e7815af9ba.png","tail_icon_fa":"http:\/\/i0.hdslb.com\/bfs\/live\/38cb2a9f1209b16c0f15162b0b553e3b28d9f16f.png","head_icon_fan":1,"tail_icon_fan":4,"background":"#FFE6BDFF","color":"#9D5412FF","highlight":"#FF6933FF","time":10},"half":{"head_icon":"http:\/\/i0.hdslb.com\/bfs\/live\/4db5bf9efcac5d5928b6040038831ffe85a91883.png","tail_icon":"","background":"#FFE6BDFF","color":"#9D5412FF","highlight":"#FF6933FF","time":8},"side":{"head_icon":"http:\/\/i0.hdslb.com\/bfs\/live\/fa323d24f448d670bcc3dc59996d17463860a6b3.png","background":"#F5EBDDFF","color":"#DA9F77FF","highlight":"#C67137FF","border":"#ECDDC0FF"},"msg_type":1,"cmd":"NOTICE_MSG","roomid":680462,"real_roomid":680462,"msg_common":"\u606d\u559c\u4e3b\u64ad<%\u662f\u5357\u6e86\u5416OvO%>\u593a\u5f97<%15:00-16:00%>\u5c0f\u65f6\u603b\u699c\u7b2c\u4e00\u540d\uff01\u8d76\u5feb\u6765\u56f4\u89c2\u5427~","msg_self":"\u606d\u559c\u4e3b\u64ad<%\u662f\u5357\u6e86\u5416OvO%>\u593a\u5f97<%15:00-16:00%>\u5c0f\u65f6\u603b\u699c\u7b2c\u4e00\u540d\uff01","link_url":"https:\/\/live.bilibili.com\/680462"}

func (bili *BiliLiveClient) parseDanMu(message []byte, isZlib bool) (err error) {
	if isZlib {
		jstr, err := bili.zlib(message)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Debug("zlib msg:", string(jstr))
		for {
			msg, sub := parse(jstr)
			log.Debug("action:", msg.Action, " data:", string(msg.Data))
			if err := bili.parseJson(msg.Data); err != nil {
				log.Error(err)
				return err
			}
			if len(sub) != 0 {
				jstr = sub
			} else {
				break
			}
		}
		return nil
	} else {
		// 未压缩的内容
		return bili.parseJson(message)
	}
}

// 解析json格式
func (bili *BiliLiveClient) parseJson(jstr []byte) (err error) {
	// <d p="23.826000213623,1,25,16777215,1422201084,0,057075e9,757076900">我从未见过如此厚颜无耻之猴</d>
	// 0:时间(弹幕出现时间)
	// 1:类型(1从右至左滚动弹幕|6从左至右滚动弹幕|5顶端固定弹幕|4底端固定弹幕|7高级弹幕|8脚本弹幕)
	// 2:字号
	// 3:颜色
	// 4:时间戳 ?
	// 5:弹幕池id
	// 6:用户hash
	// 7:弹幕id
	dic, err := simplejson.NewJson(jstr)
	if err != nil {
		log.Error(err)
		return
	}
	cmd, err := dic.Get("cmd").String()
	if err != nil {
		log.Error(err)
		return
	}
	// 弹幕升级了，弹幕cmd获得的值不是DANMU_MSG, 而是DANMU_MSG: + 版本, 例如: DANMU_MSG:4:0:2:2:2:0
	// 在这里兼容一下
	// log.Debug("========>",dic)
	if strings.HasPrefix(cmd, string(CmdDanmuMsg)) {
		cmd = string(CmdDanmuMsg)
	}
	bili.callCmdHandlerChain(&Context{RoomID: bili.roomID, Msg: dic, Cmd: CmdType(cmd)})
	return nil
}

func (bili *BiliLiveClient) callCmdHandlerChain(c *Context) {
	if len(bili.eventChan) >= defaultCap {
		log.Error("chan is full")
		return
	}
	log.Debug("eventChan len :", len(bili.eventChan))
	bili.eventChan <- *c
}

func CatchThrowHandle(handle func(err error)) {
	if p := recover(); p != nil {
		if e, ok := p.(error); ok {
			handle(e)
		} else {
			panic(p)
		}
	}
}
