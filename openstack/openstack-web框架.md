openstack旧服务的web框架采用的是paste、webob(http协议解析)、eventlet.wsgi（协程），但是这个框架添加代码比较臃肿，新增的服务或者重构的服务逐渐使用pecan来简化框架
# restful api
openstack项目都是通过restful api对外提供服务，什么是restful api？看看老外写的这篇文章：[链接](http://www.infoq.com/cn/articles/rest-introduction)言简意赅

# pecan
准确来讲，pecan只解决了route的问题，实现了PEP 333 ( [Python Web Server Gateway Interface v1.0](https://www.python.org/dev/peps/pep-0333/))，网络通信部分依赖其他库的实现。route的优化让代码更加简洁，无需手动撰写路由表，后期维护很方便。以[官方文档](http://pecan.readthedocs.io/en/latest/index.html)举例：
```
from pecan import expose

class BooksController(object):
    @expose()
    def index(self):
        return "Welcome to book section."

    @expose()
    def bestsellers(self):
        return "We have 5 books in the top 10."

class CatalogController(object):
    @expose()
    def index(self):
        return "Welcome to the catalog."

    books = BooksController()

class RootController(object):
    @expose()
    def index(self):
        return "Welcome to store.example.com!"

    @expose()
    def hours(self):
        return "Open 24/7 on the web."

    catalog = CatalogController()
```
URI定义成/catalog/books/bestsellers，即可将http请求路由到BooksController，是不是很方便？
# pecan的实现原理
每个让你惊艳的瞬间，都有神来一笔，下面摘取一点pecan的实现源码
```
# 入口，服务创建部分利用了python内置wsgiref
# pecan/commons/serve.py
from wsgiref.simple_server import make_server
host, port = conf.server.host, int(conf.server.port)
srv = make_server(
      host,
      port,
      app,
      handler_class=PecanWSGIRequestHandler,)
srv.serve_forever()

# app的创建
app = Pecan(root, **kw)

# Pecan实现pep333
# pecan/core.py
class PecanBase(object):
    #根据pep333定义的实现
    def __call__(self, environ, start_response):
        ......
        #从RootController开始递归遍历查找路由对应的类实现
        controller, args, kwargs = self.find_controller(state)
    def find_controller(self, state):
        ......
        controller, remainder = self.route(req, self.root, path)
    def def route(self, req, node, path):
        ......
        node, remainder = lookup_controller(node, path, req)
# pecan/routing.py，递归
def lookup_controller(obj, remainder, request=None):
      ......
```
# paste、pastedeploy、routes
[paste](http://blog.csdn.net/happyanger6/article/details/54518491)的功能比较完善（沉重），鉴权、session、压缩之类的都做了，不过本质上也是wsgi的实现，需要遵循协议，需要依赖网络库,配合pastedeploy它解决了第一层路由。[routes](http://blog.csdn.net/bellwhl/article/details/8956088)类库的引用尝试更优雅的解决路由的问题，看源码实现，很明显，它的路由寻径的方式和pecan不一样，它是在应用层由开发人员手动添加mapper。
paste有一个亮点，就是配置文件可以当成一个路由开关，决定加载哪些api，方便的控制可访问的接口版本
如trove模块中api-paste.ini文件：
```
[composite:trove]
use = call:trove.common.wsgi:versioned_urlmap
/: versions
/v1.0: troveapi

[app:versions]
paste.app_factory = trove.versions:app_factory

[pipeline:troveapi]
pipeline = cors http_proxy_to_wsgi faultwrapper osprofiler authtoken authorization contextwrapper ratelimit extensions troveapp
#pipeline = debug extensions troveapp

[filter:extensions]
paste.filter_factory = trove.common.extensions:factory

[filter:authtoken]
paste.filter_factory = keystonemiddleware.auth_token:filter_factory

[filter:authorization]
paste.filter_factory = trove.common.auth:AuthorizationMiddleware.factory

[filter:cors]
paste.filter_factory = oslo_middleware.cors:filter_factory
oslo_config_project = trove
```
这里要补充一点，在trove组件里，上面所示paste.filter_factory = trove.common.extensions:factory对应的实现又通过库stevedore动态加载扩展的api，有种WTF的感觉
