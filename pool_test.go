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

	err = p.Close(7)

	assert.NotNil(t, err)
	assert.Equal(t, "size is larger than the worker pool", err.Error())

}

func TestClosingAllWorkers(t *testing.T) {
	f := Input
	p, err := New(5, f)

	assert.Nil(t, err)

	assert.Equal(t, 5, p.Size())

	err = p.Close(5)

	assert.NotNil(t, err)
	assert.Equal(t, "cannot close all workers", err.Error())
}

func TestSubmit(t *testing.T) {
	t.Skip()
	f := Input

	p, err := New(5, f)
	assert.Nil(t, err)

	p.Start()

	request := Request{
		input: 1,
	}

	response := p.Submit(request)
	assert.Nil(t, response.err)
	assert.Equal(t, 2, response.output)
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
