trove支持数据库实例的备份、增量备份（目前只有mysql、mariaDB、postgresql支持增备）、恢复、实例复制，它的架构设计为了保持厂商之间的中立性和兼容性，采取了一种通用且保守的设计，意味着RPO/RTO要打折扣的

# 运行架构如下所示

![image.png](https://github.com/jwongzblog/myblog/blob/master/openstack/trove-backup-restore.png)

# 实例恢复/复制的具体过程大致如下：
- 首先创建一个备份任务，将通知到Guest Agent（缩写：GA），GA会根据数据库实例类型加载对应的驱动，然后将需要备份的文件一股脑拷贝至swift对象存储中
- 创建一个新数据库实例
- 加载对应的备份副本，拷贝至相应目录
- 完成备份点的恢复

# 一些激进的想法
我们的架构是基于ceph（其他支持快照的存储也行）存储搭建的，那么，如果备份和恢复、复制都基于ceph的快照进行，那么RPO/RTO都是分秒级的
