1.即使不参与stack生态的开发，coder也能从架构设计中学习很多东西，重复一句，这点很重要：
   佩服三点：
 - **设计模式**：接口的高度抽象、python的动态加载、驱动加载的可配置化让厂商轻松融入生态
 - **架构设计**：耗时任务有持久化的schedule机制处理，日志通过消息中间件处理来缓解I/O，各个大模块架构设计的高度相似性，CI/CD，git，gerrit......
 - **文档**：非常非常非常完善的[文档](https://docs.openstack.org/)

2.买一本书：英特尔开源小组出品的*《openstack设计与实现》*（最新版是17年5月的 [第二版](http://item.jd.com/12069413.html)），真的是一本好书，笔者曾经在菊长驻场开发，那里人手一本

3.订阅openstack-dev@lists.openstack.org
  openstack有IRC的meeting，但是meeting时间一般是每周2、3 北京时间凌晨1点-3点，大陆几乎没法参加，很多想法和发言没法及时通过IRC来沟通，只能看IRC的Log看大家讨论了什么，唯一剩下方式是通过向openstack-dev@lists.openstack.org发送邮件咨询一些问题，各个模块的leader会很热心的回答问题，笔者曾经提了一个ironic的需求，第二天就有个code reviewer回答了我的问题

4.如何修复BUG、提交代码
这个[链接](https://docs.openstack.org/infra/manual/developers.html )，非常详细的描述了整个过程：一般**小厂没时间和精力去做这个事情**，慢慢来吧

5.需求说明
这个[链接](https://specs.openstack.org/openstack)，里面归档了旧版本的需求以及正在实现的需求，以及即将要实现的需求

6.openstack的文档实在太丰富，文档的版本管理也很perfect，everthing is in the docs，完爆ceph社区

7.[oslo](https://wiki.openstack.org/wiki/Oslo)是openstack的依赖库，兼容python2、3，非常棒的填了python的坑

8.所有模块支持搭建单机版，可以通过UT触发某个接口执行，然后调试
举个例子，ironic模块的[one-step](https://docs.openstack.org/ironic/latest/contributor/dev-quickstart.html)

9.虽然是云，但是开发者很容易搭建环境，尤其是all in one vm/mac...
[devstack](https://docs.openstack.org/devstack/latest/)，跑UT也比较简单，举个ironic的[例子](https://docs.openstack.org/ironic/latest/contributor/dev-quickstart.html)
