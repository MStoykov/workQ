package workQ

import (
	"sync"
	"testing"
	"time"
)

type basicWorkerQueue struct {
	queue WorkQ
	in    chan basicWorker
	out   chan basicWorker
}

func newBasicWorkerQueue() basicWorkerQueue {
	queue := NewWorkQ()
	in := make(chan basicWorker)
	go func() {
		wrappedIn := queue.In()
		defer close(wrappedIn)
		for worker := range in {
			wrappedIn <- worker
		}
	}()
	out := make(chan basicWorker)
	go func() {
		wrappedOut := queue.Out()
		defer close(out)
		for worker := range wrappedOut {
			basicWorker, ok := worker.(basicWorker)
			if !ok {
				panic("WAT!?!?!")
			}
			out <- basicWorker
		}
	}()
	return basicWorkerQueue{
		queue: queue,
		in:    in,
		out:   out,
	}
}

func (b *basicWorkerQueue) Out() <-chan basicWorker {
	return b.out
}

func (b *basicWorkerQueue) In() chan<- basicWorker {
	return b.in
}

type basicWorker struct {
	Number int
}

func (b basicWorker) Work() {
	time.Sleep(time.Millisecond * time.Duration(b.Number%50))
}

func _TestBasicUsage(t *testing.T) {
	queue := NewWorkQ()
	in := queue.In()
	basicWorker := basicWorker{1}
	in <- basicWorker
	worker := <-queue.Out()
	if worker == nil || worker != basicWorker {
		t.Fatal("Couldn't get a single element through")
	}
}

func TestBasic50kWorkers(t *testing.T) {
	queue := NewWorkQ()
	in := queue.In()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 1; i < 50000; i++ {
			basicWorker := basicWorker{i}
			in <- basicWorker
		}
		close(in)
		wg.Done()
	}()
	go func() {
		i := 1
		for worker := range queue.Out() {
			basicWorker, ok := worker.(basicWorker)
			if !ok || basicWorker.Number != i {
				t.Errorf("something went wrong for i: %d and worker: %v", i, basicWorker)
			}
			i++
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestBasicWorkerQueueWrapper50K(t *testing.T) {
	queue := newBasicWorkerQueue()
	in := queue.In()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 1; i < 50000; i++ {
			basicWorker := basicWorker{i}
			in <- basicWorker
		}
		close(in)
		wg.Done()
	}()
	go func() {
		i := 1
		for worker := range queue.Out() {
			if worker.Number != i {
				t.Errorf("something went wrong for i: %d and worker: %v", i, worker)
			}
			i++
		}
		wg.Done()
	}()
	wg.Wait()
}
