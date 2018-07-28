trove组件总共分为trove api、trove conductor、trove taskmanager以及部署在实例内部的trove gust agent，他们之间的通信规则如下图所示

![image.png](http://upload-images.jianshu.io/upload_images/5945542-4dc91ae219b60f4b.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

先以trove show为例，讲解一下最为简单的call模式

![image.png](http://upload-images.jianshu.io/upload_images/5945542-c759429006552219.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)
下面的代码show了一把消息中间-RPC的用法，代码逻辑如下：
1.api响应。/trove/instance/service.py
```
class InstanceController(wsgi.Controller):
    ......
    def show(self, req, tenant_id, id):
        """Return a single instance."""
        LOG.info(_LI("Showing database instance '%(instance_id)s' for tenant "
                     "'%(tenant_id)s'"),
                 {'instance_id': id, 'tenant_id': tenant_id})
        LOG.debug("req : '%s'\n\n", req)
        context = req.environ[wsgi.CONTEXT_KEY]
        server = models.load_instance_with_info(models.DetailInstance,
                                                context, id)
        self.authorize_instance_action(context, 'show', server)
        return wsgi.Result(views.InstanceDetailView(server,
                                                    req=req).data(), 200)
    ......
```
2.向guest-agent发送请求。/trove/instance/models.py
```
def load_guest_info(instance, context, id):
    if instance.status not in AGENT_INVALID_STATUSES:
        guest = create_guest_client(context, id)
        try:
            #get_volume_info是基于MQ实现的RPC函数调用
            volume_info = guest.get_volume_info()
            instance.volume_used = volume_info['used']
            instance.volume_total = volume_info['total']
        except Exception as e:
            LOG.exception(e)
    return instance
```
3.向message queue发送同步（call）请求。/trove/instance/remote.py
```
def guest_client(context, id, manager=None):
    from trove.guestagent.api import API
    if manager:
        clazz = strategy.load_guestagent_strategy(manager).guest_client_class
    else:
        clazz = API
    return clazz(context, id)

```
4.guest-agent对外暴露的client。/trove/guestagent/api.py
```
    def get_volume_info(self):
        """Make a synchronous call to get volume info for the container."""
        LOG.debug("Check Volume Info on instance %s.", self.id)
        version = self.API_BASE_VERSION
        return self._call("get_filesystem_stats", AGENT_LOW_TIMEOUT,
                          version=version, fs_path=None)
    #封装消息
    def _call(self, method_name, timeout_sec, version, **kwargs):
        LOG.debug("Calling %s with timeout %s" % (method_name, timeout_sec))
        try:
            cctxt = self.client.prepare(version=version, timeout=timeout_sec)
            result = cctxt.call(self.context, method_name, **kwargs)
            LOG.debug("Result is %s." % result)
            return result
        except RemoteError as r:
            LOG.exception(_("Error calling %s") % method_name)
            raise exception.GuestError(original_message=r.value)
        except Exception as e:
            LOG.exception(_("Error calling %s") % method_name)
            raise exception.GuestError(original_message=str(e))
        except Timeout:
            raise exception.GuestTimeout()
```
5.guest-agent服务启动/trove/cmd/guest.py
```
def main():
    cfg.parse_args(sys.argv)
    logging.setup(CONF, None)
    debug_utils.setup()
    from trove.guestagent import dbaas
    #此处会把定义的响应函数添加进来，映射到RPC中
    manager = dbaas.datastore_registry().get(CONF.datastore_manager)
    ......
    from trove.common.rpc import service as rpc_service
    server = rpc_service.RpcService(
        key=CONF.instance_rpc_encr_key,
        topic="guestagent.%s" % CONF.guest_id,
        manager=manager, host=CONF.guest_id,
        rpc_api_version=guest_api.API.API_LATEST_VERSION)
    launcher = openstack_service.launch(CONF, server)
    launcher.wait()
```
6.guest-agent响应基类。/trove/guestagent/datastore/manage.py
```
class Manager(periodic_task.PeriodicTasks):
    ......
    #子类继承了改方法，在一些特殊的数据库里会override这个方法
    def get_filesystem_stats(self, context, fs_path):
        """Gets the filesystem stats for the path given."""
        # TODO(peterstac) - note that fs_path is not used in this method.
        mount_point = CONF.get(self.manager).mount_point
        LOG.debug("Getting file system stats for '%s'" % mount_point)
        return dbaas.get_filesystem_volume_stats(mount_point)
```
至此，整个trove show的流程完毕
