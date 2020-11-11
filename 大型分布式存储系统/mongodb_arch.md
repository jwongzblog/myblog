# 一、部署架构

![image.png](https://github.com/jwongzblog/myblog/blob/master/image/mongodb_arch.png)


# 二、模块作用
- mongos:提供router功能，可以找到key对应的shard server
- mongod:主进程
- shard:采用shard模式部署集群，数据可以均匀的平衡到各个服务器
- replication:为确保数据的高可用，采用一主两从的模式
- chunk:默认每个chunk 64MB，超过64MB会分裂成2个32MB的chunk；删除数据会merge chunk
- balancer:控制chunk的split、merge，集群平衡的时机，最大程度的确保不影响系统的负载

# 三、优化项
## 1.shard的模式：range sharding、hash sharding
#### range sharding

- 优点：让value接近的值存在相同或邻近的chunk，这样可以提高scan的效率
- 缺点：大量的数据插入容易造成写热点

#### hash sharding

- 优点：可以通过pre-split预分配空chunk，大量插入数据时，有效缓解写热点
- 缺点：scan时，可能由于索引的问题造成全表扫，也可能由于存储引擎未命中缓存而需要解压chunk文件造成负载过高。可以尝试关闭存储引擎的压缩功能，适当调大cache size

## 2.index
- mongodb的索引采用b-tree，适合point select，scan没用b+tree效率高。
- 设计并创建合适的索引，可以提高查询效率

## 3.mirror read
我们的mongo版本为v3.6+，在最新的v4.4版本，提供了mirror read。其主要作用是支持在read时，从secondary节点读。需要注意的是，如果使用transaction，需要确保读和写在同一个服务器

## 4.storage engine
默认使用：WiredTiger Storage Engine。By default, WiredTiger uses block compression with the snappy compression library for all collections and prefix compression for all indexes。数据量不大，可以把这两个压缩功能关了