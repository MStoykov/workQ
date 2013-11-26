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

	queue.startLoop()
	return queue
}

func (w *WorkQ) startLoop() {
	go func() {
		// signal that the queue is no longer usable
		defer close(w.out)
		//waiting group for the defer above
		wg := sync.WaitGroup{}
		// channel we wait to close before we can return the worker
		waitFor := make(chan struct{})
		// the first element can be returned immediately
		close(waitFor)
		// get next worker
		for worker := range w.in {
			wg.Add(1)
			// the channel that we will give to the next to wait for and will close when we have returned
			nextWaitFor := make(chan struct{})
			go func(waitFor <-chan struct{}, nextWaitFor chan struct{}, worker Worker) {
				// in case ... something
				defer wg.Done()
				worker.Work()
				<-waitFor
				w.out <- worker
				close(nextWaitFor)
			}(waitFor, nextWaitFor, worker)
			// swap for the next
			waitFor = nextWaitFor
		}
		wg.Wait()
	}()
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
