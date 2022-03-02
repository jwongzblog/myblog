tidb、mongodb的存储引擎暴露的接口都是key/value语义，索引的设计各有优缺点

# 数据映射
```
TiDB:
3条行存数据：
1, "TiDB", "SQL Layer", 10
2, "TiKV", "KV Engine", 20
3, "PD", "Manager", 30


tikv中的映射：
t10_r1 --> ["TiDB", "SQL Layer", 10]
t10_r2 --> ["TiKV", "KV Engine", 20]
t10_r3 --> ["PD", "Manager", 30]


索引：
t10_i1_10_1 --> null
t10_i1_20_2 --> null
t10_i1_30_3 --> null
```

```
mongodb:
{ "_id" : ObjectId("5e6664112c8519caad5f8596"), "username" : "xiaoming"}

插入的数据会自动生成并补充_id

存储引擎中一张表是一个文件：collection-171.wt，b+tree
mongodb的索引存储于index-149.wt文件中，b+tree。为了避免index分散到各个shard，index和collection建在同一存储引擎中
```

# 差异 
索引生成规则和tidb一致，但存储的位置有差异。

mongodb：列表请求会广播到所有shard，获取数据后排序，返回对应limit的数据以及下一个cursorId信息，这种触发广播的请求性能很差，且容易造成整个集群负载变高。不过索引和数据在同一个存储引擎中，命中索引后无需跨网络再请求其他存储引擎

tidb：索引和数据分属不同的region，甚至服务器。所有命中索引的请求都需要跨网络请求一次数据，单次请求增加了一次网络延迟的时间，不过好在tidb节点是无状态的，可以扩容的方式降低集群的负载，可以让每次请求的平均时延稳定在预期范围内