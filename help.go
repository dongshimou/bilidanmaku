package bilidanmaku

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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

// 获取ws服务器

func getWsServer(roomId int) (*responseDanmuConf, error) {
	// 获取ws服务器地址
	res, err := http.Get(fmt.Sprintf("https://api.live.bilibili.com/room/v1/Danmu/getConf?room_id=%d&platform=pc&player=web", roomId))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	conf := &responseDanmuConf{}
	if err := json.Unmarshal(data, conf); err != nil {
		log.Error(err)
		return nil, err
	}
	return conf, nil
}
