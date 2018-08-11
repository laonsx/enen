package game

import (
	"github.com/laonsx/gamelib/codec"
	"github.com/laonsx/gamelib/gofunc"
	"github.com/laonsx/gamelib/redis"
)

func Run() {

	var redisConf []*redis.RedisConf
	gofunc.LoadJsonConf(gofunc.CONFIGS, "redis", &redisConf)
	redis.InitRedis(codec.MsgPack, codec.UnMsgPack, redisConf...)
}
