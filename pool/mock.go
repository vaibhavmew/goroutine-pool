package pool

import (
	"errors"
	"time"
)

func Input(r Request) Response {
	return Response{
		Output: r.Input + 1,
		Err:    nil,
	}
}

func Error(r Request) Response {
	return Response{
		Err: errors.New("invalid input"),
	}
}

func Timeout(r Request) Response {
	time.Sleep(10 * time.Second)
	return Response{
		Output: r.Input + 1,
		Err:    nil,
	}
}
