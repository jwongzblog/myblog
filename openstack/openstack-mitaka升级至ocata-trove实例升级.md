 openstack Mitaka升级到Ocata后，由于service与实例之间的api（主要是参数）发生变化，之前创建并使用的trove实例必须升级，否则在运行个别功能时，实例会发生异常。此外，由于trove upgrade（使用新镜像重建实例达到升级的目的，避免人工干预升级过程）在N版本提供，所以本次升级无法使用这个功能。为此，我们准备了一系列的脚本来升级trove实例，以下是各个步骤、脚本说明、注意事项：
    
**前置条件： trove服务升级成功，确保各项配置文件是正确的，尤其是rabbitMQ、ceph的配置信息（特殊：针对trove独立部署、网络不通需要HAproxy转换）**
   
#####1、上传最新的mysql镜像，更新datastore-version的镜像ID，指向新上传的镜像
#####2、/usr/lib/python2.7/site-packages/trove/templates/mysql/目录下有个配置文件夹 ‘5.5-ga’是升级前创建的datastore，我们升级后这个文件夹不存在了，需要把当前目录下的'5.5'文件夹复制成‘5.5-ga’，所有的controller节点都需要更新
#####3、搜集trove实例相关信息，方便脚本正常运行，需要留意 跨project的问题。通过“ select * from instances where deleted = '0';”搜集到实例，与应用层给出的实例对比一下
4、选择其中一个controller节点（例如controller1），进入$work_dir/common/get_trove_instance_info/目录执行：bash -x 01_get_trove_instances_info.sh
```
#!/usr/bin/env bash
# author: wangjun
# date: 2017.9.28
# Function: get nova instances's nic_id\ip\secgroupname by trove_id,write into nova_nic_ip_secgroup.txt
# Function: add secgroup rule to make sure remote-access;write rule_id in sec_group_id.txt
set -e 
set -xsource 
../../upgrade_env
for value in $(cat $work_dir/common/get_trove_instance_info/trove_ins_ids.txt)
do 
    echo $value TROVE_ID=$(echo $value |awk -F\; '{print $1}')
# 留意参数server_id，如果升级前执行需改成compute_instance_id 
    NOVA_ID=$(trove show $TROVE_ID |awk '/^\| compute_instance_id / {print $4}') 
    NOVA_NIC_ID=$(nova interface-list $NOVA_ID |awk '/^\| ACTIVE / {print $6}') 
    NOVA_IP_ADDR=$(nova interface-list $NOVA_ID |awk '/^\| ACTIVE / {print $8}') 
    SECURITY_GROUPS=$(nova show $NOVA_ID |awk '/^\| security_groups / {print $4}')
    echo $NOVA_NIC_ID\;$NOVA_IP_ADDR\;$SECURITY_GROUPS, >> $work_dir/common/get_trove_instance_info/nova_nic_ip_secgroup.txt done

for value in $(cat $work_dir/common/get_trove_instance_info/nova_nic_ip_secgroup.txt)
do 
    echo $value LINE=$value 
    SECURITY_GROUPS_NAME=$(echo $LINE |awk -F\; '{print $3}') 
    SECURITY_GROUP_NAME=$(echo $SECURITY_GROUPS_NAME |awk -F, '{print $1}') 
    SECURITY_GROUP_ID=$(neutron security-group-show $SECURITY_GROUP_NAME |awk '/^\| id / {print $4}')
    #开放实例的22端口，方便后面的脚本执行 
    SECURITY_GROUP_RULE_ID=$(neutron security-group-rule-create \ --direction ingress --ethertype IPv4 \ --protocol tcp --port-range-min 22 --port-range-max 22 \ --remote-ip-prefix 0.0.0.0/0 
    $SECURITY_GROUP_ID |awk '/^\| id / {print $4}') 
    echo $SECURITY_GROUP_RULE_ID >> $work_dir/common/get_trove_instance_info/sec_group_id.txtdone
#拷贝文本至网络节点，实例升级脚本需在网络节点执行
sshpass -p$NETWORK_NODE_PASSWORD scp $work_dir/common/get_trove_instance_info/nova_nic_ip_secgroup.txt root@$NETWORK_NODE_IP:$work_dir/network
```
#####5.选择其中一个network节点（例如network1），进入$work_dir/network/目录执行：bash -x  08_trove_instance_upgrade.sh
```
#!/usr/bin/env bash
# author: wangjun
# date: 2017.9.25
# Function: trove instance upgrade
# warning: exc this script after all services are upgraded,act once
echo "make sure all services are upgraded,then remove the block"
:<<BLOCK
set -e 
set -x
source ../upgrade_env
for key in $(cat $work_dir/network/nova_nic_ip_secgroup.txt)
do
    echo $key
    LINE=$key
    NET_ID=$(echo $LINE |awk -F\; '{print $1}')
    INSTANCE_IP=$(echo $LINE |awk -F\; '{print $2}')
    USER=root
    PASSWORD=trove.com
    NS_COMMAND="ip netns exec qdhcp-$NET_ID "
    SSH_COMMAND="$NS_COMMAND sshpass -p$PASSWORD ssh $USER@$INSTANCE_IP "
    SCP_COMMAND="$NS_COMMAND sshpass -p$PASSWORD scp "
#将包含源码和库包的压缩包拷贝至实例
    $SCP_COMMAND $work_dir/network/trove_rpm_code.tar root@$INSTANCE_IP:/home
    $SSH_COMMAND "systemctl stop openstack-trove-guestagent"
    $SSH_COMMAND "cd /home; tar xvf trove_rpm_code.tar"
    $SSH_COMMAND "cd /home/trove_rpm_code/rpm; rpm -U *.rpm"
    $SSH_COMMAND "mv /usr/lib/python2.7/site-packages/trove-7.0.0-py2.7.egg-info /home"
    $SSH_COMMAND "mv /usr/lib/python2.7/site-packages/trove  /usr/lib/python2.7/site-packages/trove_back"
    $SSH_COMMAND "cd /home/trove_rpm_code/trove; python setup.py install"
#使用RPM包升级后的服务启动脚本路径不对，需使用压缩包的脚本覆盖
    $SSH_COMMAND "\cp  /home/trove_rpm_code/openstack-trove-guestagent.service /usr/lib/systemd/system/openstack-trove-guestagent.service"
    $SSH_COMMAND "systemctl daemon-reload"
    $SSH_COMMAND "systemctl restart openstack-trove-guestagent"
    $SSH_COMMAND "cd /home; rm trove_rpm_code.tar trove_rpm_code -r"
done
BLOCK
```

