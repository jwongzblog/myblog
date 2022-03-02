
# 一、安装软件
参考链接：https://docs.mongodb.com/manual/tutorial/install-mongodb-on-red-hat/



# 二、部署shard cluster

参考连接：https://docs.mongodb.com/manual/tutorial/deploy-shard-cluster/
上述方法只能部署单机版本，不过我们接下来的部署依赖上述的二进制包。先关闭mongodb服务：

```
systemctl disable mongod

systemctl stop mongod

systemctl status mongod
```


### 1.申请云主机

- 申请3台云主机，每台主机配3块rssd数据盘（其大小决定了数据盘的吞吐，尽量申请大一点）

```
10.9.134.47
10.9.63.80
10.9.98.185
```



- 格式化xfs文件系统，分别挂载/data1、/data2、/data3

### 2.配置config_server、mongos、shard_repl_set

以上3个服务集群均采用3节点部署，具体配置路径可以参考以下配置文件，shard_repl_set集群采用大容量数据盘

##### a. configure server

配置信息

```
# config.conf
#
# systemLog:
systemLog:
  #日志级别，0：默认值，包含“info”信息，1~5，即大于0的值均会包含debug信息
  verbosity: 0

  #日志输出目的地，可以指定为“ file”或者“syslog”，如果不指定，则会输出到标准输出
  destination: file

  #当mongod或mongos重启时，如果为true，将日志追加到原来日志文件内容末尾；如果为false，将创建一个新的日志文件
  logAppend: true

  #日志文件路径
  path: /data/config_svr/log/config.log

  #为了防止日志文件过大，此选项可指定rename和reopen
  logRotate: rename

  #ctime表示显示时间戳格式为：Wed Dec 31 18:17:54.811.
  timeStampFormat: iso8601-local

  #quiet模式限制了日志输出的数量。生产环境下不建议使用quiet模式，这会使得跟踪问题变得很困难。
  quiet: false

# storage:
storage:
  #数据库文件存放路径
  dbPath: /data/config_svr/data
  journal:
    #是否开启journal日志持久存储
    enabled: true
    #mongod进程提交journal日志到硬盘的时间间隔，即fsync的间隔，单位是毫秒
    commitIntervalMs: 2

  #是否将不同DB的数据存储在不同的目录中
  directoryPerDB: true

  #mongod使用fsync操作将数据flush到磁盘的时间间隔，默认值为60（单位：秒），强烈建议不要修改此值。
  syncPeriodSecs: 60

  #存储引擎类型，mongodb 3.0之后支持“mmapv1”、“wiredTiger”两种引擎，3.0到3.2版本默认是“mmapv1”，3.2后默认是“wiredTiger”。
  engine: wiredTiger
  wiredTiger:
    engineConfig:

      #wiredTiger缓存工作集（working set）数据最大占用的内存大小，单位：GB。用于限制Mongod对内存的使用量。
      cacheSizeGB: 5

      #journal日志的压缩算法，可选值为“none”、“snappy”、“zlib”。
      journalCompressor: snappy

    indexConfig:
      #是否对索引数据使用“前缀压缩”，可以有效的减少索引数据的内存使用量。默认值为true。
      prefixCompression: true

# processManagement:
processManagement:
  #指定为true是damon模式运行。通常情况下我们都指定为true。
  fork: true

  #mongo进程的pid文件
  pidFilePath: /data/config_svr/mongodb.pid

# net:
net:
  #mongo实例监听的TCP端口。
  port: 21000

  #mongo进程绑定的IP地址，如果要指定多个地址，需要用逗号分开。
  bindIp: 0.0.0.0

  #最大连接数，如果设置的数大于操作系统可接受的连接数，那么不生效。不要设置的值太小，否则在正常操作下也会产生错误。
  maxIncomingConnections: 10000

  #检测写入mongo数据的格式，如果格式错误则阻止写入。
  wireObjectCheck: true

  #是否使用ipv6地址
  ipv6: false

# security:
#security:
  #keyFile: /mongokey/db0/mongokey_0
  #authorization: enabled

# operationProfiling:
operationProfiling:
  slowOpThresholdMs: 100
  mode: slowOp

# replication:
replication:
  #replication操作日志的最大尺寸，单位：MB，一旦mongod创建了oplog文件，此后再次修改oplogSizeMB将不会生效。
  #此值不要设置的太小， 应该足以保存24小时的操作日志，以保证secondary有充足的维护时间。
  oplogSizeMB: 1024

  #“复制集”的名称，复制集中的所有mongd实例都必须有相同的名字，sharding分布式下，不同的sharding应该使用不同的replSetName。
  replSetName: db0_config

  #是否开启readConcern的级别为“majority”，默认为false
  enableMajorityReadConcern: true

# setParameter:
setParameter:
  enableLocalhostAuthBypass: true
  #对mongod/mongos有效；表示当前mongos或者shard与集群中其他shards链接的链接池的最大容量.
  connPoolMaxShardedConnsPerHost: 5000

  #默认值为200，对mongod/mongos有效；同上，表示mongos或者mongod与其他mongod实例之间的连接池的容量，根据host限定。
  connPoolMaxConnsPerHost: 5000

# sharding:
sharding:
   #在sharding集群中，此mongod实例的角色，可选值：configsvr、shardsvr。
   #configsvr表示此实例为config server。shardsvr：此实例为shard（分片）。
   clusterRole: configsvr
# auditLog:
# snmp:
```


