package rpc

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {

	server = new(Server)
	server.serviceMap = make(map[string]*service)
}

var server *Server

const SESSIONUID = "uid"

// Server 对象
type Server struct {
	name string

	listener   net.Listener
	opts       []grpc.ServerOption
	mux        sync.RWMutex
	serviceMap map[string]*service
	grpcServer *grpc.Server
}

// NewServer 创建Server对象
func NewServer(name string, lis net.Listener, opts []grpc.ServerOption) *Server {

	server.name = name
	server.listener = lis
	server.opts = opts
	return server
}

// Start 启动rpc服务
func (s *Server) Start() {

	grpcServer := grpc.NewServer(s.opts...)
	s.grpcServer = grpcServer

	RegisterGameServer(grpcServer, s)

	log.Printf("rpcserver(%s) listening on %s", s.name, s.listener.Addr().String())
	grpcServer.Serve(s.listener)
}

// Stop 停止rpc服务
func (s *Server) Close() {

	log.Printf("rpcserver(%s) closing", s.name)
	s.grpcServer.Stop()
}

// Call grpc server接口实现
func (s *Server) Call(ctx context.Context, in *GameMsg) (*GameMsg, error) {

	dot := strings.LastIndex(in.ServiceName, ".")
	sname := in.ServiceName[:dot]
	mname := in.ServiceName[dot+1:]
	serv, ok := s.serviceMap[sname]
	if !ok {

		return nil, errors.New("service not found")
	}

	resp, err := serv.handle(mname, in)
	if err != nil {

		err = errors.New(fmt.Sprintf("rpcserver(%s) handle %v", s.name, err))
	}

	return resp, err
}

// Stream grpc server接口实现
func (s *Server) Stream(stream Game_StreamServer) error {

	gameMsg := make(chan *GameMsg, 1)
	quit := make(chan int)

	go func() {

		defer close(quit)

		for {

			in, err := stream.Recv()
			if err == io.EOF {

				return
			}

			if err != nil {

				return
			}

			gameMsg <- in
		}
	}()

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {

		return errors.New("rpc.Stream: stream ctx error")
	}

	var session *Session

	if len(md[SESSIONUID]) > 0 {

		userID, _ := strconv.ParseUint(md[SESSIONUID][0], 10, 64)
		session = new(Session)
		session.Uid = userID
	}

	defer func() {

		close(gameMsg)
	}()

	for {

		select {

		case in := <-gameMsg:

			dot := strings.LastIndex(in.ServiceName, ".")
			sname := in.ServiceName[:dot]
			mname := in.ServiceName[dot+1:]
			serv, ok := s.serviceMap[sname]

			if !ok {

				return errors.New(fmt.Sprintf("rpcserver(%s): service(%s) not found", s.name, sname))
			}

			if in.Session == nil {

				in.Session = session
			}

			resp, err := serv.handle(mname, in)
			if err != nil {

				return errors.New(fmt.Sprintf("rpcserver(%s) handle %v", s.name, err))
			}

			if err := stream.Send(resp); err != nil {

				return errors.New(fmt.Sprintf("rpcserver(%s) streamsend, err=%v", s.name, err))
			}

		case <-quit:

			return nil
		}
	}
}

// RegisterService 注册一个服务
func RegisterService(v interface{}) {

	server.mux.Lock()
	defer server.mux.Unlock()

	if server.serviceMap == nil {

		server.serviceMap = make(map[string]*service)
	}

	s := new(service)
	s.typ = reflect.TypeOf(v)
	s.rcvr = reflect.ValueOf(v)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {

		panic("rpc.Register: no service name for type " + s.typ.String())
	}

	if _, present := server.serviceMap[sname]; present {

		panic("rpc.Register: service already defined " + sname)
	}

	s.name = sname
	s.method = suitableMethods(s.typ)
	server.serviceMap[s.name] = s
}

func suitableMethods(typ reflect.Type) map[string]reflect.Method {

	methods := make(map[string]reflect.Method)
	for m := 0; m < typ.NumMethod(); m++ {

		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name

		if method.PkgPath != "" {

			continue
		}

		if mtype.NumOut() != 1 {

			panic(fmt.Sprintf("rpc.Register: method %s has wrong number of outs: %d", mname, mtype.NumOut()))
		}

		if mtype.NumIn() != 3 {

			panic(fmt.Sprintf("rpc.Register: method %s has wrong number of ins: %d", mname, mtype.NumIn()))
		}

		methods[mname] = method
	}

	return methods
}

type service struct {
	name   string
	rcvr   reflect.Value
	typ    reflect.Type
	method map[string]reflect.Method
}

//处理客户端发送的数据包
func (s *service) handle(methodName string, in *GameMsg) (*GameMsg, error) {

	method, ok := s.method[methodName]
	if !ok {

		return nil, errors.New(fmt.Sprintf("rpc.handle: method(%s) not found", methodName))
	}

	function := method.Func
	rvs := []reflect.Value{s.rcvr, reflect.ValueOf(in.Msg), reflect.ValueOf(in.Session)}
	ret := function.Call(rvs)
	resp := ret[0].Bytes()

	return &GameMsg{ServiceName: in.ServiceName, Msg: resp}, nil
}
