package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
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
	var ch1 chan interface{}
	for _, j := range jobs {
		ch2 := make(chan interface{})
		wg.Add(1)
		go worker(wg, j, ch1, ch2)
		ch1 = ch2
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	for chVal := range in {
		wg.Add(1)
		go func(chVal interface{}, wg *sync.WaitGroup) {
			defer wg.Done()

			var strVal string
			switch val := chVal.(type) {
			case int:
				strVal = strconv.Itoa(val)
			case string:
				strVal = val
			}

			ch1 := make(chan string, 1)
			ch2 := make(chan string, 1)
			worker := func(strVal string, out chan<- string) {
				out <- DataSignerCrc32(strVal)
			}

			go worker(strVal, ch1)

			mutex.Lock()
			md5Val := DataSignerMd5(strVal)
			mutex.Unlock()

			go worker(md5Val, ch2)

			first := <-ch1
			fmt.Println("SH first: ", first)
			second := <-ch2
			fmt.Println("Sh second: ", second)

			out <- (first + "~" + second)
		}(chVal, wg)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for chVal := range in {
		wg.Add(1)
		go func(wg *sync.WaitGroup, chVal interface{}) {
			defer wg.Done()
			val, ok := chVal.(string)
			var res string
			if !ok {
				panic("Can't use val as not string")
			}

			chSlice := make([]chan string, 6, 6)
			for i := 0; i < 6; i++ {
				chSlice[i] = make(chan string, 1)
				go func(i int, val string, ch chan<- string) {
					strI := strconv.Itoa(i)
					ch <- DataSignerCrc32(strI + val)
				}(i, val, chSlice[i])
			}

			for _, val := range chSlice {
				res += <-val
			}
			fmt.Println("MultiHash data: ", chVal)
			fmt.Println("MultiHash Result: ", res)
			out <- res
		}(wg, chVal)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var strVals []string
	for chVal := range in {
		val, ok := chVal.(string)
		if !ok {
			panic("Can' not use val as not string")
		}
		strVals = append(strVals, val)
	}

	sort.Strings(strVals)
	out <- strings.Join(strVals, "_")
}

func main() {

	inputData := []int{0, 1, 1, 2, 3, 5}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("Can't convert to string")
			}
			fmt.Println(data)
		}),
	}

	start := time.Now()
	ExecutePipeline(hashSignJobs...)
	end := time.Since(start)
	fmt.Println("Time passed from: ", end)

}
