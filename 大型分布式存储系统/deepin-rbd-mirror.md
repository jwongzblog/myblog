本篇结合ceph源码，尝试解析rbd mirror进程的高可用及其同步机制

#### 进程的高可用

- rbd mirror daemon经历了J、L、M三个版本的迭代，从最初J版的单进程，到L版的一主多备，再到M版的进程多活。
- L版：首个rbd mirror daemon起来时，会把自身标记成Leader，后面起来的进程会定期获取Leader的位置是否被锁定了，一旦Leader lock被释放，自己会尝试成为Leader。此机制确保Daemon的HA，但在12.2.7的版本中，进程间启动间隔短，容易出现两个Leader
- M版：每个rbd mirror daemon都正常工作，为image建立同步任务时，会为这个任务绑定一个rbd mirror daemon，相当于一个简易的负载均衡

 

#### image的同步机制

进程会为每个image建立同步任务分配一个线程，定时获取remote ceph cluster的同名image状态。
进程先检测local集群有没有同名image，如果没有则从remote集群clone一个image。
当定时器检测到remote的journal tag与local的不一致时，同步tag
```
* <start>
*    |
*    v
* GET_REMOTE_TAG_CLASS * * * * * * * * * * * * * * * * * *
*    |                                                   * (error)
*    v                                                   *
* OPEN_REMOTE_IMAGE  * * * * * * * * * * * * * * * * * * *
*    |                                                   *
*    |/--------------------------------------------------*---\
*    v                                                   *   |
* IS_PRIMARY * * * * * * * * * * * * * * * * * * * * *   *   |
*    |                                               *   *   |
*    | (remote image primary, no local image id)     *   *   |
*    \----> UPDATE_CLIENT_IMAGE  * * * * * * * * * * *   *   |
*    |         |   ^                                 *   *   |
*    |         |   * (duplicate image id)            *   *   |
*    |         v   *                                 *   *   |
*    \----> CREATE_LOCAL_IMAGE * * * * * * * * * * * *   *   |
*    |         |                                     *   *   |
*    |         v                                     *   *   |
*    | (remote image primary)                        *   *   |
*    \----> OPEN_LOCAL_IMAGE * * * * * * * * * * * * *   *   |
*    |         |   .                                 *   *   |
*    |         |   . (image doesn't exist)           *   *   |
*    |         |   . . > UNREGISTER_CLIENT * * * * * *   *   |
*    |         |             |                       *   *   |
*    |         |             v                       *   *   |
*    |         |         REGISTER_CLIENT * * * * * * *   *   |
*    |         |             |                       *   *   |
*    |         |             \-----------------------*---*---/
*    |         |                                     *   *
*    |         v (skip if not needed)                *   *
*    |      GET_REMOTE_TAGS  * * * * * * *           *   *
*    |         |                         *           *   *
*    |         v (skip if not needed)    v           *   *
*    |      IMAGE_SYNC * * * > CLOSE_LOCAL_IMAGE     *   *
*    |         |                         |           *   *
*    |         \-----------------\ /-----/           *   *
*    |                            |                  *   *
*    |                            |                  *   *
*    | (skip if not needed)       |                  *   *
*    \----> UPDATE_CLIENT_STATE  *|* * * * * * * * * *   *
*                |                |                  *   *
*    /-----------/----------------/                  *   *
*    |                                               *   *
*    v                                               *   *
* CLOSE_REMOTE_IMAGE < * * * * * * * * * * * * * * * *   *
*    |                                                   *
*    v                                                   *
* <finish> < * * * * * * * * * * * * * * * * * * * * * * *
``` 

#### journal的tag机制
这个机制原先就有，只不过在tag的数据结构中补充了一些mirror相关的string类型信息
```
* <start>
*    |
*    v
* OPEN_JOURNALER * * * * *
*    |                   *
*    v                   *
* ALLOCATE_TAG * * * * * *
*    |                   *
*    v                   *
* APPEND_EVENT * * *     *
*    |             *     *
*    v             *     *
* COMMIT_EVENT     *     *
*    |             *     *
*    v             *     *
* STOP_APPEND <* * *     *
*    |                   *
*    v                   *
* SHUT_DOWN_JOURNALER <* *
*    |
*    v
* <finish>
```

#### 总结：

虽然不停的优化rbd mirror daemon的高可用，但是image peer的线程却没有得到保障，宿主机的内存使用完毕后，在虚拟化环境的ceph集群会特别卡，直接导致线程被system killed了，重启进程后才能继续同步，希望物理机环境不会出现这个问题吧，除非像erlang一样，线程状态也被监控

rbd mirror使用journal机制去回放op，对于日志盘、网络、cpu有一定的压力