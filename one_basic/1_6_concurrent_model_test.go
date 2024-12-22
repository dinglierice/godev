package one_basic

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
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
