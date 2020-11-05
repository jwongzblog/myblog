trove支持增量备份，上层应用封装的时候希望提供的功能更人性化，比如说我删除一个增量备份点15，trove能把依赖备份点15的增量备份点16、17保留，并且能把备份点15合并到16。需求的描述似乎只变更了一句话，但是底层的实现复杂度是指数级增加。有这几个问题：
- 如果备份点15的数据合并到16，上层应用的封装是否能及时处理相应逻辑？
- 备份点的数据合并比较耗时，上层应用是轮询状态，还是希望底层能够通知上层修改状态？
- 由于比较耗时，万一出现异常，上层如何与底层状态同步，并保持一致？
- 如果底层实现了数据合并，应该在何时合并，不与其他进程抢夺资源？
- 如果有大量的合并任务，底层的进程如何并行，如何调度，如何控制进程的异常状态？

作为程序员，有时候需要估量一下产品经理的一句话会对系统带来多大的冲击，无论如何都需要充分的评估需求，不要在评审会上打马虎。

回到标题本身，trove的备份删除策略很简单，就是深度递归，删除所有依赖备份点15的children。
```
    @classmethod
    def delete(cls, context, backup_id):
        """
        update Backup table on deleted flag for given Backup
        :param cls:
        :param context: context containing the tenant id and token
        :param backup_id: Backup uuid
        :return:
        """

        # Recursively delete all children and grandchildren of this backup.
        query = DBBackup.query()
        query = query.filter_by(parent_id=backup_id, deleted=False)
        for child in query.all():
            cls.delete(context, child.id)

        def _delete_resources():
            backup = cls.get_by_id(context, backup_id)
            if backup.is_running:
                msg = _("Backup %s cannot be deleted because it is running.")
                raise exception.UnprocessableEntity(msg % backup_id)
            cls.verify_swift_auth_token(context)
            api.API(context).delete_backup(backup_id)
```
