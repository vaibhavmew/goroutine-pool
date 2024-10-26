package pool

type Request struct {
	ch chan int //do not remove this. used internally

	//add your fields here
	Input int
}

type Response struct {
	Err error //do not remove this. used internally

	//add your fields here
	Output int
}
