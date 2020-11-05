# 背景
由于一些不可控因素导致单个AZ不可用，影响客户业务，而openstack的多AZ方案能尽可能缩短业务恢复时间，降低因服务不可用造成的损失。而不断向前迭代的ceph，强有力的支撑了openstack的高可用，通过ceph rbd mirror功能，可以提供pool、volume的跨集群同步
# 部署架构
![image.png](https://github.com/jwongzblog/myblog/blob/master/image/multi-az1.png)

![image.png](https://github.com/jwongzblog/myblog/blob/master/image/multi-az2.png)

![image.png](https://github.com/jwongzblog/myblog/blob/master/image/multi-az3.png)
#### 如图所示
- 平时AZ1对外提供服务，AZ2不对外提供服务
- AZ1的ceph1集群与AZ2的ceph2集群可以进行volume一级的同步，当VM数据下发至vol1、vol2后， ceph1集群以异步的方式将vol1、vol2的数据同步至ceph2集群上的vol1’、vol2’
- 一旦AZ1因为不可控因素导致不可用，我们可以在AZ2恢复出VM对外提供服务，此时启动VM对外提供无差异云主机
- 当AZ1恢复正常后，再将ceph2存储同步至ceph1，最后切换回AZ1对外提供服务

# 瓶颈
- AZ1、AZ2切换的稳定性
- 数据同步对网络带宽的要求，对集群的性能影响
- 同步延迟导致部分数据不可用，如何确保数据一致性
- Ceph集群间同步的并发能力
