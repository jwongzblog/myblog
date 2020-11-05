ironic的raid配置生效时机是在节点的cleaning阶段进行的，具体过程如下：
# 1.进入维护模式：
如果节点是刚创建，处于enroll阶段，那么执行如下命令使节点进入维护模式：
```
ironic --ironic-api-version 1.15 node-set-provision-state $NODE_UUID manage 
```

如果节点是active状态，那么执行如下命令把节点标记成删除状态才能继续操作：
```
ironic --ironic-api-version 1.15 node-set-provision-state $NODE_UUID deleted
```


# 2.退出维护模式：
```
ironic node-set-maintenance $NODE_UUID false
```
# 3.执行clean-step:
- 首先，配置
```
 raid configration：ironic --ironic-api-version 1.15 node-set-target-raid-config $NODE_UUID '{"logical_disks": [{"size_gb": "MAX", "disk_type": "hdd", "number_of_physical_disks": 2, "raid_level": "1"}]}'
```
- 触发clean-step：
``` 
 ironic --ironic-api-version 1.15 node-set-provision-state $NODE_UUID clean --clean-steps '[{"interface": "raid", "step": "delete_configuration"}, {"interface": "raid", "step": "create_configuration"}]'
```
- 注意事项：

HP的deploy image制作命令：
```
 disk-image-create -o proliant-agent-ramdisk ironic-agent fedora/ubuntu proliant-tools 
```
目前HP 驱动团队只在fedora/ubuntu进行过测试，再者diskimage-builder需要最新版本，否则无法将HP的RAID驱动注入到ironic python agent
raid配置参数具体看这里：[raid configration](https://docs.openstack.org/ironic/latest/admin/raid.html)

HP服务器会进入deploy system，然后创建raid，所有配置完成后将进入关机状态

# 4.进入available状态：
  ``` 
 ironic --ironic-api-version 1.15 node-set-provision-state $NODE_UUID provide
```
# 5.为什么老司机花了两周时间才搞定上面一点点的东西
**接下来讲讲排查问题遇到的坑，以及如何解决，一方面进行思维训练，另一方面也给大家排查问题提供方法、思路**
- 有次华三的产品经理说他们招了几百人的团队专门做KVM的优化，不知道是不是吹的，但我们的思路是通过ironic提供物理机来解决性能问题
- 一开始接到ironic的存储管理预研，同事给我一台all in one的devstack以及一台HP的服务器，并且告知成功创建了实例，但是按照ironic docs操作的时候我配了raid configration，始终创建node失败，maintenance reason显示系统尝试erase /sda1 /sdb1，但是由于只读所以失败了，HP服务器显示一块磁盘出于degraded状态，并且没有按照我的配置创建raid。我怀疑是磁盘的问题，由于一些原因换不了，为了验证是不是磁盘原因导致，我使用本地镜像安装了centos，排除了磁盘的影响，并且使用擦除磁盘的命令，确实没有权限
- 翻烂了docs，显示erase disk过程是默认开启的，于是通过修改ironic.conf关闭了这个操作，最终系统正常安装，并且能通过nova成功创建实例，*但raid始终不能按照我的配置进行*，还有就是为什么要默认开启erase disk？分析了一下conf文件，这个按理会调用IPA（[ironic python agent](https://github.com/openstack/ironic-python-agent)缩写）上面的驱动的清理策略，可能跟后续分析的驱动安装不正确有关系
- 开始读源码，首先按照我通常看源码的思路是先找入口，从rest-api开始分析，发现只有一个set raid configration的[PUT](https://www.w3.org/Protocols/rfc2616/rfc2616-sec9.html)操作。。。api只更新raid配置文件，不做具体操作，那raid怎么触发了的？后来从读ironic master 再到读到ironic python agent才发现真正触发raid生效的居然是通过命令输入一个[json](http://www.json.org/)格式，json的key即是函数名，ironic的agent模块会寻找合适的raid driver并调用该函数，但是新问题又来了，ironic master只有DELL的驱动实现了raid管理，但docs明明表示支持HP的raid管理。。。 
- 继续分析源码，发现其他驱动会把这个json格式发送给IPA，IPA会通过自省函数[getattr()](http://www.cnblogs.com/pylemon/archive/2011/06/09/2076862.html)去调用接送制定的函数，即create_configration()，但是IPA master全局搜了一把又没有该函数。。。于是跑去[IRC](http://blog.chinaunix.net/uid-27183448-id-3395934.html) [ironic频道](https://wiki.openstack.org/wiki/Meetings/Ironic)提问，一面常驻的code reviewer给我一个hp驱动的[git链接](https://github.com/openstack/proliantutils)，里面果然有HP的RAID管理，于是按照docs的提示，将这个驱动打包进diskimage里面，但是驱动返回的信息始终提示没有[clean_step](https://specs.openstack.org/openstack/ironic-specs/specs/not-implemented/manual-cleaning.html)。明明所有步骤都是对的，但是结果始终不对。
- 最后不得不发邮件到[openstack-dev@lists.openstack.org](openstack-dev@lists.openstack.org)求助，HP的 ilo驱动开发团队很快就联系上了我，我陆陆续续发了一些日志过去，但是依然没有头绪，看了diskimage-builder docs，HP的驱动似乎只支持ubuntu/fedora系统，而我的是centos，但他们反馈说centos他们只是没测试，让我实验成功后把测试结果给他们。。。同事帮我做了ubuntu deploy
 image后还是失败了。最后他们的一个工程师给了几个步骤去实验，第一个就是让我把[diskimage-builder](https://docs.openstack.org/diskimage-builder/latest/)升级到最新版本，然后我搜了一把，发现我们的生产环境是15年底的1.14版本，而巧合的是1.13版本刚好支持将HP驱动打包进去，所以没抛异常，而在diskimage-builder2.26版本才彻底支持HP驱动。。。
- 升级后，执行clean-step，raid按照我的配置成功创建，deploy image成功安装，node处于available状态，坐等nova创建实例，但此时，创建实例的时候user image又无法安装成功。
- 看过ironic的[troubleshooting](https://docs.openstack.org/ironic/latest/admin/troubleshooting.html)后，显示在/var/log/ironic/deploy下有image安装失败的日志。。。里面有一条是显示iscsi target创建过程中port创建失败，因为[3260端口被占用](http://www.361way.com/rhel7-iscsi/4728.html)。。。这个问题的解决思路是先执行删除占用端口的操作，然后重新创建，但是deploy-image无法由我们控制。。。
- 最后我不得不修改ironic.conf的默认3260端口，神奇的是居然成功安装user image，实例创建成功

**两周拿不出成果是很焦虑的，茶不思、饭不想，晚上做梦也在尝试解决这个问题，好在最后在同事和社区的帮助下成功解决，坐等ironic商业化**
