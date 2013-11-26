// a try at making FIFO container that runs and waits for method of it's Workers before it returns them, but doesn't block on writes.
package workQ

import "sync"

// A queue of Workers that runs the Work() function on them async,
// but return them in FIFO order after Work() has finished
//
// TODO: make it more natural to get things out
// look at the test for basic implementation of a typed wrapper
type WorkQ struct {
	in  chan Worker
	out chan Worker
	wg  sync.WaitGroup
}

// the interface that needs to be implemented
type Worker interface {
	// is the function that will be called and will be waited to finish before the worker can be returned
	Work()
}

// get a new WorkQ
func NewWorkQ() WorkQ {
	queue := WorkQ{
		in:  make(chan Worker),
		out: make(chan Worker),
	}

	go queue.loop()
	return queue
}

// call Worker.Work, return the Worker after both the Work has finished and waitFor has been closed
// and close the return channel after all that.
func (w *WorkQ) doWork(waitFor <-chan struct{}, worker Worker) chan struct{} {
	nextWaitFor := make(chan struct{})
	go func() {
		// in defer in case Work panics
		defer w.wg.Done()
		defer close(nextWaitFor)
		worker.Work()
		<-waitFor
		w.out <- worker
	}()
	return nextWaitFor
}

func (w *WorkQ) loop() {
	// signal that the queue is no longer usable
	defer close(w.out)
	// channel we wait to close before we can return the worker
	waitFor := make(chan struct{})
	// the first element can be returned immediately
	close(waitFor)
	// get next worker
	for worker := range w.in {
		w.wg.Add(1)
		// the channel that we will give to the next to wait for and will close when we have returned
		waitFor = w.doWork(waitFor, worker)
	}
	// wait for the not yet returned items
	w.wg.Wait()
}

// you write on this channel.
//
// if you close it the WorkQ will close the Out()
// channel after there are no more workers and will become unusable
func (w *WorkQ) In() chan<- Worker {
	return w.in
}

// you read from here.
//
// if the In() channel is closed this channel will close after the last element in
func (w *WorkQ) Out() <-chan Worker {
	return w.out
}
