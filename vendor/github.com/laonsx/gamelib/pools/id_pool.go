package pools

import (
	"fmt"
	"sync"
)

//IdPool id池
type IdPool struct {
	sync.RWMutex
	used      map[int64]bool
	startId   int64
	maxUsedId int64
}

//NewIdPool 创建id池
//startId id开始单位
func NewIdPool(startId int64) *IdPool {

	return &IdPool{
		used:      make(map[int64]bool, 0),
		startId:   startId,
		maxUsedId: startId,
	}
}

//Get 获取一个id
func (p *IdPool) Get() int64 {

	p.Lock()
	defer p.Unlock()

	for id := range p.used {

		delete(p.used, id)

		return id
	}

	p.maxUsedId = p.maxUsedId + 1

	return p.maxUsedId
}

//Put 放回一个id
func (p *IdPool) Put(id int64) {

	p.Lock()
	defer p.Unlock()

	if id <= p.startId || id > p.maxUsedId {

		panic(fmt.Errorf("IDPool.Put(%v): invalid value, must be in the range [%v,%v]", id, p.startId, p.maxUsedId))
	}

	if p.used[id] {

		panic(fmt.Errorf("IDPool.Put(%v): can't put value that was already recycled", id))
	}

	p.used[id] = true
}

//MaxUsedCount 同一时刻最多使用id数量
func (p *IdPool) MaxUsedCount() int64 {

	p.RLock()
	defer p.RUnlock()

	return p.maxUsedId - p.startId
}

//CurrUsedCount 当前使用id数量
func (p *IdPool) CurrUsedCount() int64 {

	p.RLock()
	defer p.RUnlock()

	return p.maxUsedId - p.startId - int64(len(p.used))
}
