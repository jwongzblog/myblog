# TiDB各组件内存分配及释放
目前TIKV内存无法下降，造成收费较高，需要想办法降下去

## pingcap提供的内存分析工具，容器中无法执行：

### tidb 的内存收集方法
go tool pprof http://127.0.0.1:10080/debug/pprof/heap --收集内存的 profile，然后会在 $HOME/pprof/ 目录下面生成类似这样 pprof......pb.gz 的文件

### tikv 的内存收集方法
```
perf record -e malloc -F 99 -p $1 -g -- sleep 10
perf script > out.perf
/opt/FlameGraph/stackcollapse-perf.pl out.perf > out.folded
/opt/FlameGraph/flamegraph.pl  --colors=mem out.folded > mem.svg

https://github.com/pingcap/tidb-inspect-tools/tree/master/tracing_tools/perf
```

## tikv的测试场景
```
TiDB v2.1.4
监控显示tikv有2T数据，tikv内存一直稳定在35GB
删除database，tikv数据降至60GB，tikv内存稳定在25GB
重启tikv容器，内存稳定在2GB左右
```

```
TiDB 3.0.5
停止导入数据后，内存未下降，4个小时后，逐个重启tikv节点，每个节点的内存最终稳定在2GB左右
```

## tikv的内存模型
TiKV 的内存使用是跟 MySQL 等数据库一样的淘汰机制，就是设置了内存阈值(block-cache-siz)，在这个阈值内 TiKV 内存用多少申请多少，用完也不会释放

因为tikv是基于rocksDB再封装了一层，所以可以根据rocksDB的内存使用机制来解释这个现象：
```
https://github.com/facebook/rocksdb/wiki/Block-Cache
https://github.com/facebook/rocksdb/wiki/Memory-usage-in-RocksDB#block-cache
```

rocksDB的内存消费大户是BlockCache、filter、index。

其中BlockCache缓存了新写入的数据，采用了LRU（淘汰最少使用）算法，所以一直占用了cache，我们的配置是block-cache-size = 2GB,但是一旦有数据需要插入block cache，rocksDB的默认参数strict_capacity_limit=false(default)会导致跳过2GB的限制，直至OOM，所以只有重启，才能把这个cache清空。

另外两个内存消费大户：filter、index，由于数据清理，会慢慢变小，官方举得例子是blockcache = 10GB，但是rocksDB使用了15GB。

strict_capacity_limit: In rare case, block cache size can go larger than its capacity. This is when ongoing reads or iterations over DB pin blocks in block cache, and the total size of pinned blocks exceeds the capacity. If there are further reads which try to insert blocks into block cache, if strict_capacity_limit=false(default), the cache will fail to respect its capacity limit and allow the insertion. This can create undesired OOM error that crashes the DB if the host don't have enough memory. Setting the option to true will reject further insertion to the cache and fail the read or iteration. The option works on per-shard basis, means it is possible one shard is rejecting insert when it is full, while another shard still have extra unpinned space.

如果严格的按照这个容量来限制，会导致fail the read or iteration


目前只能通过重启的方式来清理blockcache