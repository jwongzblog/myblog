# 下载基础镜像
# 通过virsh启动镜像
# 安装redis
```
sudo yum install redis
sudo systemctl start redis
sudo systemctl enable redis
 
#检查安装
$redis-cli ping
PONG
 
#安装redis-py，否则GA无法启动
pip2 install redis
 
# By default, redis-py will attempt to use the HiredisParser if installed.
# Using Hiredis can provide up to a 10x speed improvement in parsing responses from the Redis server.
# 安装hiredis
pip2 install hiredis
 
#修改/etc/redis.conf配置，否则无法远程登录，或者去掉此项，有configuration_group来维护redis的登录密码
protected-mode no
```

# 安装openstack-trove-guestagent

```
添加trove用户
/*useradd -u <old_uid> -g <old_gid> -d /home/trove*/否则会导致upgrade后对拷贝的文件夹没有权限，目前uid、gid都是1000
groupadd -g 1000 trove
useradd -u 1000 -g 1000 -d /home/trove -m trove
passwd -d trove
sudo usermod trove -a -G wheel

sed -i -e 's/^Defaults.*requiretty/# Defaults requiretty/g' /etc/sudoers
yum -y install openstack-trove-guestagent python-troveclient python-netifaces
```

## 卸载openstack-trove-guestagent，使用源码安装openstack-trove-guestagent
```
sudo systemctl stop openstack-trove-guestagent
sudo pip uninstall trove /*yum -y install python-pip*/
sudo python setup.py install /*如果提升git没安装，执行yum -y install git*/
sudo systemctl start openstack-trove-guestagent
```

## 配置trove-guestagent

## 修改/usr/lib/systemd/system/openstack-trove-guestagent.service 为如下：

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
 

## 执行：
```
chown trove:trove /etc/trove
chown trove:trove /usr/share/trove
chown trove:trove /var/log/trove
mkdir  /etc/trove/conf.d
cp /etc/trove/guest_info /etc/trove/conf.d/guest_info.conf
cp /etc/trove/trove-guestagent.conf /etc/trove/conf.d/trove-guestagent.conf
chmod 0755 /etc/trove/conf.d/trove-guestagent.conf
chmod 0755 /etc/trove/conf.d/guest_info.conf
chown trove:trove /etc/trove/conf.d/trove-guestagent.conf
chown trove:trove /etc/trove/conf.d/guest_info.conf
```
## 修改：/etc/trove/conf.d/guest_info.conf，添加guest_id=none

## 修改：vim /etc/trove/conf.d/trove-guestagent.conf  ，将trove-logging-guestagent.conf注释掉

在DEFAULT中添加
```
datastore_manager = redis
# Enable and start service
systemctl enable openstack-trove-guestagent.service
systemctl start openstack-trove-guestagent.service
systemctl status openstack-trove-guestagent.service
``` 

# 配置系统参数

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

## 还原yum源信息，清理导入的trove源码

## 清理日志

```
rm -rf /var/lib/cloud/*
rm  /tmp/* -r
rm  ~/.bash_history -rf
rm  ~/.viminfo -rf
rm  /var/log/*.log -rf
rm  /var/log/*.old -rf
yum clean all
history -c
```