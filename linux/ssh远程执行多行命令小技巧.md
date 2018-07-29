在openstack的ops中，经常碰到多个环境切来切去执行相关命令的情况，最纠结的是可能会出现批量的环境需要重复的执行一组操作，作为程序员，重复是最大的浪费。解决了这个问题后，总结出几个小技巧来解决这个问题。

# 背景
openstackM版升级至O版，租户的trove实例需要批量升级
# 远程执行一条命令

`ssh nick@xxx.xxx.xxx.xxx "df -h"`
# 远程执行多条命令
#### 方法一：

`ssh nick@xxx.xxx.xxx.xxx "pwd; cat hello.txt"`
作为处女座情节的非处女座程序员无法忍受这样的书写格式
#### 方法二：
```
ssh nick@xxx.xxx.xxx.xxx "pwd
> cat hello.txt
> ls
> pwd
> "
```
如果需要引用当前环境变量，可以这么做：
```
export name=nick
ssh nick@xxx.xxx.xxx.xxx "
> echo $name
> "
```
#### 方法三：
```
ssh root@xxx.xx.xx.xx > /dev/null 2>&1 << eeooff
cd /home
tar xvf trove_rpm_code.tar
cd /home/trove_rpm_code/rpm
rpm -U *.rpm
exit
eeooff
```
笔者最喜欢这种格式，稍微解释一下 。`> /dev/null 2>&1`的作用是把命令执行的错误以及输出全部吞噬不输出，`<< eeooff`的作用是把接下来的字符串全部作为参数输入，直到碰到eeooff结束

[参考一：SSH 远程执行任务](http://www.cnblogs.com/sparkdev/p/6842805.html)
[参考二：shell十三问](http://bbs.chinaunix.net/forum.php?mod=viewthread&tid=218853&page=7#pid1636825)
