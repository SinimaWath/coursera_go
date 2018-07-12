package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(wg *sync.WaitGroup, j job, in, out chan interface{}) {
	defer wg.Done()
	defer close(out)
	j(in, out)
}

// ExecutePipeline -  выполняет задачи в конвеерном режиме. По аналогии с pipeline в unix
func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	ch1 := make(chan interface{})

	for _, j := range jobs {
		ch2 := make(chan interface{})
		wg.Add(1)
		go worker(wg, j, ch1, ch2)
		ch1 = ch2
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {

}

func MultiHash(in, out chan interface{}) {

}

func CombineResults(in, out chan interface{}) {

}

func main() {

	// jobs := []job{
	// 	job(func(in, out chan interface{}) {
	// 		out <- 1
	// 		time.Sleep(5 * time.Second)
	// 		fmt.Println("Job put 1 is done")
	// 	}),
	// 	job(func(in, out chan interface{}) {
	// 		fmt.Println("Job printer start")
	// 		for val := range in {
	// 			fmt.Println("Printed value: ", val)
	// 		}
	// 	}),
	// }

	// ExecutePipeline(jobs...)

	ch := make(chan int)
	go test(ch)
	for i := 0; i < 5; i++ {
		ch <- i
	}
	// ch2 := make(chan int)
	// ch = ch2
	// go test(ch2)
	// ch2 <- 1

}

func test(ch chan int) {
	time.Sleep(time.Second)
	for val := range ch {
		fmt.Println(val)
	}
}