每个节点执行：`/usr/bin/mongod -f /root/mongo_conf_svr/mongod.conf`


$mongo --host 10.9.134.47 --port 2700

```
rs.initiate(
{
_id: "db0_config",
configsvr: true,
members: [
{ _id : 0, host : "10.9.134.47:27000" },
{ _id : 1, host : "10.9.63.80:27000" },
{ _id : 2, host : "10.9.98.185:27000" }
]
}
)

rs.status()
```




##### b. shard repl-set cluster

- 配置信息
每个主机配置三份，分别监听28000、29000、30000，注意区分replSetName

```
# data_node.conf
#
# 系统日志:
systemLog:
  #日志级别，0：默认值，包含“info”信息，1~5，即大于0的值均会包含debug信息
  verbosity: 0

  #日志输出目的地，可以指定为“ file”或者“syslog”，如果不指定，则会输出到标准输出
  destination: file

  #当mongod或mongos重启时，如果为true，将日志追加到原来日志文件内容末尾；如果为false，将创建一个新的日志文件
  logAppend: true

  #日志文件路径
  path: /data2/mongod_svr/log/mongod.log

  #为了防止日志文件过大，此选项可指定rename和reopen
  logRotate: rename

  #ctime表示显示时间戳格式为：Wed Dec 31 18:17:54.811.
  timeStampFormat: iso8601-local

  #quiet模式限制了日志输出的数量。生产环境下不建议使用quiet模式，这会使得跟踪问题变得很困难。
  quiet: false

  #为调试打印详细信息，用于支持相关的故障排除
  traceAllExceptions: false

# storage:
storage:
  #数据库文件存放路径
  dbPath: /data2/mongod_svr/data

  journal:
    #是否开启journal日志持久存储
    enabled: true

    #mongod进程提交journal日志到硬盘的时间间隔，即fsync的间隔，单位是毫秒
    commitIntervalMs: 2

  #是否将不同DB的数据存储在不同的目录中
  directoryPerDB: true

  #mongod使用fsync操作将数据flush到磁盘的时间间隔，默认值为60（单位：秒），强烈建议不要修改此值。
  syncPeriodSecs: 60

  #存储引擎类型，mongodb 3.0之后支持“mmapv1”、“wiredTiger”两种引擎，3.0到3.2版本默认是“mmapv1”，3.2后默认是“wiredTiger”。
  engine: wiredTiger

  wiredTiger:
    engineConfig:
      #wiredTiger缓存工作集（working set）数据最大占用的内存大小，单位：GB。用于限制Mongod对内存的使用量。
      cacheSizeGB: 12

      #journal日志的压缩算法，可选值为“none”、“snappy”、“zlib”。
      journalCompressor: snappy

    indexConfig:
      #是否对索引数据使用“前缀压缩”，可以有效的减少索引数据的内存使用量。默认值为true。
      prefixCompression: true

    collectionConfig:
      blockCompressor: snappy

# processManagement:
processManagement:

  #指定为true是damon模式运行。通常情况下我们都指定为true。
  fork: true

  #mongo进程的pid文件
  pidFilePath: /data2/mongod_svr/mongodb.pid

# net:
net:
  #mongo实例监听的TCP端口。
  port: 28017

  #mongo进程绑定的IP地址，如果要指定多个地址，需要用逗号分开。
  bindIp: 0.0.0.0

  #最大连接数，如果设置的数大于操作系统可接受的连接数，那么不生效。不要设置的值太小，否则在正常操作下也会产生错误。
  maxIncomingConnections: 50000

  #检测写入mongo数据的格式，如果格式错误则阻止写入。
  wireObjectCheck: true

  #是否使用ipv6地址
  ipv6: false

# security:
#security:
  #keyFile: /mongokey/db0/mongokey_0
  #authorization: enabled


# operationProfiling:
operationProfiling:
  #数据库profiler判定一个操作是“慢查询”的时间阀值，单位毫秒，mongod将会把慢查询记录到日志中，即使profiler被关闭。
  slowOpThresholdMs: 100

  #数据库profiler级别，操作的性能信息将会被写入日志文件中，可选值：off、slowOp、all。
  mode: slowOp

# replication:
replication:
  #replication操作日志的最大尺寸，单位：MB，一旦mongod创建了oplog文件，此后再次修改oplogSizeMB将不会生效。
  #此值不要设置的太小， 应该足以保存24小时的操作日志，以保证secondary有充足的维护时间。
  oplogSizeMB: 131072

  #“复制集”的名称，复制集中的所有mongd实例都必须有相同的名字，sharding分布式下，不同的sharding应该使用不同的replSetName。
  replSetName: mongo_shard_repl_1

  #是否开启readConcern的级别为“majority”，默认为false
  enableMajorityReadConcern: false

# setParameter:
setParameter:
  #true或者false，默认为true，对mongod/mongos有效；表示是否开启“localhost exception”，
  #对于sharding cluster而言，建议于在mongos上开启，在shard节点的mongod上关闭。
  enableLocalhostAuthBypass: false

  #对mongod/mongos有效；表示当前mongos或者shard与集群中其他shards链接的链接池的最大容量.
  connPoolMaxShardedConnsPerHost: 5000

  #默认值为200，对mongod/mongos有效；同上，表示mongos或者mongod与其他mongod实例之间的连接池的容量，根据host限定。
  connPoolMaxConnsPerHost: 5000

# sharding:
sharding:
   #在sharding集群中，此mongod实例的角色，可选值：configsvr、shardsvr。
   #configsvr表示此实例为config server。shardsvr：此实例为shard（分片）。
   clusterRole: shardsvr

# auditLog:
# snmp:

```


