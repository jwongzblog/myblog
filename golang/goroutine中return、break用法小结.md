最近尝试在goroutine中开启定时来完成周期性触发的任务，对一些关键字的行为不熟悉，导致碰到小坑

```
go func() {
	tk := time.NewTicker(time.Second * time.Duration(GAPTIME))
	for {
		select {
		case <-tk.C:
			bc := new(BackupConsumer)
			backupInfo, svrInfo, err := bc.RegisterStartBackup(t.session)
			if err != nil || backupInfo == nil || svrInfo == nil {
				return
			}
			bc.StartBackup(t.session, backupInfo, svrInfo)
		}
	}
}()
```
上面的代码中我启动一个协程，然后让定时器触发注册备份任务的行为，由于注册时发生错误，err非空，导致进入return逻辑，我误以为return的作用域只影响select，但其实作用域范围是 go func()，协程退出了，定时器被销毁

将return改成break后，结果才是预期的，只break了select，下次tk.C触发后会继续注册备份任务

但如果我想break for循环呢？代码就要这样写了：
```
go func() {
	tk := time.NewTicker(time.Second * time.Duration(GAPTIME))
LOOP:
	for {
		select {
		case <-tk.C:
			bc := new(BackupConsumer)
			backupInfo, svrInfo, err := bc.RegisterStartBackup(t.session)
			if err != nil || backupInfo == nil || svrInfo == nil {
				break LOOP
			}
			bc.StartBackup(t.session, backupInfo, svrInfo)
		}
	}
}()
```

同事说effective c++不建议使用类似goto的语法，其实不一定，滥用goto才是导致程序行为不可控的原因
