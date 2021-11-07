package codec

import "io"

type Header struct {
	ServiceMethod string // "Service.Method"
	Seq           uint64 // sequence number
	Error         string
}

// 对消息体进行编码的接口Codec
type Codec interface {
	io.Closer                         // 关闭接口
	ReadHeader(*Header) error         // 从io读取Header
	ReadBody(interface{}) error       // 从io读取body
	Write(*Header, interface{}) error // 将header、body写入io
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
