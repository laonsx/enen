package gate

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/laonsx/gamelib/gofunc"
	"github.com/laonsx/gamelib/rpc"
	"github.com/laonsx/gamelib/server"
	"github.com/laonsx/gamelib/timer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var agentService *agentServiceStruct

func InitAgent() {

	agentService = new(agentServiceStruct)
	agentService.agents = make(map[uint64]*agent)

	timer.AfterFunc(time.Duration(300)*time.Second, 0, func(n int) {

		agentService.clear()
	})
}

type agentServiceStruct struct {
	mux    sync.RWMutex
	agents map[uint64]*agent
}

func (as *agentServiceStruct) clear() {

	for _, a := range as.agents {

		if a.online == 0 && time.Now().Unix()-a.offlineTime.Unix() > 300 {

			as.delAgent(a)
		}
	}

	as.mux.RLock()
	defer as.mux.RUnlock()

	logrus.Infof("agent count = %d", len(as.agents))
}

func (as *agentServiceStruct) getAgent(uid uint64) *agent {

	as.mux.RLock()
	defer as.mux.RUnlock()

	return as.agents[uid]
}

func (as *agentServiceStruct) newAgent(userId uint64, secret string) *agent {

	as.mux.RLock()
	a, ok := as.agents[userId]
	as.mux.RUnlock()
	if ok {

		a.secret = secret
		if a.conn != nil {

			a.conn.Close()
		}
		a.offlineTime = time.Now()

		return a
	}

	a = new(agent)
	a.userId = userId
	a.secret = secret
	a.offlineTime = time.Now()

	a.subchans = make(map[string]int)

	stream, err := rpcStream(userId)
	if err != nil {

		return nil
	}

	a.stream = stream

	as.mux.Lock()
	as.agents[userId] = a
	as.mux.Unlock()

	logrus.Infof("agent new, uid=%d", userId)

	return a
}

func (as *agentServiceStruct) delAgent(a *agent) {

	as.mux.Lock()
	delete(as.agents, a.userId)
	as.mux.Unlock()

	a.close()

	//rpc.Call("center", "UserService.Logout", a.userId)

	logrus.Infof("agent exiting, uid=%d", a.userId)
}

func (as *agentServiceStruct) agentCount() int {

	as.mux.RLock()
	defer as.mux.RUnlock()

	return len(as.agents)
}

type agent struct {
	mux         sync.RWMutex
	userId      uint64                //玩家id
	conn        server.Conn           //玩家连接标示
	connTime    time.Time             //上线时间
	offlineTime time.Time             //离线时间
	limitTime   time.Time             //发包频率检测时间
	stream      rpc.Game_StreamClient //游戏服务流
	packetCount int                   //包计数
	online      int                   //状态 0=离线 1=在线
	secret      string                //密钥
	subchans    map[string]int        //订阅的频道
}

func (a *agent) start(conn server.Conn) {

	logrus.Infof("agent starting, uid=%d conn=%d", a.userId, conn.Id())

	a.setConn(conn)
	a.setOnline()

	defer func() {

		conn.Close()
		a.setOffline()

		logrus.Infof("agent closing, uid=%d", a.userId)
	}()

	for {

		b, err := conn.Read()
		if err != nil {

			return
		}

		a.dispatch(b)
	}
}

func (a *agent) dispatch(data []byte) {

	defer gofunc.PrintPanic()

	pnum, body, err := decode(data)
	if err != nil {

		return
	}

	switch pnum {

	case PNUM_PING: //ping

		a.conn.SetReadDeadline(time.Duration(30) * time.Second)
		a.conn.AsyncSend(pong())

	default:

		err = a.input(pnum, body)
		if err != nil {

			a.conn.AsyncSend(errorResp(pnum+1, 1))
		}
	}
}

func (a *agent) setConn(conn server.Conn) {

	if a.conn == conn {

		return
	}

	a.conn = conn
	a.connTime = time.Now()
}

