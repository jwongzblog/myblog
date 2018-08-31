# Redis
### redis自身支持replication、cluster模式
trove redis也实现了上面两种模式
- replication：https://specs.openstack.org/openstack/trove-specs/specs/liberty/redis-replication.html
- cluster：https://specs.openstack.org/openstack/trove-specs/specs/liberty/redis-cluster.html

那么我们产品形态可以自由组合
- 单机模式：纯缓存场景，不保证数据的高可用性
- 双机热备：主从切换型
- 集群模式：节点任意扩展，满足客户高负载

### trove redis备份与恢复
- https://specs.openstack.org/openstack/trove-specs/specs/liberty/redis-backup-restore.html
- 利用trove-schedule可以实现定时备份，保证数据安全

### redis-guestagent实现的接口有：
```
configuration_manager()
restart()
stop()
createbackup()
update_overrides()//主要是配置文件修改，redis的密码是写配置文件的，可以通过这一项获取/修改密码
enable_as_master()
attach_replica()//添加从节点
detach_replica()//移除从节点
enable_root()/disable_root()/get_root_password()// Queen版支持
```

### 待测试
需要测试确认是否同时支持集群模式和主从模式

# MongoDB
### trove MongoDB支持创建集群、扩展集群、新增shard
- 创建集群：https://wiki.openstack.org/wiki/Trove/Clusters-MongoDB#Secondary_Members_and_Arbiters
- 扩展集群：https://specs.openstack.org/openstack/trove-specs/specs/liberty/cluster-scaling.html
- 弹性扩展shard


### trove MongoDB单实例的备份与恢复
- https://specs.openstack.org/openstack/trove-specs/specs/liberty/mongodb-backup-restore.html
- 利用trove-schedule可以实现定时备份，保证数据安全

### MongoDB-guestagent实现的接口有：
```
add_shard()//新增shard节点
grow()//新增relica_set节点
shrink()//移除replica_set节点
change_password()
create_database()
create_user()
grant_access()//为用户配置数据库权限
enable_root()
create_backup()
create_admin_user()
is_shard_active()
```
### 阿里的架构：
![image.png](https://github.com/jwongzblog/myblog/blob/master/openstack/ali-mongo-arch.png)

### trove MongoDB允许的集群架构
- trove MongoDB创建集群时默认创建3个replica_set实例
- 根据全局conf生成若干Mongos(queryRouter)、configServer实例
- 允许继续新增shard及其replica set，每次新增shard，会新创建等量的replica_set实例

![image.png](https://github.com/jwongzblog/myblog/blob/master/openstack/trove-mongo-arch.png)


# 结论
- 从数据库的架构复杂度来讲，redis特性更简单，易集成，可以优先测试，至于MongoDB，仍需要花点时间熟悉这个数据库的底层架构及特性
- 目前还不清楚是否适配nosql的所有版本
- 没有实现upgrade，无法更新数据库
- 暂时无法评估磁盘扩容对集群的影响
- 如何配置出最佳性能（DBA）
- 没有针对应用的监控措施