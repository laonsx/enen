package redis

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	poolRedisHelper map[string][]*Redis
	encode          encodeType
	decode          decodeType
)

type encodeType func(data interface{}) ([]byte, error)
type decodeType func(data []byte, v interface{}) error

type Redis struct {
	rate int
	rp   *redis.Pool
}

type RedisConf struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port string `json:"port"`
	Rate int    `json:"rate"`
}

func NewRedisConf(name string, host string, port string, rate int) *RedisConf {

	return &RedisConf{
		Name: name,
		Host: host,
		Port: port,
		Rate: rate,
	}
}

func InitRedis(encodefunc encodeType, decodefunc decodeType, conf ...*RedisConf) {

	poolRedisHelper = make(map[string][]*Redis)
	
	encode = encodefunc
	decode = decodefunc

	rconf := make(map[string][]*RedisConf)
	for _, v := range conf {

		rconf[v.Name] = append(rconf[v.Name], v)
	}

	byrate := func(p1, p2 *RedisConf) bool {

		return p1.Rate < p2.Rate
	}

	for _, value := range rconf {

		ps := &sorter{
			data: value,
			by:   byrate,
		}
		sort.Sort(ps)

		for _, v := range value {

			if len(v.Name) == 0 || len(v.Host) == 0 || len(v.Port) == 0 {

				panic("redis conf error")
			}

			addr := v.Host + ":" + v.Port

			r := new(Redis)

			r.rp = &redis.Pool{
				Dial:    Dial(addr),
				MaxIdle: 64,
			}
			r.rate = v.Rate

			poolRedisHelper[v.Name] = append(poolRedisHelper[v.Name], r)
		}
	}

}

func Dial(addr string) func() (redis.Conn, error) {

	timeout := time.Duration(500) * time.Millisecond

	return func() (redis.Conn, error) {

		return redis.Dial("tcp", addr, redis.DialConnectTimeout(timeout), redis.DialConnectTimeout(timeout), redis.DialConnectTimeout(timeout))
	}
}

type sorter struct {
	data []*RedisConf
	by   func(p1, p2 *RedisConf) bool
}

func (s *sorter) Len() int {

	return len(s.data)
}
func (s *sorter) Swap(i, j int) {

	s.data[i], s.data[j] = s.data[j], s.data[i]
}
func (s *sorter) Less(i, j int) bool {

	return s.by(s.data[i], s.data[j])
}

func UseRedis(name string, id uint64) (*Redis, error) {

	r, ok := poolRedisHelper[name]
	if !ok {

		return nil, errors.New("redis mod err")
	}

	h := hash(id)
	var key int
	for k, v := range r {

		if h >= v.rate {

			key = k
		}
	}

	return r[key], nil
}

func UseRedisByName(name string) (*Redis, error) {

	r, ok := poolRedisHelper[name]
	if !ok {

		return nil, errors.New("redis mod err " + name)
	}

	return r[0], nil
}

func (r *Redis) Get(key string, v interface{}) (err error) {

	var data interface{}
	var conn redis.Conn

	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		data, err = conn.Do("GET", key)
		conn.Close()
		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	if err != nil {

		return err
	}

	if data != nil {

		return Decode(data, v)
	}

	return nil
}

func (r *Redis) Set(key string, data interface{}, expire ...int) (err error) {

	b, e := Encode(data)
	if e != nil {

		return e
	}

	var conn redis.Conn
	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		if len(expire) > 0 {

			_, err = conn.Do("SETEX", key, expire[0], b)
		} else {

			_, err = conn.Do("SET", key, b)
		}
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return
}

func (r *Redis) Del(key string) (err error) {

	var conn redis.Conn
	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		_, err = conn.Do("DEL", key)
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return
}

func (r *Redis) Incrby(key string, n int64) (num int64, err error) {

	var conn redis.Conn
	var ok interface{}
	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		ok, err = conn.Do("INCRBY", key, n)
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return reflect.ValueOf(ok).Int(), err
}

func (r *Redis) Hget(key string, field string, v interface{}) (err error) {

	var conn redis.Conn
	var data interface{}

	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		data, err = conn.Do("HGET", key, field)
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	if err != nil {

		return
	}

	if data != nil {

		return Decode(data, v)
	}

	return nil
}

func (r *Redis) GetRaw(key string) (interface{}, error) {

	var conn redis.Conn
	var data interface{}
	var err error

	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		data, err = conn.Do("GET", key)
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return data, err
}

