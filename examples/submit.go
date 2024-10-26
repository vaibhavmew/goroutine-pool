package examples

import (
	"fmt"
	"pool/pool"
)

//your function should follow the format mentioned in pool.Input

func Submit() {
	p, err := pool.New(0, pool.Input)
	if err != nil {
		panic(err)
	}

	p.Start()

	response := p.Submit(pool.Request{
		Input: 1,
	})

	if response.Err != nil {
		fmt.Println(response.Err)
	} else {
		fmt.Println(response.Output)
	}

	//close the pool
	p.Close()
}
