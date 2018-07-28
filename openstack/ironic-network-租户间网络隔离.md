## 1.版本支持
Ocata版本之前只支持flat网络，Ocata版本开始可以支持多租户网络。
## 2.如何配置服务来支持多租户特性
主要通过neutron来解决这个问题，具体配置步骤参考：[link](https://docs.openstack.org/ironic/latest/install/configure-tenant-networks.html#configure-tenant-networks)
## 3.生产支持多租户网络的节点
生产过程创建port时指定switch_id、port_id，具体配置参考：[link](https://docs.openstack.org/ironic/latest/install/configure-tenant-networks.html#configure-tenant-networks)
## 4.程序运行原理
整个运行流程大致如下：手工先简单将节点注册进ironic服务环境，horizon dashboard创建实例选择相应flavor和network 或者ports，ironic conductor开始接管剩下的事情，节点先被放置在生产网络进行生产（deploying  deploy-image and user-image），之后会被放置在租户网络。节点生产完毕后，不再支持PXE booting，这是因为无法访问TFTP服务器。所有在生产网络之外的节点都需要是local boot，一旦使用这个功能，实例的netboot需要被限制。
