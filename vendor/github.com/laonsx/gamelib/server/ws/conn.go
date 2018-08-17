package ws

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	defaultSendTimeout = time.Duration(100) * time.Millisecond
)

type Conn struct {
	id            uint64
	ws            *websocket.Conn
	sendChan      chan []byte
	closeChan     chan int
	closeFlag     int32
	closeCallback func(id uint64)
	msgType       int
	err           error
}

func (c *Conn) Ping() func(string) error {

	return func(s string) error {

		c.pong(deadline)

		return nil
	}
}

func (c *Conn) Id() uint64 {

	return c.id
}

func (c *Conn) AsyncSend(b []byte) error {

	if c.IsClosed() {

		return errors.New("conn closed")
	}

	select {

	case c.sendChan <- b:

	case <-time.After(defaultSendTimeout):

		return errors.New("send timeout")
	}

	return nil
}

func (c *Conn) SetMsgType(t int) {

	c.msgType = t
}

func (c *Conn) Send(b []byte) error {

	if c.IsClosed() {

		return errors.New("conn closed")
	}

	return c.ws.WriteMessage(c.msgType, b)
}

func (c *Conn) Read() ([]byte, error) {

	_, b, err := c.ws.ReadMessage()

	return b, err
}

func (c *Conn) Close() error {

	if atomic.CompareAndSwapInt32(&c.closeFlag, 0, 1) {

		close(c.closeChan)
		c.closeCallback(c.id)

		return c.ws.Close()
	}

	return nil
}

func (c *Conn) Error() error {

	return c.err
}

func (c *Conn) IsClosed() bool {

	return atomic.LoadInt32(&c.closeFlag) == 1
}

func (c *Conn) SetReadDeadline(t time.Duration) {

	c.ws.SetReadDeadline(time.Now().Add(t))
}

func (c *Conn) RemoteAddr() net.Addr {

	return c.ws.RemoteAddr()
}

func newConn(id uint64, ws *websocket.Conn, closeCallback func(id uint64)) *Conn {

	c := &Conn{
		id:            id,
		ws:            ws,
		sendChan:      make(chan []byte, 32),
		closeChan:     make(chan int),
		closeCallback: closeCallback,
		msgType:       websocket.BinaryMessage,
	}

	ws.SetReadLimit(32768)

	go c.sendLoop()

	return c
}

func (c *Conn) sendLoop() {

	for {

		select {

		case msg := <-c.sendChan:

			err := c.Send(msg)
			if err != nil {

				c.err = errors.New(fmt.Sprintf("send wsconn id=%d err=%v", c.id, err))
				c.Close()

				return
			}

		case <-c.closeChan:

			return
		}
	}
}

func (c *Conn) pong(deadline time.Duration) {

	c.ws.SetReadDeadline(time.Now().Add(deadline))
	c.sendMsg(websocket.PongMessage, []byte("pong"))
}

func (c *Conn) sendMsg(t int, b []byte) error {

	if c.IsClosed() {

		return errors.New("conn closed")
	}

	return c.ws.WriteMessage(t, b)
}
