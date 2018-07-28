最近在做mysql5.7的产品化预研，碰到一点坑，在此说明一下。社区里面merge了一个patch(https://review.openstack.org/#/c/526728/)，代码改动量不大，其实改动最大的是mysql本身，所以在制作mysql5.7镜像的时候需要注意很多地方

#my.cnf的改动
- log_err变量名称改成log_error
- myisam-recover变量名称改成myisam-recover-options
- ……
#安全性
- mysql5.5、mysql5.6部署完毕后root密码为空，因此trove-guest-agent初始化的代码逻辑是默认mysql的root密码是空的，但是mysql5.7部署完毕后root密码是随机的，所以我们需要重新将root密码置为空，否则需要将随机生成的密码写进代码中
```
$mysql> update user set authentication_string = password(''), password_expired = 'N', password_last_changed = now() where user = 'root';
```
- mysql5.7默认安全validate_password plugin插件，这个会导致trove所有创建或者修改密码的字符串都要严格遵循密码安全等级，为此我们可以卸载这个插件
```
$mysql>  uninstall plugin validate_password;
```
#备份
trove创建从节点、备份等功能都依赖percona-xtrabackup工具，但是percona-xtrabackup在2.4.8及其以上版本才支持mysql5.7，因此我们需要更新该工具至合适的版本

#mysqld_safe
mysql5.7不再安装这个工具，但是trove创建从节点的时候却需要使用这个工具
```
For some Linux platforms, MySQL installation from RPM or Debian packages includes systemd support for managing MySQL server startup and shutdown.
On these platforms, [**mysqld_safe**](https://dev.mysql.com/doc/refman/5.7/en/mysqld-safe.html "4.3.2 mysqld_safe — MySQL Server Startup Script") is not installed because it is unnecessary. 
For more information, see [Section 2.5.10, “Managing MySQL Server with systemd”](https://dev.mysql.com/doc/refman/5.7/en/using-systemd.html "2.5.10 Managing MySQL Server with systemd").
```
```
    def _start_mysqld_safe_with_init_file(self, init_file, err_log_file):
        child = pexpect.spawn("sudo mysqld_safe"
                              " --skip-grant-tables"
                              " --skip-networking"
                              " --init-file='%s'"
                              " --log-error='%s'" %
                              (init_file.name, err_log_file.name)
                              )
```
略

#slave_running
slave_running被在5.7被移除了，需要特别处理一下，代码里面以此作为判断主从状态，因此需要调整成其他字段来过去状态

#detach replica
还有点不一样的在于从节点被移除后，主从状态的判断字段被清空，原来的代码实现逻辑不再试用