#####6.回到controller1节点，进入$work_dir/common/get_trove_instance_info/目录执行：bash -x 02_clear_secgroup_rule.sh   关闭实例22端口
```

#!/usr/bin/env bash
# author: wangjun
# date: 2017.9.28
# Function: after all,clear secgroup rule
source ../../upgrade_env
for value in $(cat $work_dir/common/get_trove_instance_info/sec_group_id.txt)
do
    echo $value
    SECURITY_GROUP_ID=$value
    neutron security-group-rule-delete $SECURITY_GROUP_ID
done
```

#####7.回到controller1节点，进入$work_dir/common/get_trove_instance_info/目录执行：bash -x 03_trove_instance_upgrade.sh 使用upgrade彻底升级实例，除去人工痕迹
```

#!/usr/bin/env bash
# author: wangjun
# date: 2017.10.10
# Function: upgrade all trove instance
set -e 
set -x
source ../../upgrade_env
for value in $(cat $work_dir/common/get_trove_instance_info/trove_ins_ids.txt)
do
    echo $value
    TROVE_ID=$(echo $value |awk -F\; '{print $1}')
    TROVE_DATASTORE_VERSION=$(echo $value |awk -F\; '{print $2}')
    trove upgrade $TROVE_ID $TROVE_DATASTORE_VERSION
done
```

#####8、验证操作旧 trove 实例：

trove list
trove show ${ID}
trove secgroup-list
trove secgroup-show
trove backup-list
trave backup-list-instance
trove backup-show
trove user-create ${instance-id} wangjun trove.com --host xxxx  --databases xxxx
trove user-list ${instance-id}
新建trove datastore/trove实例等验证：
trove create ${instance-name} ${flavor} --size 10 --datastore ${datastore-name} --datastore_version ${datastore-version} --nic net-id=${net-id}
trove create ${instance-replica-name} ${flavor} --size 10 --datastore ${datastore-name} --datastore_version ${datastore-version} --nic net-id=${net-id}   --replica_of  ${master-id}
验证trove实例升级：
验证数据库是否正常运行，验证数据库表及数据
select * from instances where deleted = '0';
其他说明：
O 版本新特性：
specs: [https://specs.openstack.org/openstack/trove-specs/specs/newton/instance-upgrade.html](https://specs.openstack.org/openstack/trove-specs/specs/newton/instance-upgrade.html)
BP: [https://blueprints.launchpad.net/trove/+spec/instance-upgrade](https://blueprints.launchpad.net/trove/+spec/instance-upgrade)
Commit : [https://review.openstack.org/#/c/326064/](https://review.openstack.org/#/c/326064/)
