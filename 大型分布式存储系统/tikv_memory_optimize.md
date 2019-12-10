# RocksDB参数调优
## 编译benchmark工具
参考：https://github.com/facebook/rocksdb/blob/master/INSTALL.md#supported-platforms

## 让rocksdb/tools/db_bench_tool.cc工具支持strict_capacity_limit、use_clock_cache
- 官方提示clock_cache比LRU_cache算法有更好的提升[link](https://github.com/facebook/rocksdb/wiki/Block-Cache)，
- 默认情况下，strict_capacity_limit=false，如果需要优化控制内存使用，需要使这个数值为true
- 但benchmark工具只支持LRU cache，因此我们需要改点编译脚本让其支持clock cache
```
## centos环境下支持clock cache，还依赖intel的一个库

yum install tbb,tbb-devl
```
```
## SUPPORT_CLOCK_CACHE=1
## 需要处理gflag环境变量：CPATH=/usr/local/include/gflags/ LIBRARY_PATH=${LIBRARY_PATH}:/usr/local/lib/ 
## vim /etc/ld.so.conf
## >> include ld.so.conf.d/*.conf
## >> /usr/local/lib

## 执行：ldconfig

make -j 4 db_bench SUPPORT_CLOCK_CACHE=1
```
- benchmark不支持strict_capacity_limit，因此我们需要改点代码支持这个参数
```
diff --git a/tools/db_bench_tool.cc b/tools/db_bench_tool.cc
index d0540cd..2f69d8c 100644
--- a/tools/db_bench_tool.cc
+++ b/tools/db_bench_tool.cc
@@ -305,6 +305,9 @@ DEFINE_bool(enable_numa, false,
 DEFINE_int64(db_write_buffer_size, rocksdb::Options().db_write_buffer_size,
              "Number of bytes to buffer in all memtables before compacting");
 
+DEFINE_bool(strict_capacity_limit, false,
+            "strict_capacity_limit.");
+
 DEFINE_bool(cost_write_buffer_to_cache, false,
             "The usage of memtable is costed to the block cache");
 
@@ -2442,7 +2445,8 @@ class Benchmark {
     }
     if (FLAGS_use_clock_cache) {
       auto cache = NewClockCache(static_cast<size_t>(capacity),
-                                 FLAGS_cache_numshardbits);
+                                 FLAGS_cache_numshardbits,
+                                FLAGS_strict_capacity_limit);
       if (!cache) {
         fprintf(stderr, "Clock cache not supported.");
         exit(1);
@@ -2451,7 +2455,7 @@ class Benchmark {
     } else {
       return NewLRUCache(
           static_cast<size_t>(capacity), FLAGS_cache_numshardbits,
-          false /*strict_capacity_limit*/, FLAGS_cache_high_pri_pool_ratio);
+          FLAGS_strict_capacity_limit /*strict_capacity_limit*/, FLAGS_cache_high_pri_pool_ratio);
     }
   }
 
@@ -3818,9 +3822,9 @@ class Benchmark {
     if (FLAGS_row_cache_size) {
       if (FLAGS_cache_numshardbits >= 1) {
         options.row_cache =
-            NewLRUCache(FLAGS_row_cache_size, FLAGS_cache_numshardbits);
+            NewLRUCache(FLAGS_row_cache_size, FLAGS_cache_numshardbits, FLAGS_strict_capacity_limit);
       } else {
-        options.row_cache = NewLRUCache(FLAGS_row_cache_size);
+        options.row_cache = NewLRUCache(FLAGS_row_cache_size, -1, FLAGS_strict_capacity_limit);
       }
     }
     if (FLAGS_enable_io_prio) {

```

## 调整benchmark压测参数
32线程read、1线程write，`--use_clock_cache=true`

#### 控制索引占用的内存大小
让索引内存使用block cache，只开启L0的索引

`./db_bench --db=/data/pirlo/facebook/rocksdbtest --num_levels=6 --key_size=20 --prefix_size=20 --keys_per_prefix=0 --value_size=100 --cache_size=2147483648 --cache_numshardbits=6 --compression_type=none --compression_ratio=1 --min_level_to_compress=-1 --disable_seek_compaction=1 --hard_rate_limit=2 --write_buffer_size=134217728 --max_write_buffer_number=2 --level0_file_num_compaction_trigger=8 --target_file_size_base=134217728 --max_bytes_for_level_base=1073741824 --disable_wal=0 --wal_dir=/data/pirlo/facebook/rocksdb_wal/WAL_LOG --sync=0 --verify_checksum=1 --delete_obsolete_files_period_micros=314572800 --max_background_compactions=4 --max_background_flushes=0 --level0_slowdown_writes_trigger=16 --level0_stop_writes_trigger=24 --statistics=0 --stats_per_interval=0 --stats_interval=1048576 --histogram=0 --use_plain_table=0 --open_files=-1 --mmap_read=1 --mmap_write=0 --bloom_bits=10 --bloom_locality=1 --duration=7200 --benchmarks=readwhilewriting --use_existing_db=1 --num=524288000 --threads=16 --use_clock_cache=true --max_write_buffer_number_to_maintain=-1 --bytes_per_sync=10 --strict_capacity_limit=true --optimize_filters_for_hits=false --cache_index_and_filter_blocks=true  --pin_l0_filter_and_index_blocks_in_cache=true`

让索引内存使用block cache，关闭L0的索引
`--pin_l0_filter_and_index_blocks_in_cache=false`

参考：https://github.com/facebook/rocksdb/wiki/Memory-usage-in-RocksDB

#### 索引使用partition index
比cache_index_and_filter_blocks拥有更好像的cache管理

`./db_bench --db=/data/pirlo/facebook/rocksdbtest --num_levels=6 --key_size=20 --prefix_size=20 --keys_per_prefix=0 --value_size=100 --cache_size=2147483648 --cache_numshardbits=6 --compression_type=none --compression_ratio=1 --min_level_to_compress=-1 --disable_seek_compaction=1 --hard_rate_limit=2 --write_buffer_size=134217728 --max_write_buffer_number=2 --level0_file_num_compaction_trigger=8 --target_file_size_base=134217728 --max_bytes_for_level_base=1073741824 --disable_wal=0 --wal_dir=/data/pirlo/facebook/rocksdb_wal/WAL_LOG --sync=0 --verify_checksum=1 --delete_obsolete_files_period_micros=314572800 --max_background_compactions=4 --max_background_flushes=0 --level0_slowdown_writes_trigger=16 --level0_stop_writes_trigger=24 --statistics=0 --stats_per_interval=0 --stats_interval=1048576 --histogram=0 --use_plain_table=0 --open_files=-1 --mmap_read=1 --mmap_write=0 --bloom_bits=10 --bloom_locality=1 --duration=7200 --benchmarks=readwhilewriting --use_existing_db=1 --num=524288000 --threads=16 --use_clock_cache=true --max_write_buffer_number_to_maintain=-1 --bytes_per_sync=10 --strict_capacity_limit=true --optimize_filters_for_hits=false --cache_index_and_filter_blocks=true  --pin_l0_filter_and_index_blocks_in_cache=true  --partition_index=true --partition_index_and_filters=true --use_block_based_filter=false --metadata_block_size=4096 --bloom_bits=0 --pin_top_level_index_and_filter=true --cache_high_pri_pool_ratio=1`

参考：https://github.com/facebook/rocksdb/wiki/Partitioned-Index-Filters

## 测试结果对比
#### 条件
- SST文件为130MB大小
- 通`pmap -X $pid`观察内存占用
- 通过`iostat -xm 1`观察磁盘iops
- benchmark本身输出read/write ops

#### 开了L0层索引
SST内存在20M左右，差不多占文件大小的20%

#### rocksdb压测时，关了L0层的索引
hdd盘的util 100%了，不清缓存的情况下内存一直增长，但可以通过echo 1 > /proc/sys/vm/drop_caches 清理缓存，不过接下来读的ops会在1分30秒内下降一半。除了正在写入的sst和merge的SST内存占用较大，其余的内存占用在5M以内

#### 开启partition_index
HDD磁盘，OPS不变，缓存依然用的非常厉害，但是清空缓存的动作几乎没影响OPS。再者每个SST文件130M，但是索引会占SST文件的10%左右，也就是虽然降低了内存，但是随着SST文件越来越多，分区索引也无法降低内存的使用。新合并的sst文件内存占用几乎和文件大小一致，旧有的SST文件内存占用在10%左右

#### 结论
block cache size为2GB

###### partition index
待文件merge后，内存最低值在17GB左右

###### 非 partition index
- 关闭L0 索引
待文件merge后，内存最低值在5GB左右

- 开启L0 索引
待文件merge后，内存最低值在22GB左右

无论是否严格控制block cache size大小，rockdb内存的消耗大户是索引和merge、compression过程

# Tikv的内存调优
此处省略若干字

综合上述情况，insert QPS下降不到10%，其他QPS不变的情况下，内存下降到原来的1/4-1/5，收益相当可观