func (r *Redis) Hdel(key string, args string) error {

	var conn redis.Conn
	var err error

	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		_, err = conn.Do("HDEL", []interface{}{key, args}...)
		conn.Close()

		if err == nil {

			return nil
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return err
}

func (r *Redis) Hgetall(key string) (map[string]interface{}, error) {

	var conn redis.Conn
	var err error

	//var data interface{}
	var data []interface{}
	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		data, err = redis.Values(conn.Do("HGETALL", key))
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	if err != nil {

		return nil, err
	}

	if len(data)%2 != 0 {

		return nil, errors.New("redis: StringMap expects even number of values result")
	}

	m := make(map[string]interface{}, len(data)/2)

	for i := 0; i < len(data); i += 2 {

		key, okKey := data[i].([]byte)
		value, okValue := data[i+1].([]byte)
		if !okKey || !okValue {

			return nil, errors.New("redigo: ScanMap key not a bulk string value")
		}

		m[string(key)] = value
	}

	return m, nil
}

func (r *Redis) HmgetByKey(key string, args string) (interface{}, error) {

	var conn redis.Conn
	var err error
	var data interface{}

	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		data, err = conn.Do("HMGET", []interface{}{key, args}...)
		conn.Close()

		if err == nil {

			return data, nil
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return nil, err
}

func (r *Redis) ToStrings(value interface{}, err error) ([]string, error) {

	return redis.Strings(value, err)
}

func (r *Redis) ToString(value interface{}, err error) (string, error) {

	return redis.String(value, err)
}

func (r *Redis) ToStringMap(value interface{}, err error) (map[string]string, error) {

	return redis.StringMap(value, err)
}

func (r *Redis) HgetallToMap(value interface{}, v interface{}) error {

	return Decode(value, v)
}

func (r *Redis) Hmget(key string, fields []interface{}, v ...interface{}) (err error) {

	fieldsN := len(fields)
	args := make([]interface{}, fieldsN+1)

	args[0] = key
	copy(args[1:], fields)

	var conn redis.Conn
	var data []interface{}

	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		data, err = redis.Values(conn.Do("HMGET", args...))
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	if err != nil {

		return
	}

	dataN := len(data)

	if len(v) != fieldsN {

		return errors.New("hmget params error")
	}

	if fieldsN != dataN {

		return errors.New("hmget error data")
	}

	for i, fv := range data {

		if fv == nil {

			continue
		}

		err = Decode(fv, v[i])
		if err != nil {

			s := fmt.Sprintf("Decode f=%s err=%s", fields[i], err.Error())

			return errors.New(s)
		}
	}

	return nil
}

func (r *Redis) Hset(key string, field string, data interface{}) (err error) {

	b, e := Encode(data)
	if e != nil {

		return e
	}

	var conn redis.Conn
	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		_, err = conn.Do("HSET", key, field, b)
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return
}

func (r *Redis) Hmset(key string, data map[string]interface{}) (err error) {

	var args []interface{}
	args = append(args, key)
	for k, v := range data {

		b, e := Encode(v)
		if e != nil {

			return e
		}

		args = append(args, k, b)
	}

	var conn redis.Conn
	var ok interface{}
	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		ok, err = conn.Do("HMSET", args...)
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	if err != nil {

		return
	}

	if reflect.ValueOf(ok).String() != "OK" {

		return errors.New("hmset err")
	}

	return nil
}

func (r *Redis) Hincrby(key string, field string, n int64) (num int64, err error) {

	var conn redis.Conn
	var ok interface{}
	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		ok, err = conn.Do("HINCRBY", key, field, n)
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	if err != nil {

		return
	}

	return reflect.ValueOf(ok).Int(), err
}

func (r *Redis) Hmincrby(key string, data map[string]int64) (ret map[string]int64, err error) {

	conn := r.rp.Get()
	defer conn.Close()

	var fields []string
	for k, v := range data {

		err = conn.Send("HINCRBY", key, k, v)
		if err != nil {

			return
		}

		fields = append(fields, k)
	}

	err = conn.Flush()
	if err != nil {

		return
	}

	ret = make(map[string]int64)
	for _, f := range fields {

		var i interface{}
		i, err = conn.Receive()
		if err != nil {

			return
		}

		ret[f] = reflect.ValueOf(i).Int()
	}

	if len(ret) != len(data) {

		err = errors.New("hmincrby reply err")
	}

	return
}

func (r *Redis) Hgetraw(key string, field string) (data interface{}, err error) {

	var conn redis.Conn
	for i := 0; i < 2; i++ {

		conn = r.rp.Get()
		data, err = conn.Do("HGET", key, field)
		conn.Close()

		if err == nil {

			break
		}

		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return
}

func (r *Redis) Zadd(key string, id interface{}, score interface{}) error {

	conn := r.rp.Get()
	defer conn.Close()

	_, err := conn.Do("ZADD", key, score, id)

	return err
}

func (r *Redis) Zrem(key string, id interface{}) error {

	conn := r.rp.Get()
	defer conn.Close()

	_, err := conn.Do("ZREM", key, id)

	return err
}

func (r *Redis) Zscore(key string, id interface{}) (int64, error) {

	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("ZSCORE", key, id)
	if err != nil {

		return 0, err
	}

	if data != nil {

		b := data.([]byte)
		return strconv.ParseInt(string(b), 10, 64)
	}

	return 0, nil
}

func (r *Redis) Zincrby(key string, id interface{}, n int) (int64, error) {

	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("ZINCRBY", key, n, id)
	if err != nil {

		return 0, err
	}

	if data != nil {

		b := data.([]byte)
		return strconv.ParseInt(string(b), 10, 64)
	}

	return 0, nil
}

func (r *Redis) Zrank(key string, id interface{}) (int64, error) {

	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("ZRANK", key, id)
	if err != nil {

		return 0, err
	}

	return data.(int64), err
}

func (r *Redis) Zrevrank(key string, id interface{}) (int64, error) {

	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("ZREVRANK", key, id)
	if err != nil {

		return 0, err
	}
	if data == nil {

		return -1, nil
	}

	return data.(int64), err
}

func (r *Redis) Zrevrange(key string, start int, end int) ([][]string, error) {

	conn := r.rp.Get()
	defer conn.Close()

	var args []interface{}
	args = append(args, key)
	args = append(args, start)
	args = append(args, end)
	//if withscores {
	args = append(args, "WITHSCORES")
	//}
	data, err := conn.Do("ZREVRANGE", args...)
	if err != nil {

		return nil, err
	}

	if data != nil {

		var val [][]string
		v := data.([]interface{})
		l := len(v)
		i := 0
		for {

			if i+2 > l {

				break
			}

			k := string(v[i].([]byte))
			score := string(v[i+1].([]byte))
			val = append(val, []string{k, score})
			i += 2
		}

		return val, nil
	}

	return nil, nil
}

func (r *Redis) ZrangeByScore(key string, params ...interface{}) ([]string, error) {

	conn := r.rp.Get()
	defer conn.Close()

	var args []interface{}
	args = append(args, key)
	args = append(args, params...)
	data, err := conn.Do("ZRANGEBYSCORE", args...)
	if err != nil {

		return nil, err
	}

	if data != nil {

		var arr []string
		v := data.([]interface{})
		for _, s := range v {

			arr = append(arr, string(s.([]byte)))
		}

		return arr, nil
	}

	return nil, nil
}

func (r *Redis) ZrevrangeByScore(key string, params ...interface{}) ([]string, error) {

	conn := r.rp.Get()
	defer conn.Close()

	var args []interface{}
	args = append(args, key)
	args = append(args, params...)
	data, err := conn.Do("ZREVRANGEBYSCORE", args...)
	if err != nil {

		return nil, err
	}

	if data != nil {

		var arr []string
		v := data.([]interface{})
		for _, s := range v {

			arr = append(arr, string(s.([]byte)))
		}

		return arr, nil
	}
	return nil, nil
}

func (r *Redis) Zcard(key string) (int64, error) {
	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("ZCARD", key)
	if err != nil {

		return 0, err
	}

	if data == nil {

		return 0, nil
	}

	return data.(int64), err
}

func (r *Redis) Zcount(key string, s, e interface{}) (int64, error) {

	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("ZCOUNT", key, s, e)
	if err != nil {

		return 0, err
	}

	if data == nil {

		return 0, nil
	}

	return data.(int64), err
}

func (r *Redis) Subscribe(channels []string, cb func([]byte)) error {

	conn := r.rp.Get()
	defer conn.Close()

	psc := redis.PubSubConn{conn}
	for _, cl := range channels {

		psc.Subscribe(cl)
	}

	for {

		switch v := psc.Receive().(type) {

		case redis.Message:

			cb(v.Data)

		case redis.Subscription:

			log.Println("[info] subscribe", v.Channel, v.Kind, v.Count)

		case error:

			return v

		}
	}
}

func (r *Redis) Publish(channel string, msg []byte) error {

	conn := r.rp.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", msg)

	return err
}

func (r *Redis) Lpush(key string, v interface{}) error {

	conn := r.rp.Get()
	defer conn.Close()

	b, err := Encode(v)
	if err != nil {

		return err
	}

	_, err = conn.Do("LPUSH", key, b)

	return err
}

func (r *Redis) Rpush(key string, v interface{}) error {

	conn := r.rp.Get()
	defer conn.Close()

	b, err := Encode(v)
	if err != nil {

		return err
	}

	_, err = conn.Do("RPUSH", key, b)

	return err
}

func (r *Redis) Lpop(key string, v interface{}) error {

	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("LPOP", key)
	if data != nil && v != nil {

		return Decode(data, v)
	}

	return err
}
func (r *Redis) Rpop(key string, v interface{}) error {

	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("RPOP", key)
	if data != nil && v != nil {

		return Decode(data, v)
	}

	return err
}

func (r *Redis) Lrange(key string, start interface{}, offset interface{}) ([]interface{}, error) {

	conn := r.rp.Get()
	defer conn.Close()

	var args []interface{}
	args = append(args, key)
	args = append(args, start)
	args = append(args, offset)
	data, err := conn.Do("LRANGE", args...)
	if data != nil {

		return data.([]interface{}), nil
	}

	return nil, err
}

func (r *Redis) Llen(key string) (int64, error) {

	conn := r.rp.Get()
	defer conn.Close()

	data, err := conn.Do("LLEN", key)
	if err != nil {

		return 0, err
	}

	if data == nil {

		return 0, nil
	}

	return data.(int64), err
}

func hash(id uint64) int {

	return int(id % uint64(128))
}

func Encode(data interface{}) (b []byte, err error) {

	switch v := data.(type) {

	case []byte:

		b = data.([]byte)

	case string:

		b = []byte(v)

	case int:

		var dst []byte
		b = strconv.AppendInt(dst, int64(v), 10)

	case int8:

		var dst []byte
		b = strconv.AppendInt(dst, int64(v), 10)

	case int16:

		var dst []byte
		b = strconv.AppendInt(dst, int64(v), 10)

	case int32:

		var dst []byte
		b = strconv.AppendInt(dst, int64(v), 10)

	case int64:

		var dst []byte
		b = strconv.AppendInt(dst, v, 10)

	case uint:

		var dst []byte
		b = strconv.AppendUint(dst, uint64(v), 10)

	case uint8:

		var dst []byte
		b = strconv.AppendUint(dst, uint64(v), 10)

	case uint16:

		var dst []byte
		b = strconv.AppendUint(dst, uint64(v), 10)

	case uint32:

		var dst []byte
		b = strconv.AppendUint(dst, uint64(v), 10)

	case uint64:

		var dst []byte
		b = strconv.AppendUint(dst, v, 10)

	default:

		b, err = encode(data)

		return b, err
	}

	return
}

func Decode(data interface{}, iv interface{}) error {

	b := data.([]byte)

	switch v := iv.(type) {

	case *[]byte:

		*v = b

	case *string:

		*v = string(b)

	case *int:

		s := string(b)
		i, e := strconv.Atoi(s)
		if e != nil {

			return e
		}

		*v = i

	case *int8:

		s := string(b)
		i, e := strconv.Atoi(s)
		if e != nil {

			return e
		}

		*v = int8(i)

	case *int16:

		s := string(b)
		i, e := strconv.Atoi(s)
		if e != nil {

			return e
		}

		*v = int16(i)

	case *int32:

		s := string(b)
		i, e := strconv.Atoi(s)
		if e != nil {

			return e
		}

		*v = int32(i)

	case *int64:

		s := string(b)
		i, e := strconv.ParseInt(s, 10, 64)
		if e != nil {

			return e
		}

		*v = i

	case *uint:

		s := string(b)
		i, e := strconv.ParseUint(s, 10, 64)
		if e != nil {

			return e
		}

		*v = uint(i)

	case *uint8:

		s := string(b)
		i, e := strconv.ParseUint(s, 10, 64)
		if e != nil {

			return e
		}

		*v = uint8(i)

	case *uint16:

		s := string(b)
		i, e := strconv.ParseUint(s, 10, 64)
		if e != nil {

			return e
		}

		*v = uint16(i)

	case *uint32:

		s := string(b)
		i, e := strconv.ParseUint(s, 10, 64)
		if e != nil {

			return e
		}

		*v = uint32(i)

	case *uint64:

		s := string(b)
		i, e := strconv.ParseUint(s, 10, 64)
		if e != nil {

			return e
		}

		*v = i

	default:

		err := decode(b, iv)

		return err
	}

	return nil
}
