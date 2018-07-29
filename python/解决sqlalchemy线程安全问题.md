sqlalchemy是一个非常强大的python orm库，功能完善，BUG少，版本发布频繁，缺点就是代码的可读性略差，我估计过不了pep8、pylint的检查
# ORM的好处
- 帮忙解决sql注入的问题
- 将操作SQL语句变成操作对象，无需像原生那样用下标取值，也就没有新增字段需要调整下标的问题
- 在不变更逻辑实现的情况下替换数据库引擎
- 无需编写大量的重复SQL语句
- 简便的升级机制alembic、migrate
......

# 处理线程安全
你的进程形如RPC或者webservice，如果是阻塞式的，那么你可以放心使用，用完注意close上下文的资源即可（如果不及时处理连接池，会因为mysql长时间没有访问，造成client端连接还存在的假象，实际上长连接已被mysql service释放，再次访问时会出现‘mysql has gone away’的异常。mysql安装默认是8小时移除连接，RDS是两小时）
但如果你的进程是多线程的，那么你可要当心了，分分钟让你的程序的数据库操作异常，因为**sqlalchemy并不是线程安全的**。官方文档为此特别出了一篇说明：[链接](http://docs.sqlalchemy.org/en/latest/orm/contextual.html#unitofwork-contextual)
举例一个webservice的正确使用姿势：
```
Web Server          Web Framework        SQLAlchemy ORM Code
--------------      --------------       ------------------------------
startup        ->   Web framework        # Session registry is established
                    initializes          Session = scoped_session(sessionmaker())

incoming
web request    ->   web request     ->   # The registry is *optionally*
                    starts               # called upon explicitly to create
                                         # a Session local to the thread and/or request
                                         Session()

                                         # the Session registry can otherwise
                                         # be used at any time, creating the
                                         # request-local Session() if not present,
                                         # or returning the existing one
                                         Session.query(MyClass) # ...

                                         Session.add(some_object) # ...

                                         # if data was modified, commit the
                                         # transaction
                                         Session.commit()

                    web request ends  -> # the registry is instructed to
                                         # remove the Session
                                         Session.remove()

                    sends output      <-
outgoing web    <-
response
```
# scoped_session是如何解决线程安全问题？
我贴两处源码大家应该就懂了
```
class scoped_session(object):
    def __init__(self, session_factory, scopefunc=None):
        """Construct a new :class:`.scoped_session`.

        :param session_factory: a factory to create new :class:`.Session`
         instances. This is usually, but not necessarily, an instance
         of :class:`.sessionmaker`.
        :param scopefunc: optional function which defines
         the current scope.   If not passed, the :class:`.scoped_session`
         object assumes "thread-local" scope, and will use
         a Python ``threading.local()`` in order to maintain the current
         :class:`.Session`.  If passed, the function should return
         a hashable token; this token will be used as the key in a
         dictionary in order to store and retrieve the current
         :class:`.Session`.

        """
        self.session_factory = session_factory
        if scopefunc:
            self.registry = ScopedRegistry(session_factory, scopefunc)
        else:
            self.registry = ThreadLocalRegistry(session_factory)
```
```
class ThreadLocalRegistry(ScopedRegistry):
    """A :class:`.ScopedRegistry` that uses a ``threading.local()``
    variable for storage.

    """

    def __init__(self, createfunc):
        self.createfunc = createfunc
        self.registry = threading.local()
```
进程创建了一个全局的session，但通过scoped_session，每个线程会在一个独立的threading.local()空间保留一个registry
# 补充
- remove不要遗漏，否则依然会出现mysql has gone away的错误
- 代码比较丑陋，比较舒服的是openstack的处理方式：
```
def service_destroy(context, service_id):
    session = get_session()
    with session.begin():
        service_ref = _service_get(context, service_id, session=session)
        service_ref.delete(session=session)
```
每个数据库操作都是原子的，oslo_db.sqlalchemy.session的改造，使得with会在函数退出时主动释放session，代码会很优雅。从出现的时间上，go语言的defer设计应该是参考python的with

- flask-sqlalchemy
flask框架也处理的很优雅，但是和框架耦合的比较紧密，无法移植
```
@app.teardown_appcontext
def shutdown_session(response_or_exc):
      if app.config['SQLALCHEMY_COMMIT_ON_TEARDOWN']:
          if response_or_exc is None:
             self.session.commit()

      self.session.remove()
      return response_or_exc
```
装饰器是flask的一个函数，它可以将这个shutdown_session注册到flask_app上下文中，当一个请求退出时，flask会主动调用这个函数
