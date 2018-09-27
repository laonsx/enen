package robot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"enen/common/pb"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/laonsx/gamelib/crypt"
	"github.com/laonsx/gamelib/timer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var CenterAddr string

func Test() {

	test := viper.GetString("test.robot")
	cmds := strings.Split(test, "_")
	if len(cmds) != 2 {

		panic("test.robot cmd = " + test)
	}

	id := cmds[0]
	count, err := strconv.Atoi(cmds[1])
	if err != nil {

		panic("test.robot cmd err = " + err.Error())
	}

	defer func() {

		logrus.Println("done")
	}()

	timer.AfterFunc(100*time.Millisecond, count, func(n int) {

		token := fmt.Sprintf("%s_%d", id, n)
		start(token)

		logrus.Info(token + " login")
	})
}

func start(token string) {

	addr, secret, uid := login(token)

	logrus.Debugf("connecting to gate", uid)

	d := websocket.Dialer{}
	conn, _, err := d.Dial(fmt.Sprintf("ws://%s/ws", addr), nil)
	if err != nil {

		logrus.Fatal("dial: ", err)

		return
	}

	req := new(pb.GateRequest)
	req.Secret = secret
	req.Uid = uid

	b, _ := proto.Marshal(req)
	_, code, _, err := request(conn, 3, b)
	if err != nil || code != 0 {

		return
	}

	logrus.Debugf("gate auth finished", uid)

	u := newUser(uid, conn)
	u.pnum(1001)

	//挂机开始
	timer.AfterFunc(10*time.Second, 0, func(n int) {

		u.pnum(1)
	})

	timer.AfterFunc(2*time.Second, 0, func(n int) {

		req := &pb.HelloRequest{}
		req.ReqMsg = "哈哈。。hello..服务端。。nice to meet you"
		data, err := proto.Marshal(req)
		if err != nil {

			logrus.Errorf("marshal 1001 err = %v", err)
			return
		}
		u.send(pack(1001, data))
	})

}

func (u *user) printInfo(pnum uint16, body []byte) {

	switch pnum {

	case 1002:

		resp := pb.HelloResponse{}
		err := proto.Unmarshal(body, &resp)
		if err != nil {

			logrus.Errorf("unmarshal 1002 err = %v", err)
		}

		logrus.Debugf("1002 return =>", resp.RespMsg)
	}
}

func login(token string) (string, string, uint64) {

	logrus.Debugf("connecting to center")

	d := websocket.Dialer{}
	c, _, err := d.Dial(fmt.Sprintf("ws://%s/ws", CenterAddr), nil) //"ws://127.0.0.1:8002/ws", nil)
	if err != nil {

		logrus.Fatal("dial: ", err)
	}

	defer c.Close()

	_, challenge, err := c.ReadMessage()
	if err != nil {

		logrus.Fatal(err)
	}

	privatekey := crypt.Randomkey()
	clientkey := crypt.DHExchange(privatekey)

	err = c.WriteMessage(websocket.TextMessage, []byte(clientkey.String()))
	if err != nil {

		logrus.Fatal(err)
	}

	_, serverkey, err := c.ReadMessage()
	if err != nil {

		logrus.Fatal(err)
	}

	exchangekey := new(big.Int)
	exchangekey.SetString(string(serverkey), 10)
	secret := crypt.DHSecret(privatekey, exchangekey)

	mac := hmac.New(sha256.New, challenge)
	mac.Write([]byte(secret.String()))
	out := mac.Sum(nil)
	clientmac := base64.StdEncoding.EncodeToString(out)
	err = c.WriteMessage(websocket.TextMessage, []byte(clientmac))
	if err != nil {

		logrus.Fatal(err)
	}

	_, ok, err := c.ReadMessage()
	if string(ok) != "ok" {

		logrus.Fatal("climac err")
	}

	c.WriteMessage(websocket.TextMessage, []byte(token+":test"))

	_, info, err := c.ReadMessage()
	if err != nil {

		logrus.Fatal(err)
	}

	v := make(map[string]interface{})
	err = json.Unmarshal(info, &v)
	if err != nil {

		logrus.Println(err.Error())
	}

	logrus.Debugf("center finished", v)

	uid := v["uid"].(float64)

	return v["addr"].(string), secret.String(), uint64(uid)
}

type user struct {
	uid      uint64
	sendChan chan []byte
	conn     *websocket.Conn
}

func newUser(uid uint64, conn *websocket.Conn) *user {

	u := new(user)
	u.uid = uid
	u.conn = conn
	u.sendChan = make(chan []byte, 8)

	go u.readLoop()
	go u.sendLoop()

	return u
}

func (u *user) send(b []byte) {

	u.sendChan <- b
}

func (u *user) pnum(pnum uint16) {

	u.send(pack(pnum, nil))
}

func (u *user) sendLoop() {

	for {

		select {

		case msg := <-u.sendChan:

			err := u.conn.WriteMessage(websocket.BinaryMessage, msg)
			if err != nil {

				logrus.Printf("send wsconn uid=%d err=%v", u.uid, err)

				return
			}
		}
	}
}

func (u *user) readLoop() {

	for {

		_, resp, err := u.conn.ReadMessage()
		if err != nil {

			logrus.Println("read", u.uid, err)

			return
		}

		pnum, code, body := unpack(resp)
		if code != 0 {

			logrus.Printf("uid=%d pnum=%d code=%d", u.uid, pnum, code)
		} else {

			u.printInfo(pnum, body)
		}
	}
}

func request(conn *websocket.Conn, reqpnum uint16, b []byte) (pnum uint16, code uint16, body []byte, err error) {

	err = conn.WriteMessage(websocket.BinaryMessage, pack(reqpnum, b))
	if err != nil {

		logrus.Println("write", err)

		return
	}

	logrus.Debugf("request, pnum=%d", reqpnum)

	var resp []byte
	_, resp, err = conn.ReadMessage()
	if err != nil {

		logrus.Println("read", err)

		return
	}

	pnum, code, body = unpack(resp)
	logrus.Debugf("response, pnum=%d code=%d", pnum, code)

	return
}

func pack(pnum uint16, b []byte) []byte {

	data := make([]byte, len(b)+2)
	binary.BigEndian.PutUint16(data[0:2], pnum)
	copy(data[2:], b)

	return data
}

func unpack(b []byte) (pnum uint16, code uint16, body []byte) {

	pnum = binary.BigEndian.Uint16(b[0:2])
	code = binary.BigEndian.Uint16(b[2:4])
	body = b[4:]

	return
}
