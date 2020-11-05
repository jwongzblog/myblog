曾经在研究存储快照备份的时候有个比较难处理的技术点，就是我可以通过ssh command支配外置存储挂载卷到生产环境，但是就是不知道该如何刷新这个scsi信息以及分配磁盘符、找到磁盘符对应scsi的映射信息。好在后来发现udev工具能操作这个场景，并且是开源的，但如何把有用的代码集成至我的功能代码了？看源码......

我知道udev解决这些问题的命令，于是从入口开始。
**ps:这个经验不适合解读具备reflect能力的编程语言，笔者曾经读一段openstack代码的时候懵逼了很久，类的实例化居然依赖前端传过来的相对路径，然后动态加载类并实例化......**

代码是非常难读的，因为有几十个if ('xxx' == 'xxx') else...，这些c代码里面还插入了swicth......不仅费眼而且费脑子......但是为了完整的解析掉所有的命令，又不得不这样写。

17年初google开源的python库 python-fire很大程度上解决了这个问题，举个官方的例子：
```
import fire

def add(x, y):
  return x + y

def multiply(x, y):
  return x * y

if __name__ == '__main__':
  fire.Fire({
      'add': add,
      'multiply': multiply,
  })
```
我们就可以这样执行代码
```
$ python example.py add 10 20
30
$ python example.py multiply 10 20
200
```
也可以封装在类里：
```
import fire

class Calculator(object):

  def add(self, x, y):
    return x + y

  def multiply(self, x, y):
    return x * y

if __name__ == '__main__':
  calculator = Calculator()
  fire.Fire(calculator)
```
执行：
```
$ python example.py add 10 20
30
$ python example.py multiply 10 20
200
```
这样一来，你设计并实现的command工具可以极少的写判断语句，不仅添加功能方便，维护简单，而且代码量也很少，代码少BUG就少，一举多得。

## 参考：

[github代码链接](https://github.com/google/python-fire)

[readme](https://github.com/google/python-fire/blob/master/docs/guide.md)
