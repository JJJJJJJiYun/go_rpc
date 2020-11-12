package codec

import (
	"io"
)

// 请求头
type Header struct {
	// 方法名
	ServiceMethod string
	// 请求id
	Seq uint64
	// 错误信息
	Error string
}

// Codec 打解包器
type Codec interface {
	io.Closer
	ReadHeader(header *Header) error
	ReadBody(body interface{}) error
	Write(header *Header, body interface{}) error
}

type NewCodecFunc func(rwc io.ReadWriteCloser) Codec

type Type string

const (
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[JsonType] = newJsonCodec
}
