package robot

import (
	"enen/common/pb"
	"github.com/golang/protobuf/proto"
)

//todo 改为配置信息
var order bool = true

type requestCmd struct {
	id   int
	pnum uint16
	data proto.Message
}

//todo reqCmdList 可改从配置表中读取（过滤日志信息生成当前玩家所有操作命令生成配置表模拟用户线上操作）

var reqCmdList = map[int]*requestCmd{
	1: &requestCmd{id: 1, pnum: 1001, data: &pb.HelloRequest{ReqMsg: "hi,hi,hi,hi,hi,hi,hi."}},
	//2: &requestCmd{id: 2, pnum: 1001, data: &pb.HelloRequest{ReqMsg: "hi,nice to meet you."}},
	//3: &requestCmd{id: 3, pnum: 1051, data: &pb.MailListRequest{}},
	//4: &requestCmd{id: 4, pnum: 1053, data: &pb.MailDelRequest{}},
}

var respPbList = map[uint16]proto.Message{
	1002: &pb.HelloResponse{},
	1052: &pb.MailListResponse{},
	1054: &pb.MailDelResponse{},
}
