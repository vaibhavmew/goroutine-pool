package main

import (
	"fmt"
	"pool/pool"
	"sync"
)

func main() {
	p, err := pool.New(0, pool.Input)
	if err != nil {
		panic(err)
	}

	p.Start()

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
			fmt.Println("error for input")
		} else {
			fmt.Println(k.Output)
		}
	}

	//close pool
	p.Close()

}
