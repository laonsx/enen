package game

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"enen/common"
	"enen/game/service"

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

	conf, ok := serverConfs[viper.GetString("game.name")]
	if !ok {

		panic(fmt.Sprintf("server name(%s) not found", viper.GetString("game.name")))
	}

	if viper.GetBool("game.debug") {

		gofunc.Pprof(conf.PprofAddr)
	}

	log.InitLogrus(fmt.Sprintf(conf.LogDir, viper.GetString("game.name")), viper.GetBool("game.debug"))

	//redis
	var redisConf []*redis.RedisConf
	gofunc.LoadJsonConf(gofunc.CONFIGS, "redis", &redisConf)

	redis.InitRedis(codec.MsgPack, codec.UnMsgPack, redisConf...)

	//rpc客户端
	rpcConf := new(common.RpcConf)
	gofunc.LoadJsonConf(gofunc.CONFIGS, "rpc", rpcConf)

	node := make(map[string]string)
	for _, v := range rpcConf.Node {

		if v.Name == viper.GetString("game.name") {

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
	lis, err := net.Listen("tcp", conf.RpcAddr)
	if err != nil {

		panic(err)
	}

	rpcServer := rpc.NewServer(viper.GetString("game.name"), lis, serverOpts)

	go rpcServer.Start()

	service.Init()

	defer func() {

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
