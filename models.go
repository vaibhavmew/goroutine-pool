package pool

// request
type Request struct {
	ch chan int //do not remove this. used internally

	//add your fields here
	input int
}

type Response struct {
	err error //do not remove this. used internally

	//add your fields here
	output int
}
