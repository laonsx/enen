package gate

import (
	"enen/common/pb"

	"github.com/golang/protobuf/proto"
	"github.com/laonsx/gamelib/rpc"
	"github.com/sirupsen/logrus"
)

func init() {

	rpc.RegisterService(&GateService{})
}

//GateService 内部接口
type GateService struct{}

//Login 登录接口 center->gate
func (gateService *GateService) Login(data []byte, session *rpc.Session) []byte {

	req := &pb.GateRequest{}
	err := proto.Unmarshal(data, req)
	if err != nil {

		//return nil
	}

	uid := req.GetUid()

	logrus.Infof("login uid=%d", uid)

	agentService.newAgent(uid, req.GetSecret())

	return nil
}

//Kick 踢下线
func (gateService *GateService) Kick(data []byte, session *rpc.Session) []byte {

	req := &pb.GateRequest{}
	err := proto.Unmarshal(data, req)
	if err != nil {

		//return nil
	}

	if agent := agentService.getAgent(req.GetUid()); agent != nil {

		agent.kick()
	}

	return nil
}

//ReloadGameServiceConf 重新加载协议号对应的游戏服务配置
func (gateService *GateService) ReloadGameServiceConf(data []byte, session *rpc.Session) []byte {

	reloadGameServiceConf()

	return nil
}
