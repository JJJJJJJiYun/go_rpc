package opt

import (
	"github.com/JJJJJJJiYun/go_rpc/codec"
)

const (
	MagicNumber = 717
)

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.JsonType,
}
