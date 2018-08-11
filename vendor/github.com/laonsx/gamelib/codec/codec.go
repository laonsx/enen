package codec

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

var (
	msgpackHandle codec.MsgpackHandle
)

func MsgPack(v interface{}) ([]byte, error) {
	byteBuf := new(bytes.Buffer)
	enc := codec.NewEncoder(byteBuf, &msgpackHandle)
	err := enc.Encode(v)
	return byteBuf.Bytes(), err
}

func UnMsgPack(data []byte, v interface{}) error {
	dec := codec.NewDecoder(bytes.NewReader(data), &msgpackHandle)
	return dec.Decode(v)
}
