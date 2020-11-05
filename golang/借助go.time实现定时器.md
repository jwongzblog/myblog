方式一：
```
package main

import (
	"os"
	"time"
)

func doSomething() (ok bool){
	return true
}

func main() {
	now_timestamp := time.Now().Unix()
	check_timeout := now_timestamp + 1800
	for {
		if doSomething(){
			break
		}
		
		now_timestamp = time.Now().Unix()
		if now_timestamp > check_timeout {
			os.Exit(1)
		}

		time.Sleep(2 * time.Second)
	}
}
```
上面的实现会有一个问题，如果doSomething阻塞，那么这个进程永远卡在这，接下来看这个优雅的方式

方式二：
```
func main() {
	callTicker := time.NewTicker(10 * time.Second)
	defer callTicker.Stop()
	timeoutTicker := time.NewTicker(1800 * time.Second)
	defer timeoutTicker.Stop()

loop:
	for {
		select {
		case <-callTicker.C:
			if doSomething(){
				break loop
			}

		case <-timeoutTicker.C:
			lib.ERROR("service state is not Available until timeout")
			os.Exit(1)
		}
	}
}
```
doSomething()即使很耗时，也不会阻塞 timeoutTicker