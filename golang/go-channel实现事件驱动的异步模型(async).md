历史不应只成为被遗忘的过去，因为此刻是由历史塑造--by 《罗辑思维》

被go的一段代码惊艳到之后，本来想写这篇文章的，为了让此篇读者也能感受到惊艳，我需要先介绍背景。结果发现这个背景很冗长，有些内容居然能独立成篇，所以请先移目至这里《[浅谈api设计的耗时因素](https://www.jianshu.com/p/93e0a85a7246)》以及这里《[浅谈异步处理](https://www.jianshu.com/p/ac84add16d53)》。

如果理解了异步原语，我们就来看看几种语言是如何处理异步的，由于笔者只会c/c++、python、go，所以用三者举例。欢迎留言补充其他语言的神之一手。

**c++98：**
我用的最多的版本，要实现异步，通常会使用线程来处理
```
class UserThread:thread {
public: 
        void GetUser(User &user);
        bool IsDone ();
private:
        User _user;
}

void UserThread::start()
{
    //rpc_call
}
```
有多少个数据类型(class)，就有多少个线程类的实现，while里面轮询调用IsDone()来获取结果是否完毕。比如：
```
void main () {
    for(i = 0; i < len(threads); i++){
        threads[i].IsDone();
        //......
    }
    //......
}
```
无疑，这样的代码量是巨大的，而且线程的状态很难维护

**我们再看看python2.x的实现**
```
import threading

def getUser():
    #rpc_call

if __name__ == "__main__":
    t = threading.Tread(target=getUser)
    t.start()
    t.join()
    #......
```
代码量少很多了，但是这个方案依然需要一个类似sqlite的缓存层获取结果，如果想在当前的执行流程中拿到结果，也需要模仿c++98的那种实现

**python3.x引入了关键字async**
我们知道python有个GIL的全局锁，所以再多的thread都只能保证一个线程在单个CPU上跑，thread的出现只是让你有种异步的**感觉**，而已，处理并发的能力是很差的。所以async的本质是什么了？协程！而python的协程又是什么了？一段压栈的代码片段而已。**go也有协程，但是go的协程是内核态的线程。而python的协程是一个进程里面不断压栈的代码片段，看起来是通常意义的并发的外观，但实际上是通过一个进程上的可执行代码片把一个cpu跑满，跑透来提高利用率**，代码更简洁易读了，我们看看这段代码：
```
async def do_some_work(x):
    print('Waiting {}'.format(x))
    return 'Done after {}s'.format(x)
 
start = now()
 
coroutine = do_some_work(2)
loop = asyncio.get_event_loop()
task = asyncio.ensure_future(coroutine)
loop.run_until_complete(task)
 
print('Task ret: {}'.format(task.result()))
print('TIME: {}'.format(now() - start))
```
结果输出：
```
Waiting:  2
Task ret:  Done after 2s
TIME:  0.0003650188446044922
```
task是个future类型，可以通过task.result()拿到协程的处理结果，并且是通过注册事件来唤醒。
**c++11终于让我看到这门老大哥语言开始放下身段，吸收这些新型语言的长处，更新速度开始快起来了。我们看看c++11提供的async和future类：**
```
bool is_prime(int x)
{
    for(int i=2;i<x;i++)
    {
        if(x%i==0)
            return false;

        return true;
    }
}

int main()
{
    std::future <bool> fut = std::async(is_prime,4444444444444444443);

    std::cout<<"wait,Checking";
    std::chrono::milliseconds span(10);
    while(fut.wait_for(span)==std::future_status::timeout)
        std::cout<<'.'<<std::flush;
        bool x = fut.get();
        std::cout<<"\n4444444444444444443"<<(x?" is":"is not") << " prime.\n";
        return 0;
}
```
[这里](http://blog.csdn.net/xiangxianghehe/article/details/76359214)的async调用函数返回的fut是future类，并且做了超时处理。但是请注意，future实例调用get()方法后依然会阻塞在这里，这样一套实现只是延迟了阻塞的时机而已，相比于事件驱动还差了点意思。

终于要介绍go的**事件驱动**解决异步问题了，最近在阅读docker源码，被一段代码吸引住了，因为一开始我没看懂，不知道代码为什么这样写，可读性很差，但是看懂之后才明白这段代码的精巧之处。先贴代码：
```
// Events returns a stream of events in the daemon. It's up to the caller to close the stream
// by cancelling the context. Once the stream has been completely read an io.EOF error will
// be sent over the error channel. If an error is sent all processing will be stopped. It's up
// to the caller to reopen the stream in the event of an error by reinvoking this method.
func (cli *Client) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {

	messages := make(chan events.Message)
	errs := make(chan error, 1)

	started := make(chan struct{})
	go func() {
		defer close(errs)

		query, err := buildEventsQueryParams(cli.version, options)
		if err != nil {
			close(started)
			errs <- err
			return
		}

		resp, err := cli.get(ctx, "/events", query, nil)
		if err != nil {
			close(started)
			errs <- err
			return
		}
		defer resp.body.Close()

		decoder := json.NewDecoder(resp.body)

		close(started)
		for {
			select {
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			default:
				var event events.Message
				if err := decoder.Decode(&event); err != nil {
					errs <- err
					return
				}

				select {
				case messages <- event:
				case <-ctx.Done():
					errs <- ctx.Err()
					return
				}
			}
		}
	}() //实现匿名函数并立马执行，传参为空
	<-started //block，直至started被关闭

	return messages, errs
}
```
* started是一个channel，go语言里的channel是协程之间负责通信处理的，一个协程可以往channel里面写数据，另一个协程可以读数据，如果调用了读操作，而那边没写入，程序就会阻塞在这里。当然，close这个channel也可以结束阻塞
* 程序启动了一个协程去执行匿名函数时，"go func() {} ()"，这不会阻塞Events函数，直到"<-started"处开始阻塞
* 匿名函数在执行过程中执行到“close(started)”，此时Events函数不再阻塞，返回一个入口处初始化的 messages，这个messages也是一个channel，在这里你应该明白了吧，上层调用Events函数的那块逻辑，也可以自由控制messages了
* 协程中的匿名函数发起一个http请求，等响应完毕后把数据塞到messages channel中，上层调用Events函数的那块逻辑可以读取管道的数据，例如
```
var result  = <- messages 
```
一旦func所处的协程写完数据，result那块的逻辑被唤醒，继续程序的逻辑。这样一来就可以用同步的语法实现异步的功能，代码写起来很舒服。

现代语言把那些解决问题的优秀方案吸纳到语言自身，让编程者降低心智负担，提高生产率，这就是趋势。
