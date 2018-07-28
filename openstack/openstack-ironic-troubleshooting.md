1.deploy过程失败，maintance reason显示错误为：wipe /sda1 /sdb1 error, read only
解决方法：
2.执行clean_step提示IPA不存在该函数实现
解决方法：
3.如果deploy-image过程完毕后进入安装生成环境镜像失败，除非是网络原因导致日志传输失败，一般失败日志都会保留在/var/log/ironic/deploy
解决方法：
4.DELL 服务器节点生产时，驱动加载错误，错误日志：“2017-08-11 19:55:40.771 11110 ERROR ironic.conductor.manager     boot_device = next(key for (key, value) in _BOOT_DEVICES_MAP.items()  //2017-08-11 19:55:40.771 11110 ERROR ironic.conductor.manager StopIteration”
解决方法：
5.DELL 服务器节点生产时，idrac页面显示“PXE-E51: No DHCP or proxyDHCP offers were received”，服务器尝试多次后关机，clean error:time out
解决方法：
6.如果选择PXE方式，节点启动时出现“tftp open timeout”
解决方法：
7.如果PXE安装deploy image成功，但过了一段时间后关机并提示clean failed:timeout
解决方法：
8.虚拟机删除后重新生产节点会把port删掉
解决方法：
9.出现如下错误：
```
2017-08-22 01:40:15.207 10756 ERROR ironic.drivers.modules.deploy_utils [req-36e4af9d-dba4-45a5-8531-3c0f92fac50f - - - - -] Command: sudo ironic-rootwrap /etc/ironic/rootwrap.conf parted -a optimal -s /dev/disk/by-path/ip-192.168.20.11:3261-iscsi-iqn.2008-10.org.openstack:ec0b8d4f-f798-409b-a274-301317805f90-lun-1 -- unit MiB mklabel msdos mkpart primary  1 512001
2017-08-22 01:40:15.208 10756 ERROR ironic.drivers.modules.deploy_utils [req-36e4af9d-dba4-45a5-8531-3c0f92fac50f - - - - -] StdOut: u''
2017-08-22 01:40:15.208 10756 ERROR ironic.drivers.modules.deploy_utils [req-36e4af9d-dba4-45a5-8531-3c0f92fac50f - - - - -] StdErr: u'Error: The location 512001 is outside of the device /dev/sdb.\n'
```
解决方法：
10.raid创建完毕后进入manageable状态，通过provide使其进入available状态时出现如下错误：
```

94f43b4d56215a6fc9027 - - -] Exiting old state 'manageable' in response to event 'provide' on_exit /usr/lib/python2.7/site-packages/ironic/common/states.py:228
2017-08-23 03:40:31.059 16065 DEBUG ironic.common.states [req-a0306ea2-3d9d-4ab5-96e8-3e74cf856632 2697fc6a3e0f4dda9ba5537d87d1a8ca 6f8f6310a2994f43b4d56215a6fc9027 - - -] Entering new state 'cleaning' in response to event 'provide' on_enter /usr/lib/python2.7/site-packages/ironic/common/states.py:234
2017-08-23 03:40:31.066 16065 INFO ironic.conductor.task_manager [req-a0306ea2-3d9d-4ab5-96e8-3e74cf856632 2697fc6a3e0f4dda9ba5537d87d1a8ca 6f8f6310a2994f43b4d56215a6fc9027 - - -] Node e8d3f6b5-bc7f-4de6-83f5-84778851662f moved to provision state "cleaning" from state "manageable"; target provision state is "available"
2017-08-23 03:40:31.068 16065 DEBUG ironic.conductor.manager [req-a0306ea2-3d9d-4ab5-96e8-3e74cf856632 2697fc6a3e0f4dda9ba5537d87d1a8ca 6f8f6310a2994f43b4d56215a6fc9027 - - -] Starting automated cleaning for node e8d3f6b5-bc7f-4de6-83f5-84778851662f _do_node_clean /usr/lib/python2.7/site-packages/ironic/conductor/manager.py:896
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager [req-a0306ea2-3d9d-4ab5-96e8-3e74cf856632 2697fc6a3e0f4dda9ba5537d87d1a8ca 6f8f6310a2994f43b4d56215a6fc9027 - - -] Failed to prepare node e8d3f6b5-bc7f-4de6-83f5-84778851662f for cleaning: 'NoneType' object is not iterable
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager Traceback (most recent call last):
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager   File "/usr/lib/python2.7/site-packages/ironic/conductor/manager.py", line 928, in _do_node_clean
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager     prepare_result = task.driver.deploy.prepare_cleaning(task)
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager   File "/usr/lib/python2.7/site-packages/ironic_lib/metrics.py", line 61, in wrapped
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager     result = f(*args, **kwargs)
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager   File "/usr/lib/python2.7/site-packages/ironic/drivers/modules/drac/deploy.py", line 47, in prepare_cleaning
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager     in node.driver_internal_info.get('clean_steps', [])
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager TypeError: 'NoneType' object is not iterable
2017-08-23 03:40:31.069 16065 ERROR ironic.conductor.manager
```
11.nova boot有时候会出现“No valid host was found. There are not enough hosts available”的错误
解决方法：
12.nova boot失败，偶尔会出现ironic node deploy failed的状态，此时需要通过nova-set-provision-state deleted使其恢复到available的状态
13.安装win2008镜像时，偶尔会出现no such file or directory的错误而中断整个虚机创建过程
解决方法：

想知道答案？不好意思，公司花了钱雇我，而我花大量的时间解决的，你通过简单的搜索就想立马解决这个问题，身为程序员居然还是用百度。。。不要太爽啊。。。自己花点时间，实在解决不了私聊我
