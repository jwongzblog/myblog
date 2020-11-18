# 安装软件
参考链接：https://docs.mongodb.com/manual/tutorial/install-mongodb-on-red-hat/

# 部署shard cluster
参考[连接](https://docs.mongodb.com/manual/tutorial/deploy-shard-cluster/)，
上述方法只能部署单机版本，不过我们接下来的部署依赖上述的二进制包。先关闭mongodb服务：
```
systemctl disable mongod

systemctl stop mongod

systemctl status mongod
```

## 申请云主机
- 申请3台云主机，每台主机配3块rssd数据盘（其大小决定了数据盘的吞吐，尽量申请大一点）
```
10.9.134.47
10.9.63.80
10.9.98.185
```

- 格式化xfs文件系统，分别挂载/data1、/data2、/data3

## 配置config_server、mongos、shard_repl_set
以上3个服务集群均采用3节点部署，具体配置路径可以参考以下配置文件，shard_repl_set集群采用大容量数据盘

### configure server
#### 配置信息
```
systemLog:
  destination: file
  logAppend: true
  path: /mongo_conf_svr/mongo_log/mongod.log

# Where and how to store data.
storage:
  dbPath: /mongo_conf_svr/mongo_data
  journal:
    enabled: true
#  engine:
#  wiredTiger:

# how the process runs
processManagement:
  fork: true  # fork and run in background
  pidFilePath: /mongo_conf_svr/mongod.pid  # location of pidfile
  timeZoneInfo: /usr/share/zoneinfo

# network interfaces
net:
  port: 27000
  bindIp: 10.9.134.47  # Enter 0.0.0.0,:: to bind to all IPv4 and IPv6 addresses or, alternatively, use the net.bindIpAll setting.

sharding:
  clusterRole: configsvr
replication:
  replSetName: mongo_config_svr
```
#### 启动进程
每个节点执行：`/usr/bin/mongod -f /mongo_conf_svr/mongo_conf/mongod.conf`

#### 初始化
$mongo --host 10.9.134.47 --port 2700

```
rs.initiate(
  {
    _id: "mongo_config_svr",
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

### shard repl-set cluster
#### 配置信息
每个主机配置三份，分别监听28000、29000、30000，注意区分replSetName
```
systemLog:
  destination: file
  logAppend: true
  path: /data1/mongo_log/mongod.log

# Where and how to store data.
storage:
  dbPath: /data1/mongo_shard_data
  journal:
    enabled: true

    commitIntervalMs: 2

  directoryPerDB: true
  
  syncPeriodSecs: 60
  
  engine: wiredTiger
  
  wiredTiger:
    engineConfig:
      cacheSizeGB: 4
  
      journalCompressor: snappy

    indexConfig:
      prefixCompression: true

    collectionConfig:
      blockCompressor: snappy

# how the process runs
processManagement:
  fork: true  # fork and run in background
  pidFilePath: /data1/mongo_shard_conf/mongod.pid  # location of pidfile
  timeZoneInfo: /usr/share/zoneinfo

# network interfaces
net:
  port: 28000
  bindIp: 10.9.134.47  # Enter 0.0.0.0,:: to bind to all IPv4 and IPv6 addresses or, alternatively, use the net.bindIpAll setting.

sharding:
  clusterRole: shardsvr
replication:
  replSetName: mongo_shard_repl_1

```

#### 每个节点分别执行：
```
/usr/bin/mongod -f /data1/mongo_shard_conf/mongod.conf
/usr/bin/mongod -f /data2/mongo_shard_conf/mongod.conf
/usr/bin/mongod -f /data3/mongo_shard_conf/mongod.conf
```

#### 初始化3个shard repl set
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

### mongos cluster
#### 配置信息
```
systemLog:
  destination: file
  logAppend: true
  path: /mongos_svr/mongo_log/mongod.log

# how the process runs
processManagement:
  fork: true  # fork and run in background
  pidFilePath: /mongos_svr/mongod.pid  # location of pidfile
  timeZoneInfo: /usr/share/zoneinfo

# network interfaces
net:
  port: 26000
  bindIp: 10.9.134.47  # Enter 0.0.0.0,:: to bind to all IPv4 and IPv6 addresses or, alternatively, use the net.bindIpAll setting.

sharding:
  configDB: mongo_config_svr/10.9.134.47:27000,10.9.63.80:27000,10.9.98.185:27000

```
#### 拉起进程
每个节点执行：`/usr/bin/mongos -f /mongos_svr/mongo_conf/mongod.conf`

#### 将shard repl set加入路由
```
$mongo --host 10.9.63.80 --port 26000
$sh.addShard( "mongo_shard_repl_1/10.9.134.47:28000,10.9.63.80:28000,10.9.98.185:28000")
$sh.addShard( "mongo_shard_repl_2/10.9.134.47:29000,10.9.63.80:29000,10.9.98.185:29000")
$sh.addShard( "mongo_shard_repl_3/10.9.134.47:30000,10.9.63.80:28000,10.9.98.185:30000")
```

# sysbench
## 使用雅虎开源的压测工具：YCSB
工具安装：https://github.com/brianfrankcooper/YCSB/tree/master/mongodb

## 使用hash sharding
```
sh.enableSharding("ycsb")
db.usertable.createIndex( { _id: "hashed" } )
sh.shardCollection("ycsb.usertable",{"_id": "hashed"})
```

## 编辑YCSB模板
workloads/workloada

## 执行
./bin/ycsb load mongodb -s -P workloads/workloada -p mongodb.url=mongodb://10.9.134.47:26000,10.9.63.80:26000,10.9.98.185:26000/ycsb?w=0