package pb

import (
	"encoding/binary"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

func Response(pb proto.Message) []byte {

	var b []byte
	if pb != nil {

		var err error

		b, err = proto.Marshal(pb)
		if err != nil {

			logrus.Errorf("proto error %s", err.Error())

			return nil
		}
	}

	data := make([]byte, len(b)+2)
	binary.BigEndian.PutUint16(data[0:2], OK)
	copy(data[2:], b)

	return data
}

func Error(code ErrorCode, v ...interface{}) []byte {

	var s string

	l := len(v)
	for i := 0; i < l; i++ {

		s += fmt.Sprintf(",%v", v[i])
	}

	if code < 500 {

		logrus.Errorf("errcode=%d%s", code, s)
	} else {

		logrus.Warnf("errcode=%d%s", code, s)
	}

	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data[0:2], code)

	return data
}

func FromResultData(data []byte) (ErrorCode, []byte) {

	if len(data) < 2 {

		return FORMATE, nil
	}

	return binary.BigEndian.Uint16(data[0:2]), data[2:]
}
