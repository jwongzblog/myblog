产品的用户管理需要兼容windows AD，所以使用了openldap进行二次开发导入windows AD组织单位及域用户。windows AD和openldap都是基于LDAP协议实现，所以兼容不成问题。windows SDK自带域控访问API，perl、python、php我都见过有库函数去操作AD。

- 老外写了一份demo程序获取域控信息，基本可用，先使用系统管理员bind LD，再设置optional，然后使用simple-search即可遍历，search可以写查询语句。不过这个demo程序使用的是旧版本接口，在编译的时候需要处理宏。
- 碰到问题可以把ldap-error打出来，stackoverflow很多牛逼的程序员有作答，貌似还有谁整理了一篇error FAQ，只怪我当时预研的时候没留下来
- 使用openldap修改密码可以使用modify这个接口，但是由于权限问题是不能直接调用接口修改的，需要把域控的证书导出来放进你的系统
- 如果域控没做设置，默认一次只能搜索出1000条信息，这个值是可以在域控制器设置的，具体方法可以搜到
- 有一点很奇怪，这个库能获取到windows AD域的信息，但是我使用它来获取openldap搭建的域控制器时却始终bind不了，由于没这样的需求后来没细究了
- 参考官方的文档，实在不行去stackoverflow提问。

几年前的项目了，细节就不详述
