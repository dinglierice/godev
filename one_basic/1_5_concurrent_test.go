package one_basic

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 常见的并发模型
// 1、多线程并发模型
// 2、基于消息的并发模型 --CSP理论

// goroutine之间是共享内存的

// goroutine和系统线程之间的关系
// 1、更小的栈 -> 系统级线程的栈在2-4m之间, 大小固定; goroutine的栈默认在2-4k之间, 栈大小可变
// 2、独立的调度器 -> m - n 模型

// 原子操作
// 最小的不可并发的操作
// sync.Mutex -> 粗粒度地模仿通过CPU指令实现的锁

var total struct {
	sync.Mutex
	value int
}

func workerNoLock(wg *sync.WaitGroup) {
	defer wg.Done()

	// 无所并发
	for i := 0; i < 10000; i++ {
		total.value++
	}
}

func worker(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 10000; i++ {
		total.Lock()
		total.value++
		total.Unlock()
	}
}

func Test(t *testing.T) {
	fmt.Println("现在是没有锁的情况")
	var wg sync.WaitGroup
	wg.Add(2)

	go workerNoLock(&wg)
	go workerNoLock(&wg)
	wg.Wait()

	println(total.value)

	total.value = 0
	fmt.Println("现在是有锁的情况")
	var wg2 sync.WaitGroup
	wg2.Add(2)
	go worker(&wg2)
	go worker(&wg2)
	wg2.Wait()
	println(total.value)
}

// 可以得到输出
// === RUN   Test
// 现在是没有锁的情况
// 12789
// 现在是有锁的情况
// 20000
// --- PASS: Test (0.00s)
// PASS

// 不建议使用互斥锁来保护共享资源
// 使用标准库的包，提供了原子化的操作能力
var total2 uint64

func workerWithStdAtomicMethod(wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 10000; i++ {
		atomic.AddUint64(&total2, 1)
	}
}

func TestWithStdAtomicMethod(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	go workerWithStdAtomicMethod(&wg)
	go workerWithStdAtomicMethod(&wg)
	wg.Wait()
	println(total2)
}

// 单例模式
type Singleton struct {
}

var (
	instance     *Singleton
	instanceFlag uint32
	m            sync.Mutex
)

// 实现了一个双重检查模式的单例模式
func GetInstance() *Singleton {
	if atomic.LoadUint32(&instanceFlag) == 1 {
		return instance
	}

	// 假设此刻有多个线程等待
	m.Lock()
	defer m.Unlock()

	// 如果不判断, 第两个线程会先后创建单例
	if instance == nil {
		defer atomic.StoreUint32(&instanceFlag, 1)
		instance = &Singleton{}
	}
	return instance
}

// 标准库有现成的实现
var (
	singleton2 *Singleton
	once       sync.Once
)

func GetInstance2() *Singleton {
	// 顾名思义 只会执行一次的操作
	once.Do(func() {
		singleton2 = &Singleton{}
	})
	return singleton2
}

// 针对复杂对象的原子操作
// 下面一个例子, 多线程操作共享配置信息
var cf atomic.Value

func TestSingletonProviderAndService(t *testing.T) {
	cf.Store(map[string]string{"key": "value"})

	// 启动一个协程, 每秒钟更新一次s
	go func() {
		for {
			time.Sleep(time.Second)
			cf.Store(map[string]string{"key": "value"})
		}
	}()

	go func() {
		for r := range requests() {
			cfg := cf.Load()
			processRequest(r, cfg)
			// ...
		}
	}()
}

func processRequest(r struct{}, cfg any) {

}

func requests() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		for {
			c <- struct{}{}
		}
	}()
	return c
}

// 顺序一致的内存模型
var (
	str  string
	flag bool
)

func stepUp() {
	str = "hello"
	flag = true
}

func TestOrderConsistency(t *testing.T) {
	go stepUp()
	// 只要flag显示没有完成, 持续阻塞主协程

	// Go 语言并不保证在 main 函数中观测到的对 done 的写入操作发生在对字符串 a 的写入的操作之后
	// 因此程序很可能打印一个空字符串。
	// 更糟糕的是，因为两个线程之间没有同步事件，setup线程对 done 的写入操作甚至无法被 main 线程看到，main函数有可能陷入死循环中
	for !flag {
		//前提：不可排序

		//如果在一个 Goroutine 中顺序执行 a = 1; b = 2; 两个语句，虽然在当前的 Goroutine 中可以认为 a = 1; 语句先于 b = 2; 语句执行
		//但是在另一个 Goroutine 中 b = 2; 语句可能会先于 a = 1; 语句执行，甚至在另一个 Goroutine 中无法看到它们的变化（可能始终在寄存器中）。
		//也就是说在另一个 Goroutine 看来, a = 1; b = 2;两个语句的执行顺序是不确定的。
		fmt.Println("flag is not true")
		// ...
	}
	println(str)
}

// CLAUDE解释为什么会发生内存一致性问题
// 同一个 goroutine 内的操作确实是按顺序执行的。然而，问题出在不同 goroutine 之间的交互上。让我解释一下为什么即使在这种情况下也可能出现内存一致性问题：

// 1、编译器和 CPU 优化：
// 编译器或 CPU 可能会对指令进行重排序以提高性能。虽然这种重排序在单个 goroutine 内保证不会改变程序的行为，但它可能影响其他 goroutine 观察到的执行顺序。
// 2、内存可见性：
// 在没有proper synchronization的情况下，一个 goroutine 对变量的修改可能不会立即对其他 goroutine 可见。这是由于现代计算机架构中的内存模型造成的。
// 3、缓存一致性：
// 在多核处理器上，每个核心可能有自己的缓存。一个核心上的写操作可能不会立即反映在其他核心的缓存中。
// 4、读取的顺序：
// 主 goroutine 可能在 flag 变为 true 后立即读取 str，但由于缓存或内存可见性问题，可能还没有看到 str 的最新值。

// 通过通信来给原子事件进行排序
// channel 操作不仅提供了同步，还建立了一个"happens-before"关系。这意味着在 channel 发送操作之前的所有内存写入，对接收操作之后的代码都是可见的。
func TestOrderConsistencyByChannel(t *testing.T) {
	done := make(chan int)
	var stringToBePrint string

	go func() {
		stringToBePrint = "nice"
		done <- 1
	}()

	<-done
	println(stringToBePrint)
}

// 通过互斥锁也可以实现
func TestOrderConsistencyByMutex(t *testing.T) {
	var m sync.Mutex
	var stringToBePrint string

	m.Lock()
	go func() {
		stringToBePrint = "nice"
		m.Unlock()
	}()
	m.Lock()
	println(stringToBePrint)
}

// 深入理解goroutine初始化的顺序
// pkg1 -> pkg2 -> pkg3 -> const -> var -> init -> main
// 所有的 init中的goroutine 和main是并发执行的
// 所有的 init 函数和 main 函数都是在主线程完成，它们也是满足顺序一致性模型的

// 在go中一个并发执行的demo
func TestChannelCocurrent(t *testing.T) {
	c := make(chan int, 3)
	works := []func(){
		func() { println("1"); time.Sleep(1 * time.Second) },
		func() { println("2"); time.Sleep(1 * time.Second) },
		func() { println("3"); time.Sleep(1 * time.Second) },
		func() { println("4"); time.Sleep(1 * time.Second) },
		func() { println("5"); time.Sleep(1 * time.Second) },
	}

	for i, work := range works {
		fmt.Println("work" + strconv.Itoa(i))
		go func(w func()) {
			c <- 1
			w()
			<-c
		}(work)
	}

	fmt.Println("main")

	select {}
}
