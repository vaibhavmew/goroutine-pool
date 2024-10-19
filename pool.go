package pool

import (
	"context"
	"errors"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	DefaultPoolSize = runtime.NumCPU()
	DefaultTimeout  = 30 * time.Second
)

var (
	ErrInvalidPoolSize  = errors.New("pool size cannot be negative")
	ErrPoolSizeMismatch = errors.New("size is larger than the worker pool")
	ErrClosePool        = errors.New("cannot close all workers")
	ErrPoolNotActive    = errors.New("pool has not started")
)

type Pool struct {
	requestCh  chan Request
	responseCh map[int]chan Response
	f          func(Request) Response
	log        *log.Logger
	timeout    time.Duration
	isActive   bool

	mu    sync.Mutex
	size  int
	ids   map[int]bool
	close chan struct{}
}

func Input(r Request) Response {
	time.Sleep(30 * time.Second)

	return Response{
		output: r.input + 1,
		err:    nil,
	}

	//error state
	// return Response{
	// 	err: errors.New("error while processing"),
	// }
}

//start the goroutines at runtime?

//write blog on how to use go generate
//write benchmarks => less memory. check for interface conversion speed

func New(size int, f func(Request) Response) (*Pool, error) {
	if size < 0 {
		return &Pool{}, ErrInvalidPoolSize
	}

	if size == 0 {
		size = DefaultPoolSize
	}

	p := Pool{
		size:       size,
		f:          f,
		requestCh:  make(chan Request, 1),
		responseCh: make(map[int]chan Response),
		log:        log.Default(),
		ids:        make(map[int]bool),
		timeout:    DefaultTimeout,
	}

	return &p, nil
}

func (p *Pool) Start() {
	for i := 0; i < p.size; i++ {
		go p.Worker()
	}
	p.isActive = true
}

func (p *Pool) Tune(size int) error {
	if size <= p.size {
		return nil
	}

	if !p.isActive {
		return ErrPoolNotActive
	}

	for i := 0; i < size-p.size; i++ {
		go p.Worker()
	}

	p.size = size

	return nil
}

func (p *Pool) List() []int {
	var ids []int

	p.mu.Lock()
	for k := range p.ids {
		ids = append(ids, k)
	}
	p.mu.Unlock()

	return ids
}

func (p *Pool) Size() int {
	return p.size
}

func (p *Pool) Worker() {
	id := id()
	p.log.Println("worker started on goroutine no: " + strconv.Itoa(id))
	time.Sleep(10 * time.Millisecond)

	p.mu.Lock()
	p.ids[id] = false
	p.mu.Unlock()

	for {
		select {
		case <-p.close:
			delete(p.ids, id) //decrease size from here or close?
			p.log.Println("closed worker on goroutine no: " + strconv.Itoa(id))
			time.Sleep(100 * time.Millisecond)
			return
		case k := <-p.requestCh:
			p.ids[id] = true
			k.ch <- id
			p.responseCh[id] <- p.f(k)
			p.ids[id] = false
		default:
			//check
		}
	}
}

func (p *Pool) Close(size int) error {
	if size > p.size {
		return ErrPoolSizeMismatch
	}

	if p.size == size {
		return ErrClosePool
	}

	p.mu.Lock()

	p.close = make(chan struct{}, size)

	for i := 0; i < size; i++ {
		p.close <- struct{}{}
	}

	p.size -= size

	p.mu.Unlock()

	return nil
}

func (p *Pool) Submit(r Request) Response {
	wait := make(chan int)

	r.ch = wait
	p.requestCh <- r

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	select {
	case response := <-p.responseCh[<-wait]:
		return response
	case <-ctx.Done():
		return Response{
			err: ctx.Err(),
		}
	}
}

func (p *Pool) SubmitAndAggregate(r Request, wg *sync.WaitGroup, ch chan Response) {
	wait := make(chan int)

	r.ch = wait
	p.requestCh <- r

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	select {
	case response := <-p.responseCh[<-wait]:
		ch <- response
	case <-ctx.Done():
		ch <- Response{
			err: ctx.Err(),
		}
	}

	wg.Done()
}

func (p *Pool) Active() int {
	var count int
	p.mu.Lock()
	for _, v := range p.ids {
		if v {
			count++
		}
	}
	p.mu.Unlock()
	return count
}

// goroutine id
func id() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(err)
	}
	return id
}