每个节点分别执行：

```
/usr/bin/mongod -f /data1/mongo_shard_conf/mongod.conf
/usr/bin/mongod -f /data2/mongo_shard_conf/mongod.conf
/usr/bin/mongod -f /data3/mongo_shard_conf/mongod.conf
```



初始化3个shard repl set
$mongo --host 10.9.134.47 --port 28000

```
rs.initiate(
{
_id: "mongo_shard_repl_1",
members: [
{ _id : 0, host : "10.9.134.47:28000" },
{ _id : 1, host : "10.9.63.80:28000" },
{ _id : 2, host : "10.9.98.185:28000" }
]
}
)

rs.status()
```



初始化后均为secondary状态，等待片刻，会选出primary节点
$mongo --host 10.9.134.47 --port 29000

```
rs.initiate(
{
_id: "mongo_shard_repl_2",
members: [
{ _id : 0, host : "10.9.134.47:29000" },
{ _id : 1, host : "10.9.63.80:29000" },
{ _id : 2, host : "10.9.98.185:29000" }
]
}
)

rs.status()
```




```
rs.initiate(
{
_id: "mongo_shard_repl_3",
members: [
{ _id : 0, host : "10.9.134.47:30000" },
{ _id : 1, host : "10.9.63.80:30000" },
{ _id : 2, host : "10.9.98.185:30000" }
]
}
)

rs.status()
```




