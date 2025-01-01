package one_basic

import (
	"context"
	"fmt"
)

// 素数筛选法
// 以下代码使用了context实现了goroutine的关闭操作
// 但是这一版代码会产生死锁的问题, 后续来修复
func genNaturalNumber(ctx context.Context) <-chan int {
	out := make(chan int)
	go func() {
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

func filterPrime(ctx context.Context, in <-chan int, prime int) <-chan int {
	out := make(chan int)
	go func() {
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

func Cpf1() {
	// 通过 Context 控制后台 Goroutine 状态
	ctx, cancel := context.WithCancel(context.Background())

	ch := genNaturalNumber(ctx) // 自然数序列: 2, 3, 4, ...
	for i := 0; i < 100; i++ {
		prime := <-ch // 新出现的素数
		fmt.Printf("%v: %v\n", i+1, prime)
		ch = filterPrime(ctx, ch, prime) // 基于新素数构造的过滤器
	}

	cancel()
}
