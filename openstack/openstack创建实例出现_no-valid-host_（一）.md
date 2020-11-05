测试环境经常出现创建实例失败的现象，命令行show的时候只展示出"no valid host"的异常，其实这个只是表面现象，只能说明你创建实例的参数不满足nova scheduler的匹配算法，异常信息很笼统、抽象。要解决这个问题还是只能从日志入手，通过失败实例的nova id去检索日志，找出上下文，openstack这点很棒，日志很健全，基本能定位问题。本主题针对"no valid host"现象，逐一解决这个问题，此篇是我创建RDS实例出现的。

# 日志
```
2018-04-08 10:01:17.398 25269 DEBUG nova.filters [req-91b63cbe-2c92-4ecb-b9ea-652e6a192e2e - - - - -] 
Availability Zone 'nova' requested.
 (yc-demo-compute-3, yc-demo-compute-3) ram: 76194MB 

2018-04-08 10:01:17.398 25269 DEBUG nova.filters [req-91b63cbe-2c92-4ecb-b9ea-652e6a192e2e - - - - -] 
Filter RetryFilter returned 3 host(s) get_filtered_objects /usr/lib/python2.7/site-packages/nova/filters.py:104
2018-04-08 10:01:17.398 25269 DEBUG nova.scheduler.filters.aggregate_instance_extra_specs
 [req-91b63cbe-2c92-4ecb-b9ea-652e6a192e2e - - - - -] (yc-demo-compute-1, yc-demo-compute-1) ram: -47710MB 
disk: 95017984MB io_ops: 0 instances: 308 fails instance_type extra_specs requirements. 
'set([u'has-c', u'ecs-p1-c1', u'ecs-p2-m', u'ecs-gn-gn1'])' 
do not match 'rds-c' host_passes /usr/lib/python2.7/site-packages/nova/scheduler/filters/aggregate_instance_extra_specs.py:74

2018-04-08 10:01:17.400 25269 DEBUG nova.scheduler.filters.availability_zone_filter [req-91b63cbe-2c92-4ecb-b9ea-652e6a192e2e - - - - -] 
Availability Zone 'nova' requested. (yc-demo-compute-3, yc-demo-compute-3)
 ram: 76194MB disk: 95017984MB io_ops: 0 instances: 0 has AZs: set([u'az2']) 
host_passes /usr/lib/python2.7/site-packages/nova/scheduler/filters/availability_zone_filter.py:59

2018-04-08 10:01:17.400 25269 DEBUG nova.filters [req-91b63cbe-2c92-4ecb-b9ea-652e6a192e2e - - - - -]
 Filter AvailabilityZoneFilter returned 1 host(s) get_filtered_objects /usr/lib/python2.7/site-packages/nova/filters.py:104


10:01:17.401 25269 DEBUG nova.scheduler.filters.ram_filter [req-91b63cbe-2c92-4ecb-b9ea-652e6a192e2e - - - - -] 
(yc-demo-compute-2, yc-demo-compute-2) ram: -195166MB disk: 95017984MB io_ops: 0 
instances: 212 does not have 2048 MB usable ram, it only has 1395.0 MB usable ram. 
host_passes /usr/lib/python2.7/site-packages/nova/scheduler/filters/ram_filter.py:59
2018-04-08 10:01:17.402 25269 INFO nova.filters [req-91b63cbe-2c92-4ecb-b9ea-652e6a192e2e - - - - -] 
Filter RamFilter returned 0 hosts

```
# 原因
通过日志上下文可以看出
- 发现了3台computer节点
- RDS flavor匹配的aggregate没有computer-1节点
- computer-2节点内存不够
- computer-3节点的AvailabilityZone不匹配

# 解决方式
- 清理computer-2节点的资源
- aggregate增加computer-1节点
