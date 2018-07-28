压测ceph集群以便评估几家厂商的送测设备，其中一家的设备很奇怪，fio同时压测所有硬盘能达到机械硬盘的极限值，但是fio压ceph rbd就是跑不满，为此调整了几个Linux内核参数，每块盘的iops显著提升，总的来讲TCP参数的调整带来的改善要大得多，这个跟ceph的数据模型设计有大的关系吧。

目标文件/etc/sysctl.conf，参数调整如下：
#net.ipv4.tcp_mtu_probing=1

为了提高ceph集群节点间的数据同步性能，我们开启了jump frame，并且设定值为9000（标准的报文传输量是1500，超过这个值报文就会被分段，影响性能，因此巨帧可以减少报文分段），但是巨帧（jump frame）有个缺陷，就是一旦出现错误或者丢包，重新组织并发送这个包就会大大影响传输效率。因此这个参数开启了PLPMTUD（Packetization Layer Path MTU Discovery），相比于之前的PMTUD（最大传输单元路径发现），大大提升了健壮性。参考至rfc4821

 

#net.netfilter.nf_conntrack_max = 1048576

调高conntrack最大值，conntrack（连接跟踪）用来跟踪和记录一个连接的状态

 

#kernel.pid_max = 4194303

调整创建的进程数上限

 

#vm.swappiness = 10

使用磁盘swap分区的比例，默认值为60，但是在内存足够大时，调成10可以提高性能，调成0就只使用物理内存

 

#vm.vfs_cache_pressure = 50

默认值是100，超过100时，内核会高优先清理缓存

 

#net.core.rmem_max = 56623104

最大的TCP数据接收缓冲


#net.core.wmem_max = 56623104

最大的TCP数据发送缓冲


#net.core.rmem_default = 56623104
#net.core.wmem_default = 56623104

同上，默认值


#net.core.optmem_max = 40960

每个套接字允许的最大缓冲区大小


#net.ipv4.tcp_rmem = 4096 87380 56623104

为自动调优定义socket使用的内存。第一个值是为socket接收缓冲区分配的最少字节数；第二个值是默认值（该值会被rmem_default覆盖），缓冲区在系统负载不重的情况下可以增长到这个值；第三个值是接收缓冲区空间的最大字节数（该值会被rmem_max覆盖）


#net.ipv4.tcp_wmem = 4096 65536 56623104

为自动调优定义socket使用的内存。第一个值是为socket发送缓冲区分配的最少字节数；第二个值是默认值（该值会被wmem_default覆盖），缓冲区在系统负载不重的情况下可以增长到这个值；第三个值是发送缓冲区空间的最大字节数（该值会被wmem_max覆盖）


#net.core.somaxconn = 1024

系统中每一个端口最大的监听队列的长度，默认是128


Increase number of incoming connections backlog, default is 1000
#net.core.netdev_max_backlog = 50000

在每个网络接口接收数据包的速率比内核处理这些包的速率快时，允许送到队列的数据包的最大数目


Maximum number of remembered connection requests, default is 128
#net.ipv4.tcp_max_syn_backlog = 30000

对于还未获得对方确认的连接请求，可保存在队列中的最大数目。如果服务器经常出现过载，可以尝试增加这个数字


Increase the tcp-time-wait buckets pool size to prevent simple DOS attacks, default is 8192
#net.ipv4.tcp_max_tw_buckets = 2000000

表示系统同时保持TIME_WAIT的最大数量，如果超过这个数字，TIME_WAIT将立刻被清除并打印警告信息


Recycle and Reuse TIME_WAIT sockets faster, default is 0 for both
net.ipv4.tcp_tw_recycle = 1

#表示开启TCP连接中TIME-WAIT sockets的快速回收，默认为0，表示关闭


#net.ipv4.tcp_tw_reuse = 1

表示开启重用。允许将TIME-WAIT sockets重新用于新的TCP连接，默认为0，表示关闭。为了对NAT设备更友好，建议设置为0


Decrease TIME_WAIT seconds, default is 30 seconds
#net.ipv4.tcp_fin_timeout = 10

修改系統默认的 TIMEOUT 时间


Tells the system whether it should start at the default window size only for TCP connections that have been idle for too long, default is 1
#net.ipv4.tcp_slow_start_after_idle = 0

设置为0，一个tcp连接在空闲后不进入slow start阶段，即每次收发数据都直接使用高速通道，平均延时43毫秒，跟计算的理论时间一致

 
#If your servers talk UDP, also up these limits, default is 4096
#net.ipv4.udp_rmem_min = 8192

UDP接收缓存区最小值


#net.ipv4.udp_wmem_min = 8192

UDP输出缓存区最小值

 

Disable source redirects
Default is 1
#net.ipv4.conf.all.send_redirects = 0

禁止转发重定向报文


#net.ipv4.conf.all.accept_redirects = 0

禁止接收路由重定向报文，防止路由表被恶意更改


Disable source routing, default is 0
#net.ipv4.conf.all.accept_source_route = 0

禁止包含源路由的ip包
