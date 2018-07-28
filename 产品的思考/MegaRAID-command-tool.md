如果web是用户最好的交互，那么对于运维来讲CLI command tool则是最好的交互。**如果一个开源项目或者大型的闭源系统后台不提供CLI command tool，我认为是减分严重的，甚至是不成熟的**，毕竟它能大幅度的提高生产效率。

举个例子，因为ceph的特性，我们只需要将每块HDD硬盘设置成RAID0，SSD设置成non-raid，既利用了raid卡的cache，又能保证数据的高可靠。但是，假如我们新增3台DELL的服务器，如果通过idrac 图形界面操作去初始化存储，大半天时间浪费了，而且需要肉眼去check所有配置项是否正确，很容易犯错。随着数据中心规模越来越大，这样的生产效率无疑是低下的。

DELL自身也提供CLI command tool，但是一个数据中心的设计不可能被一个厂商绑架，好在RAID存储适配器的选择并不多，这样一来我们的脚本只需要适配少数几款RAID卡即可。

MegaRAID command tool仅仅支持LSI Logic SAS RAID存储适配器，下面简单介绍一下几个功能

- 查看硬盘信息
```
# /opt/MegaRAID/MegaCli/MegaCli64 -PDList -aALL
```
- 查看单块磁盘的详细信息
```
# /opt/MegaRAID/MegaCli/MegaCli64 -pdInfo -PhysDrv[252:3] -aALL
```
- 创建RAID0
```
# /opt/MegaRAID/MegaCli/MegaCli64 -CfgLdAdd -r0[252:5] WT Direct -a0
```
- 启动RAID卡的JBOD模式（non-raid）
```
# /opt/MegaRAID/MegaCli/MegaCli64 -AdpSetProp -EnableJBOD -1 -aALL

将一块盘设置成non-raid
# /opt/MegaRAID/MegaCli/MegaCli64 -PDMakeJBOD -PhysDrv[252:5] -a0 
```
- 删除raid
```
# /opt/MegaRAID/MegaCli/MegaCli64 -CfgLdDel -L3 -a0
```
更多CLI command请参考《[LSI SAS RAID卡配置中文版](https://wenku.baidu.com/view/79f0d0482b160b4e767fcfb0)》

**在这里多说几句，虽然生成命令行工具的库有很多，但是google brain开源的[python-fire](https://github.com/google/python-fire)库可以更简略的生成命令行**
# 参考
《[MegaRAID管理磁盘](http://blog.csdn.net/shengyyyyyy/article/details/78951747)》

《[MegaCli 监控raid状态](http://blog.chinaunix.net/uid-25135004-id-3139293.html)》
