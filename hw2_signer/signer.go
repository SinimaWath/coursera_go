package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
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
	for chVal := range in {
		var strVal string
		switch val := chVal.(type) {
		case int:
			strVal = strconv.Itoa(val)
		case string:
			strVal = val
		}
		res := DataSignerCrc32(strVal) + "~" + DataSignerCrc32(DataSignerMd5(strVal))
		fmt.Println("SingleHash data:", strVal)
		fmt.Println("SingleHash Result: ", res)
		out <- res
	}
}

func MultiHash(in, out chan interface{}) {
	for chVal := range in {
		val, ok := chVal.(string)
		var res string
		if !ok {
			panic("Can' not use val as not string")
		}

		for i := 0; i < 6; i++ {
			strI := strconv.Itoa(i)
			res += DataSignerCrc32(strI + val)
		}
		fmt.Println("MultiHash data: ", chVal)
		fmt.Println("MultiHash Result: ", res)
		out <- res
	}
}

func CombineResults(in, out chan interface{}) {
	var res string
	for chVal := range in {
		val, ok := chVal.(string)
		if !ok {
			panic("Can' not use val as not string")
		}
		res += val
		res += "_"
	}

	if strings.HasSuffix(res, "_") {
		out <- strings.TrimSuffix(res, "_")
	} else {
		out <- res
	}
}

func main() {

	inputData := []int{0, 1}

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

	ExecutePipeline(hashSignJobs...)

}
