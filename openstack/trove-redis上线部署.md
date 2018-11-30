### 各AZ上传redis镜像
### 镜像制作参考[此处](https://github.com/jwongzblog/myblog/blob/master/openstack/trove-redis%E9%95%9C%E5%83%8F%E5%88%B6%E4%BD%9C.md)

### 创建datastore及datastore-version
```
su -s /bin/sh -c "trove-manage  --config-file /etc/trove/trove.conf datastore_update redis ''" trove
```
 
### 各AZ对应的镜像不一样
```
su -s /bin/sh -c "trove-manage --config-file /etc/trove/trove.conf  datastore_version_update   redis 4.0-az1 redis 'f3749524-fea7-4754-86b1-f26f2efc1f77'  ''  1" trove
```

### 修改/etc/trove/trove.conf
```
volume_support = False
```

### 创建redis.cloudinit
```
将mysql.cloudinit内容拷贝给redis.cloudinit
```

### 创建redis flavor
务必指定ephemeral
```
nova  flavor-create redis.c2.small redis.c2.small 4096 40 4 --ephemeral 20
```

### 初始化redis的configration_group
redis可以设置protected_mode=no来免密登录，但是此方式不安全，所以最好是每个redis创建一个configration_group，里面单独设置redis的密码（requirepass）
```
#每个datastore-version都需要初始化
sudo trove-manage db_load_datastore_config_parameters redis 4.0-az1 /usr/lib/python2.7/site-packages/trove/templates/redis/validation-rules.json
```

### 为master节点创建configration_group
```
trove  configuration-create redis-conf-m '{"requirepass":"a123456"}' --datastore redis --datastore_version 43a2a9ff-d78f-4d5f-ba96-87b60b675125
```

### 为replica节点创建configration_group
```
#主从同步，需要添加主节点的密码masterauth
trove  configuration-create redis-conf-repl '{"requirepass":"a123456","masterauth":"a123456"}' --datastore redis --datastore_version 43a2a9ff-d78f-4d5f-ba96-87b60b675125
```

### 创建主从模式的redis
```
trove create redis1010 redis.c2.small  --datastore redis --datastore_version 4.0-az1 --nic net-id=1acccaf1-a300-46f1-9766-5bfc749bc0b7 --availability_zone nova  --volume_type yc-beta-ceph-hdd-2 --configuration redis-conf-m
 
trove create redis1010-repl redis.c2.small  --datastore redis --datastore_version 4.0-az1 --nic net-id=1acccaf1-a300-46f1-9766-5bfc749bc0b7 --availability_zone nova  --volume_type yc-beta-ceph-hdd-2  --replica_of  redis1010 --configuration redis-conf-repl
```

### Redis集群模式
- trove.conf  trove-taskmanager.conf开放3个端口6379,16379,10000，供集群间同步使用
- 修改/etc/trove/trove.conf，volume_support = True，此处和replication版本冲突，python-trove-client要求volume必须添加，trove-api又要求quota不为None，所以暂时填了True，但是trove显示的volume使用量是和flavor的ephemeral的数据量是冲突的，实例实际使用的是ephemeral的值。需继续调研，如果要统一，要分析一下flavor的ephemeral为空、volume_support = True的情况下，为什么无法创建实例
```
trove cluster-create redis-cluster redis 4.0-az1  --instance "flavor=redis.c2.small,volume=10,volume_type=yc-beta-ceph-hdd-2,availability_zone=nova,nic='net-id=1acccaf1-a300-46f1-9766-5bfc749bc0b7'"  --instance "flavor=redis.c2.small,volume=10,volume_type=yc-beta-ceph-hdd-2,availability_zone=nova,nic='net-id=1acccaf1-a300-46f1-9766-5bfc749bc0b7'"  --instance "flavor=redis.c2.small,volume=10,volume_type=yc-beta-ceph-hdd-2,availability_zone=nova,nic='net-id=1acccaf1-a300-46f1-9766-5bfc749bc0b7'"

```
- 测试集群
```
$redis-cli -c -h $IP -p $PORT
$IP:$PORT>get hello
"world"
```
- 集群扩容
```
trove cluster-grow ${cluster-id}   --instance "flavor=redis.c2.small,volume=10,volume_type=yc-beta-ceph-hdd-2,availability_zone=nova,nic='net-id=1acccaf1-a300-46f1-9766-5bfc749bc0b7'"
```
- 集群缩容
```
trove cluster-shrink ${cluster-id} ${instance-id}
```

### Redis扩容
只支持local storage（flavor.ephemeral，系统盘做了一个分区），resize-volume会抛错，提示没有卷，所以单机和主从模式会面临一个扩容问题
*如果考虑扩容，建议使用cluster模式，可以通过grow/shrink来扩、缩容集群*

### Redis备份与恢复
单机、replication、cluster模式均可选择其中一个节点进行备份
*但是，只能通过备份ID创建单机的redis进行恢复，无法通过备份创建一个新集群出来*
```
trove backup-create 4eb653e0-dc1a-42b8-b05b-d5bf9134792c rb1
 
trove create redis1130-bk redis.c2.small --size 10  --datastore redis --datastore_version 4.0-az1 --nic net-id=1acccaf1-a300-46f1-9766-5bfc749bc0b7 --availability_zone nova  --volume_type yc-beta-ceph-hdd-2 --backup 395b5d5a-8c54-42a1-94ec-5ecfa45eea19
```