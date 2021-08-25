mongodb的慢查询问题排查

# 一、mongod慢查询
#### 通过kibana查看mongod慢查询日志，可以通过排序知道慢查询的分布情况
- 大部分情况都是其中一个分片出现慢查询而影响到整个集群
- 如果没有kibana，可以通过物理机负载的监控，判断故障分片的具体位置
- 如果连物理监控也没有，需要逐个查阅mongod的日志，看故障时间段是否出现慢查询

#### mongod没有出现大量慢查询
大概率是mongos出现问题了，查看后面关于mongos慢查询排查的表述

#### 如果mongod log中出现大量慢查询，通常可以查看最初那条慢查询附近30条日志来分析
mongod出现慢查询的原因有：
- 全表扫；关键字："planSummary":"COLLSCAN"、getMore。通常是备份、需要全表扫的业务触发，需要将请求发送至secondary，避免影响primary节点
- 大量删除；大量删除会让db性能下降30%，并且内存的占用也会上升，可以适当调低删除的并发量
- 列表请求；关键字：sort。部分列表请求会发送给mongodb，通常是secondary，但由于配置错误，请求发送到primary节点而引发问题，需及时调整。
- 同一台服务器的其他分片抢占cpu（刷脏页）或者内存资源，导致本进程出现大量慢查询
- 日志中打印了大量的moveChunk日志，由于开启了balance导致，可以通过sh.status()确认是否开启，建议关闭。如果数据分布不均匀，建议创建collection时，将range模式改成hash模式
- 没有建立索引；关键字："planSummary":"COLLSCAN"；务必创建索引


结合grafana、物理机的监控可以验证上述几点。grafana重点查看ops曲线、wiredtiger cache置换速度，如果wiredtiger cache置换速度read指标过快，需要考虑调大cache size；物理机的监控重点查看cpu的负载和磁盘的ioutil，是否达到瓶颈。


# 二、mongos慢查询
#### 如果mongos没有慢查询
- 需要考虑业务与mongos的网络通信（ip+port）问题
- 连接mongos hang住了，通常是连接configure svr出现问题，可以通过重启configure svr解决

#### mongos出现慢查询的问题：
mongod出现慢查询，mongos必然出现大量慢查询，只不过没有mongod的慢查询日志详细。mongod的慢查询会显示是否走索引，lock的具体数量，而mongos没有。
mongos这一层的慢查询通常由以下原因造成：
- mongod脑裂，日志中打印的同一个分片有两个primary节点。可以通过重启分片进程解决
- configure svr不通。通常可以看到日志中打印的configure svr type为unkown，或者connect ip+port timeout。可以通过重启configure svr解决
- mongos未配置taskExecutorPoolSize为0。从4.4版本开始，该值默认为1，严重影响mongos性能，修改配置后重启生效
