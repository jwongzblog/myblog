### 各AZ上传redis镜像
### 镜像制作参考制作trove-redis镜像（基于我们自己的Centos7系统）方法

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

### 未完成
由于其他更重要的任务需要处理，暂时未调测cluster模式