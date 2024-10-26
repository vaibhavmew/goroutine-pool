# `goroutine pool`

A simple and efficient pool to recycle goroutines instead of creating new ones on the fly.
Supports aggregating data into a buffered channel. The caller will wait for all the workers 
to complete and then get all the responses from every worker into the buffered channel. 

# Getting Started

## Installing
To start using `goroutine-pool`, install Go and run `go get`:

```sh
$ go get -u github.com/vaibhavmew/goroutine-pool
```

## Usage 
Check the following examples to understand the usage
[submit](https://github.com/vaibhavmew/goroutine-pool/blob/main/examples/submit.go)
[submit and aggregate ](https://github.com/vaibhavmew/goroutine-pool/blob/main/examples/submitandaggregate.go)

## Features
1. Get the response from p.Submit(), once the request is submitted.
2. Aggregate data from all the workers. Pass a buffered channel to the pool. The data will be inserted into it.
3. Supports timeout. No worker is blocked infinitely.
4. The pool doesn't use queues, sync.Pool or any other algorithm to fetch and return the workers.
5. No use of interface{} or it's conversion anywhere.


Increase the pool size at runtime using 
```go
err = p.Tune(10) //here 10 refers to the new pool size
if err != nil {
    panic(err)
}
```

Close the pool once the workers are done
```go
err = p.Close()
if err != nil {
    panic(err)
}
```