package main

import (
	"fmt"
	"pool/pool"
)

func main() {
	p, err := pool.New(0, pool.Input)
	if err != nil {
		panic(err)
	}

	p.Start()

	size := 100
	aggregate := make(chan pool.Response, size)
	waitCh := make(chan chan int, size)

	for i := 0; i < size; i++ {
		p.SubmitAndAggregate(pool.Request{
			Input: i,
		}, waitCh)
	}

	p.Wait(aggregate, waitCh, size)

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
