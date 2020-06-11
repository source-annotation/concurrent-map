# concurrent-map 阅读版

源码阅读版 (带着问题去读源码) 。 

[中文文档](README-zh.md) | [Document](README-en.md) 

# Sched   

- [ ] [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines) 
- [x] fanIn/fanOut   
- [x] crypto hash/ non-crypto hash    
- [ ] range-sharding vs hash-sharding  
- [ ] Q3 存疑部分解答 
- [x] 自己实现一遍 concurrent-map 

# Q&A

## #0 为什么会出现这个项目 ?  

go 有原生的 map 和 sync.Map，为什么还要有 concurrent-map 这个项目，而且还有这么多 star？  

首先，在 Go 1.9 以下是没有 sync.Map 的。而原生的 map 在并发读写的情况下会 panic，所以开发者在使用 map 时一般都会给 map 加一个锁以保证 map 是 synchronize access 的。 于是作者提供了一个支持并发访问的 map 的解决方案，同时还做了分片以提高性能。 

Go 1.9 以后 Go 官方支持了并发读写的 map，就是 sync.Map。

## #1 有了 sync.Map，concurrent-map 还有用的必要吗？  

可以看看：
1. concurrent-map 的 readme  
3. sync.Map 不支持取 map 长度给开发者带来的困扰： https://github.com/golang/go/issues/20680 和 http://xiaorui.cc/archives/4972 
4. [深度解密Go语言之sync.map](https://mp.weixin.qq.com/s/mXOU8TElP8bbqaybRKN8eA)   

结论是：sync.Map 一般够用了，但 sync.Map 适用于读多写少场景。在多写的场景中， sync.Map 锁冲突变多，而且 sync.Map 没有做 shard(原因见：https://github.com/golang/go/issues/20360)。 所以这样看来，用于"写多"场景的 map，用 concurrent-map 会更好一些。 

## #2 项目里有个 fanin 函数，那么什么是 fan in /fan out ? 
fan in / fan out 常用于流水线模型，流水线模型是一个数据由 upstream 流向 downstream 的过程。

上游生产者把数据塞到下一 stage，而下一个 stage 有多个 consumer 共同处理数据。 数据像一个扇形流出(扇出 : fan out)。

假设当前 stage 由多个 consumer 共同处理数据，而下一个 stage 则只有 1 个 consumer 集中处理数据。 那么此时当前 stage 数据流向下一 stage 的操作就是 fan in，数据流像扇形一样进入到下一个 stage。

## #3 为了降低锁冲突，可以把 map 切成 1000000000 个 shard 吗？ `TODO`

我们知道，锁粒度越小，锁冲突几率越小。比如在 MySQL 中发生锁冲突几率 ： 行锁 <  表锁。 所以理论上我们增加 shard 数量就可以降低锁冲突。 那么我们把 shard 数量改为 1000000000 个会有什么问题吗？ 我们每个 key 一个 shard 会有什么问题吗？

其实这样做理论上确实会把锁冲突几率降到最低。 

最开始我想的是：如果把 32 个 shard 改为 len(key) 个 shard ：空间复杂度由 O(1) 变为 O(n) 整体上升了一个量级，但时间上的提升(锁等待时间减少) 效果并没有想象的那么理想。总的成本其实是增加了不少的。  

但是看了 [issue/20360](https://github.com/golang/go/issues/20360) 中提到的 : "at the moment all newly-added keys are guarded by a single Mutex" ，这下有点不敢确定。所以存疑。 标记个 TODO 后面再研究研究。   

## #4 `Cryptographic Hash` vs `Non-Cryptographic Hash`

concurrment-map 里用到的 fnv32 是一个 non-cryptographic hash 算法，而且是直接从`$GOSRC/hash/fnv` 里的 fnv32 抄过来的 。那什么是 cryptographic hash，什么又是 non-cryptographic hash 呢？  根据维基百科：
> A cryptographic hash function (CHF) is a hash function that is suitable for use in cryptography.
 
一篇好文章 : [Cryptographic and Non-Cryptographic Hash Functions](https://dadario.com.br/cryptographic-and-non-cryptographic-hash-functions/) 

一般 cryptographic hash 就是用于加密比如 sha1 等。而 non-cryptographic hash 多用于哈希映射之类的非加密场景，比如 fnv, murmurhash 等。 
## #5 `range-sharding` vs `hashed(key)-sharding`  

concurrent-map 用的是 hash-based sharding(或 key-based sharding)，而一般还有一种 range-based sharding。

可以读一读 :  

1. [Understanding Database Sharding](https://www.digitalocean.com/community/tutorials/understanding-database-sharding) 
2. [MongoDB hashed-sharding](https://docs.mongodb.com/manual/core/hashed-sharding/) / [MongoDB ranged sharding](https://docs.mongodb.com/manual/core/ranged-sharding/)

