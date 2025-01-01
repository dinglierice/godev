package one_basic

import (
	"context"
	"fmt"
	"sync"
)

func genNaturalNumber2(ctx context.Context, wg *sync.WaitGroup) <-chan int {
	out := make(chan int)
	go func() {
		// 如果某个 filterPrime2 goroutine 阻塞在 <-in，而 in 通道没有被关闭，那么这个 goroutine 永远无法退出，导致死锁。
		defer wg.Done()
		for i := 2; ; i++ {
			select {
			case out <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func filterPrime2(ctx context.Context, prime int, wg *sync.WaitGroup, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer wg.Done()
		for {
			if i := <-in; i%prime != 0 {
				select {
				case <-ctx.Done():
					return
				case out <- i:
				}
			}
		}
	}()
	return out
}

func Main2() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	ch := genNaturalNumber2(ctx, &wg) // 自然数序列: 2, 3, 4, ...
	wg.Add(1)
	for i := 0; i < 100; i++ {
		prime := <-ch // 新出现的素数
		fmt.Printf("%v: %v\n", i+1, prime)
		wg.Add(1)
		ch = filterPrime2(ctx, prime, &wg, ch) // 基于新素数构造的过滤器
	}
	cancel()
	wg.Wait()
}
