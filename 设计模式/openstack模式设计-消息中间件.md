openstack的通信设计成两类：项目之间通过restful api进行通信；项目内部，不同服务进程之间必须通过消息中间件。这样一来可以保证接口的可扩展性和可靠性，用以支持大规模部署。
openstack通过oslo.messaging库来封装消息中间件，这样一来开发者只需要关注接口开发，而不需要关心底层的中间件是kafka还是rabbitMQ等开源消息中间件。
##oslo.messaging将消息中间件封装成两种方式来解决通信问题
- 远程过程调用（RPC）
- 事件通知
#####RPC
-通过远程过程调用，一个服务进程可以调用其他远程服务进程的方法，并且用两种调用方式：call和cast。
- 通过call的方式调用，远程方法会被同步执行，调用者会被阻塞直到结果返回
- 通过cast的方式调用，远程方法会被异步执行，结果并不会立即执行，调用者也不会被阻塞，但是调用者需要利用其他方式查询这次远程调用的结果
#####事件通知
某个服务进程可以把事件通知发送到消息中间件上，所有对此类事件感兴趣的服务进程都可以获得此事件的通知，并进一步处理，处理的结果并不会返回给事件发送者。通过这种通信方式，不但可以在同一个OS项目内部的各个服务之间发送通知，跨项目之间也能通知发送

##再谈网络编程
《UNIX网络编程》里面列出了五种I/O模型，感兴趣的可以阅读原著，也可以看看这篇[文章](http://www.cnblogs.com/chy2055/p/5220793.html)。把消息中间件封装成RPC的设计着实让我惊艳了一把，因为在RPC上，我吃了很多苦头。
- 自己实现RPC的苦
刚入职场加入了一个企业项目，属于 N客户端-单server的架构设计，没有架构师，所以我们把另一个成熟项目的架构复制过来了，立项时老板跟我说我们的单个企业客户不会顶多400-500用户，绝对不会超过1000，结果发布产品后就卖了一个接近1W用户的客户，性能瓶颈很快就出来了。
当时Windows客户端与server端的通信是公司架构组利用ACE封装的RPC实现，不支持集群，不支持高可用，而且还是消息阻塞式的，前面的消息没消费，后面的要等，那个1W员工的企业8点上班，服务器卡到爆
客户端也没法封装成异步获取消息，一个消息卡，客户端就假死了
- apache thrift
 文档很少
 支持多语言
 支持服务端多线程非阻塞式I/O，可以通过callback方式实现异步客户端
 市面上还有一些利用zookeeper服务发现的方式实现HA
 TCP/IP协议传输
- HTTP协议
交互方式简单，通用
只能通过轮询（poll）的方式获取耗时任务执行状态
websocket没考虑过

- gRPC
 http/2协议
 由于支持http，所以支持HAPROXY实现HA，LB
 大数据流的传输和控制
 TLS的支持

所有这些通信方式都是基于cache的方式，可能出现进程异常时丢失消息
消息中间件的方式性能上比thrift、gRPC差一点，但是有规模优势


参考：[openstack设计与实现](https://item.jd.com/12069413.html)
参考：[深入了解gRPC:协议](https://mp.weixin.qq.com/s?__biz=MzI3NDIxNTQyOQ==&mid=2247484946&idx=2&sn=f5d52103e363f9ca6a5facfa9ce55fb5&chksm=eb162178dc61a86e03eadc2eeb3f2a15831ae0bd178558ede3b98d507908ffd54dc25c3642c3&mpshare=1&scene=24&srcid=0920NqVWVat2y84W9W3DZGBF#rd)
参考：[gRPC 官方文档中文版](http://doc.oschina.net/grpc?t=56831)
参考：[TiDB与gRPC的那点事](http://www.infoq.com/cn/articles/tidb-and-grpc?utm_source=tuicool&utm_medium=referral)
