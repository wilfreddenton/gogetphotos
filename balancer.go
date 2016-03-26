package main

import (
	"container/heap"
	"gopkg.in/cheggaaa/pb.v1"
	"strconv"
)

type Balancer struct {
	pool Pool
	bars []*pb.ProgressBar
	in   chan *Resource
	done chan *WorkerDeposit
}

func (b *Balancer) dispatch(res *Resource) {
	w := heap.Pop(&b.pool).(*Worker)
	w.resources <- res
	w.load += res.Size
	heap.Push(&b.pool, w)
}

func (b *Balancer) completed(depo *WorkerDeposit) {
	w := depo.worker
	r := depo.resource
	w.load -= r.Size
	heap.Remove(&b.pool, w.index)
	heap.Push(&b.pool, w)
}

func (b *Balancer) Balance() {
	for {
		select {
		case res := <-b.in:
			b.dispatch(res)
		case worker := <-b.done:
			b.completed(worker)
		}
	}
}

func NewBalancer(name string, numWorkers int, in chan *Resource, out chan *Resource, end chan *Resource) *Balancer {
	pool := make(Pool, numWorkers)
	done := make(chan *WorkerDeposit)
	bars := make([]*pb.ProgressBar, numWorkers)
	for i := 0; i < numWorkers; i += 1 {
		bars[i] = pb.New(0).Prefix(name + " " + strconv.Itoa(i+1) + ":")
		pool[i] = NewWorker(i, bars[i])
		go pool[i].work(out, in, end, done)
	}
	return &Balancer{pool, bars, in, done}
}
