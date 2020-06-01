package bilidanmaku

import (
	"encoding/binary"
	"net"
)

// [body len] [head+ver] [action ] [param  ]
// [0 0 1 78] [0 16 0 0] [0 0 0 5] [0 0 0 0]
const (
	HeadLength  = 16
	HeartLength = 4

	Body    = 4
	Head    = 2
	Version = 2
	Action  = 4
	Param   = 4
)

var (
	WsHeart = string("[object Object]")
)

type LiveMsg struct {
	Hlen    int
	Blen    int
	Version int
	Action  int
	Param   int
	Head    []byte
	Data    []byte
}

func (v *LiveMsg) isZlib() bool {
	return v.Version == 2
}

func (v *LiveMsg) isHeart() bool {
	return v.Version == 1
}

func (v *LiveMsg) isRawJson() bool {
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

func parseHead(data []byte) (res *LiveMsg) {
	res = &LiveMsg{}
	res.Head = data[:Body+Head+Version+Action+Param]
	pos := 0
	// 前4个为长度 0 0 1 78
	// 后12个为标识 0 16 0 0 0 0 0 5 0 0 0 0
	res.Blen = b4int(data[pos : pos+Body])
	pos += Body
	res.Hlen = b2int(data[pos : pos+Head])
	pos += Head
	res.Version = b2int(data[pos : pos+Version])
	pos += Version
	res.Action = b4int(data[pos : pos+Action])
	pos += Param
	res.Param = b4int(data[pos : pos+Param])
	return res
}

func parse(data []byte) (res *LiveMsg, surplus []byte) {
	res = parseHead(data[:Body+Head+Version+Action+Param])
	res.Data = data[Body+Head+Version+Action+Param : res.Blen]
	surplus = data[res.Blen:]
	return
}

func ioParse(conn net.Conn) (res *LiveMsg, err error) {
	return nil, nil
}