此处如果添加了`configsvr: true`，会导致集群加入路由失败，可以采用删除data目录，重新初始化即可

##### c. mongos cluster

配置信息

```

# route.conf
#
# systemLog:
systemLog:
  #日志级别，0：默认值，包含“info”信息，1~5，即大于0的值均会包含debug信息
  verbosity: 0

  #日志输出目的地，可以指定为“ file”或者“syslog”，如果不指定，则会输出到标准输出
  destination: file

  #当mongod或mongos重启时，如果为true，将日志追加到原来日志文件内容末尾；如果为false，将创建一个新的日志文件
  logAppend: true

  #日志文件路径
  path: /data2/mongos_svr/log/route.log

  #为了防止日志文件过大，此选项可指定rename和reopen
  logRotate: rename

  #ctime表示显示时间戳格式为：Wed Dec 31 18:17:54.811.
  timeStampFormat: iso8601-local

  #quiet模式限制了日志输出的数量。生产环境下不建议使用quiet模式，这会使得跟踪问题变得很困难。
  quiet: false

  #为调试打印详细信息，用于支持相关的故障排除
  traceAllExceptions: false


# processManagement:
processManagement:
  #指定为true是damon模式运行。通常情况下我们都指定为true。
  fork: true

  #mongos进程的pid文件
  pidFilePath: /data2/mongos_svr/mongodb0.pid

# net:
net:
  #mongo实例监听的TCP端口。
  port: 20030

  #mongo进程绑定的IP地址，如果要指定多个地址，需要用逗号分开。
  bindIp: 0.0.0.0

  #最大连接数，如果设置的数大于操作系统可接受的连接数，那么不生效。不要设置的值太小，否则在正常操作下也会产生错误。
  maxIncomingConnections: 50000

  #检测写入mongo数据的格式，如果格式错误则阻止写入。
  wireObjectCheck: true

  #是否使用ipv6地址
  ipv6: false

# security:
#security:
  #keyFile: /mongokey/db0/mongokey_0
  #authorization: disabled

# setParameter:
setParameter:
  #true或者false，默认为true，对mongod/mongos有效；表示是否开启“localhost exception”，
  #对于sharding cluster而言，我们倾向于在mongos上开启，在shard节点的mongod上关闭。
  enableLocalhostAuthBypass: true

  connPoolMaxShardedConnsPerHost: 5000

  #默认值为200，对mongod/mongos有效；同上，表示mongos或者mongod与其他mongod实例之间的连接池的容量，根据host限定。
  connPoolMaxConnsPerHost: 5000
  ShardingTaskExecutorPoolMinSize: 5

  ShardingTaskExecutorPoolRefreshRequirementMS: 600000
  taskExecutorPoolSize: 0

# sharding:
sharding:

   #是否开启sharded collections的自动分裂，仅对mongos有效。默认为true
   #autoSplit: true

   #设定config server的地址列表，每个server地址之间以“,”分割，通常sharded集群中指定1或者3个config server。
   configDB: db0_config/10.65.89.9:21000,10.65.89.8:21000,10.65.89.7:21000

   #sharded集群中每个chunk的大小，单位：MB，默认为64, 太小会导致数据迁移频繁，太大会导致分裂不均匀。
   #chunkSize: 512

#   clusterRole: configsvr
# auditLog:
# snmp

```

拉起进程：
- 每个节点执行：`/usr/bin/mongos -f /mongos_svr/mongo_conf/mongod.conf`

- 将shard repl set加入路由

```
$mongo --host 10.9.63.80 --port 26000
$sh.addShard( "mongo_shard_repl_1/10.9.134.47:28000,10.9.63.80:28000,10.9.98.185:28000")
$sh.addShard( "mongo_shard_repl_2/10.9.134.47:29000,10.9.63.80:29000,10.9.98.185:29000")
$sh.addShard( "mongo_shard_repl_3/10.9.134.47:30000,10.9.63.80:28000,10.9.98.185:30000")
```

