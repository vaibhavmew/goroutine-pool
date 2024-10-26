package main

import (
	"fmt"
	"pool/pool"
	"sync"
	"time"
)

func main() {
	p, err := pool.New(0, pool.Input)
	if err != nil {
		panic(err)
	}

	p.Start()

	time.Sleep(100 * time.Millisecond)

	var wg sync.WaitGroup
	aggregate := make(chan pool.Response, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		r := pool.Request{
			Input: i,
		}
		go p.SubmitAndAggregate(r, &wg, aggregate)
	}

	wg.Wait()
	close(aggregate)

	for k := range aggregate {
		if k.Err != nil {
			fmt.Println("error for input", k.Input)
		} else {
			fmt.Println(k.Input, k.Output)
		}
	}

}
