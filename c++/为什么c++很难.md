深受Scott Meyers的《Effective c++》影响，每学一门语言我就迫不及待的去查阅其effective系列，避免入坑。对比后发现其他语言的effective大多讲的是不常用的特性如何避坑，而《Effective c++》里面几乎每一条都是c++常用的特性如何避坑......如果你感兴趣，还有《More effective c++》、《Exceptional c++》、《More Exceptional c++》、《深入探索c++对象模型》、《Effective STL》、《泛型编程与STL》......等着你去征服，但愿你的毛发能扛得住。决定在简历上填写精通c++前，建议先搞定这些系列，除非面试你的技术官也不是很精通c++。

此篇就不对比语言的框架，库生态，以及槽点了，我挑几条《Effective c++》里面的避坑手法，让大家明白，**c++如何让你轻而易举的犯错，而且有些坑，Meyers自己都无法解决**，我觉得一门理想的编程语言，应该是学懂基本的语法就能够让编程者轻松的写出正确的代码，让程序员把更多的精力放在业务逻辑上，架构梳理上。掌握c++很累，如果不彻底了解c++执行的原理，就难以利用c++的优势，也难以写出健壮的代码。如果做不到，那么选择c++的意义又何在了？

# 尽量以const、enum、inline替换define
关键字define是c语言的范畴，但是c++设计之初就把c作为c++的子集给兼容了，举一个define的错误示范，程序的结果不符合编码者的预期：
```
//以a、b的较大值调研函数f
#define CALL_WITH_MAX(a, b) f ((a) > (b) ? (a) : (b))

in a = 5,b = 0;
CALL_WITH_MAX(++a, b);        //a被累加二次
CALL_WITH_MAX(++a, b + 10);   //a被累加一次
```
**a的递增次数居然取决于它和谁比较......其实，新出的编程语言，已经没有出现“++”这种设计了，无论是大学试卷还是面试题，“++”居然被设计成难以驾驭的考试控分点。在《Effective STL》里也介绍了循环中，“++“符号出现在变量前后，结果是如何不一样的......**，我似乎跑偏~~

# 为多态基类声明virtual析构函数
```
class TimeKeeper {
public:
   TimeKeeper();
   ~TimeKeeper();
   ...
};

class WaterClock:public TimeKeeper{...};  //水钟
class AppleWatch:public TimeKeeper{...};  //苹果表
......


TimeKeeper* ptk = getTimeKeeper ();   //从继承体系获取动态对象
...
delete ptk
``` 
我们之所以选择使用面向对象的c++，我们就是希望能把不同类型的业务但具有大量类似行为场景抽象化，然后通过操作实例化对象达到抽象工厂的效果，但是当子类对象经由一个父类指针被删除，而父类带着一个non-virtual析构函数，那么实际执行时通常是子对象不会被销毁，造成内存泄露

上面正确的做法是：
```
class TimeKeeper {
public:
   TimeKeeper();
   virtual ~TimeKeeper();
   ...
};

```

# 别让异常脱离析构函数
程序员们很懒，非常喜欢利用析构函数被自动调用的机制让程序去释放一些资源，比如在析构函数里面delete DBconnect之类的，但是一旦析构函数里面抛出异常，那么这个实例化对象就不能被正常释放，造成内存泄露。如果非得用这个特性，那么唯一能做的就是在析构函数里面把内存吞了。但是把异常吞了，又会造成程序的不可控行为，因为不知道delete DBconnect有没有成功，调用者不知道行为结果，所以有需要一些额外的机制让调用者知道到底发生了什么，程序越来越难看了。

# 在重载运算符“=”时
处理自我赋值也是个大坑，需要小心翼翼的处理深拷贝的问题，作者比较推荐的处理方式是copy and swap

# 在资源管理类中提供对原始资源的访问
c++作为一个没有垃圾回收的语言（GC），内存释放是一个比较痛苦的事情，所以从c++11开始提供智能指针，帮助程序员有效管理内存，但是使用的过程中依然有些别扭，智能指针类型与RAII类依然不是同一个类型，所以需要实现get或者隐式转换的方式对外提供被智能指针包裹的原始指针
举例：
```
std::shared_ptr<Investment> pInv(createInvestment());

int dayHeld(const Investment* pi);

int days = dayHeld(pInv);       //如果这么调用就会出错

//正确的做法
int days = dayHeld(pInv.get());
```

# 以独立语句将newed对象置入智能指针
举例就明白了：
```
int priority();
void processWidget(std::shared_ptr<Widget> pw, int priority);
//调用时
processWidget(new Widget, priority());
```
上面的调用效果，相当于processWidget(std::shared_ptr<Widget> (new Widget), priority()),一旦“new Widget”抛错，有肯能会造成内存泄露
正确的做法：
```
std::shared_ptr<Widget> pw(new Widget);
processWidget(pw, priority());
```
**我想说的是，以上所有的逻辑简化下来都是一样的，就是调用顺序不一样而已，但是一个一不小心的异常行为，就轻松的造成内存泄露，而程序员唯一能做的就是小心翼翼的设计代码，避免别人调用或者调用别人代码时的错误传导**

# 必须返回对象时，别妄想返回reference
看以下二者的区别：
```
const Rational& operator* (const Rational&lhs, const Rational&rhs)
const Rational operator* (const Rational&lhs, const Rational&rhs)
```
类里面尝试重载运算符”*“，我们一直被教导一定要传引用来代替传值，引用会节约内存，但是一个函数无论怎么实现，对象的释放都将是一个灾难。c++给你一锅菜的原材料和配方让你炒，不告诉你具体怎么做，结果一些细微的差别让你炒出来的东西难以下咽，不像方便面，倒一杯开水就能吃。

# 少做类型转换
类型转换是比较危险的操作，但是我TMD做不到啊，由于兼容c语言，大量的库都是c-style的，如果要使用这些**上古**的库包，类型转换是少不了的，不然你就得自己造轮子，这点恶心到我了。

# 避免遮掩继承而来的名称
出个题恶心你一下，答出下列分别调用的是哪个类的函数
```
class Bass {
private:
   int x;
public:
   virtual void mf1() = 0;
   virtual void mf1(int);
   virtual void mf2();
   void mf3();
   void mf3(double);
}

class Derived:public Bass{
public:
   using Bass::mf1;
   using Bass::mf3;
   virtual void mf1();
   void mf3();
   void mf4();
}

Derived d;
int x;
...
d,mf1();
d.mf1(x);
d.mf2();
d.mf3();
d.mf3(x);
```
常做c++的题库，有助于防止老年痴呆症

#绝不重新定义继承而来的non-virtual函数
再做个智力题：
```
class B{
public:
   void mf();
};

class D:public B{
...
};

D x;
B* pB = &x;
pB->mf();

D* pD = &x;
pD->mf();
```
请问以上二者的运行结果一致吗？由于non-virtual是静态绑定的，所以如果D实现了自己版本的mf()，上面二者都会调用到D实现的mf()。

回想一下上面的教条，一个关键字或者符号的差异，程序运行的结果完全不一样，所以我偷偷的修改了刚毕业时的简历，把精通c++改成了熟练使用c++