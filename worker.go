package main

import (
	"gopkg.in/cheggaaa/pb.v1"
	"sync"
)

type Worker struct {
	resources chan *Resource
	id        int
	load      int
	index     int
	bar       *pb.ProgressBar
	mutex     *sync.Mutex
}

type WorkerDeposit struct {
	worker   *Worker
	resource *Resource
}

func (w *Worker) work(send, resend, sendErr chan *Resource, done chan *WorkerDeposit) {
	for res := range w.resources {
		go func(res *Resource) {
			res, err := res.fn(w.bar, w.mutex)
			if err != nil {
				if res.attempts < 3 {
					resend <- res
				} else {
					sendErr <- res
				}
			} else {
				send <- res
			}
			done <- &WorkerDeposit{w, res}
		}(res)
	}
}

func NewWorker(id int, bar *pb.ProgressBar) *Worker {
	return &Worker{make(chan *Resource), id, 0, 0, bar, new(sync.Mutex)}
}
