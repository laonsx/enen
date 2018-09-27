package service

import (
	"enen/common/pb"
	"github.com/golang/protobuf/proto"
	"github.com/laonsx/gamelib/rpc"
	"github.com/sirupsen/logrus"
)

func init() {

	rpc.RegisterService(&MailService{})
}

type MailService struct{}

func (MailService *MailService) MailList(data []byte, session *rpc.Session) []byte {

	req := pb.MailListRequest{}
	err := proto.Unmarshal(data, &req)
	if err != nil {

		return pb.Error(pb.PBUNMARSHAL, "MailService.MailList", err)
	}
	logrus.WithFields(logrus.Fields{
		"uid": session.Uid,
		"req": req,
	}).Debug("MailService.MailList")

	resp := &pb.MailListResponse{}

	return pb.Response(resp)
}

func (MailService *MailService) MailDel(data []byte, session *rpc.Session) []byte {

	req := pb.MailDelRequest{}
	err := proto.Unmarshal(data, &req)
	if err != nil {

		return pb.Error(pb.PBUNMARSHAL, "MailService.MailDel", err)
	}
	logrus.WithFields(logrus.Fields{
		"uid": session.Uid,
		"req": req,
	}).Debug("MailService.MailDel")

	resp := &pb.MailDelResponse{}

	return pb.Response(resp)
}
