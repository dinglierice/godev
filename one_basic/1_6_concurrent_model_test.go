package one_basic

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/gatefs"
)

// TestConcurrentModelWithMutex 并发编程的核心概念是同步通信，但是同步的方式却有多种
// 1. 互斥锁
func TestConcurrentModelWithMutex(t *testing.T) {
	var mu sync.Mutex

	mu.Lock()
	go func() {
		fmt.Println("hello World")
		mu.Unlock()
	}()

	mu.Lock()
}

// 2. 通道
// 对于从无缓冲 Channel 进行的接收，发生在对该 Channel 进行的发送完成之前
func TestConcurrentModelWithChannel(t *testing.T) {
	cha := make(chan int, 5)

	go func() {
		fmt.Println("hello World")
		cha <- 1
	}()

	<-cha
}

func TestShadow(t *testing.T) {
	startTime := time.Now()
	for i := 0; i < 10; i++ {
		fmt.Println("hello world")
	}
	endTime := time.Now()
	fmt.Println("time taken: " + strconv.Itoa(int(endTime.Sub(startTime).Microseconds())))
}

func TestConcurrentModelWithChannel2(t *testing.T) {
	startTime := time.Now()
	cha := make(chan int, 10)
	for i := 0; i < cap(cha); i++ {
		fmt.Println("hello world")
		cha <- 1
	}

	for i := 0; i < cap(cha); i++ {
		<-cha
	}

	endTime := time.Now()
	fmt.Println("time taken: " + strconv.Itoa(int(endTime.Sub(startTime).Microseconds())))
}

// 使用sync.WaitGroup实现等待的效果
func TestConcurrentModelWithWaitGroup(t *testing.T) {
	startTime := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			fmt.Println("hello world")
			wg.Done()
		}()
	}
	wg.Wait()
	endTime := time.Now()
	fmt.Println("time taken: " + strconv.Itoa(int(endTime.Sub(startTime).Microseconds())))
}

// producer consumer model

// producer
func Producer(factor int, out chan<- int) {
	for i := 0; ; i++ {
		out <- i * factor
	}
}

// consumer
func Consumer(in <-chan int) {
	for v := range in {
		fmt.Println("整数倍结果为: " + strconv.Itoa(v))
	}
}

func TestProducerConsumer(t *testing.T) {
	cha := make(chan int, 64)

	go Producer(2, cha)
	go Producer(5, cha)
	go Consumer(cha)

	time.Sleep(time.Second * 5)
}

// publisher subscriber model
// see 1_6_pub_sub_model.go

// control concurrency
func TestControllConcurrency(t *testing.T) {
	// 带有最大并发阻塞的虚拟文件系统
	fs := gatefs.New(vfs.OS("/path"), make(chan bool, 8))
	fs.Open("/path/file")
}

func GenerateNatural() chan int {
	ch := make(chan int)
	go func() {
		for i := 2; ; i++ {
			ch <- i
		}
	}()
	return ch
}

func PrimeFilter(in <-chan int, prime int) chan int {
	out := make(chan int)
	go func() {
		// 持续消费提供的管道
		for {
			if i := <-in; i%prime != 0 {
				out <- i
			}
		}
	}()
	return out
}

// 我们先是调用 GenerateNatural() 生成最原始的从 2 开始的自然数序列。
// 然后开始一个 100 次迭代的循环，希望生成 100 个素数。在每次循环迭代开始的时候，管道中的第一个数必定是素数，我们先读取并打印这个素数。
// 然后基于管道中剩余的数列，并以当前取出的素数为筛子过滤后面的素数。不同的素数筛子对应的管道是串联在一起的。
func TestPrimeFilter(t *testing.T) {
	ch := GenerateNatural()
	for i := 0; i < 100; i++ {
		prime := <-ch
		fmt.Println(prime)
		ch = PrimeFilter(ch, prime)
	}
}

// 并发的安全退出
func TestConcurrentSafeExit(t *testing.T) {
	ch := make(chan int)
	select {
	case v := <-ch:
		fmt.Println("hello world " + strconv.Itoa(v))
	case <-time.After(time.Second * 5):
		fmt.Println("timeout")
	default:
	}
}

// goroutine 退出控制
func worker4ExitControl(cancel chan bool) {
	for {
		select {
		default:
			fmt.Println("hello world")
		case v := <-cancel:
			fmt.Println("receive cancel signal " + strconv.FormatBool(v))
			fmt.Println("exit")
			//return
		}
	}
}

// 存在交替执行的问题
// 由于 select 的随机选择特性和 goroutine 的并发执行，"hello world" 的打印和取消信号的打印可能会交错
// 例如，在发送取消信号的瞬间，select 可能已经选择了 default 分支，因此会先打印 "hello world"，然后再处理取消信号。
func TestWorker4ExitControl(t *testing.T) {
	cancel := make(chan bool)
	go worker4ExitControl(cancel)
	time.Sleep(time.Second * 2)
	cancel <- true
}

// 关闭通道
// 广播关闭通道
func TestWorker4ExitControlWithCloseCmd(t *testing.T) {
	cancel := make(chan bool)
	for i := 0; i < 10; i++ {
		go worker4ExitControl(cancel)
	}
	time.Sleep(time.Second * 2)
	close(cancel)
}

// 因为 main 线程并没有等待各个工作 Goroutine 退出工作完成的机制
// 结合上面乱序输出的问题, 测试使用sync.WaitGroup
func workerWithWg(wg *sync.WaitGroup, cancel chan bool) {
	defer wg.Done()
	for {
		select {
		default:
			fmt.Println("hello world")
		case v := <-cancel:
			fmt.Println("receive cancel signal " + strconv.FormatBool(v))
			fmt.Println("exit")
			return
		}
	}
}

// 测试使用sync.WaitGroup
func TestWorker4ExitControlWithWg(t *testing.T) {
	cancel := make(chan bool)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go workerWithWg(&wg, cancel)
	}

	time.Sleep(time.Second * 2)
	close(cancel)
	wg.Wait()
}

// 并发withContext
// 标准库增加了一个 context 包，用来简化对于处理单个请求的多个 Goroutine 之间与请求域的数据、超时和退出等操作
// 我们可以用 context 包来重新实现前面的线程安全退出或超时的控制
func workWithContext(wg *sync.WaitGroup, ctx context.Context) error {
	defer wg.Done()
	for {
		select {
		default:
			fmt.Println("hello world")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func TestWithWorkWithContext(t *testing.T) {
	var wg sync.WaitGroup
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go workWithContext(&wg, ctx)
	}
	time.Sleep(time.Second * 3)
	cancelFunc()

	wg.Wait()
}
