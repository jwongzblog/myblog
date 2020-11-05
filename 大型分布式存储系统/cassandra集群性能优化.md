2008年开源的项目，官方文档至今没有完善。。。花几十美刀买的文档才足够详细，也是醉了。
# 硬件环境优化
- commit logs、data files放在不同的disk上，如果在一起会导致操作倍阻塞。他们的I/O模型也不一样，commit log是append-only write，SSTable是随机存。如果全是SSD，可放一起。如果两种硬盘都有，那么把commit log 放普通硬盘，SSTable放SSD盘
- 由于本身分布式的特性，坏个节点没多大关系，所以采用RAID0读写性能俱佳
- XFS文件系统最佳
- SCSI is better than SATA,SSD is better than SCSI
- 16G RAM以上最佳
- 8个CPU以上最佳
- 网络越快越好
- 关掉swap memor
```
sudo swapoff --all
```
- 确保时钟同步。NTP
- disk readahead设置成512
```
$ sudo blockdev --setra 512 /dev/<device>
 ```
# 配置优化
- num_tokens：默认256，如果有128GB RAM或者8核CPU以上可以设置成1024
- relica strategy：
   SimpleStrategy：默认选项，不智能
  NetworkTopologyStrategy：powerful and robust
- 多数据中心鉴权
- commitlog_sync强制I/O落盘时间
- commitlog_total_space_in_mb指定commit log在内存中占用的虚拟地址空间，32-bit JVM默认32MB，64-bit JAM默认是1024MB
- Java heap：依据你的内存大小，按比例调整
- java Garbage collection设置
```
# GC tuning options
JVM_OPTS="$JVM_OPTS -XX:+UseParNewGC"
JVM_OPTS="$JVM_OPTS -XX:+UseConcMarkSweepGC"
JVM_OPTS="$JVM_OPTS -XX:+CMSParallelRemarkEnabled"
JVM_OPTS="$JVM_OPTS -XX:SurvivorRatio=8"
JVM_OPTS="$JVM_OPTS -XX:MaxTenuringThreshold=1"
JVM_OPTS="$JVM_OPTS -XX:CMSInitiatingOccupancyFraction=75"
JVM_OPTS="$JVM_OPTS -XX:+UseCMSInitiatingOccupancyOnly"
JVM_OPTS="$JVM_OPTS -XX:+UseTLAB"
```
- 备份和恢复：支持创建快照，基于快照的备份，以及备份副本的恢复
- 负载均衡：
```
$ tools/bin/token-generator 3 3
```
 
# 数据库写入、查询调优
主要是要不要压缩，要不要cache的CQL操作

参考数目：[Mastering Apache Cassandra, 2nd Edition](https://www.amazon.com/Mastering-Apache-Cassandra-Nishant-Neeraj/dp/1784392618/)
