package pool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	f := Input
	p, err := New(5, f)

	assert.Nil(t, err)

	assert.Equal(t, 5, p.Size())
}

func TestInvalidPoolSize(t *testing.T) {
	f := Input
	_, err := New(-1, f)

	assert.NotNil(t, err)
	assert.Equal(t, "pool size cannot be negative", err.Error())
}

func TestTunePool(t *testing.T) {
	f := Input
	p, err := New(5, f)

	assert.Nil(t, err)

	assert.Equal(t, 5, p.Size())

	err = p.Tune(7)

	assert.NotNil(t, err)
	assert.Equal(t, "pool has not started", err.Error())
}

func TestClosePoolInvalidSize(t *testing.T) {
	f := Input
	p, err := New(5, f)

	assert.Nil(t, err)

	assert.Equal(t, 5, p.Size())

	err = p.Stop(7)

	assert.NotNil(t, err)
	assert.Equal(t, "size is larger than the worker pool", err.Error())

}

func TestClosingAllWorkers(t *testing.T) {
	f := Input
	p, err := New(5, f)

	p.Start()

	assert.Nil(t, err)

	assert.Equal(t, 5, p.Size())

	err = p.Stop(5)

	assert.Nil(t, err)
}

func TestSubmit(t *testing.T) {
	f := Input

	p, err := New(5, f)
	assert.Nil(t, err)

	p.Start()

	request := Request{
		Input: 1,
	}

	response := p.Submit(request)
	assert.Nil(t, response.Err)
	assert.Equal(t, 2, response.Output)
}

func TestSubmitTimeout(t *testing.T) {
	f := Timeout

	p, err := New(5, f)
	assert.Nil(t, err)

	p.Start()

	request := Request{
		Input: 1,
	}

	response := p.Submit(request)
	assert.NotNil(t, response.Err)
	assert.Equal(t, "context deadline exceeded", response.Err.Error())
}

func TestSubmitAndAggregate(t *testing.T) {
	f := Input

	p, err := New(5, f)
	assert.Nil(t, err)

	p.Start()

	size := 10
	aggregate := make(chan Response, size)
	waitCh := make(chan chan int, size)

	for i := 0; i < size; i++ {
		p.SubmitAndAggregate(Request{
			Input: i,
		}, waitCh)
	}

	p.Wait(aggregate, waitCh, size)

	close(aggregate)

	assert.Equal(t, 10, len(aggregate))

	for response := range aggregate {
		assert.Nil(t, response.Err)
	}
}

func TestSubmitAndAggregateError(t *testing.T) {
	f := Error

	p, err := New(5, f)
	assert.Nil(t, err)

	p.Start()

	size := 10
	aggregate := make(chan Response, size)
	waitCh := make(chan chan int, size)

	for i := 0; i < size; i++ {
		p.SubmitAndAggregate(Request{
			Input: i,
		}, waitCh)
	}

	p.Wait(aggregate, waitCh, size)

	close(aggregate)

	assert.Equal(t, 10, len(aggregate))

	for response := range aggregate {
		assert.NotNil(t, response.Err)
		assert.Equal(t, "invalid input", response.Err.Error())
	}
}

func TestStartPool(t *testing.T) {
	f := Input

	p, err := New(5, f)
	assert.Nil(t, err)

	p.Start()
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 5, len(p.List()))
	assert.True(t, p.isActive)
}
