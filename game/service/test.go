package service

import "github.com/laonsx/gamelib/rpc"

func init() {

	rpc.RegisterService(&TestService{})
}

type TestService struct{}

func (testService *TestService) Hello(data []byte, session *rpc.Session) []byte {

	return nil
}
