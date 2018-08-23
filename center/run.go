package center

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"enen/common"

	"github.com/laonsx/gamelib/g"
	"github.com/laonsx/gamelib/gofunc"
	"github.com/laonsx/gamelib/log"
	"github.com/laonsx/gamelib/rpc"
	"github.com/laonsx/gamelib/server"
	"github.com/laonsx/gamelib/server/ws"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func Run() {

	serverConfs := make(common.ServiceConf)
	gofunc.LoadJsonConf(gofunc.CONFIGS, "server", &serverConfs)

	serverConf, ok := serverConfs[viper.GetString("center.name")]
	if !ok {

		panic(fmt.Sprintf("server name(%s) not found", viper.GetString("game.name")))
	}

	if viper.GetBool("center.debug") {

		gofunc.Pprof(serverConf.PprofAddr)
	}

	log.InitLogrus(fmt.Sprintf(serverConf.LogDir, viper.GetString("center.name")), viper.GetBool("center.debug"))

	//rpc客户端
	rpcConf := new(common.RpcConf)
	gofunc.LoadJsonConf(gofunc.CONFIGS, "rpc", rpcConf)

	node := make(map[string]string)
	for _, v := range rpcConf.Node {

		if v.Name == viper.GetString("center.name") {

			continue
		}

		node[v.Name] = v.Addr
	}

	var methods [][]string
	for _, v := range rpcConf.Method {

		methods = append(methods, []string{v.Id, v.Name})
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	rpc.InitClient(node, methods, opts)

	//rpc服务
	var serverOpts []grpc.ServerOption
	lis, err := net.Listen("tcp", serverConf.RpcAddr)
	if err != nil {

		panic(err)
	}

	rpcServer := rpc.NewServer(viper.GetString("center.name"), lis, serverOpts)

	go rpcServer.Start()

	conf := &server.Config{}
	conf.Addr = serverConf.WebSocketAddr
	conf.MaxConn = 10000
	gs := ws.NewServer(viper.GetString("center.name"), conf)
	gs.SetHandler(NewLoginServer())

	go gs.Start()

	defer func() {

		gs.Close()
		rpcServer.Close()
		g.Close()
	}()

	handleSignal()
}

func handleSignal() {

	sigstop := syscall.Signal(15)
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, sigstop, syscall.SIGINT, syscall.SIGTERM)

	<-ch
}
