package gmt

import (
	"fmt"

	"enen/common"
	"enen/gmt/router"
	"github.com/gin-gonic/gin"
	"github.com/laonsx/gamelib/codec"
	"github.com/laonsx/gamelib/g"
	"github.com/laonsx/gamelib/gofunc"
	"github.com/laonsx/gamelib/log"
	"github.com/laonsx/gamelib/redis"
	"github.com/spf13/viper"
)

func Run() {

	serverConfs := make(common.ServiceConf)
	gofunc.LoadJsonConf(gofunc.CONFIGS, "server", &serverConfs)

	conf, ok := serverConfs[viper.GetString("gmt.name")]
	if !ok {

		panic(fmt.Sprintf("server name(%s) not found", viper.GetString("game.name")))
	}

	if viper.GetBool("gmt.debug") {

		gofunc.Pprof(conf.PprofAddr)
	} else {

		gin.SetMode(gin.ReleaseMode)
	}

	log.InitLogrus(conf.Log, viper.GetBool("gmt.debug"))

	//redis
	var redisConf []*redis.RedisConf
	gofunc.LoadJsonConf(gofunc.CONFIGS, "redis", &redisConf)

	redis.InitRedis(codec.MsgPack, codec.UnMsgPack, redisConf...)

	defer func() {

		g.Close()
	}()

	router.Run(conf.HttpAddr)
}
