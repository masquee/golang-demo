package main

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/atomic"
)

func main() {
	inited := atomic.NewBool(false)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		inited.CompareAndSwap(false, true)
		time.Sleep(time.Second * 10)
		wg.Done()
	}()

	time.Sleep(time.Second)
	startTime := time.Now().Second()
	res := inited.CompareAndSwap(false, true)
	fmt.Println("res: ", res)
	endTime := time.Now().Second()
	fmt.Println(endTime - startTime)
	wg.Wait()
}
