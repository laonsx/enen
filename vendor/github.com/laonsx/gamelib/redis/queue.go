package redis

import (
	"log"
	"time"

	"github.com/laonsx/gamelib/g"
)

//RegisterQueueHandler 注册一个队列处理方法
func RegisterQueueHandler(id string, handler func([]byte)) {

	qstart(id, handler)
}

//QPush 添加数据到队列
func QPush(id string, v []byte) error {

	r, err := UseRedisByName("queue")
	if err != nil {

		return err
	}

	key := "queue-" + id
	err = r.Rpush(key, v)

	return err
}

func qstart(id string, callback func([]byte)) {

	r, err := UseRedisByName("queue")
	if err != nil {

		log.Panicf("redis [queue] err=%s", err.Error())
	}

	key := "queue-" + id

	log.Printf("queue[%s] starting", id)

	g.Go(func() {

		defer log.Printf("queue[%s] exiting", id)

		for {

			select {

			case <-g.Quit():

				return

			default:

				var v []byte
				if err := r.Lpop(key, &v); err != nil {

					time.Sleep(time.Duration(1) * time.Second)
					r, err = UseRedis("queue", 0)
					if err != nil {

						log.Println("redis queue connection refused")
					}

					continue
				}

				if len(v) == 0 {

					time.Sleep(time.Duration(1) * time.Second)

					continue
				}

				callback(v)
			}
		}
	})
}
