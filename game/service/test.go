package service

import (
	"fmt"

	"enen/common/pb"

	"github.com/golang/protobuf/proto"
	"github.com/laonsx/gamelib/rpc"
	"github.com/sirupsen/logrus"
)

func init() {

	rpc.RegisterService(&TestService{})
}

type TestService struct{}

func (testService *TestService) Hello(data []byte, session *rpc.Session) []byte {

	req := pb.HelloRequest{}
	err := proto.Unmarshal(data, &req)
	if err != nil {

		return pb.Error(pb.PBUNMARSHAL, "TestService.Hello", err)
	}

	logrus.WithFields(logrus.Fields{
		"uid": session.Uid,
		"req": req,
	}).Debug("TestService.Hello")

	resp := pb.HelloResponse{}
	resp.RespMsg = fmt.Sprintf("%s too.", req.ReqMsg)

	return pb.Response(&resp)
}
