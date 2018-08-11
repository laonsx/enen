package redis

import "fmt"

//IRank 接口
type IRank interface {
	Key() string
	Name() string
}

//Rank 排名结构
type Rank struct {
	IRank
}

//NewRank 初始化一个rank
func NewRank(ir IRank) *Rank {

	return &Rank{ir}
}

//GetName 排名榜名
func (rank *Rank) GetName() string {

	return rank.Name()
}

//GetKey 排名榜key
// 在调用rank中方法时，必须首先调用这个函数获取一个key
// key := rank.Key(20160621)
// rank.SetRankScore(100)
func (rank *Rank) GetKey(k ...interface{}) string {

	var key = rank.Key()
	for _, v := range k {

		key += fmt.Sprintf("%v", v)
	}

	return MakeKey("rank", key)
}

//SetRankScore 设置score
//score float64  -9007199254740992 - 9007199254740992
func (rank *Rank) SetRankScore(key string, id interface{}, score interface{}) error {

	r, err := UseRedisByName("rank")
	if err != nil {

		return err
	}

	return r.Zadd(key, id, score)
}

//IncrbyRankScore 增加排行榜score
func (rank *Rank) IncrbyRankScore(key string, id interface{}, score int) (int64, error) {

	r, err := UseRedisByName("rank")
	if err != nil {

		return 0, err
	}

	return r.Zincrby(key, id, score)
}

//GetRankScore 获取score
func (rank *Rank) GetRankScore(key string, id interface{}) (int64, error) {

	r, err := UseRedisByName("rank")
	if err != nil {

		return 0, err
	}

	return r.Zscore(key, id)
}

//GetRankByPage 按页获取排名 每页num条数据
func (rank *Rank) GetRankByPage(key string, page int, num int) ([][]string, error) {

	r, err := UseRedisByName("rank")
	if err != nil {

		return nil, err
	}

	start := (page - 1) * num
	end := start + num - 1

	return r.Zrevrange(key, start, end)
}

//GetRank 获取排名
//返回0，nil：无排名信息
func (rank *Rank) GetRank(key string, id interface{}) (int64, error) {

	r, err := UseRedisByName("rank")
	if err != nil {

		return 0, err
	}

	n, err := r.Zrevrank(key, id)
	if err == nil && n >= 0 {

		n++
	}
	if n == -1 {

		return 0, nil
	}

	return n, err
}

//Del 删除排名
func (rank *Rank) Del(key string) error {

	r, err := UseRedisByName("rank")
	if err != nil {

		return err
	}

	return r.Del(key)
}

//RemRank 删除成员排名
func (rank *Rank) RemRankScore(key string, id uint64) error {

	r, err := UseRedisByName("rank")
	if err != nil {

		return err
	}

	return r.Zrem(key, id)
}

//RangeByScore 获取积分范围内的成员
func (rank *Rank) RangeByScore(key string, s interface{}, e interface{}) ([]string, error) {

	r, err := UseRedisByName("rank")
	if err != nil {

		return nil, err
	}

	return r.ZrevrangeByScore(key, s, e)
}

func (rank *Rank) RangeByRank(key string, s int, e int) ([][]string, error) {

	r, err := UseRedisByName("rank")
	if err != nil {

		return nil, err
	}

	return r.Zrevrange(key, s, e)
}

//SetCachePageRank 按页缓存数据
func (rank *Rank) SetCachePageRank(key string, page int, v interface{}) error {

	r, err := UseRedisByName("rank")
	if err != nil {

		return err
	}

	key = MakeKey(key+"cache:page", page)

	return r.Set(key, v)
}

//GetCachePageRank 获取缓存数据
func (rank *Rank) GetCachePageRank(key string, page int, v interface{}) error {

	r, err := UseRedisByName("rank")
	if err != nil {

		return err
	}

	key = MakeKey(key+"cache:page", page)

	return r.Get(key, v)
}

//GetCacheRank 获取成员缓存排名
func (rank *Rank) GetCacheRank(key string, id uint64, v interface{}) error {

	r, err := UseRedisByName("rank")
	if err != nil {

		return err
	}

	key = MakeKey(key+"cache:rank", id)

	return r.Get(key, v)
}

//SetCacheRank 设置成员缓存排名
func (rank *Rank) SetCacheRank(key string, id uint64, v interface{}) error {

	r, err := UseRedisByName("rank")
	if err != nil {

		return err
	}

	key = MakeKey(key+"cache:rank", id)

	return r.Set(key, v)
}

func MakeKey(key string, v interface{}) string {

	return fmt.Sprintf("%s:%v", key, v)
}
