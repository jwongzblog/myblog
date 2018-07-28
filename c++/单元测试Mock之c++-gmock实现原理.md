google开源的单元测试框架gtest、gmock被广泛应用在c++语境中，具体用法参照[玩转Google开源C++单元测试框架Google Test系列(gtest)(总)](http://www.cnblogs.com/coderzh/archive/2009/04/06/1426755.html)，相当不错的文章

按惯例，只谈设计的精巧之处。前面讲到[python mock的实现原理](http://www.jianshu.com/p/18d37ded2619)，执行用例时，python mock直接运行时改变object id，相当于c++里面直接把指针指向另一个类实例......写惯了c++代码后初次接触这样的用法，简直不可思议，不过深入python的实现原理，看似粗暴的实现，其实也有其道理，而且还是比较高级的道理。

python所有的类的共同祖先（基类）是object，这才是万物皆对象啊，连函数都是对象......这种设计思路和c++面向对象编程的终极设计COM组件神似（我之所以称之为c++的终极设计，是从《COM本质论》和设计模式的角度观察的，通过COM组件，无需烧脑的处理你的设计模式，可以避免不精通c++的继承、多态而带来的犯错，有时间再聊聊这个话题），COM组件的设计也是所有类的基类是IUnkown。

gmock的用法也很相似，mock出来的子类继承自需要替代的父类，只不过需要在编译阶段，已经确定调用的对象在runtime时指向了mock类。大致原理如此，但是gmock的实现设计也十分讨巧，下面顺藤摸瓜从一个宏来解析一下gmock的实现原理

gmock的用法[举例](http://blog.csdn.net/breaksoftware/article/details/51384083)，看看例子里面的这个宏
```
EXPECT_CALL(*mockDepInterface, test1(testing::_))
                .WillOnce (testing::Throw (Exception (_T("ncIDepInterface::test1 has an exception"))));
```
这个宏是用来指定逻辑层调用test1方法时的返回值，将宏展开其实很容易就能明白它做了什么。 WillOnce内部核心就是：
```
untyped_actions_.push_back(new Action<F>(action));
```
 也就是说它将返回值塞进了一个vector容器中。

那么调用mock的时候返回值是怎么整出来的了，我们看这个宏：
```
#define MOCK_METHOD0(m, F)  GMOCK_METHOD0_(, , , m, F)
```
再看这个：
```
#define GMOCK_METHOD0_(tn, constness, ct, Method, F) \
  GMOCK_RESULT_(tn, F) ct Method() constness { \
    GTEST_COMPILE_ASSERT_(::std::tr1::tuple_size< \
        tn ::testing::internal::Function<F>::ArgumentTuple>::value == 0, \
        this_method_does_not_take_0_arguments); \
    GMOCK_MOCKER_(0, constness, Method).SetOwnerAndName(this, #Method); \
    return GMOCK_MOCKER_(0, constness, Method).Invoke(); \
  } \
  ::testing::MockSpec<F>& \
      gmock_##Method() constness { \
    GMOCK_MOCKER_(0, constness, Method).RegisterOwner(this); \
    return GMOCK_MOCKER_(0, constness, Method).With(); \
  } \
  mutable ::testing::FunctionMocker<F> GMOCK_MOCKER_(0, constness, Method)
```
其实我们调用mock的时候是调用
```
GMOCK_MOCKER_(0, constness, Method).Invoke();
```
他其实是一个FunctionMocker类的对象调用invoke,
invoke的实现如下：
```
  Result InvokeWith(const ArgumentTuple& args) {
    return static_cast<const ResultHolder*>(
        this->UntypedInvokeWith(&args))->GetValueAndDelete();
  }
```
再看看
```
GetValueAndDelete（）
  T GetValueAndDelete() const {
    T retval(value_);
    delete this;
    return retval;
  }
```
也就是说它从之前说的那个vector中取出返回的结果被删除自身，这样子j就实现了每次调用同一个接口可以返回不一样的结果。
当然，里面运用了大量的模版和c++11中的元组来实现泛型。

总而言之，就是EXPECT_CALL把要返回的结果塞进一个vector<void*>中，当mock类对象调用方法时，会从容器中取第一个值并删除自身。
