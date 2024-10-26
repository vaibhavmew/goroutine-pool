package examples

import (
	"pool/pool"
)

func Tune() {
	p, err := pool.New(0, pool.Input)
	if err != nil {
		panic(err)
	}

	p.Start()

	size := 15

	//if the size is less than the pool size, then no new workers will be spawned
	//otherwise the poolsize - size no of workers will be spawned
	//if 10 workers are already active and 15 is sent then 5 new workers will be created.
	//safe for concurrent use with other goroutines
	p.Tune(size)

}
