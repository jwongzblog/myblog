### 功能简介
* ceph从Jewel版本开始支持跨集群的pool、image同步，提高存储的容灾能力。主要实现原理是启动多线程监听pool或者image。
* Luminous版本支持rbd mirror 在多台服务器上部署进程，并保证其中一个环境的进程成为Leader，确保进程的scale-out。支持延迟同步。
* Minic版本支持每个mirror进程都能工作，分担负载，还支持延迟删除。

### 安装部署
本文档结合Redhat的[最佳实践](https://access.redhat.com/documentation/en-us/red_hat_ceph_storage/3/html/block_device_guide/block_device_mirroring)（jewel版本），在环境测试后总结，由于我用的是luminous版本，和Redhat的最佳实践还是有出入。此篇为了精炼表达原理，只描述了单向同步的步骤，至于双向同步，无非是两个集群按照以下步骤无差别重复一遍。

集群一：mirror-ceph1

集群二：mirror-ceph2

安装rbd-mirror
jewel版本中，一个集群只能部署一个rbd-mirror进程，但luminous版本已经允许部署多个进程来横向扩容
```
#mirror-ceph2
yum install rbd-mirror
```
为rbd-mirror进程创建独立账户
```
#mirror-ceph1、mirror-ceph2
ceph auth get-or-create client.rbd-mirror.{unique id} mon 'profile rbd' osd 'profile rbd' -o /etc/ceph/ceph.client.rbd-mirror.{unique id}.keyring
{unique id}:unique name，在这里我命名为ceph2，即创建账户：client.rbd-mirror.ceph2
```
启动rbd-mirror Daemon
```
#mirror-ceph1
systemctl enable ceph-rbd-mirror.target
systemctl enable ceph-rbd-mirror@rbd-mirror.ceph1
systemctl start ceph-rbd-mirror@rbd-mirror.ceph1
 
#mirror-ceph2
systemctl enable ceph-rbd-mirror.target
systemctl enable ceph-rbd-mirror@rbd-mirror.ceph2
systemctl start ceph-rbd-mirror@rbd-mirror.ceph2
```
将key拷贝给mirror-ceph2（如果需要双向同步，反向操作ceph1）
```
#mirror-ceph1
 
# scp /etc/ceph/ceph.conf <user>@mirror-ceph2-host:/etc/ceph/mirror-ceph1.conf
# scp /etc/ceph/ceph.client.rbd-mirror.ceph1.keyring <user>@mirror-ceph2-host:/etc/ceph/mirror-ceph1.client.rbd-mirror.ceph1.keyring
```
在两个ceph集群中分别创建同名pool
```
ceph osd pool create mirror-pool-test 64 64
```
两个集群开启同步模式：pool/image

在jewel版本中，只要开启mirror-pool-test的pool模式，其下面的image会自动同步。但在luminous版本中感觉二者差异不大，比如说开启pool模式后，需要开启image的journaling才同步，而开启image模式，除了开启image的journaling，还需要enable image mirror才开始同步，多了一个步骤而已。具体的区别后续再深入剖析

pool模式：
```
#mirror-ceph1、mirror-ceph2
rbd  mirror pool enable mirror-pool-test pool
```
image模式：
```
#mirror-ceph1、mirror-ceph2
rbd  mirror pool enable mirror-pool-test image
```
创建peer服务进程，监听mirror-ceph1中mirror-pool-test的变化
```
#mirror-ceph2
rbd  mirror pool peer add mirror-pool-test client.rbd-mirror.ceph1@mirror-ceph1
rbd mirror pool info mirror-pool-test
```
*移除peer：*
```
rbd mirror pool peer remove  {pool-name} {peer-uuid}
```
测试效果：
```
#mirror-ceph1
rbd -p mirror-pool-test create image1 --size 1024
rbd feature enable mirror-pool-test/image1 journaling
 
#如果是image模式，需要再执行：
rbd mirror image enable  mirror-pool-test/image1
```
```
#mirror-ceph2查看同步效果
rbd -p mirror-pool-test list
rbd mirror image status mirror-pool-test/image1
```
提升slave image为primary

上面的步骤可以看到image1的状态，在mirror-ceph1中，image1的primary为true，意味这个image是可以读写操作，而mirror-ceph2的image1的primary为false，意味着这个image是只读。当mirror-ceph1发生故障时，我们可以通过promotion操作，把mirror-ceph2的image1提升为主（primary状态）
```
rbd  mirror image promote mirror-pool-test/image1
```
脑裂或者故障恢复后，可以强制resync，重新同步
```
rbd mirror image resync mirror-pool-test/image1
#从L版的执行过程来看，应该是先把从节点的卷删除，重新同步了一次，并不是增量同步（从节点执行）
```

### 总结：
对照Redhat最佳实践，在进程启动的key处理以及pool/image模式的处理有点出入。cinder实现了replication机制，确保创建volume时，备份集群创了一个mirror卷，实时同步的。它的调研文档我就偷懒不写了，反正过程中问题很多，趟了很多坑，最终把功能都调通了，有需求的可以私我。由于成熟度不高，达不到我们想要的主机的高可用，需要太多人为的操作才能恢复主机。我分别向社区提交了三个需求，不知道猴年马月能完成

https://blueprints.launchpad.net/cinder/+spec/rbd-replication-failover-by-force
https://blueprints.launchpad.net/cinder/+spec/rbd-replication-promote-old-master-vol-when-failback
https://blueprints.launchpad.net/cinder/+spec/make-vm-work-again-after-failover