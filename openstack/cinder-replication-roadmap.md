# Cinder Replication

### [Juno](https://specs.openstack.org/openstack/cinder-specs/specs/juno/volume-replication.html)
- 通过快照创建volume，如果新卷指定的是replication类型，则会额外创建replication卷
- 通过克隆创建volume，如果新卷指定的是replication类型，则会额外创建replication卷
- 删除replication-type卷，会清理掉non-primary卷
- re-type volume，需移除或新增replication卷
- cmd：
```
  cinder replication-promote:将单个卷主从切换
  cinder replication-reenable:将inactive、active-stopped、error状态重置成active
```
  
### [Liberty](https://specs.openstack.org/openstack/cinder-specs/specs/liberty/replication_v2.html)
调整了接口和参数，便于扩展，方便各厂商实现driver层

### [Ocata](https://specs.openstack.org/openstack/cinder-specs/specs/ocata/ha-aa-replication.html)
- Cinder Volume Active/Active support
- cmd:
```
  cinder failover-host:将replication-type的所有master卷全部迁移至replication集群
```  
  
### [Pike](https://specs.openstack.org/openstack/cinder-specs/specs/pike/replication-group.html)
- 实现了replication-group，租户可以批量操作replication-group内的volume进行主从切换

### [Rocky](https://specs.openstack.org/openstack/cinder-specs/specs/rocky/cheesecake-promote-backend.html)
- 实现了fail-back功能，fail-over之后恢复存储集群，如果要恢复replication功能，R版之前需要手动修复cinder-db中的数据
- cmd:
```
  cinder-manage reset-active-backend replication_status=<status> <active_backend_id> <backend-host>
```  