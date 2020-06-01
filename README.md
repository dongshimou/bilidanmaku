# bilidanmaku

B 站直播弹幕 Go 版。
在[原项目](https://github.com/lyyyuna/gobilibili) 基础上作了以下修改:

* 更换tcp(直播姬)的方式为websocket
* 部分修改

## 安装
```
go get github.com/dongshimou/bilidanmaku
``` 
## 示例

### 实时打印弹幕

```go
package main

import "github.com/dongshimou/bilidanmaku"

func main() {
	bili := bilidanmaku.NewBiliBiliClient()
	bili.RegHandleFunc(bilidanmaku.CmdAll, bilidanmaku.DefaultHandler)
	bili.ConnectServer(102)
}
```
#### 事件订阅
如果你希望订阅不同的事件，请尝试`bilidanmaku.Cmd*`开头的一系列常量。
以下是一些示例,你也可以随时在example目录下查看.

*订阅弹幕事件，并输出弹幕信息*

```go
bili := bilidanmaku.NewBiliBiliClient()
bili.RegHandleFunc(bilidanmaku.CmdDanmuMsg, func(c *bilidanmaku.Context) {
	dinfo := c.GetDanmuInfo()
	log.Printf("[%d]%d 说: %s\r\n", c.RoomID, dinfo.UID, dinfo.Text)
})
```

*进入房间*

```go
bili.RegHandleFunc(bilidanmaku.CmdWelcome, func(c *bilidanmaku.Context) {
	winfo := c.GetWelcomeInfo()
	if winfo.Uname != "" {
		log.Printf("[%d]%s 进入了房间\r\n", c.RoomID, winfo.Uname)
	} else {
		log.Printf("[%d]%d 进入了房间\r\n", c.RoomID, winfo.UID)
	}
})
```
*投喂礼物*

```go
bili.RegHandleFunc(bilidanmaku.CmdSendGift, func(c *bilidanmaku.Context) {
	gInfo := c.GetGiftInfo()
	log.Printf("[%d]%s %s 了 %s x %d (价值%.3f)\r\n", c.RoomID, gInfo.Uname, gInfo.Action, gInfo.GiftName, gInfo.Num, float32(gInfo.Price*gInfo.Num)/1000)
})
```
*在线人数变动*

```go
bili.RegHandleFunc(bilidanmaku.CmdOnlineChange, func(c *bilidanmaku.Context) bool {
	online := c.GetOnlineNumber()
	log.Printf("[%d]房间里当前在线：%d\r\n", c.RoomID, online)
})
```
*状态切换为直播开始*

```go
bili.RegHandleFunc(bilidanmaku.CmdLive, func(c *bilidanmaku.Context) bool {
	online := c.GetOnlineNumber()
	log.Println("主播诈尸啦!")
})
```
*状态切换为准备中*

```go
bili.RegHandleFunc(bilidanmaku.CmdPreparing, func(c *bilidanmaku.Context) bool {
	online := c.GetOnlineNumber()
	log.Println("主播正在躺尸")
})
```
*返回值*
Handler和HandleFunc的返回值用于控制调用链是否继续向下执行。 
如果你希望其它调用链不响应,调用`c.Abort()`

## 消息调试
通过注册 `bilidanmaku.DebugHandler`,可以在收到直播消息时查看原始消息。 

```go
package main

import "github.com/dongshimou/bilidanmaku"

func main() {
	bili := bilidanmaku.NewBiliBiliClient()
	bili.RegHandleFunc(bilidanmaku.CmdAll, bilidanmaku.DebugHandler)
	bili.ConnectServer(102)
}
```
运行后,当直播间发生事件时,将会输出类似格式的JSON输出:

```json
{
  "cmd": "DANMU_MSG",
  "info": [
    [
      0,
      1,
      25,
      16777215,
      1517402685,
      -136720455,
      0,
      "c42d0814",
      0
    ],
    "干嘛不播啦",
    [
      30731115,
      "Ed在",
      0,
      0,
      0,
      10000,
      1,
      ""
    ],
    [],
    [
      1,
      0,
      9868950,
      "\u003e50000"
    ],
    [],
    0,
    0,
    {
      "uname_color": ""
    }
  ]
}
```
以上示例的是一个弹幕消息.
其中`"cmd": "DANMU_MSG"`中的`"DANMU_MSG"`,就是调用`bili.RegHandleFunc`时需要传入的`cmd`参数。
你可以通过`bilidanmaku.CmdType("嘿,我是CmdType")`,将string转换为CmdType.
在这之后，你可以使用 bili.RegHandleFunc 或 bili.RegHandler 注册这个CmdType.

## 扩展
通过读取gobilibili.Context传入的Msg,可以处理尚未进行支持的事件.
请搭配上一节的消息调试进行食用。
以下是DefaultHandler的实现。

```go
func DefaultHandler(c *Context) bool {
	cmd, err := c.Msg.Get("cmd").String()
	if err != nil {
		return true
	}
	if cmd == "LIVE" {
		fmt.Println("直播开始。。。")
		return false
	}
	if cmd == "PREPARING" {
		fmt.Println("房主准备中。。。")
		return false
	}
	if cmd == "DANMU_MSG" {
		commentText, err := c.Msg.Get("info").GetIndex(1).String()
		if err != nil {
			fmt.Println("Json decode error failed: ", err)
			return false
		}

		commentUser, err := c.Msg.Get("info").GetIndex(2).GetIndex(1).String()
		if err != nil {
			fmt.Println("Json decode error failed: ", err)
			return false
		}
		fmt.Println(commentUser, " say: ", commentText)
		return false
	}
	return false
}
```






