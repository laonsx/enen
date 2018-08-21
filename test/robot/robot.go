package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"mynet/lib/crypt"
	"mynet/lib/gofunc"
	"mynet/lib/timer"
	"mynet/pb"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

var quit = make(chan os.Signal, 1)

func main() {

	id := flag.String("id", "robot", "-id")
	num := flag.Int("num", 0, "-num")
	flag.Parse()

	if *num == 0 {
		fmt.Println("./robot -id=robot -num=100")
		return
	}

	defer func() {
		log.Println("done")
	}()

	timer.AfterFunc(1*time.Second, *num, func(n int) {
		token := fmt.Sprintf("%s%d", *id, n)
		start(token)
	})

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func login(token string) (string, string, uint64) {
	log.Println("connecting to login")
	d := websocket.Dialer{}
	c, _, err := d.Dial("ws://119.29.6.220:10002/ws", nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	defer c.Close()

	_, challenge, err := c.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}

	privatekey := crypt.Randomkey()
	clientkey := crypt.DHExchange(privatekey)
	err = c.WriteMessage(websocket.TextMessage, []byte(clientkey.String()))
	if err != nil {
		log.Fatal(err)
	}

	_, serverkey, err := c.ReadMessage()
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	_, ok, err := c.ReadMessage()
	if string(ok) != "ok" {
		log.Fatal("climac err")
	}

	c.WriteMessage(websocket.TextMessage, []byte(token+":test"))

	_, info, err := c.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}

	v := make(map[string]interface{})
	err = json.Unmarshal(info, &v)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("login finished", v)

	uid := v["uid"].(float64)
	return v["addr"].(string), secret.String(), uint64(uid)
}

func start(token string) {
	addr, secret, uid := login(token)

	log.Println("connecting to gate", uid)

	d := websocket.Dialer{}
	conn, _, err := d.Dial(fmt.Sprintf("ws://%s/ws", addr), nil)
	if err != nil {
		log.Fatal("dial: ", err)
		return
	}

	req := new(pb.AuthRequest)
	req.Secret = proto.String(secret)
	req.Uid = proto.Uint64(uid)
	b, _ := proto.Marshal(req)
	_, code, _, err := request(conn, 3, b)
	if err != nil || code != 0 {
		return
	}

	log.Println("auth finished", uid)

	u := newUser(uid, conn)
	u.pnum(1001)

	//挂机开始
	timer.AfterFunc(10*time.Second, 0, func(n int) {
		u.pnum(1)
		u.sync()
	})

	timer.AfterFunc(30*time.Second, 0, func(n int) {
		//u.chat(fmt.Sprintf("%d", gofunc.TimeNowUnix()))
	})

	time.Sleep(5 * time.Second)

	timer.AfterFunc(10*time.Second, 0, func(n int) {
		u.pnum(1001)
	})

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
	//go u.ping()

	return u
}

func (u *user) send(b []byte) {
	u.sendChan <- b
}

func (u *user) sync() {
	req := new(pb.SyncRequest)
	req.SceneMon = proto.Int32(1)
	b, _ := proto.Marshal(req)
	u.send(pack(1005, b))
}

func (u *user) chat(msg string) {
	req := new(pb.ChatSendMsgRequest)
	req.Id = proto.Int32(1)
	req.Msg = proto.String(msg)
	b, _ := proto.Marshal(req)
	u.send(pack(1701, b))
}

func (u *user) create() {
	k := gofunc.RandInt(1, 9999)
	name := fmt.Sprintf("%d%d", gofunc.TimeNowUnix(), k)
	req := new(pb.UserCreateRequest)
	req.Name = proto.String(name)
	req.Avatar = proto.String("boy#aboy1_head_icon_png#aboy1_head,aboy1_body,boy1_foot")
	req.Code = proto.String("1234")

	b, _ := proto.Marshal(req)
	u.send(pack(1011, b))
}

func (u *user) pnum(pnum uint16) {
	u.send(pack(pnum, nil))
}

func (u *user) ping() {
	for {
		u.send(pack(1, nil))
		time.Sleep(10 * time.Second)
	}
}

func (u *user) readLoop() {
	for {
		_, resp, err := u.conn.ReadMessage()
		if err != nil {
			log.Println("read", u.uid, err)
			return
		}
		pnum, code, body := unpack(resp)
		if code != 0 {
			log.Printf("uid=%d pnum=%d code=%d", u.uid, pnum, code)
		} else {
			u.printInfo(pnum, body)
		}
	}
}

func (u *user) sendLoop() {
	for {
		select {
		case msg := <-u.sendChan:
			err := u.conn.WriteMessage(websocket.BinaryMessage, msg)
			if err != nil {
				log.Printf("send wsconn uid=%d err=%v", u.uid, err)
				return
			}
		}
	}
}

func (u *user) printInfo(pnum uint16, body []byte) {
	switch pnum {
	// case 900:
	// 	resp := &pb.NotifyChatMsg{}
	// 	err := proto.Unmarshal(body, resp)
	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}
	// 	log.Println("chat", resp)
	case 1002:
		userInfo := &pb.UserInfoResponse{}
		err := proto.Unmarshal(body, userInfo)
		if err != nil {
			log.Println(err)
			return
		}
		if userInfo.GetState() == 0 {
			u.create()
		}
	case 1012:
		resp := &pb.UserCreateResponse{}
		err := proto.Unmarshal(body, resp)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("create ok", resp.Data.GetUid())
	}
}

func request(conn *websocket.Conn, reqpnum uint16, b []byte) (pnum uint16, code uint16, body []byte, err error) {
	err = conn.WriteMessage(websocket.BinaryMessage, pack(reqpnum, b))
	if err != nil {
		log.Println("write", err)
		return
	}

	log.Printf("request, pnum=%d", reqpnum)

	var resp []byte
	_, resp, err = conn.ReadMessage()
	if err != nil {
		log.Println("read", err)
		return
	}
	pnum, code, body = unpack(resp)
	log.Printf("response, pnum=%d code=%d", pnum, code)
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
