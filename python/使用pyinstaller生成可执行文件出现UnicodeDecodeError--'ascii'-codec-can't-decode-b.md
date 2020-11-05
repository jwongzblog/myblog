使用pyinstaller生成可执行文件出现  UnicodeDecodeError: 'ascii' codec can't decode byte 0xb3 in position 12: ordinal
 not in range(128)

要解决这个问题从两个方面出发，一种是本身程序有问题，二个环境有问题。

我碰到这个问题的时候，直接执行python代码是正常的，但使用pyinstaller老是出这个问题，无论改成ascii编码还是utf-8还是utf-8无BOM，因此我怀疑是环境问题。

至于环境问题我使用的是python2.7.8win64，其他几个依赖环境也是使用当前最新版64位的。由于时间关系，我就放弃尝试了。

最后我选择py2exe打包python程序，一切顺利。

具体使用方法请参考下面这个blog:

http://www.cnblogs.com/jans2002/archive/2006/09/30/519393.html
