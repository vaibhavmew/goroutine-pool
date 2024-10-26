package examples

import (
	"fmt"
	"pool/pool"
	"sync"
)

// default pool size is runtime.NumCPU()
func SubmitAndAggregate() {
	p, err := pool.New(0, pool.Input)
	if err != nil {
		panic(err)
	}

	p.Start()

	var wg sync.WaitGroup
	aggregate := make(chan pool.Response, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go p.SubmitAndAggregate(pool.Request{
			Input: i,
		}, &wg, aggregate)
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
