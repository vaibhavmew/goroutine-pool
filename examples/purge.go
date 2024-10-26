package examples

import (
	"pool/pool"
)

func Purge() {
	p, err := pool.New(0, pool.Input)
	if err != nil {
		panic(err)
	}

	p.Start()

	//closes 5 workers from the pool
	//safe for concurrent use with other goroutines
	p.Close(5)
}
