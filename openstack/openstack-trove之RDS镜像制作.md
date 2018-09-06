自从Tesora被以色列公司收购后，原先tesora的CTO（trove的PTL）出走verizon，trove就几乎冷门了下来，官方的镜像制作被delete，只有在github上还有人clone了一些分支在维护，但这些脚本多多少少有问题，在此，我将贡献出我花费了两周时间反复制作、调试，最后通过所有测试用例的手动制作镜像流程，为了社区的发展，我还是决定共享出来。

# 下载系统iso镜像，准备virsh、qemu环境（本文使用centos7.3）
生成qcow文件，配置xml，启动虚拟机，配置网络
# 安装mysql服务
##### mysql5.5
```
[mysql55-community]
name=MySQL 5.5 Community Server
baseurl=http://repo.mysql.com/yum/mysql-5.5-community/el/7/$basearch/
enabled=1
gpgcheck=0
gpgkey=file:/etc/pki/rpm-gpg/RPM-GPG-KEY-mysql
```
##### mysql5.6
```
[mysql56-community]
name=MySQL 5.6 Community Server
baseurl=http://repo.mysql.com/yum/mysql-5.6-community/el/7/$basearch/
enabled=1
gpgcheck=0
gpgkey=file:/etc/pki/rpm-gpg/RPM-GPG-KEY-mysql
```
###### 执行安装
```
yum -y install mysql-server
sudo mkdir -p /etc/mysql/conf.d/
sudo chown mysql:mysql -R /etc/mysql
```
##### 配置适合自身环境的my.cnf文件
##### 设置mysql开机启动
```
systemctl enable mysqld.service
systemctl start mysqld.service
systemctl status mysqld.service
```
# 部署openstack-trove-guestagent服务
##### 添加trove用户
```
/*useradd -u <old_uid> -g <old_gid> -d /home/trove*/否则会导致upgrade后对拷贝的文件夹没有权限，目前uid、gid都是1000
groupadd -g 1000 trove
useradd -u 1000 -g 1000 -d /home/trove -m trove
passwd -d trove
sudo usermod trove -a -G wheel
```

如果继续使用centos官方源，那么这个服务将会是最新的Q版，尽管trove的源码很久没大规模变动了，但如果trove其他组件还是较老的版本，那么trove依赖包会有一些问题，比如我们定制的代码依赖了tenant id去命名，结果最新的openstack把tenant_id给过滤了，导致代码无法正常工作，此时就需要改动yum源信息，指向公司内部归档的yum源码
###### 改动yum.repos.d，指向内部源
执行
```
sed -i -e 's/^Defaults.*requiretty/# Defaults requiretty/g' /etc/sudoers
yum -y install openstack-trove-guestagent python-troveclient python-netifaces
```
##### 卸载openstack-trove-guestagent，使用定制代码安装openstack-trove-guestagent
```
sudo systemctl stop openstack-trove-guestagent
sudo pip uninstall trove
sudo python setup.py install /*如果提升git没安装，执行yum -y install git*/
sudo systemctl start openstack-trove-guestagent
```

##### 配置trove-guestagent服务启动，修改/usr/lib/systemd/system/openstack-trove-guestagent.service
```
[Unit]                                                                                                                                                                                         
Description=OpenStack Trove Guest
After=syslog.target network.target cloud-init.service
[Service]
Type=simple
User=trove
ExecStart=/usr/bin/trove-guestagent --config-file /etc/trove/conf.d/trove-guestagent.conf --config-file /etc/trove/conf.d/guest_info.conf
Restart=on-failure
RestartSec=2s
[Install]
WantedBy=multi-user.target
```
##### 由于环境缺少一些东西，执行以下命令
```
chown trove:trove /etc/trove
chown trove:trove /usr/share/trove
chown trove:trove /var/log/trove
mkdir  /etc/trove/conf.d
cp /etc/trove/guest_info /etc/trove/conf.d/guest_info.conf
cp /etc/trove/trove-guestagent.conf /etc/trove/conf.d/trove-guestagent.conf
chmod 0755 /etc/trove/conf.d/guest_info.conf
chown trove:trove /etc/trove/conf.d/guest_info.conf
```
修改：/etc/trove/conf.d/guest_info.conf，添加guest_id=none

修改：
- vim /etc/trove/conf.d/trove-guestagent.conf  ，将trove-logging-guestagent.conf注释掉
- 在DEFAULT中添加datastore_manager = mysql
##### trove-guestagent启动
```
# Enable and start service
systemctl enable openstack-trove-guestagent.service
systemctl start openstack-trove-guestagent.service
systemctl status openstack-trove-guestagent.service
```
##### 安装xtrabackup，mysql备份依赖此工具
```
#yum -y install percona-xtrabackup
```
##### 配置系统参数
```
sed -i -r 's/^\s*#(net\.ipv4\.ip_forward=1.*)/\1/' /etc/sysctl.conf
echo 1 > /proc/sys/net/ipv4/ip_forward
 
TEMPFILE=`mktemp`
echo "trove ALL=(ALL) NOPASSWD:ALL" > $TEMPFILE
chmod 0440 $TEMPFILE
sudo chown root:root $TEMPFILE
sudo mv $TEMPFILE /etc/sudoers.d/60_trove_guest
 
echo 'GRUB_CMDLINE_LINUX="no_timer_check"' > /etc/default/grub
grub2-mkconfig > /boot/grub2/grub.cfg
```
# 清理日志
- 还原yum源信息，清理导入的trove源码
- 删除日志
```
#rm -rf /var/lib/cloud/*
#rm  /tmp/* -r
#rm  ~/.bash_history -rf
#rm  ~/.viminfo -rf
#rm  /var/log/*.log -rf
#rm  /var/log/*.old -rf
#yum clean all
#history -c
```
# 使用virt-sysprep工具对镜像做处理
##### 在做这个命令之前一定要备份镜像文件，因为这个命令很可能会损坏镜像，关闭虚拟机，执行：
```
#virt-sysprep  -a  $image
```
# 上传镜像至glance
