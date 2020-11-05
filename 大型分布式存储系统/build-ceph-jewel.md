- 只适用jewel版本，因为每个大版本间的编译变化比较大。
- 应该还有其他方式也能编译出来，我们只是依赖了官方比较权威的编译参数。
- 由于编译后的文件夹将近89G，需要分配足够大的磁盘空间。

### debug版本

##### 下载源码
```
git clone git://github.com/ceph/ceph
```
##### 初始化环境依赖
```
./install-deps.sh
```
##### github的代码还少了一些工具包的源码，需要执行以下命令下载
```
./autogen.sh
```
##### 生成make文件，此编译参数是从SRC-RPM configure_log弄出来的
```
./configure --build=x86_64-redhat-linux-gnu --host=x86_64-redhat-linux-gnu --program-prefix= --disable-dependency-tracking --prefix=/usr --exec-prefix=/usr --bindir=/usr/bin --sbindir=/usr/sbin --sysconfdir=/etc --datadir=/usr/share --includedir=/usr/include --libdir=/usr/lib64 --libexecdir=/usr/lib --localstatedir=/var --sharedstatedir=/var/lib --mandir=/usr/share/man --infodir=/usr/share/info CPPFLAGS=" -I/usr/lib/jvm/java/include -I/usr/lib/jvm/java/include/linux" --prefix=/usr --libexecdir=/usr/lib --localstatedir=/var --sysconfdir=/etc --with-systemdsystemunitdir=/usr/lib/systemd/system --docdir=/usr/share/doc/ceph --with-man-pages --mandir=/usr/share/man --with-nss --without-cryptopp --with-debug --enable-cephfs-java --with-selinux --with-librocksdb-static=check --with-radosgw CFLAGS="-O2 -g -pipe -Wall -Wp,-D_FORTIFY_SOURCE=2 -fexceptions -fstack-protector-strong --param=ssp-buffer-size=4 -grecord-gcc-switches -m64 -mtune=generic" CXXFLAGS="-O2 -g -pipe -Wall -Wp,-D_FORTIFY_SOURCE=2 -fexceptions -fstack-protector-strong --param=ssp-buffer-size=4 -grecord-gcc-switches -m64 -mtune=generic"
```
##### 编译
```
make
```
*需要留意一点，configure.ac文件里面定义了，如果检测到.git文件，则把git log的commit uuid截取一段作为版本号，所以编译前，configure命令会修改源码，编译出来的二进制文件的版本变成10.2.2-g6ab137a，导致进程无法启动*

### release版本

官方提供了rpm-build的编译步骤，但是编译的过程还是会有一些坑

##### 参考第一点第2小点安装依赖
```
./install-deps.sh
```
##### 安装rpm工具包及设置环境变量
```
yum install rpm-build rpmdevtools
rpmdev-setuptree
```
##### 获取源码包
```
wget -P ~/rpmbuild/SOURCES/ http://ceph.com/download/ceph-<version>.tar.bz2
```
*请注意，这个压缩包和github上的tar包有差异，这个bz2的包没有cmake相关的文件，但包含了像SPDK这样的源码。另外就是bz2的文件所有者是jenkins-build，其ID是1001，可以如下创建账户*
```
useradd -u 1001 -g 1001 -d /home/jenkins-build -m jenkins-build
```
*请注意，解压文件再压缩（jcvf)文件会导致文件属性的改变，甚至有一些软链接在windows下会变成实体文件，文件内容会影响编译，比如变成"#include <../../rados.h>"，这个是没法编译通过的，因此推荐7zip打开压缩包，直接把优化的源码拖拽进去，最小化改变源码包*

##### 获取spec文件
```
tar --strip-components=1 -C ~/rpmbuild/SPECS/ --no-anchored -xvjf ~/rpmbuild/SOURCES/ceph-<version>.tar.bz2 "ceph.spec"
```
##### 开始编译
```
rpmbuild -ba ~/rpmbuild/SPECS/ceph.spec
```
- "-ba"是编译src.rpm和二进制包，详情参考rpmbuild命令
- "-bb"只编译出二进制包
- 如果编译出错，重新执行该命令会删掉已经编译好的文件从头开始，不像make那样可以继续编译

*请注意，编译过程是自动的，无需人工参与，但是configure.am文件定义如何生成Makefile文件，里面有一些参数依赖于你的编译环境，比如在编译librockdb的时候，如果环境有lz4的压缩工具，则编译出来的so会依赖liblz4.so，但是官方给出的rpm二进制包是没有这个依赖的，所以如有必要，需要修改configure.am来控制生成的Makefile文件*

### 进程启动

ceph-osd替换后需要给755权限，否则进程无法启动