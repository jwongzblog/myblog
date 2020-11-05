package runner

import (
	"../common"
	"log"
	"../product"
	"strings"
	"sync"
)

type Runner struct {
	uploadDir, product string
	threadCount        int
}

func NewRunner(uploadDir, product string, threadCount int) *Runner {
	return &Runner{uploadDir: uploadDir, product: product, threadCount: threadCount}
}

func (r *Runner) Run() {
	fileMap, err := common.GetFileMap(r.uploadDir)
	if err != nil {
		log.Print(err)
	}

	concurrentChan := make(chan error, r.threadCount)
	for i := 0; i != r.threadCount; i++ {
		concurrentChan <- nil
	}

	wg := &sync.WaitGroup{}
	for path, objName := range fileMap {
		uploadErr := <-concurrentChan //最初允许启动 {$threadCount} 个 goroutine，超出{$threadCount}个后，有分片返回才会开新的goroutine.
		if uploadErr != nil {
			err = uploadErr
			break // 中间如果出现错误立即停止继续上传
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			var objPro product.Product
			if strings.EqualFold(r.product, common.Us3) {
				objPro = product.NewUs3(path, objName)
			}
			e := objPro.Upload()
			concurrentChan <- e //跑完一个 goroutine 后，发信号表示可以开启新的 goroutine。
		}()
	}
	wg.Wait()       //等待所有任务返回
	if err == nil { //再次检查剩余上传完的分片是否有错误
	loopCheck:
		for {
			select {
			case e := <-concurrentChan:
				err = e
				if err != nil {
					break loopCheck
				}
			default:
				break loopCheck
			}
		}
	}
	close(concurrentChan)
}
