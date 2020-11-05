based on tornado 4.5.2

1999 年，Dan Kegel 向网络服务器提出了一个骇人听闻的难题:**是时候让网络服务器去同时应对 10000 个客户端了，你觉得呢？毕竟网络已经变得很普及了。**这就是著名的 [C10K](http://www.kegel.com/c10k.html) 问题，如今nginx、apache等开源软件已经能够轻松应对10K+，而facebook在2009年的时候开源的python tornado，也号称解决了该问题。

tornado代码不足一万五千行，麻雀虽小五脏俱全，使用他的异步模型，代码复杂一点点，但无需依赖nginx就具备高并发能力，这得益于他的coroutine（协程）装饰器以及epoll网络模型，coroutine可以让我们使用同步的语义实现异步的功能，代码可读性大大提升，具体原理后面再讲。

tornado网络通信采用epoll模型，简单介绍一下epoll，在《[openstack模式设计-消息中间件](https://github.com/jwongzblog/myblog/blob/master/%E8%AE%BE%E8%AE%A1%E6%A8%A1%E5%BC%8F/openstack%E6%A8%A1%E5%BC%8F%E8%AE%BE%E8%AE%A1-%E6%B6%88%E6%81%AF%E4%B8%AD%E9%97%B4%E4%BB%B6.md)》中简单提到了I/O模型，epoll是Linux下多路复用IO接口select/poll的增强版本，它能显著提高程序在大量并发连接中只有少量活跃的情况下的系统CPU利用率。举个简单的例子，比如你在京东上想买个东西，结果没货，你会怎么做？肯定不会是每隔一阵子去看看到货了没，正确的做法是点击“到货通知”，让京东的系统在货物到仓时及时短信通知你。epoll也是这样，server端在处理网络请求是，无需等待数据是否写满buffer，只需将这些socket以红黑树（一种数据结构）的方式写进内核里，另外还有epoll_wait会定时检查这些注册的事件对应的数据是否准备就绪（链表存储），如果好了就唤醒对应的回调函数。

回到协程，看看tornado建议程序员的写法
```
# tornado异步处理的写法
class AsyncHandler(RequestHandler):
    @asynchronous
    def get(self):
        http_client = AsyncHTTPClient()
        http_client.fetch("http://example.com",
                          callback=self.on_fetch)

    def on_fetch(self, response):
        do_something_with_response(response)
        self.render("template.html")
```
```
#tornado协程的写法，同步的写法达到异步的效果
class GenAsyncHandler(RequestHandler):
    @gen.coroutine
    def get(self):
        http_client = AsyncHTTPClient()
        response = yield http_client.fetch("http://example.com")
        do_something_with_response(response)
        self.render("template.html")
```
下面的源码是coroutine的实现，tornado在python2.7里，协程实现的本质是通过yield将调用的函数变成一个生成器（generator），等调用next的时候唤醒函数本身（在这里再多说一句yield，yield原来是python2.7的关键字，专门用来处理iterator的语义，他的实现原理是将yield关键字后面的对象（函数也是对象）的上下文记录在堆栈，然后将返回值交给外部逻辑，不会阻塞进程，由外部决定何时唤醒这段对象/程序，据说一个线程占8M内存，而协程不到1KB）
```
def coroutine(func, replace_callback=True):
    return _make_coroutine_wrapper(func, replace_callback=True)

def _make_coroutine_wrapper(func, replace_callback):
    wrapped = func
    if hasattr(types, 'coroutine'):
        func = types.coroutine(func)

    @functools.wraps(wrapped)
    def wrapper(*args, **kwargs):
        future = TracebackFuture()

        if replace_callback and 'callback' in kwargs:
            callback = kwargs.pop('callback')
            IOLoop.current().add_future(
                future, lambda future: callback(future.result()))

        try:
            result = func(*args, **kwargs)
        except (Return, StopIteration) as e:
            result = _value_from_stopiteration(e)
        except Exception:
            future.set_exc_info(sys.exc_info())
            return future
        else:
            if isinstance(result, GeneratorType):
                try:
                    orig_stack_contexts = stack_context._state.contexts
                    yielded = next(result)
                    if stack_context._state.contexts is not orig_stack_contexts:
                        yielded = TracebackFuture()
                        yielded.set_exception(
                            stack_context.StackContextInconsistentError(
                                'stack_context inconsistency (probably caused '
                                'by yield within a "with StackContext" block)'))
                except (StopIteration, Return) as e:
                    future.set_result(_value_from_stopiteration(e))
                except Exception:
                    future.set_exc_info(sys.exc_info())
                else:
                    _futures_to_runners[future] = Runner(result, future, yielded)
                yielded = None
                try:
                    return future
                finally:
                    future = None
        future.set_result(result)
        return future

    wrapper.__wrapped__ = wrapped
    wrapper.__tornado_coroutine__ = True
    return wrapper
```
上面贴出来coroutine的实现，通过next取出generator后丢给了Runner
```
_futures_to_runners[future] = Runner(result, future, yielded)
```
我们再看看Runner的实现
```
class Runner(object):
    def __init__(self, gen, result_future, first_yielded):
        if self.handle_yield(first_yielded):
            gen = result_future = first_yielded = None
            self.run()
    def handle_yield(self, yielded):
        # Lists containing YieldPoints require stack contexts;
        # other lists are handled in convert_yielded.
        if _contains_yieldpoint(yielded):
            yielded = multi(yielded)

        if isinstance(yielded, YieldPoint):
            # YieldPoints are too closely coupled to the Runner to go
            # through the generic convert_yielded mechanism.
            self.future = TracebackFuture()

            def start_yield_point():
                try:
                    yielded.start(self)
                    if yielded.is_ready():
                        self.future.set_result(
                            yielded.get_result())
                    else:
                        self.yield_point = yielded
                except Exception:
                    self.future = TracebackFuture()
                    self.future.set_exc_info(sys.exc_info())

            if self.stack_context_deactivate is None:
                # Start a stack context if this is the first
                # YieldPoint we've seen.
                with stack_context.ExceptionStackContext(
                        self.handle_exception) as deactivate:
                    self.stack_context_deactivate = deactivate

                    def cb():
                        start_yield_point()
                        self.run()
                    self.io_loop.add_callback(cb)
                    return False
            else:
                start_yield_point()
        else:
            try:
                self.future = convert_yielded(yielded)
            except BadYieldError:
                self.future = TracebackFuture()
                self.future.set_exc_info(sys.exc_info())

        if not self.future.done() or self.future is moment:
            def inner(f):
                # Break a reference cycle to speed GC.
                f = None # noqa
                self.run()
            self.io_loop.add_future(
                self.future, inner)
            return False
        return True
    def run(self):
        """Starts or resumes the generator, running until it reaches a
        yield point that is not ready.
        """
        if self.running or self.finished:
            return
        try:
            self.running = True
            while True:
                future = self.future
                if not future.done():
                    return
                self.future = None
                try:
                    orig_stack_contexts = stack_context._state.contexts
                    exc_info = None

                    try:
                        value = future.result()
                    except Exception:
                        self.had_exception = True
                        exc_info = sys.exc_info()
                    future = None

                    if exc_info is not None:
                        try:
                            yielded = self.gen.throw(*exc_info)
                        finally:
                            # Break up a reference to itself
                            # for faster GC on CPython.
                            exc_info = None
                    else:
                        yielded = self.gen.send(value)

                    if stack_context._state.contexts is not orig_stack_contexts:
                        self.gen.throw(
                            stack_context.StackContextInconsistentError(
                                'stack_context inconsistency (probably caused '
                                'by yield within a "with StackContext" block)'))
                except (StopIteration, Return) as e:
                    self.finished = True
                    self.future = _null_future
                    if self.pending_callbacks and not self.had_exception:
                        # If we ran cleanly without waiting on all callbacks
                        # raise an error (really more of a warning).  If we
                        # had an exception then some callbacks may have been
                        # orphaned, so skip the check in that case.
                        raise LeakedCallbackError(
                            "finished without waiting for callbacks %r" %
                            self.pending_callbacks)
                    self.result_future.set_result(_value_from_stopiteration(e))
                    self.result_future = None
                    self._deactivate_stack_context()
                    return
                except Exception:
                    self.finished = True
                    self.future = _null_future
                    self.result_future.set_exc_info(sys.exc_info())
                    self.result_future = None
                    self._deactivate_stack_context()
                    return
                if not self.handle_yield(yielded):
                    return
                yielded = None
        finally:
            self.running = False
```
这个Runner不停的在检索result是否OK，一旦generator有返回值，服务通过set_result设置检索future的状态。

但是，**请注意**，如果被调用的函数本身是阻塞的，加了coroutine装饰器也没有，基于tornado的web服务依然会阻塞，所以在处理耗时操作时需要依赖tornado的ioloop来实现一些异步的操作，比如这些[库](https://github.com/tornadoweb/tornado/wiki/Links)，实现了mongo、redis、mysql等的client端，以便你的tornado web service调用。
以下是来自官方的高并发建议：
```
In general, you should think about IO strategies for tornado apps in this order:

Use an async library if available (e.g. AsyncHTTPClient instead of requests).

Make it so fast you don't mind doing it synchronously and blocking the IOLoop. This is most appropriate for things like memcache and database queries that are under your control and should always be fast. If it's not fast, make it fast by adding the appropriate indexes to the database, etc.

Do the work in a ThreadPoolExecutor. Remember that worker threads cannot access the IOLoop (even indirectly) so you must return to the main thread before writing any responses.

Move the work out of the tornado process. If you're sending email, for example, just write it to the database and let another process (whose latency doesn't matter) read from the queue and do the actual sending.

Block the IOLoop anyway. This is the lazy way out but may be acceptable in some cases.
```

到此tornado单进程的协程用法展示完毕，那么如何在这个单进程里启动多线程来处理协程了？此处为胡俊同学整理，
![image.png](https://github.com/jwongzblog/myblog/blob/master/python/python-tornado.png)

上面的代码利用concurrent.futures，在python2.7中需要download库，但是在python3中集成了该库，这也是官方推荐的处理异步的方式

上面利用了tornado的with_timeout，下面是这个函数的实现，从调试结果来看，运行过程中，concurrent.futures会替换掉tornado内部实现的Future类
```
def with_timeout(timeout, future, quiet_exceptions=()):
    """Wraps a `.Future` (or other yieldable object) in a timeout.

    Raises `tornado.util.TimeoutError` if the input future does not
    complete before ``timeout``, which may be specified in any form
    allowed by `.IOLoop.add_timeout` (i.e. a `datetime.timedelta` or
    an absolute time relative to `.IOLoop.time`)

    If the wrapped `.Future` fails after it has timed out, the exception
    will be logged unless it is of a type contained in ``quiet_exceptions``
    (which may be an exception type or a sequence of types).

    Does not support `YieldPoint` subclasses.

    .. versionadded:: 4.0

    .. versionchanged:: 4.1
       Added the ``quiet_exceptions`` argument and the logging of unhandled
       exceptions.

    .. versionchanged:: 4.4
       Added support for yieldable objects other than `.Future`.
    """
    # TODO: allow YieldPoints in addition to other yieldables?
    # Tricky to do with stack_context semantics.
    #
    # It's tempting to optimize this by cancelling the input future on timeout
    # instead of creating a new one, but A) we can't know if we are the only
    # one waiting on the input future, so cancelling it might disrupt other
    # callers and B) concurrent futures can only be cancelled while they are
    # in the queue, so cancellation cannot reliably bound our waiting time.
    future = convert_yielded(future)
    result = Future()
    chain_future(future, result)
    io_loop = IOLoop.current()

    def error_callback(future):
        try:
            future.result()
        except Exception as e:
            if not isinstance(e, quiet_exceptions):
                app_log.error("Exception in Future %r after timeout",
                              future, exc_info=True)

    def timeout_callback():
        result.set_exception(TimeoutError("Timeout"))
        # In case the wrapped future goes on to fail, log it.
        future.add_done_callback(error_callback)
    timeout_handle = io_loop.add_timeout(
        timeout, timeout_callback)
    if isinstance(future, Future):
        # We know this future will resolve on the IOLoop, so we don't
        # need the extra thread-safety of IOLoop.add_future (and we also
        # don't care about StackContext here.
        future.add_done_callback(
            lambda future: io_loop.remove_timeout(timeout_handle))
    else:
        # concurrent.futures.Futures may resolve on any thread, so we
        # need to route them back to the IOLoop.
        io_loop.add_future(
            future, lambda future: io_loop.remove_timeout(timeout_handle))
    return result
```
到此，tornado的epoll及协程分析完毕，打开正确的使用姿势让你的web service轻松应对访问峰值，tornado的协程实现的比较讨巧，下次分析另外一个在openstack中广泛使用的网络库eventlet，它的协程是自己用c语言实现的堆栈管理。openstack web框架那种非一站式的，非打包式的，各组件插件随意插拔的架构，和openstack本身那种不想绑定厂商的架构设计气质很像。下次聊
