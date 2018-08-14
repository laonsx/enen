package rpc

import (
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"gamelib/gofunc"

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
	Name string

	listener   net.Listener
	opts       []grpc.ServerOption
	mux        sync.RWMutex
	serviceMap map[string]*service
	grpcServer *grpc.Server
}

// NewServer 创建Server对象
func NewServer(name string, lis net.Listener, opts []grpc.ServerOption) *Server {

	server.Name = name
	server.listener = lis
	server.opts = opts
	return server
}

// Start 启动rpc服务
func (s *Server) Start() {

	grpcServer := grpc.NewServer(s.opts...)
	s.grpcServer = grpcServer

	RegisterGameServer(grpcServer, s)

	log.Printf("%s rpcserver listening on %s", s.Name, s.listener.Addr().String())
	grpcServer.Serve(s.listener)
}

// Stop 停止rpc服务
func (s *Server) Close() {

	log.Printf("rpcserver closing")
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

		log.Println("rpcserver stream ctx err")

		return errors.New("stream ctx error")
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

				return errors.New("service not found")
			}

			if in.Session == nil {

				in.Session = session
			}

			resp, err := serv.handle(mname, in)
			if err != nil {

				log.Printf("rpcserver handle %v", err)
				return err
			}

			if err := stream.Send(resp); err != nil {

				log.Printf("rpcserver streamsend, err=%s", err.Error())

				return err
			}

		case <-quit:

			return nil
		}
	}
}

// RegisterService 注册一个服务
func RegisterService(v interface{}) error {

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

		s := "rpc.Register: no service name for type " + s.typ.String()
		log.Println(s)

		return errors.New(s)
	}

	if _, present := server.serviceMap[sname]; present {

		return errors.New("rpc: service already defined: " + sname)
	}

	s.name = sname
	s.method = suitableMethods(s.typ)
	server.serviceMap[s.name] = s

	return nil
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

		if mtype.NumOut() != 2 {

			log.Println("method", mname, "has wrong number of outs:", mtype.NumOut())
			
			continue
		}

		if mtype.NumIn() != 3 {

			log.Println("method", mname, "has wrong number of ins:", mtype.NumIn())
			
			continue
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

	defer gofunc.PrintPanic()

	method, ok := s.method[methodName]
	if !ok {

		return nil, errors.New("method not found")
	}

	function := method.Func
	rvs := []reflect.Value{s.rcvr, reflect.ValueOf(in.Msg), reflect.ValueOf(in.Session)}
	ret := function.Call(rvs)
	resp := ret[0].Bytes()
	errInter := ret[1].Interface()

	if errInter != nil {

		return nil, errInter.(error)
	}

	return &GameMsg{ServiceName: in.ServiceName, Msg: resp}, nil
}
