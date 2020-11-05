ironic-ui是ironic小组开发的horizon插件，这样一来可以通过web进行ironic节点的制作、状态维护等，笔者在预研的过程中差不多卡了一周多才部署成功，分享一些踩过的坑以帮助其他horizon plugin的部署。

官方[文档](https://docs.openstack.org/ironic-ui/latest/install/installation.html)非常简单，简单到不知所措，但会在一些细节上卡很久
- 首先，文档要求从git上down代码，由于实现环境不允许连外网，所以我下了一个zip包通过ftp传进demo环境，此时出现的问题是由于下载的是master版本，但demo环境是ocata环境，所以master的依赖包一升级导致horizon直接无法访问，出现‘500 internal server error’错误，排查了两天后无法恢复环境只得重装openstack(新环境安装快，但reinstall的话就特别麻烦，最后同事手动一个个服务部署成功)
- 由于实验不能联网，公司内部部署了pipy源，但源里面只有最新的库，所以基本找不到老版本的依赖库，而例如python-novaclient这样的库新版本居然把一些接口删了，所以一些库一旦升级，环境必崩
- 生产环境根本没有[virtualenv](https://virtualenv.pypa.io/en/stable/userguide/)的库，但是和ironic-ui开发者交流，他一直强调要装，当然，其实不装没关系，只要依赖包正确，不污染生产环境就行，但是生产环境不允许传递源码用 pbr打包吧。。。
- 在执行pip install -e . 去部署时一直抛错，错误如下
> Obtaining file:///home/wong_test/ironic-ui-stable-ocata
>  Complete output from command python setup.py egg_info:
>  ERROR:root:Error parsing
>  Traceback (most recent call last):
>  File "/usr/lib/python2.7/site-packages/pbr/core.py", line 111, in pbr
>        attrs = util.cfg_to_args(path, dist.script_args)
>      File "/usr/lib/python2.7/site-packages/pbr/util.py", line 249, in > cfg_to_args
>        pbr.hooks.setup_hook(config)
>      File "/usr/lib/python2.7/site-packages/pbr/hooks/__init__.py", line 25, > in setup_hook
>        metadata_config.run()
>     File "/usr/lib/python2.7/site-packages/pbr/hooks/base.py", line 27, in > run
>        self.hook()
>      File "/usr/lib/python2.7/site-packages/pbr/hooks/metadata.py", line >26, in hook
>        self.config['name'], self.config.get('version', None))
>      File "/usr/lib/python2.7/site-packages/pbr/packaging.py", line 755, in get_version
>        name=package_name))
>    Exception: Versioning for this project requires either an sdist tarball, or access to an upstream git repository. It's also possible that there is a mismatch between the package name in setup.cfg and the argument given to pbr.version.VersionInfo. Project name ironic-ui was given, but was not able to be found.
 >   error in setup command: Error parsing /home/wong_test/ironic-ui-stable-ocata/setup.cfg: Exception: Versioning for this project requires either an sdist tarball, or access to an upstream git repository. It's also possible that there is a mismatch between the package name in setup.cfg and the argument given to pbr.version.VersionInfo. Project name ironic-ui was given, but was not able to be found.

最后分析error log 才发现pbr在打包部署的时候居然依赖git的信息，所以只能在本地使用git clone将代码down下来，再使用git checkout对应的tag，最后将代码传递到实验环境的服务器，安装成功。关键是为什么要有这种机制和git强绑定？什么时候能发展到不通过源码部署插件？
