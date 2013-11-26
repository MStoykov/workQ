package workQ

import (
	"sync"
	"testing"
	"time"
)

type sleepWorkerQueue struct {
	queue WorkQ
	in    chan sleepWorker
	out   chan sleepWorker
}

func newSleepWorkerQueue() sleepWorkerQueue {
	queue := NewWorkQ()
	in := make(chan sleepWorker)
	go func() {
		wrappedIn := queue.In()
		defer close(wrappedIn)
		for worker := range in {
			wrappedIn <- worker
		}
	}()
	out := make(chan sleepWorker)
	go func() {
		wrappedOut := queue.Out()
		defer close(out)
		for worker := range wrappedOut {
			sleepWorker, ok := worker.(sleepWorker)
			if !ok {
				panic("WAT!?!?!")
			}
			out <- sleepWorker
		}
	}()
	return sleepWorkerQueue{
		queue: queue,
		in:    in,
		out:   out,
	}
}

func (b *sleepWorkerQueue) Out() <-chan sleepWorker {
	return b.out
}

func (b *sleepWorkerQueue) In() chan<- sleepWorker {
	return b.in
}

type sleepWorker struct {
	Number int
}

func (b sleepWorker) Work() {
	time.Sleep(time.Millisecond * time.Duration(b.Number%50))
}

func _TestBasicUsage(t *testing.T) {
	queue := NewWorkQ()
	in := queue.In()
	sleepWorker := sleepWorker{1}
	in <- sleepWorker
	worker := <-queue.Out()
	if worker == nil || worker != sleepWorker {
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
			sleepWorker := sleepWorker{i}
			in <- sleepWorker
		}
		close(in)
		wg.Done()
	}()
	go func() {
		i := 1
		for worker := range queue.Out() {
			sleepWorker, ok := worker.(sleepWorker)
			if !ok || sleepWorker.Number != i {
				t.Errorf("something went wrong for i: %d and worker: %v", i, sleepWorker)
			}
			i++
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestSleepWorkerQueueWrapper50K(t *testing.T) {
	queue := newSleepWorkerQueue()
	in := queue.In()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 1; i < 50000; i++ {
			sleepWorker := sleepWorker{i}
			in <- sleepWorker
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
