package gate

import (
	"log"
	"time"

	"enen/common"
	"enen/common/pb"
	"github.com/golang/protobuf/proto"
	"github.com/laonsx/gamelib/rpc"
	"github.com/laonsx/gamelib/timer"
	"github.com/spf13/viper"
)

var (
	data []byte

	serverConf *common.Config
)

func loggingCenter() {

	req := pb.CenterRequest{}
	req.Gate = &pb.GateInfo{
		Addr:   serverConf.WebSocketAddr,
		Weight: serverConf.Weight,
		State:  pb.GateState_Online,
		Name:   viper.GetString("gate.name"),
	}

	var err error
	data, err = proto.Marshal(&req)
	if err != nil {

		log.Println("LoggingCenterHandle start Marshal err", err)

		return
	}

	timer.AfterFunc(time.Second*3, 0, func(n int) {

		_, err := rpc.Call(serverConf.CenterNodeName, "CenterService.GateLogging", data, nil)

		if err != nil {

			log.Println("LoggingCenterHandle Call err", err)
		}
	})
}

func finishLoggingCenter() {

	req := pb.CenterRequest{}
	req.Gate = &pb.GateInfo{
		Addr:   serverConf.WebSocketAddr,
		Weight: serverConf.Weight,
		State:  pb.GateState_Close,
		Name:   viper.GetString("gate.name"),
	}

	var err error
	data, err = proto.Marshal(&req)
	if err != nil {

		log.Println("LoggingCenterHandle finish Marshal err", err)

		return
	}

	_, err = rpc.Call(serverConf.CenterNodeName, "CenterService.GateLogging", data, nil)
	if err != nil {

		log.Println("LoggingCenterHandle finish Call err", err)
	}
}
