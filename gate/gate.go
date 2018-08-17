package gate

import (
	"encoding/binary"
	"errors"
	"time"

	"enen/common/pb"

	"github.com/golang/protobuf/proto"
	"github.com/laonsx/gamelib/server"
	"github.com/laonsx/gamelib/server/ws"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var gateServerHandler *GateServerHandler

type GateServerHandler struct {
	server server.GateServer
	state  int
}

func NewGateServer(addr string, originAllow string) server.GateServer {

	conf := &server.Config{
		Addr:        addr,
		MaxConn:     10000,
		OriginAllow: originAllow,
	}
	wsGateServer := ws.NewServer(viper.GetString("gate.name"), conf)

	gateServerHandler = &GateServerHandler{server: wsGateServer}
	wsGateServer.SetHandler(gateServerHandler)

	return wsGateServer
}

func agentStart(conn server.Conn) {

	agent, err := auth(conn)
	if err != nil {

		logrus.Warnf("auth err, conn=%d err=%s", conn.Id(), err.Error())

		return
	}

	agent.start(conn)
}

//握手认证
func auth(conn server.Conn) (a *agent, err error) {

	defer func() {

		if err != nil {

			conn.Close()
		}
	}()

	var b []byte
	b, err = conn.Read()
	if err != nil {

		return
	}

	var pnum uint16
	var body []byte
	pnum, body, err = decode(b)
	if err != nil {

		return
	}

	if pnum != PNUM_AUTH && pnum != PNUM_RECONN {

		err = errors.New("pnum err")

		return
	}

	req := new(pb.AuthRequest)
	err = proto.Unmarshal(body, req)
	if err != nil {

		return
	}

	secret := req.Secret
	uid := req.Uid

	a = agentService.getAgent(uid)
	if a == nil {

		err = errors.New("no agent")
		conn.Send(errorResp(pnum+1, 10))

		return
	}

	if secret != a.secret {

		err = errors.New("secret err")
		conn.Send(errorResp(pnum+1, 11))

		return
	}

	conn.SetReadDeadline(time.Duration(30) * time.Second)
	conn.AsyncSend(encode(pnum+1, 0, nil))

	return
}

func (gate *GateServerHandler) Open(conn server.Conn) {

	conn.SetReadDeadline(time.Duration(30) * time.Second)
	logrus.Infof("new conn id=%d addr=%s", conn.Id(), conn.RemoteAddr().String())

	go agentStart(conn)
}

func (gate *GateServerHandler) Close(conn server.Conn) {

	logrus.Infof("conn closing id=%d err=%v", conn.Id(), conn.Error())
}

func (gate *GateServerHandler) Count() int {

	return gate.server.Count()
}

func pack(pnum uint16, code uint16, b []byte) []byte {

	data := make([]byte, len(b)+4)

	binary.BigEndian.PutUint16(data[0:2], pnum)
	binary.BigEndian.PutUint16(data[2:4], code)

	copy(data[4:], b)

	return data
}

func pong() []byte {

	return pack(PNUM_PING+1, 0, nil)
}

func encode(pnum uint16, code uint16, b []byte) []byte {

	return pack(pnum, code, b)
}

func decode(b []byte) (uint16, []byte, error) {

	pnum := binary.BigEndian.Uint16(b[0:2])
	body := b[2:]

	return pnum, body, nil
}

func errorResp(pnum uint16, code uint16) []byte {

	return pack(pnum, code, nil)
}
