# Ironic存储管理
上周对ironic的存储进行了预研，我们非常期待ironic尽快能够集成cinder
### 1.现状
通过研读O版的源码，发现目前的版本还不支持cinder管理卷，主要表现在两个方面：
- nova的ironic驱动没有实现attach_volume基类方法
- ironic模块没有实现卷的管理，比如暴露volume connector、target，mount volume，device manager不过好在富士通和日立的工程师在推进ironic的cinder卷管理，目前master的commit已经能看到日立工程师的patch，里面实现了volume connector、target的创建、获取、删除。

### 2.计划
通过咨询openstack-dev@lists.openstack.org ironic的负责人之一回复了邮件，具体内容如下：

----------------------------------------------------------------------

Message: 1
Date: Thu, 6 Jul 2017 09:00:08 -0400
From: Julia Kreger <juliaashleykreger@gmail.com>
To: "OpenStack Development Mailing List (not for usage questions)"
<openstack-dev@lists.openstack.org>
Subject: Re: [openstack-dev] I want to use nova api to attach volume
to ironic node
Message-ID:
<CAF7gwdj5h1YYxmpekKF4jLDJejD9sTfKn6E6+unqvMiisbTUfw@mail.gmail.com>
Content-Type: text/plain; charset="UTF-8"

> Greetings!

> This work is presently underway, although is largely an effort in
> Ironic as we must orchestrate volume attachments and detachments
> with-in the lifecycle of the machine. We presently have weekly status
> meetings[1] which happens to occur in approximately 3 hours,  and we
> have an Etherpad[2] tracking our present status and the patches
> related to this effort. I've included two additional links to the
> Ironic specifications [3][4] related to this subject.

> Please feel free to reach out to the ironic team, or visit us in
> openstack-ironic if you have any questions.

> Julia

[1]: http://eavesdrop.openstack.org/#Ironic_Boot_from_Volume_meeting
[2]: https://etherpad.openstack.org/p/Ironic-BFV
[3]: https://specs.openstack.org/openstack/ironic-specs/specs/not-implemented/volume-connection-information.html
[4]: https://specs.openstack.org/openstack/ironic-specs/specs/not-implemented/boot-from-volume-reference-drivers.html

2017-07-06 5:42 GMT-04:00 王俊 <wangjun@yovole.com>:
> Hi,all:
>
> my customer want use nova api attach volume to the node, cause they
> are afraid of oneday the node storage is useless, is somebody has a plan to
> provide this?
>
>
>
> __________________________________________________________________________
> OpenStack Development Mailing List (not for usage questions)
> Unsubscribe: OpenStack-dev-request@lists.openstack.org?subject:unsubscribe
> http://lists.openstack.org/cgi-bin/mailman/listinfo/openstack-dev
>

其中[需求列表]( https://specs.openstack.org/openstack/ironic-specs/specs/not-implemented/volume-connection-information.html) 详细介绍了volume manager的需求以及接口设计，从回复中可以看到ironic正在讨论如何设计 volume与node的lifecycle。

### 3.预期
以目前的社区进度，应该无法在近两个版本内提供完善的cinder卷管理功能。目力所能及的要解决这几个问题：
- 第一步是ironic本身需要提供restful api，完成volume的链接管理，挂载等。（顺利的话最快-
能在Pike版本提供）
- 第二步是nova模块需要实现ironic的virt driver。
- 第三个就是进度，目前社区这种协作方式推进项目较慢。

可以想见的未来是ironic一定有完善的node生命周期的存储管理，但进度较慢，还有一种方式是我们也主动参与进去，参与社区的讨论、接口的定制、代码的实现来加速推进整个项目进度

不太推荐的方式是为了进度，我们自己提供一套实现而不等待社区回复，但这会面临一个问题就是未来还是要合并代码，将我们的实现覆盖，而且有可能出现一些考虑不周的BUG

### 4.冲击力
ironic一旦兼容cinder，将在这几方面改善：
- 存储扩容，不受限于插槽的数量
- 存储备份，cinder backup的能力可以集成进来
- 我最爱的ceph可以提供分布式块存储过来
