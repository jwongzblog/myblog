最近解决了一个比较坑的问题，希望给看到本文的人一个启示，或者给焦头烂额的coder一个小帮助。

为了以后从容演变成分布式架构，我选择了单点无状态的设计，但是本组的coder嫌弃将图片转base64存数据库太难搞，于是我们选择了阿里云的OSS作为存储，这样数据库只需存储图片的URL路径即可。由于选择的是Python于是我们用了python的 oss2 sdk，库很方便，开发起来速度很快，但cx_freeze打包的时候始终出现了runtimeerror:maximum recursion depth exceeded的错误，我怀疑oss引入包的时候陷入死循环了，但是由不知道问题出在哪。于是我用sys.setrecursionlimit(10000000)让递归深度变大，结果出现了更能判断问题的错误，即no file named sys（crcmod.crcmod.sys），于是我在setup.py里面includecrcmod.crcmod.sys，但是又出现了crcmod.crcmod.crcmod.crcmod.sys，到此，基本断定crcmod库有问题。

把源码包打开，发现__init__导出对象时用了 from crcmod.crcmod import *，于是我参照其他包处理方式改成from .crcmod import *的方式，重新install这个库包，成功构出release环境。

项目时间紧迫，组员coder打算转api的方式实现，但我预感到时间根本来不及，于是还是采用了从打包入手去解决这个问题。早上在阿里云发帖至今没回复，小失望，幸好通过分析手段排查出问题