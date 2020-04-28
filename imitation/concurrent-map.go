// 仿写版 concurrent-map
package imap

import (
	"fmt"
	"hash/fnv"
	"sync"
)

const (
	SHARD_NUM = 32
)

type ConcurrentMap []*ConcurrentmapShard

type ConcurrentmapShard struct {
	items map[string]interface{}
	mutex sync.RWMutex
}

func New() ConcurrentMap {
	m := make(ConcurrentMap, SHARD_NUM)
	for i := 0; i < SHARD_NUM; i++ {
		m[i] = &ConcurrentmapShard{
			items:make(map[string]interface{}),
		}
	}
	fmt.Println("len(m): ", len(m))
	return m
}

func (m ConcurrentMap) Get(key string) (interface{}, bool) {
	shard := m.GetShard(key)
	shard.mutex.RLock()
	value, ok := shard.items[key]
	shard.mutex.RUnlock()
	return value, ok
}

func (m ConcurrentMap) Set(key string, value interface{})  {
	shard := m.GetShard(key)
	shard.mutex.Lock()
	shard.items[key] = value
	shard.mutex.Unlock()
}

func (m ConcurrentMap) Del(key string) {
	shard := m.GetShard(key)
	shard.mutex.Lock()
	delete(shard.items, key)
	shard.mutex.Unlock()
}

func (m ConcurrentMap) Length() int {
	total := 0
	for i:=0; i< SHARD_NUM; i++  {
		shard := m[i]
		shard.mutex.RLock()
		total+=len(shard.items)
		shard.mutex.RUnlock()
	}
	return total
}

func (m ConcurrentMap) Has(key string) bool {
	shard := m.GetShard(key)

	shard.mutex.RLock()
	_, ok := shard.items[key]
	shard.mutex.RUnlock()

	return ok
}

func (m ConcurrentMap) Keys () []string {
	wg := &sync.WaitGroup{}
	wg.Add(SHARD_NUM)

	ch := make(chan string, m.Length())
	go func() {
		for i:=0; i<SHARD_NUM; i++ {
			go func(index int) {
				shard := m[i]
				shard.mutex.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.mutex.RUnlock()
				wg.Done()
			}(i)
		}
		close(ch)
		wg.Wait()
	}()

	result := []string{}
	for key := range ch {
		result = append(result, key)
	}
 	return result
}

type KVTuple struct {
	key string
	value interface{}
}
func (m ConcurrentMap) Iter () <-chan KVTuple {
	return m.fanin()
}

func (m ConcurrentMap) fanin() <-chan KVTuple {
	out := make(chan KVTuple)

	wg := &sync.WaitGroup{}
	wg.Add(SHARD_NUM)
	go func() {
		for i := 0; i< SHARD_NUM; i++ {
			go func(index int) {
				shard := m[index]
				for key, value := range shard.items {
					out <- KVTuple{key, value}
				}
				wg.Done()
			}(i)
		}
	}()

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func GetShardIndex(key string) uint32 {
	f := fnv.New32()
	f.Write([]byte(key))
	seed := f.Sum32()

	return uint32(seed) % uint32(SHARD_NUM)
}

func (m ConcurrentMap) GetShard(key string) *ConcurrentmapShard {
	return m[GetShardIndex(key)]
}