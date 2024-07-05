# ossfs
by aliyun
## ossfs实现机制：
- 在/tmp目录创建文件句柄，然后unlink文件，但进程不释放文件句柄，ls看不到这些文件，但是lsof命令可以看到这些文件
```
ossfs     799982 907211           root   42u      REG              253,0 1073741824  134362712 /tmp/s3fstmp.l207iY (deleted)
ossfs     799982 907211           root   43u      REG              253,0 1073741824  134362713 /tmp/s3fstmp.IDQ4Yo (deleted)
ossfs     799982 907211           root   44u      REG              253,0 1073741824  134362714 /tmp/s3fstmp.7d67EP (deleted)
ossfs     799982 907211           root   45u      REG              253,0 1073741824  134362715 /tmp/s3fstmp.6qBjlg (deleted)
ossfs     799982 907211           root   46u      REG              253,0 1073741824  134362716 /tmp/s3fstmp.OJ4z1G (deleted)
ossfs     799982 907211           root   47u      REG              253,0 1073741824  134362717 /tmp/s3fstmp.kAQNH7 (deleted)
```

- 继续由下载请求往这个deleted文件句柄里面写文件，所以/tmp所属系统盘的ioutil负载100%, 220MB/s的顺序写，32线程没有命中缓存的情况下，fio的 read iops 在100+

- 等文件完全下载完成后，此后的随机读是读buffer/cache，读吞吐可以高达600MB/s

- 通过drop cache命令释放，read iops又掉到100+

## 优缺点

- ossfs默认使用/tmp目录作为缓存，其io路径无论是上传和下载，最终都需要在/tmp目录的文件系统进行读写
- 所以备份和下载的场景下，多了一层本地磁盘IO，性能较差
- 性能瓶颈在这个缓存目录，随机读的性能提升在于有没有命中操作系统的pagecache，
- fio 32线程随机读前期性能差，ossfs 20MB步长下载文件，写入/tmp，此时会把/tmp对应磁盘负载打满，预热一段时间后读吞吐可以慢慢增长到600MB/s
- 通过drop cache释放内存的缓存，fio就直接读/tmp所在磁盘了，其瓶颈在于/tmp目录的存储介质
- 应用close文件句柄，/tmp模式下会释放空间；如果应用open的文件句柄没有及时close，会出现空间不足的问题。
- 可以使用--use_cache指定缓存目录，区别在于，close文件句柄后，这个目录下的文件能被“ls”出来

# goofys
- goofys的上传：只支持顺序写满5MB buffer再分片上传到对象存储，需要主动调用flush来解决buffer没有攒满的情况，防止断电丢数据
- ossfs会修改/tmp文件，标注哪些字节被修改，异步分片上传，降低断电丢数据的风险
- goofys的下载：顺序读，会预读2GB(会有读放大的问题)；随机读，会从当前offset读完剩余数据（改了点代码，可以支持只读4K）
- goofys支持catfs映射一个本地目录作为缓存，没有用到pagecache，随机读取决于磁盘介质。catfs没有GA，不建议上生产，测试中经常hang死挂载目录，需要重启服务器

# 对比
## ls的行为：
ossfs直接发送到对象存储，goofys会在本机缓存目录树。意味着goofys多点挂载有数据一致性的风险，但性能会有所提升。取决于应用的scan操作的频次和规模

# 使用建议（最好是根据AI使用场景构造数据形态测试一下）：
- 如果只是备份、下载，建议使用goofys
- 如果是小数据量的训练，建议ossfs
