tornado是一个完整的轻量级web框架，利用tornado即可完成一个网站的搭建，而openstack的web架构却采用另一种思路来设计。这个设计思路我之前提到过，也就是openstack的所有设计，都是插件式，组合式的。比如trove采用的是eventlet处理网络请求，paste、router处理路由和http协议，webob处理http协议返回值；而新项目如ironic，依然是eventlet处理网络请求，webob处理返回值，但是使用pecan来处理路由。总体上来讲是围绕wsgi协议的web框架，各自分工很明确，也可以被替代

# tornado与eventlet的比较
相同点：
- 是他们都是采用epoll网络模型
- 采用协程处理网络请求
- 很多地方没有重复造轮子，通过import python原生库来实现
- **本身不能实现异步的处理，也就是会blocking**

不同点：
- tornado利用yield来实现协程，而eventlet采用greenlet来实现协程，yield是python的关键字，而greenlet是c语言实现的函数上下文管理，比yield有更多的功能
- **tornado可以利用一些库来实现异步处理**，比如mongo的motor就是基于tornado的ioloop实现的异步客户端；而目前为止，我没发现evenlet有异步处理的实践
- tornado启动多个worker（进程）来监听不同的端口，需要借助nginx等中间件实现负载均衡，而eventlet的worker机制是主进程创建socket，其余的worker进程共享socket的方式（抢占式）实现负载

*在此，我在这里纠正一个我的错误，也许你们也有这个误解，就是我一直以为协程和线程一样，是并发式的、抢占式的，但是深入研究发现，协程是阻塞式的，队列式的，只不过通过轻松的切换协程的执行，看起来像是并发的，所以他的目标是跑满CPU*

# eventlet的用法
举个简单的创建webservice的例子
```
import eventlet
from eventlet import wsgi

def hello_world(env, start_response):
    if env['PATH_INFO'] != '/':
        start_response('404 Not Found', [('Content-Type', 'text/plain')])
        return ['Not Found\r\n']
    start_response('200 OK', [('Content-Type', 'text/plain')])
    return ['Hello, World!\r\n']

wsgi.server(eventlet.listen(('', 8090)), hello_world)
```
上面的例子是不是很简单？但是在工程实践中，需要考虑很多元素，以trove模块为例我们来看看openstack是怎么做的。

##### 主进程中socket的创建
trove/common/base-wsgi.py:
```
class Service(service.Service):
def __init__(self, application, port,
                 host='0.0.0.0', backlog=4096, threads=1000):
        ......
        self._socket = self._get_socket(host, port, self._backlog)

    def _get_socket(self, host, port, backlog):
        info = socket.getaddrinfo(host,
                                  port,
                                  socket.AF_UNSPEC,
                                  socket.SOCK_STREAM)[0]
        family = info[0]
        bind_addr = info[-1]
        sock = None

        while not sock and time.time() < retry_until:
            try:
                sock = eventlet.listen(bind_addr,
                                       backlog=backlog,
                                       family=family)
                if sslutils.is_enabled(CONF):
                    sock = sslutils.wrap(CONF, sock)
```
##### oslo.service
oslo.service库封装了eventlet进程的创建，根据配置文件的worker值创建子进程，如果没有配置，就默认为CPU数。
如/oslo.service/oslo_srevice/service.py：
```
    def launch_service(self, service, workers=1):
        ......
        while self.running and len(wrap.children) < wrap.workers:
            self._start_child(wrap)

    def _child_process(self, service):
        self._child_process_handle_signal()

        eventlet.hubs.use_hub()

        # Close write to ensure only parent has it open
        os.close(self.writepipe)
        # Create greenthread to watch for parent to close pipe
        eventlet.spawn_n(self._pipe_watcher)

        # Reseed random number generator
        random.seed()

        launcher = Launcher(self.conf, restart_method=self.restart_method)
        launcher.launch_service(service)
        return launcher

    def _start_child(self, wrap):
        if len(wrap.forktimes) > wrap.workers:
            if time.time() - wrap.forktimes[0] < wrap.workers:
                LOG.info('Forking too fast, sleeping')
                time.sleep(1)

            wrap.forktimes.pop(0)

        wrap.forktimes.append(time.time())

        pid = os.fork()
        if pid == 0:
            self.launcher = self._child_process(wrap.service)
            while True:
                self._child_process_handle_signal()
                status, signo = self._child_wait_for_exit_or_signal(
                    self.launcher)
                if not _is_sighup_and_daemon(signo):
                    self.launcher.wait()
                    break
                self.launcher.restart()

            os._exit(status)

        wrap.children.add(pid)
        self.children[pid] = wrap

        return pid
```
###### eventlet处理socket
几个进程accept同一个socket，谁抢到谁处理
```
def server(sock, site,
           log=None,
           environ=None,
           max_size=None,
           max_http_version=DEFAULT_MAX_HTTP_VERSION,
           protocol=HttpProtocol,
           server_event=None,
           minimum_chunk_size=None,
           log_x_forwarded_for=True,
           custom_pool=None,
           keepalive=True,
           log_output=True,
           log_format=DEFAULT_LOG_FORMAT,
           url_length_limit=MAX_REQUEST_LINE,
           debug=True,
           socket_timeout=None,
           capitalize_response_headers=True):

    ......
    try:
        serv.log.info('({0}) wsgi starting up on {1}'.format(serv.pid, socket_repr(sock)))
        while is_accepting:
            try:
                client_socket, client_addr = sock.accept()
                client_socket.settimeout(serv.socket_timeout)
                serv.log.debug('({0}) accepted {1!r}'.format(serv.pid, client_addr))
                connections[client_addr] = connection = [client_addr, client_socket, STATE_IDLE]
                (pool.spawn(serv.process_request, connection)
                    .link(_clean_connection, connection))
            except ACCEPT_EXCEPTIONS as e:
                if support.get_errno(e) not in ACCEPT_ERRNO:
                    raise
            except (KeyboardInterrupt, SystemExit):
                serv.log.info('wsgi exiting')
                break
    finally:    
```
##### eventlet协程处理
上面的代码留意这一段：
```
pool.spawn(serv.process_request, connection)
```
这里是实现是将accept到的请求塞进协程里，然后塞进协程池里，而协程池由一个hub module集中管控处理，也有timeout处理。协程没消费掉，是不会让出CPU的，除非人为的调用eventlet.sleep()，把CPU让出来。
协程的切换通过switch函数来实现，举例如下：
```
from greenlet import greenlet

def test1():
    print(12)
    gr2.switch()
    print(34)

def test2():
    print(56)
    gr1.switch()
    print(78)

gr1 = greenlet(test1)
gr2 = greenlet(test2)
gr1.switch()
```
**The last line jumps to test1, which prints 12, jumps to test2, prints 56, jumps back into test1, prints 34; and then test1 finishes and gr1 dies. At this point, the execution comes back to the original gr1.switch() call. Note that 78 is never printed.**
至此，eventlet的核心功能分析完毕，有兴趣深入的同学可以继续看源码。我看过的所有timeout的处理都是在while里面自检。。。有更好的想法么？
