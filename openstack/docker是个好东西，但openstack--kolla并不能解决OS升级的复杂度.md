最近团队打算把openstack从M版升级到O版，在合并代码和测试的时候碰到很多问题，所以想引入kolla项目来降低升级的复杂度，我比较怀疑投入、产出收益比

首先，docker有如下优势：
- 测试环境与生产环境一致，运维团队部署的复杂度几何倍数降低
- 进程级别的虚拟化，比虚拟机创建的速度快，单台对比虚拟机占用资源小
- 需要大规模并发处理的环境时，启动快的时效性收益能体现的淋漓尽致

再来说说劣势：
- 管理的复杂度比虚拟机大，比如数据持久化、端口映射
- 如果你设计的软件能吃透硬件性能，docker产生的相关进程反而带来额外内存、CPU、存储容量（一个镜像的依赖库不小）损耗
- 问题排查远比虚拟机要复杂

所以使用kolla并不能解决升级时合并代码的复杂度、减小POC测试的工作量
