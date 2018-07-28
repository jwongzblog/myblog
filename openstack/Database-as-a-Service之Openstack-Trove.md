业务上使用过阿里云的ECS或者AWS的EC2，那么大概率上必然使用过他们的RDS，trove的定位其实就是openstack的RDS，trove的架构设计利用了openstack的各个组件的优势来确保[SLA](https://baike.baidu.com/item/SLA/2957862?fr=aladdin)，十分讨巧，因此，也会有其劣势，接下来我会通过一系列的trove架构剖析，尽可能的还原trove本质，为接下来的商业化铺路

# DBaaS的益处
- 分秒级提供数据库服务
- 轻便的确保所有配置文件的一致性
- 自动化运维
- 自动扩容
- 提升敏捷开发，迅速部署提供数据库实例，也能分秒级的删除数据库实例
- 更好的资源利用
- 定义角色权限

# trove适合的业务场景（支持多租户）
- 公有云
- 私有云

# Pike版本中trove支持的系统和数据库类型
## fedora
- mariadb
- mongodb
- mysql
- percona(mysql的一个分支)
- postgresql
- redis
## ubuntu
- cassandra
- couchbase
- couchdb
- db2
- mariadb
- mongodb
- mysql
- percona
- postgresql
- redis
- vertica
- pxc（[Percona XtraDB Cluster](http://www.baidu.com/link?url=ENm62Qjf3J8t3u3EXmmzKZn-w-PyM0jpuR_sZz9Kl9jb0gvNXDbny1EDBH41fu17fOJQf7go4s35fejt8LinxBwDGK_dG7q2GTGtt-ZmJe7)）

# 逐步支持各类数据库集群，截至pike版，trove支持的数据库集群
- cassandra
- galrera(mysql)
- mongodb
- redis
- vertica
# 一张简要的架构图，解释一下我为什么说trove的架构简单粗暴

![image.png](http://upload-images.jianshu.io/upload_images/5945542-483a720902e4b994.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)
看上图trove所处的位置。。。
所有的特性包括租户隔离，高可用，autoscaling，高并发......都依赖openstack子模块的能力，似乎他们也具备这样的能力，但是子模块能力的极限，也就是trove的极限
