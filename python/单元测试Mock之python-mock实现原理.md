前面聊了一下[要不要做单元测试](https://github.com/jwongzblog/myblog/blob/master/%E7%BC%96%E7%A8%8B%E6%80%9D%E6%83%B3/%E6%98%AF%E5%90%A6%E9%9C%80%E8%A6%81%E6%8A%8A%E5%8D%95%E5%85%83%E6%B5%8B%E8%AF%95%EF%BC%88unit-test%EF%BC%89%E5%8A%A0%E5%85%A5%E9%A1%B9%E7%9B%AE%E4%B8%AD.md)，接来聊聊单元测试的原理。按惯例，不聊用法聊原理。几乎每种语言都有各自的UT框架，不详细聊这个简单的东西了。自动触发test case的基本功能有：
- 递归执行每个目录下的用例
- 基类TEST的成员函数有用例启动时初始化资源、结束时释放资源的函数
- assert的语法
- 异常时不中断整个递归
- 结束时统计执行用例的数量，成功及失败的统计
- 最最重要的Mock，模拟返回值
- ......

## Mock
刚接触Mock，不知道哪位同事称呼其为打桩，后面大家都喜欢说成打桩，我觉得还是很形象的，但在外面跟人说打桩，估计没人理解是什么意思，mock的原意是模拟，模拟返回值的意思。我想聊mock的原因在于有些test case在执行的时候总会出现类似远程调用之类的逻辑而环境不具备，因此mock很好的解决了这个问题，模拟一些无法执行的函数或者类，让上层的调用顺利执行下去，测试上层的逻辑是否正确，每种程序语言的mock实现充分利用了其语言特性，非常精巧，我先聊聊python mock
## python mock资源
- mock on pypi [链接](https://pypi.python.org/pypi/mock)
- mock code on github[链接](https://github.com/testing-cabal/mock/tree/master/mock)
- mock pdf[链接](http://www.voidspace.org.uk/downloads/mock-1.0.1.pdf)
## python mock原理
代码不多，2000多行，充分利用动态语言的特性来实现mock，Python mock的使用大体分成两类
- Mock class
这类用法是把调用对象的object id直接指向了mock，也就是说你的程序在调用某个函数或者类的时候，其实调的是Mock类初始化出来的object（试想一下c++的指针能指来指去吗？），例如:
```
class PersonTest(TestCase):
    def test_should_get_age(self):
        p = Person()       
         //不mock时，get_age应该返回10
        self.assertEqual(p.get_age(), 10)
        // mock掉get_age方法，让它返回20
        p.get_age = Mock(return_value=20)
        self.assertEqual(p.get_age(), 20)
```
具体实现原理我摘取几个代码片段，一目了然：
```
#对外暴露的Mock继承至两个类
class Mock(CallableMixin, NonCallableMock):
#基类的实现
class CallableMixin(Base):

    def __init__(self, spec=None, side_effect=None, return_value=DEFAULT,
                 wraps=None, name=None, spec_set=None, parent=None,
                 _spec_state=None, _new_name='', _new_parent=None, **kwargs):
        self.__dict__['_mock_return_value'] = return_value
    #此处甚妙
    def __call__(_mock_self, *args, **kwargs):
        # can't use self in-case a function / method we are mocking uses self
        # in the signature
        _mock_self._mock_check_sig(*args, **kwargs)
        return _mock_self._mock_call(*args, **kwargs)
    def _mock_call(_mock_self, *args, **kwargs):
        ......
        if ret_val is DEFAULT:
            ret_val = self.return_value
        return ret_val
```
- patch
通常被用作装饰器，原理上没有改变类的申明，执行的过程中成员函数调用被修改了
举例：
```
class PersonTest(TestCase):
    mock_get_class_name = Mock(return_value='Guy')
 
    # 在patch中给出定义好的Mock的对象，好处是定义好的对象可以复用
    @patch('your.package.module.Person.get_class_name', mock_get_class_name)
    def test_should_get_class_name(self):
        self.assertEqual(Person.get_class_name(), 'Guy')
```
实现原理如下：
```
#对外暴露的函数
def patch(
        target, new=DEFAULT, spec=None, create=False,
        spec_set=None, autospec=None, new_callable=None, **kwargs
    ):
    getter, attribute = _get_target(target)
    return _patch(
        getter, attribute, new, spec, create,
        spec_set, autospec, new_callable, kwargs
    )
#层层调用
def _patch_object(
        target, attribute, new=DEFAULT, spec=None,
        create=False, spec_set=None, autospec=None,
        new_callable=None, **kwargs
    ):
    getter = lambda: target
    return _patch(
        getter, attribute, new, spec, create,
        spec_set, autospec, new_callable, kwargs
    )
#实际上是调用的一个类
class _patch(object):
#只不过override了 __call__():，被当成装饰器使用了
def __call__(self, func):
        if isinstance(func, ClassTypes):
            return self.decorate_class(func)
        return self.decorate_callable(func)
#类的申明被完整的复制出来
def decorate_class(self, klass):
        for attr in dir(klass):
            if not attr.startswith(patch.TEST_PREFIX):
                continue

            attr_value = getattr(klass, attr)
            if not hasattr(attr_value, "__call__"):
                continue

            patcher = self.copy()
            setattr(klass, attr, patcher(attr_value))
        return klass
#在enter里面，返回值被修改成MagicMock
def __enter__(self):
    ......
    Klass = MagicMock
    new.return_value = Klass(_new_parent=new, _new_name='()',
                                         **_kwargs)
```
