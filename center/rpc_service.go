package center

import (
	"enen/common/pb"

	"github.com/golang/protobuf/proto"
	"github.com/laonsx/gamelib/rpc"
	"github.com/sirupsen/logrus"
)

func init() {

	rpc.RegisterService(&CenterService{})
}

//CenterService 对外接口
type CenterService struct{}

//UserLogout 登出
func (centerService *CenterService) UserLogout(data []byte, session *rpc.Session) []byte {

	req := pb.CenterRequest{}
	err := proto.Unmarshal(data, &req)
	if err != nil {

		return pb.Error(pb.PBUNMARSHAL, "CenterService.UserLogout", err)
	}

	for _, uid := range req.Uid {

		centerServer.delUser(uid)

		logrus.Infof("logout uid=%d", uid)
	}

	return pb.Response(nil)
}

//UserOnline 用户上线
func (centerService *CenterService) UserOnline(data []byte, session *rpc.Session) []byte {

	req := pb.CenterRequest{}
	err := proto.Unmarshal(data, &req)
	if err != nil {

		return pb.Error(pb.PBUNMARSHAL, "CenterService.UserOnline", err)
	}

	for _, uid := range req.Uid {

		err = centerServer.setOnline(uid)

		logrus.Infof("online uid=%d err=%v", uid, err)
	}

	return pb.Response(nil)
}

//UserOffline 用户离线
func (centerService *CenterService) UserOffline(data []byte, session *rpc.Session) []byte {

	req := pb.CenterRequest{}
	err := proto.Unmarshal(data, &req)
	if err != nil {

		return pb.Error(pb.PBUNMARSHAL, "CenterService.UserOffline", err)
	}

	for _, uid := range req.Uid {

		err = centerServer.setOffline(uid)

		logrus.Infof("offline uid=%d err=%v", uid, err)
	}

	return pb.Response(nil)
}

//UserLineStateList 获取在线状态
func (centerService *CenterService) UserLineStateList(data []byte, session *rpc.Session) []byte {

	req := pb.CenterRequest{}
	err := proto.Unmarshal(data, &req)
	if err != nil {

		return pb.Error(pb.PBUNMARSHAL, "CenterService.UserLineStateList", err)
	}

	reply := centerServer.onlineList(req.Uid)

	return pb.Response(&pb.CenterResponse{Online: reply})
}
