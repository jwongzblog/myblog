mysql proxy是mysql官方开发的一款代理软件，主要达到对mysql server读写分离以及负载均衡，不过没有权重配置。源码4年前已停止更新，官方详细文档地址：[https://downloads.mysql.com/docs/mysql-proxy-en.pdf](https://downloads.mysql.com/docs/mysql-proxy-en.pdf)

github源码地址：https://github.com/mysql/mysql-proxy

# 工作原理如图所示：
![image.png](https://github.com/jwongzblog/myblog/blob/master/image/mysql-proxy.png)


# 源码安装
### 依赖
• libevent 1.4 or higher.
• lua 5.1.x only. lua 5.2.x is not currently supported.
• glib2 2.16.0 or higher.
• pkg-config.
• libtool 1.5.x or higher.
• autoconf 2.56 or higher.
• automake 1.10 or higher, up to 1.12.x. It is not possible to use version 1.13 or higher.
• flex 2.5.37.
• gtk-doc 1.18.
• MySQL 5.0.x or higher developer files.
### 安装
```
# tar zxf mysql-proxy-0.8.5.tar.gz
# cd mysql-proxy-0.8.5
# ./configure
# make
# make check
# make install
```
# 使用方式
### 命令行方式
```
# mysql-proxy --proxy-read-only-backend-addresses=192.168.0.1:3306 --proxy-read-only-backend-address=192.168.0.2:3306
or
# mysql-proxy --proxy-read-only-backend-addresses=192.168.0.1:3306,192.168.0.2:3306
# mysql-proxy --proxy-backend-addresses 192.168.0.3:3306 --proxy-backend-addresses 192.168.0.4:3306
or
# mysql-proxy --proxy-backend-addresses 192.168.0.3:3306,192.168.0.4:3306
```
### 配置文件方式：
```
#mysql-proxy  $configuration-file --defaults-file $file-location

[mysql-proxy]

user=root #运行mysql-proxy用户
admin-username=proxy #主从mysql共有的用户
admin-password=123.com #用户的密码
proxy-address=192.168.0.204:4000 #mysql-proxy运行ip和端口，不加端口，默认4040
proxy-read-only-backend-addresses=192.168.0.203 #指定后端从slave读取数据
proxy-backend-addresses=192.168.0.202 #指定后端主master写入数据
proxy-lua-script=/usr/local/mysql-proxy/lua/rw-splitting.lua #指定读写分离配置文件位置
admin-lua-script=/usr/local/mysql-proxy/lua/admin-sql.lua #指定管理脚本
log-file=/usr/local/mysql-proxy/logs/mysql-proxy.log #日志位置
log-level=info #定义log日志级别，由高到低分别有(error|warning|info|message|debug)
daemon=true    #以守护进程方式运行
keepalive=true  #负载较高时，mysql-proxy进程容易崩溃，使用keepalive确保HA
```
# 集成lua脚本
集成自己编写的lua脚本可以处理sql注入、分库分表的问题，有一定的开发量
```
if string.lower(command) == "show" and string.lower(option) == "querycounter" then
---
-- proxy.PROXY_SEND_RESULT requires
--
-- proxy.response.type to be either
-- * proxy.MYSQLD_PACKET_OK or
-- * proxy.MYSQLD_PACKET_ERR
--
-- for proxy.MYSQLD_PACKET_OK you need a resultset
-- * fields
-- * rows
--
-- for proxy.MYSQLD_PACKET_ERR
-- * errmsg
proxy.response.type = proxy.MYSQLD_PACKET_OK
proxy.response.resultset = {
              fields = {
                        { type = proxy.MYSQL_TYPE_LONG, name = "global_query_counter", },
                        { type = proxy.MYSQL_TYPE_LONG, name = "query_counter", },
},
              rows = {
                        { proxy.global.query_counter, query_counter }
              }
}
    -- we have our result, send it back
    return proxy.PROXY_SEND_RESULT
elseif string.lower(command) == "show" and string.lower(option) == "myerror" then
    proxy.response.type = proxy.MYSQLD_PACKET_ERR
    proxy.response.errmsg = "my first error"
    return proxy.PROXY_SEND_RESULT
```
参考至：

https://downloads.mysql.com/docs/mysql-proxy-en.pdf

http://blog.jobbole.com/94606/
