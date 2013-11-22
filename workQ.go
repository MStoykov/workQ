// a try at making FIFO container that runs and waits for method of it's Items before it returns them, but doesn't block on writes.
package workQ

import "sync"

// A queue of Items that runs the Work() function on them async,
// but return them in FIFO order after Work() has finished
//
// TODO: make it more natural to get things out
// look at the test for basic implementation of a typed wrapper
type WorkQ struct {
	in  chan Item
	out chan Item
}

// the interface that needs to be implemented
type Item interface {
	// is the function that will be called and will be waited to finish before the item can be returned
	Work()
}

type itemWrapper struct {
	item Item
}

func (b itemWrapper) start(start <-chan bool) <-chan bool {
	cl := make(chan bool)
	go func() {
		defer close(cl)
		b.item.Work()
		select {
		case <-start:
		}
	}()
	return cl
}

// get a new WorkQ
func NewWorkQ() WorkQ {
	queue := WorkQ{
		in:  make(chan Item),
		out: make(chan Item),
	}

	queue.startLoop()
	return queue
}

func (w *WorkQ) startLoop() {
	go func() {
		defer close(w.out)
		wg := sync.WaitGroup{}
		nextCanReturn := make(chan bool)
		close(nextCanReturn)
		for item := range w.in {
			canReturn := itemWrapper{item}.start(nextCanReturn)
			nextCanReturn = make(chan bool)
			wg.Add(1)
			go func(canReturn <-chan bool, nextCanReturn chan bool, item Item) {
				defer wg.Done()
				<-canReturn
				w.out <- item
				close(nextCanReturn)
			}(canReturn, nextCanReturn, item)
		}
		wg.Wait()
	}()
}

// you write on this channel.
//
// if you close it the WorkQ will close the Out()
// channel after there are no more items and will become unusable
func (w *WorkQ) In() chan<- Item {
	return w.in
}

// you read from here.
//
// if the In() channel is closed this channel will close after the last element in
func (w *WorkQ) Out() <-chan Item {
	return w.out
}
