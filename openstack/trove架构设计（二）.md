在[trove架构设计（一）](https://github.com/jwongzblog/myblog/blob/master/openstack/trove%E6%9E%B6%E6%9E%84%E8%AE%BE%E8%AE%A1%EF%BC%88%E4%B8%80%EF%BC%89.md)中讲举例了一个最简单的call模式（同步），这篇介绍复杂点的cast模式（异步模式），看看trove在处理耗时、等待的任务请求时如何处理
看看下图的工作流

![image.png](https://github.com/jwongzblog/myblog/blob/master/openstack/trove-arch3.png)

上面的图不顺畅的话再看看这个时序图

![image.png](https://github.com/jwongzblog/myblog/blob/master/openstack/trove-arch4.png)
我们举个创建实例的例子（trove create）
1.接受创建实例的请求。/trove/instance/service.py
```
class InstanceController(wsgi.Controller):
    ......
    def create(self, req, body, tenant_id):
        #此处省略若干参数准备的代码
        ......
        instance = models.Instance.create(context, name, flavor_id,
                                          image_id, databases, users,
                                          datastore, datastore_version,
                                          volume_size, backup_id,
                                          availability_zone, nics,
                                          configuration, slave_of_id,
                                          replica_count=replica_count,
                                          volume_type=volume_type,
                                          modules=modules,
                                          locality=locality,
                                          region_name=region_name)
        view = views.InstanceDetailView(instance, req=req)
        return wsgi.Result(view.data(), 200)
```
2.逻辑处理放在了同级目录的models.py，给程序员举了一个代码可读性的例子，同级目录还有个views.py，是专门处理返回值的，大家可以参考一下。/trove/instance/models.py
```
class Instance(BuiltInstance):
   @classmethod
    def create(cls, context, name, flavor_id, image_id, databases, users,
               datastore, datastore_version, volume_size, backup_id,
               availability_zone=None, nics=None,
               configuration_id=None, slave_of_id=None, cluster_config=None,
               replica_count=None, volume_type=None, modules=None,
               locality=None, region_name=None):
        #由于创建新实例、创建从节点、根据备份ID创建实例都集中在此api中，所以此处的逻辑判断略复杂
        ......
        #向taskmanager服务发送创建实例的请求
        task_api.API(context).create_instance(
            instance_id, instance_name, flavor, image_id, databases, users,
            datastore_version.manager, datastore_version.packages,
            volume_size, backup_id, availability_zone, root_password,
            nics, overrides, slave_of_id, cluster_config,
            volume_type=volume_type, modules=module_list,
            locality=locality)
```
3.taskmanager是一个基于MQ的RPC服务。/trove/taskmanager/api.py封装了RPC的client端
```
def create_instance(self, instance_id, name, flavor,
                    image_id, databases, users, datastore_manager,
                    packages, volume_size, backup_id=None,
                    availability_zone=None, root_password=None,
                    nics=None, overrides=None, slave_of_id=None,
                    cluster_config=None, volume_type=None,
                    modules=None, locality=None):
    LOG.debug("Making async call to create instance %s " % instance_id)
    version = self.API_BASE_VERSION
    #此处是一个异步请求，return
    self._cast("create_instance", version=version,
               instance_id=instance_id, name=name,
               flavor=self._transform_obj(flavor),
               image_id=image_id,
               databases=databases,
               users=users,
               datastore_manager=datastore_manager,
               packages=packages,
               volume_size=volume_size,
               backup_id=backup_id,
               availability_zone=availability_zone,
               root_password=root_password,
               nics=nics,
               overrides=overrides,
               slave_of_id=slave_of_id,
               cluster_config=cluster_config,
               volume_type=volume_type,
               modules=modules, locality=locality)
```
4.taskmanager service在接受到请求后，开始触发创建实例，完成调度后return
```
#/trove/taskmanager/manager.py
class Manager(periodic_task.PeriodicTasks):
    ......
    def create_instance(self, context, instance_id, name, flavor,
                        image_id, databases, users, datastore_manager,
                        packages, volume_size, backup_id, availability_zone,
                        root_password, nics, overrides, slave_of_id,
                        cluster_config, volume_type, modules, locality):
        with EndNotification(context,
                             instance_id=(instance_id[0]
                                          if isinstance(instance_id, list)
                                          else instance_id)):
            self._create_instance(context, instance_id, name, flavor,
                                  image_id, databases, users,
                                  datastore_manager, packages, volume_size,
                                  backup_id, availability_zone,
                                  root_password, nics, overrides, slave_of_id,
                                  cluster_config, volume_type, modules,
                                  locality)
    ......
    #创建从节点还是主节点
    def _create_instance(self, context, instance_id, name, flavor,
                         image_id, databases, users, datastore_manager,
                         packages, volume_size, backup_id, availability_zone,
                         root_password, nics, overrides, slave_of_id,
                         cluster_config, volume_type, modules, locality):
        if slave_of_id:
            self._create_replication_slave(context, instance_id, name,
                                           flavor, image_id, databases, users,
                                           datastore_manager, packages,
                                           volume_size,
                                           availability_zone, root_password,
                                           nics, overrides, slave_of_id,
                                           backup_id, volume_type, modules)
        else:
            if type(instance_id) in [list]:
                raise AttributeError(_(
                    "Cannot create multiple non-replica instances."))
            instance_tasks = FreshInstanceTasks.load(context, instance_id)
            scheduler_hints = srv_grp.ServerGroup.build_scheduler_hint(
                context, locality, instance_id)
            instance_tasks.create_instance(flavor, image_id, databases, users,
                                           datastore_manager, packages,
                                           volume_size, backup_id,
                                           availability_zone, root_password,
                                           nics, overrides, cluster_config,
                                           None, volume_type, modules,
                                           scheduler_hints)
            timeout = (CONF.restore_usage_timeout if backup_id
                       else CONF.usage_timeout)
            instance_tasks.wait_for_instance(timeout, flavor)
#/trove/taskmanager/models.py此处可以慢慢的处理实例创建，一点不着急
class FreshInstanceTasks(FreshInstance, NotifyMixin, ConfigurationMixin):
    #......创建各个模块的client端，开始调度，此处留意一个细节，就是guest_info.conf的内容修改在/instance/models.py中实现了
    def create_instance(self, flavor, image_id, databases, users,
                        datastore_manager, packages, volume_size,
                        backup_id, availability_zone, root_password, nics,
                        overrides, cluster_config, snapshot, volume_type,
                        modules, scheduler_hints):
        ........
        #此处最为关键，实例创建好以后往消息队列发送prepare消息，以便实例启动后guest-agent启动后会监听到prepare的请求，开始连上trove-conductor工作
        self._guest_prepare(flavor['ram'], volume_info,
                            packages, databases, users, backup_info,
                            config.config_contents, root_password,
                            overrides,
                            cluster_config, snapshot, modules)
    #发送cast请求
    def _guest_prepare(self, flavor_ram, volume_info,
                       packages, databases, users, backup_info=None,
                       config_contents=None, root_password=None,
                       overrides=None, cluster_config=None, snapshot=None,
                       modules=None):
        LOG.debug("Entering guest_prepare")
        # Now wait for the response from the create to do additional work
        self.guest.prepare(flavor_ram, packages, databases, users,
                           device_path=volume_info['device_path'],
                           mount_point=volume_info['mount_point'],
                           backup_info=backup_info,
                           config_contents=config_contents,
                           root_password=root_password,
                           overrides=overrides,
                           cluster_config=cluster_config,
                           snapshot=snapshot, modules=modules)
```
5.trove-guestagent服务接受到prepare请求后，开始触发各类数据库驱动工作，比如mysql
```
#trove guestagent服务启动 /trove/cmd/guest.py
def main():
    ......
    from trove.guestagent import dbaas
    #此处决定guestagent加载哪种类型数据库的驱动
    manager = dbaas.datastore_registry().get(CONF.datastore_manager)
    ......
#mysql  /trove/guestagent/dbaas.py
defaults = {
    'mysql':
    'trove.guestagent.datastore.mysql.manager.Manager',
    ......
}
#mysql使用了基类的prepare方法/trove/guestagent/datastore/manager.py
class Manager(periodic_task.PeriodicTasks):
    ......
    def prepare(self, context, packages, databases, memory_mb, users,
                device_path=None, mount_point=None, backup_info=None,
                config_contents=None, root_password=None, overrides=None,
                cluster_config=None, snapshot=None, modules=None):
        """Set up datastore on a Guest Instance."""
        with EndNotification(context, instance_id=CONF.guest_id):
            self._prepare(context, packages, databases, memory_mb, users,
                          device_path, mount_point, backup_info,
                          config_contents, root_password, overrides,
                          cluster_config, snapshot, modules)
    def _prepare(self, context, packages, databases, memory_mb, users,
                 device_path, mount_point, backup_info,
                 config_contents, root_password, overrides,
                 cluster_config, snapshot, modules):
        ......
        #此处会通过conductor-MQ-RPC开放的client去修改状态
        self.status.end_install(error_occurred=self.prepare_error,
                                    post_processing=post_processing)
#/trove/guestagent/datastore/service.py
    def set_status(self, status, force=False):
        """Use conductor to update the DB app status."""
        if force or self.is_installed:
            LOG.debug("Casting set_status message to conductor "
                      "(status is '%s')." % status.description)
            context = trove_context.TroveContext()
            heartbeat = {'service_status': status.description}
            conductor_api.API(context).heartbeat(
                CONF.guest_id, heartbeat, sent=timeutils.float_utcnow())
            LOG.debug("Successfully cast set_status.")
            self.status = status
        else:
            LOG.debug("Prepare has not completed yet, skipping heartbeat.")
```
6.conductor服务监听到消息后修改数据库数据。/trove/conductor/manager.py
```

class Manager(periodic_task.PeriodicTasks):
    def heartbeat(self, context, instance_id, payload, sent=None):
        LOG.debug("Instance ID: %(instance)s, Payload: %(payload)s" %
                  {"instance": str(instance_id),
                   "payload": str(payload)})
        status = inst_models.InstanceServiceStatus.find_by(
            instance_id=instance_id)
        if self._message_too_old(instance_id, 'heartbeat', sent):
            return
        if payload.get('service_status') is not None:
            status.set_status(ServiceStatus.from_description(
                payload['service_status']))
        status.save()
```
至此，整个异步模式的流程梳理完毕。如果你看懂了，那么openstack其他模块的架构你也就懂的差不多了。openstack从开源起，逐渐PK掉其他的开源云平台，与python的易读性有很大的关系，它只是一个资源管理平台而已，核心还是在各个模块的深层领域知识。我个人觉得这个架构的最大优点是能够从容的应对同步、异步模型，并且领域划分清晰，符合松耦合的特征
