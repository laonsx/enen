package gate

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"enen/common"

	"github.com/laonsx/gamelib/codec"
	"github.com/laonsx/gamelib/g"
	"github.com/laonsx/gamelib/gofunc"
	"github.com/laonsx/gamelib/log"
	"github.com/laonsx/gamelib/redis"
	"github.com/laonsx/gamelib/rpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func Run() {

	serverConfs := make(common.ServiceConf)
	gofunc.LoadJsonConf(gofunc.CONFIGS, "server", &serverConfs)

	serverConf, ok := serverConfs[viper.GetString("gate.name")]
	if !ok {

		panic(fmt.Sprintf("server name(%s) not found", viper.GetString("game.name")))
	}

	if viper.GetBool("gate.debug") {

		gofunc.Pprof(serverConf.PprofAddr)
	}

	log.InitLogrus(serverConf.Log, viper.GetBool("gate.debug"))

	//redis
	var redisConf []*redis.RedisConf
	gofunc.LoadJsonConf(gofunc.CONFIGS, "redis", &redisConf)

	redis.InitRedis(codec.MsgPack, codec.UnMsgPack, redisConf...)

	//rpc客户端
	rpcConf := new(common.RpcConf)
	gofunc.LoadJsonConf(gofunc.CONFIGS, "rpc", rpcConf)

	node := make(map[string]string)
	for _, v := range rpcConf.Node {

		if v.Name == viper.GetString("gate.name") {

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

	rpcServer := rpc.NewServer(viper.GetString("gate.name"), lis, serverOpts)

	go rpcServer.Start()

	gs := NewGateServer(serverConf.WebSocketAddr, serverConf.OriginAllow)

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

func reloadGameServiceConf() {

	rpcConf := new(common.RpcConf)
	gofunc.LoadJsonConf(gofunc.CONFIGS, "rpc", rpcConf)

	var methods [][]string
	for _, v := range rpcConf.Method {

		methods = append(methods, []string{v.Id, v.Name})
	}

	rpc.ReloadMethodConf(methods)
}
