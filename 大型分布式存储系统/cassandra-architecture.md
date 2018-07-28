Apache Cassandra是一个开源的、分布式、无中心、弹性可扩展、高可用、容错、一致性可调、面向列的数据库，它基于Amazon Dynamo的分布式设计和Google BigTable的数据模型。
#分布式无中心
- 可以在多节点，多机架（有关于机架的数据结构），多数据中心部署
- 每个节点是对等的（peer to peer的模式设计），去中心化，不会存在单点失效。相反，MongoDB采用的是主从设计，主节点坏了，整个数据库无法继续正常运行
- 通过gossip协议来维护节点的死活
![image.png](http://upload-images.jianshu.io/upload_images/5945542-da58ad1af0dbc6a4.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)
#弹性扩容
cassandra增加或者缩减节点非常方便，无需大幅修改整个集群的配置，无需重启进程
# 高可用与容错
剩余的节点可以扛到节点恢复，跨数据中心部署还可以应对不可抗拒因素
#可调节一致性
Cassandra的一致性设计是遵照CAP理论（一致性、可用性、分区耐受性），选取的是最终一致性策略，即强化AP。Partition tolerance是指集群里，由于网络故障等，导致被分成多个分区，依然可以提供服务，并且节点恢复后，依然能达到最终一致
- 严格一致性：每次写入要求所有读操作都是返回最新的写结果，要做到这一点需要加一个全局锁来确保读写顺序一致性，这样会导致写入性能非常差
- 因果一致性：写入操作被顺序读出，确保同一个被操作的对象因果关系是能被区分开来的
- 弱一致性（最终一致性）：所有更新将传播到整个分布式系统的每个角落，但需要时间。最终，所有副本都会是一致性的。
![image.png](http://upload-images.jianshu.io/upload_images/5945542-e0d7fd40e5de7c7c.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)
#面向列
传统的关系型数据库的存储是行式的设计，比如mysql，必须先设计数据表结构，如果某个字段没有数据，但存储依然会预留这个位置的空间。而Cassandra的数据存储结构是多维哈希表，允许随时随地增加或减少字段
#LSM
传统的关系型数据库的存储数据结构是B tree，而Cassandra这类对写入性能有要求的选择了LSM数据结构存储
#snitches/告密者们
snitch的工作是告诉决策者，读写操作应该落到哪个节点
#Rings/环
把节点连成一个环状，数据切片后落到哪个节点，通过ring做hash计算来决定
![image.png](http://upload-images.jianshu.io/upload_images/5945542-2d74f629081cf042.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

#partitioners
这个模块是用来计算数据分发到那个节点的
#replication strategy
比如配置3副本，那么会有三个节点会拥有一份相同的拷贝数据，正如上面所提到的，Cassandra是最终一致性的
```
一致性：
一致性意味着一个事务不会让数据库进入不合法状态，不会违反完整性约束。一致性是关系型数据库里的事务的关键方面，是ACID（原子性，一致性，隔离线，持久性）属性之一。
一致性程度如下衡量：
N=存放数据副本的节点数
W=在写操作成功返回之前必须确认写入成功的副本数
R=在读操作访问数据对象时，需要获得的最少副本数
W+R>N=强一致性
W+R<=N=最终一致性
```

#副本同步机制
反熵（anti-entropy）机制借鉴了Amazon Dynamo的同步模型，运行原理是在主压紧期间，会与邻节点进行会话，交换merkle tree，如果两个节点的树不匹配，那么必须协商（修复），确保二者是最新数据

#query and coordinator node/协调者
每个被查询的节点既是协调者，可以决定query的动作落到哪个副本
![image.png](http://upload-images.jianshu.io/upload_images/5945542-368fc04b34bb2b90.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)
#Memtables,SSTables,commit logs
写操作会直接写入至commit log，才被标记写成功了，commit log不会被client直接读，只有节点恢复时，需要恢复数据的数据才从commit log读取。写操作一般被记录在memtable，当达到一定的阈值时，才被flush至SSTable。此时commit log响应数据被标记成0，当达到阈值时清理掉。SSTable借鉴了google bigtable，它是压缩的数据，不再被应用修改直至被merged（压紧操作时，SSTable的键值合并，列被组合，丢弃  tombstone/假删除，创建新索引），后续的读（read）操作会结合SSTable和memtable的数据给出结果
![image.png](http://upload-images.jianshu.io/upload_images/5945542-28de14a50c749f59.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

#hinted handoff/提示移交
出现故障时，部分节点无法响应，这个时候Cassandra会创建一个提示备忘，即要求节点恢复时通知请求者，那时，请求者会重新发送write opration。此机制借鉴至Amazon Dynamo
#bloom filter
把数据集映射为 位数组，将大数据集计算出摘要，存于内存中，当访问磁盘是，先去bloom filter查看元素是否存在，再进一步判断是否访问磁盘
#lightweight Transaction based on Paxos
2.0开始提供了基于paxos原理的轻量级事务，提供行级别的一致性
#分阶段事件驱动架构/Staged Event-Driven Architecture (SEDA)
这个理念是一种高并发互联网服务设计的通用架构，即把一个工作从线程开始，完成一个阶段后再移交给另一个线程，如此往复，但并不是当前线程来决定是否移交。阶段是任务的基本单位，一个操作内部可能会有不同阶段之间的状态迁移，因为各阶段由不同的线程池处理。所有这些东西由控制器负责调度和线程分配。Cassandra内部好几个服务是采用这个机制实现高并发的：
- Read (local reads)
- Mutation (local writes)
- Gossip
- Request/response (interactions with other nodes)
- Anti-entropy (nodetool repair)
- Read repair
- Migration (making schema changes)
- Hinted handoff

参考书目：[Cassandra: The Definitive Guide, 2nd Edition](http://shop.oreilly.com/product/0636920043041.do)
