作为一个不间断的服务进程，如何让运维、开发迅速定位问题、排查问题，那就是提供一套日志系统。作为曾经的c++程序员，自己封装过一个日志组件，也用过一些开源的日志组件，但是都无法做到像ceph的日志系统那样，设计得如此优雅。

## 功能层面：
##### 配置
ceph的各个进程在启动时使用默认参数，如果需要刻意修改参数，可以在配置文件指定对应参数。默认情况下，各类模块的日志等级为1，即输出最基本的信息。如果程序员想通过日志DEBUG程序，可以把日志开到20档，函数的调用都能打印。ceph的各个进程在启动时使用默认参数，如果需要刻意修改参数，可以在配置文件指定对应参数。默认情况下，各类模块的日志等级为1，即输出最基本的信息。如果程序员想通过日志DEBUG程序，可以把日志开到20档，函数的调用都能打印。

##### 在线修改日志的输出等级：
```
ceph daemon /var/run/ceph/ceph-osd.0.asok config set debug_osd 15/15
```
嫌弃日志输出影响性能，可以关闭日志：
```
ceph daemon /var/run/ceph/ceph-osd.0.asok config set debug_osd 0/0
```
##### 子系统概览
各子系统都有日志级别用于分别控制其输出日志、和暂存日志，你可以分别为这些子系统设置不同的记录级别。 Ceph 的日志级别从 1 到 20 ， 1 是简洁、 20 是详尽。通常，内存驻留日志不会发送到输出日志，除非：
- 致命信号冒出来了
- 源码中的 assert 被触发
- 明确要求发送
调试选项允许用单个数字同时设置日志级别和内存级别，会设置为一样。比如，如果你指定 debug ms = 5 ， Ceph 会把日志级别和内存级别都设置为 5 。也可以分别设置，第一个选项是日志级别、后一个是内存级别，二者必须用斜线（ / ）分隔。假如你想把 ms 子系统的调试日志级别设为 1 、内存级别设为 5 ，可以写为 debug ms = 1/5 

##### 输出syslog协议格式
举例：
```
2019-02-22 16:30:08.876939 osd.46 10.3.13.68:6810/3722505 4160 : cluster [WRN] slow request 31.767952 seconds old, received at 2019-02-22 16:29:37.108770: osd_op(client.3128660.0:771979 9.70ebd1ee rbd_object_map.2ff4bb670932db [call lock.assert_locked,call rbd.object_map_update] snapc 0=[] ack+ondisk+write+known_if_redirected e6485) currently waiting for rw locks
```

## 代码层面：
比较让我惊艳的是ceph的代码对日志的封装，对程序员实在太友好了。举例：
我要输出一个等级为10的告警，我只需要像下面这样写一句代码即可，dout是一个宏，把所有的细节都封装好了，我们用起来只需像拼字符串那样就行
```
dout(10) << "error completing split on " << path << ": "
	     << cpp_strerror(r) << dendl;		       
```
dout宏的具体实现：：
```
#define dout(v) ldout((g_ceph_context), v)

#define ldout(cct, v)  dout_impl(cct, dout_subsys, v) dout_prefix

#define dout_prefix *_dout

#define dout_impl(cct, sub, v)						\
  do {									\
  if (cct->_conf->subsys.should_gather(sub, v)) {			\
    if (0) {								\
      char __array[((v >= -1) && (v <= 200)) ? 0 : -1] __attribute__((unused)); \
    }									\
    static size_t _log_exp_length=80; \
    ceph::log::Entry *_dout_e = cct->_log->create_entry(v, sub, &_log_exp_length);	\
    ostream _dout_os(&_dout_e->m_streambuf);				\
    CephContext *_dout_cct = cct;					\
    std::ostream* _dout = &_dout_os;
// flush	
#define dendl std::flush;				\
  _ASSERT_H->_log->submit_entry(_dout_e);		\
    }						\
  } while (0)
```