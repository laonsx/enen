package server

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type Handler interface {
	Open(c Conn)
	Close(c Conn)
}

type GateServer interface {
	SetMaxConn(n int)
	Start()
	Close()
	Count() int
	SetHandler(handler Handler)
}

var (
	WS_MSG_STRING = websocket.TextMessage
	WS_MSG_BINARY = websocket.BinaryMessage
)

type Conn interface {
	Id() uint64
	AsyncSend(b []byte) error
	Send(b []byte) error
	Read() (b []byte, err error)
	Close() error
	SetMsgType(t int)
	SetReadDeadline(t time.Duration)
	RemoteAddr() net.Addr
	Error() error
}

type Config struct {
	Addr        string
	MaxConn     int
	OriginAllow string
}
