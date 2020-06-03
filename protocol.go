package bilidanmaku

import (
	"encoding/binary"
	"net"
)

// [body len] [head+ver] [action ] [param  ]
// [0 0 1 78] [0 16 0 0] [0 0 0 5] [0 0 0 0]
const (
	_HeadLength  = 16
	_HeartLength = 4

	_Body    = 4
	_Head    = 2
	_Version = 2
	_Action  = 4
	_Param   = 4
)

var (
	_WsHeart = string("[object Object]")
)

type liveMsg struct {
	Hlen    int
	Blen    int
	Version int
	Action  int
	Param   int
	Head    []byte
	Data    []byte
}

func (v *liveMsg) isZlib() bool {
	return v.Version == 2
}

func (v *liveMsg) isHeart() bool {
	return v.Version == 1
}

func (v *liveMsg) isRawJson() bool {
	return v.Version == 0
}

func b4int(src []byte) int {
	return int(binary.BigEndian.Uint32(src))
}

func b2int(src []byte) int {
	raw := make([]byte, 2)
	raw = append(raw, src...)
	return int(binary.BigEndian.Uint32(raw))
}

func parseHead(data []byte) (res *liveMsg) {
	res = &liveMsg{}
	res.Head = data[:_Body+_Head+_Version+_Action+_Param]
	pos := 0
	// 前4个为长度 0 0 1 78
	// 后12个为标识 0 16 0 0 0 0 0 5 0 0 0 0
	res.Blen = b4int(data[pos : pos+_Body])
	pos += _Body
	res.Hlen = b2int(data[pos : pos+_Head])
	pos += _Head
	res.Version = b2int(data[pos : pos+_Version])
	pos += _Version
	res.Action = b4int(data[pos : pos+_Action])
	pos += _Param
	res.Param = b4int(data[pos : pos+_Param])
	return res
}

func parse(data []byte) (res *liveMsg, surplus []byte) {
	res = parseHead(data[:_Body+_Head+_Version+_Action+_Param])
	res.Data = data[_Body+_Head+_Version+_Action+_Param : res.Blen]
	surplus = data[res.Blen:]
	return
}

func ioParse(conn net.Conn) (res *liveMsg, err error) {
	return nil, nil
}