func (a *agent) input(pnum uint16, data []byte) error {

	logrus.Infof("agent input, uid=%d pnum=%d", a.userId, pnum)

	//发包频率控制
	if a.packetCount > 0 && a.packetCount%50 == 0 {

		ts := time.Since(a.limitTime)
		if ts.Seconds() <= 5 {

			a.kick()

			return errors.New("kick")
		}

		a.limitTime = time.Now()
	}

	a.packetCount++

	//todo 只保留getname函数，且移除node
	//通过协议号获取服务名
	serviceName, err := rpc.GetName(pnum)
	if err != nil {

		logrus.Errorf("agent input, uid=%d err=%s", a.userId, err.Error())

		return err
	}

	//发送数据到游戏逻辑服
	msg := &rpc.GameMsg{ServiceName: serviceName, Msg: data}
	if err = a.stream.Send(msg); err != nil {

		for i := 0; i < 5; i++ {

			err = a.resetStream()
			if err == nil {

				err = a.stream.Send(msg)

				break
			}

			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {

		logrus.Warnf("agent stream send failed, uid=%d err=%s", a.userId, err.Error())

		return err
	}

	//获取逻辑服回应
	result, err := a.stream.Recv()
	if err == io.EOF {

		return err
	}
	if err != nil {

		logrus.Warnf("agent stream recv failed, uid=%d err=%s", a.userId, err.Error())

		return err
	}

	err = a.output(pnum+1, result.Msg)
	if err != nil {

		logrus.Warnf("agent output, uid=%d err=%s", a.userId, err.Error())
	}

	return err
}

func (a *agent) output(pnum uint16, data []byte) error {

	if a.online == 0 {

		return errors.New("offline")
	}

	if len(data) < 2 {

		return errors.New("output data err")
	}

	code := binary.BigEndian.Uint16(data[0:2])
	err := a.conn.AsyncSend(encode(pnum, code, data[2:]))

	logrus.Infof("agent output, uid=%d pnum=%d err=%v", a.userId, pnum, err)

	return err
}

func (a *agent) close() {

	a.stream.CloseSend()
}

func (a *agent) resetStream() error {

	a.stream.CloseSend()

	time.Sleep(time.Duration(100) * time.Millisecond)

	stream, err := rpcStream(a.userId)
	if err != nil {

		return err
	}

	a.stream = stream

	logrus.Infof("agent reset stream, uid=%d", a.userId)

	return nil
}

//添加订阅频道
func (a *agent) addSub(chanid string, subid int) bool {

	a.mux.Lock()
	defer a.mux.Unlock()

	if a.subchans == nil {

		return false
	}

	a.subchans[chanid] = subid

	return true
}

//根据频道id获取订阅者id
func (a *agent) getSubId(chanid string) (subid int) {

	if a.subchans == nil {

		return
	}

	subid = a.subchans[chanid]

	return
}

//删除订阅频道
func (a *agent) delSub(chanid string) {

	a.mux.Lock()
	defer a.mux.Unlock()

	delete(a.subchans, chanid)
}

func (a *agent) setOnline() {

	a.online = 1

	if a.subchans == nil {

		a.subchans = make(map[string]int)
	}

	//rpc.Call("center", "UserService.UserOnline", a.userId)
}

func (a *agent) setOffline() {

	a.online = 0
	a.offlineTime = time.Now()

	//Multicast.UnSubscribeAll(a.userId, a.subchans)

	a.subchans = nil

	//rpc.Call("center", "UserService.UserOffline", a.userId)
}

func (a *agent) isOnline() bool {

	return a.online == 1
}

func (a *agent) kick() {

	if a.conn != nil {

		a.conn.Close()
		agentService.delAgent(a)

		logrus.Infof("agent kick uid=%d", a.userId)
	}
}

func rpcStream(userId uint64) (stream rpc.Game_StreamClient, err error) {

	md := make(map[string]string)
	md["uid"] = strconv.FormatUint(userId, 10)
	md["name"] = "agent"

	stream, err = rpc.Stream(strings.Replace(viper.GetString("gate.name"), "t", "m", 1), md)

	return
}
