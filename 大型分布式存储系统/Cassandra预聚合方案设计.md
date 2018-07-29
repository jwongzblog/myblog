在预研Cassandra作为monasca数据存储的方案时，它的分布式架构设计非常适合作为大型数据中心的nosql存储
优点：
* 去中心化的部署架构
* LSM tree数据结构存储，非常适合**大数据量的写入**
* 内置CQL工具，非常接近SQL查询
* 社区非常活跃

但是，如果作为时序数据库的方式使用，其缺点也比较明显：
* **查询速度非常慢**，因为没有索引机制（Spotify 开源的Heroic是在Cassandra之上使用elastic search建立索引，优化查询速度）
* **没有预聚合机制**，像influxDB内置了预聚合功能，可以定时触发计算，这样查询时就可以查看计算好的结果
* monasca的Cassandra驱动实现机制有严重的问题，一方面它没有利用Cassandra自身的聚合计算查询，另一方面它选择把所有查询结果放入内存中分时间片计算，导致不足3万条数据的情况下每次查询都超时了
 
基于上述考虑，纯粹使用目前的方案会有严重的查询性能缺陷，我们需要其他方案来实现
* 考虑是否只使用Cassandra作为TSDB，有kairosDB、Heroic等基于Cassandra的时序数据库方案，目前还不清楚能否解决当前问题，需要继续调研，而且一旦选择这些方案，monasca的数据驱动层就需要重新实现，这样一来也可以把openTSDB考虑进来
* 实现定时器，准时触发计算程序，实现预聚合
 
实现定时器也有两个方案：
* 利用Linux系统自带的cron软件，定时触发我们实现的预聚合脚本，这个方案有两个缺点：一个是部署不够灵活，几乎不存在定制开发的可能；二个是软件服务被停止后，这段时间需要触发的数据就会遗漏，计算出来的数据就不准确
* 自己实现一个通用定时器

之前我有需求实现一个定时功能，经过一番调研没有发现合适的python开源工具，pycron也只是实现了一个类似的代理层，将python程序的定时任务转发至linux cron的功能，所以新设计的定时器要足够**通用**，具体设计思路如下：
![image.png](https://github.com/jwongzblog/myblog/tree/master/image/cassandra-join.png)

如图所示，框架供分为五部分
* produce_timer：定时任务发起者，按协议格式向消息中间件发布一个时钟，让消费者计算完毕后唤醒任务执行
* consume_timer_policy：时钟消息消费者，通过订阅，接收到时钟请求后消费这条消息，并发送给compute_policy，要求计算下个周期的任务什么时候执行
* compute_policy：时钟计算进程，拿到consume_timer_policy发过来的消息后传入自己的计算队列，当时间到达后，把对应要执行的task信息发布到MQ
* MQ：消息中间件
* excute_task：消费compute_policy传过来的task，执行task
消息协议设计：
```

步骤1的协议格式：
queue："timer"
message：
{
    "action": "create_timer ", //edit_timer、delete_timer
    "timer_policy":"$cron_protocol",
    "task_id": "$uuid",
    "contex": "$serialize(contex)"
}
 
步骤4的协议设计：
queue: "$task_id"
message：
{
    "contex": "$serialize(contex)"
}
```
*cron_protocol的设计我们可以参考Linux cron协议

通过这个设计，定时器的计算无需知道生产者具体的任务类型，只需要专注于时间的计算，到时间点发给队列，让task的创建者自行消费


预聚合查询策略：
通常来讲，我们可以把结果按2分钟，30分钟，2小时，一周，一月的长度来切割时间，算出平均值后重新写入一张新表，grafana按照我们指定的策略获取数据并制图，这样一来可以大大节省查询时间
