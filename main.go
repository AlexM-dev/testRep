package main

import (
	"fmt"
	"sync"
)

const messagesAmountPerGoroutine int = 5

func demultiplexing(dataSourceChan chan int, amount int) ([]chan int, <-chan int) {
	var out = make([]chan int, amount)
	var done = make(chan int)
	for i := range out {
		out[i] = make(chan int)
	}
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			for v := range dataSourceChan {
				for _, c := range out {
					c <- v
				}
			}
			wg.Done()
		}()
		wg.Wait()
		close(done)
	}()
	return out, done
}

func multiplexing(done <-chan int, channels ...chan int) <-chan int {
	var wg sync.WaitGroup
	multiplexedChan := make(chan int)
	multiplex := func(c <-chan int) {
		defer wg.Done()
		for {
			select {
			case i := <-c:
				multiplexedChan <- i
			case <-done:
				return
			}
		}
	}
	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}
	go func() {
		wg.Wait()
		close(multiplexedChan)
	}()
	return multiplexedChan
}

func main() {
	startDataSource := func() chan int {
		c := make(chan int)
		go func() {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 1; i <= messagesAmountPerGoroutine; i++ {
					c <- i
				}
			}()
			wg.Wait()
			close(c)
		}()
		return c
	}
	consumers, done := demultiplexing(startDataSource(), 5)
	c := multiplexing(done, consumers...)
	for data := range c {
		fmt.Println(data)
	}
}
