package bilidanmaku

import (
	"fmt"
)

// DefaultHandler print cmd msg log
func DefaultHandler(c *Context) {
	cmd, err := c.Msg.Get("cmd").String()
	if err != nil {
		c.Abort()
	}
	switch cmd {
	case "LIVE":
		fmt.Println("直播开始。。。")

	case "PREPARING":
		fmt.Println("房主准备中。。。")
	case "DANMU_MSG":
		commentText, err := c.Msg.Get("info").GetIndex(1).String()
		if err != nil {
			fmt.Println("Json decode error failed: ", err)
			break
		}

		commentUser, err := c.Msg.Get("info").GetIndex(2).GetIndex(1).String()
		if err != nil {
			fmt.Println("Json decode error failed: ", err)
			break
		}
		fmt.Println(commentUser, " say: ", commentText)
		break
	}
}

// DebugHandler debug msg info
func DebugHandler(c *Context) {
	jbytes, _ := c.Msg.EncodePretty()
	fmt.Println(string(jbytes))
}
