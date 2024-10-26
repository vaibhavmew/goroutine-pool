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
	DefaultPoolSize   = runtime.NumCPU()
	DefaultTimeout    = 1 * time.Second
	DefaultIdleTime   = 1 * time.Minute
	DefaultPercentage = 50
)

var (
	ErrInvalidPoolSize  = errors.New("pool size cannot be negative")
	ErrPoolSizeMismatch = errors.New("size is larger than the worker pool")
	ErrClosePool        = errors.New("cannot close all workers")
	ErrPoolNotActive    = errors.New("pool has not started")
)

type Pool struct {
	requestCh   chan Request
	responseCh  map[int]chan Response
	f           func(Request) Response
	log         *log.Logger
	timeout     time.Duration
	maxIdleTime time.Duration
	percentage  int
	isActive    bool

	mu    sync.Mutex
	size  int
	ids   map[int]worker //all data about the workers. use sharded map or array?
	close chan struct{}
}

type worker struct {
	isActive   bool
	lastUsedAt time.Time
}

//start the goroutines at runtime?

//write blog on how to use go generate
//write benchmarks => less memory. check for interface conversion speed
//close the worker that has not been used for a long time. configure this
//used request and response pool?

func New(size int, f func(Request) Response) (*Pool, error) {
	if size < 0 {
		return &Pool{}, ErrInvalidPoolSize
	}

	if size == 0 {
		size = DefaultPoolSize
	}

	p := Pool{
		size:        size,
		f:           f,
		requestCh:   make(chan Request, 50),
		responseCh:  make(map[int]chan Response, 1),
		log:         log.Default(),
		ids:         make(map[int]worker),
		timeout:     DefaultTimeout,
		maxIdleTime: DefaultIdleTime,
	}

	return &p, nil
}

func (p *Pool) Start() {
	for i := 0; i < p.size; i++ {
		go p.Worker()
	}
	p.isActive = true
	time.Sleep(10 * time.Millisecond)
}

func (p *Pool) Close() error {

	err := p.Stop(p.Size())
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Millisecond)

	return nil
}

func (p *Pool) Tune(size int) error {
	if size <= p.size {
		return nil
	}

	if !p.isActive {
		return ErrPoolNotActive
	}

	p.mu.Lock()
	for i := 0; i < size-p.size; i++ {
		go p.Worker()
	}
	p.mu.Unlock()

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
	p.mu.Lock()
	count := len(p.ids)
	p.mu.Unlock()
	return count
}

func (p *Pool) Worker() {
	id := id()
	p.log.Println("worker started on goroutine no: " + strconv.Itoa(id))

	p.add(id)
	p.initRequestCh(id)

	for {
		select {
		case <-p.close:
			p.remove(id)
			p.log.Println("closed worker on goroutine no: " + strconv.Itoa(id))
			return
		case k := <-p.requestCh:
			p.active(id)
			k.ch <- id
			p.responseCh[id] <- p.f(k)
			p.inactive(id)
		}
	}
}

func (p *Pool) remove(workerID int) {
	p.mu.Lock()
	delete(p.ids, workerID)
	p.mu.Unlock()
}

func (p *Pool) add(workerID int) {
	p.mu.Lock()
	p.ids[workerID] = worker{}
	p.mu.Unlock()
}

func (p *Pool) active(workerID int) {
	p.mu.Lock()
	p.ids[workerID] = worker{
		isActive:   true,
		lastUsedAt: time.Now(),
	}
	p.mu.Unlock()

}

func (p *Pool) inactive(workerID int) {
	p.mu.Lock()
	lastUsedAt := p.ids[workerID].lastUsedAt
	p.ids[workerID] = worker{
		isActive:   false,
		lastUsedAt: lastUsedAt,
	}
	p.mu.Unlock()

}

func (p *Pool) initRequestCh(workerID int) {
	p.mu.Lock()
	p.responseCh[workerID] = make(chan Response)
	p.mu.Unlock()
}

// stop particular number of workers
func (p *Pool) Stop(size int) error {
	if size > p.size {
		return ErrPoolSizeMismatch
	}

	p.mu.Lock()

	p.close = make(chan struct{}, size)

	for i := 0; i < size; i++ {
		p.close <- struct{}{}
	}

	p.mu.Unlock()

	if p.size == size {
		p.log.Println("closing all workers...")
	}

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
			Err: ctx.Err(),
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
			Err: ctx.Err(),
		}
	}

	wg.Done()
}

func (p *Pool) Active() int {
	var count int
	p.mu.Lock()
	for _, v := range p.ids {
		if v.isActive {
			count++
		}
	}
	p.mu.Unlock()
	return count
}

//policy for purging. maximum how many workers can be closed
//if less no of workers exist and there is too much load. increase the worker
//do not overpurge

// remove workers that have been inactive for a long time
func (p *Pool) Purge() {

	for {
		p.purge()
		time.Sleep(1 * time.Hour)
	}

}

func (p *Pool) purge() {
	size := p.percentage * p.size / 100
	p.Stop(size)
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
