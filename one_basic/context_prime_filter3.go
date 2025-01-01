package one_basic

import (
	"context"
	"sync"
)

// 直接通过修改源代码的方式来避免死锁
func genNaturalNumber3(ctx context.Context, wg *sync.WaitGroup) <-chan int {
	out := make(chan int)
	go func() {
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

func filerPrime3(ctx context.Context, wg *sync.WaitGroup, prime int, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case i := <-in:
				if i%prime != 0 {
					select {
					case <-ctx.Done():
						return
					case out <- i:
					}
				}
			}
		}
	}()
	return out
}
