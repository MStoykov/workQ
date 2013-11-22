package workQ

import (
	"sync"
	"testing"
	"time"
)

type basicItemQueue struct {
	queue WorkQ
	in    chan basicItem
	out   chan basicItem
}

func newBasicItemQueue() basicItemQueue {
	queue := NewWorkQ()
	in := make(chan basicItem)
	go func() {
		wrappedIn := queue.In()
		defer close(wrappedIn)
		for item := range in {
			wrappedIn <- item
		}
	}()
	out := make(chan basicItem)
	go func() {
		wrappedOut := queue.Out()
		defer close(out)
		for item := range wrappedOut {
			basicItem, ok := item.(basicItem)
			if !ok {
				panic("WAT!?!?!")
			}
			out <- basicItem
		}
	}()
	return basicItemQueue{
		queue: queue,
		in:    in,
		out:   out,
	}
}

func (b *basicItemQueue) Out() <-chan basicItem {
	return b.out
}

func (b *basicItemQueue) In() chan<- basicItem {
	return b.in
}

type basicItem struct {
	Number int
}

func (b basicItem) Work() {
	time.Sleep(time.Millisecond * time.Duration(b.Number%50))
}

func _TestBasicUsage(t *testing.T) {
	queue := NewWorkQ()
	in := queue.In()
	basicItem := basicItem{1}
	in <- basicItem
	item := <-queue.Out()
	if item == nil || item != basicItem {
		t.Fatal("Couldn't get a single element through")
	}
}

func TestBasic50kItems(t *testing.T) {
	queue := NewWorkQ()
	in := queue.In()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 1; i < 50000; i++ {
			basicItem := basicItem{i}
			in <- basicItem
		}
		close(in)
		wg.Done()
	}()
	go func() {
		i := 1
		for item := range queue.Out() {
			basicItem, ok := item.(basicItem)
			if !ok || basicItem.Number != i {
				t.Errorf("something went wrong for i: %d and item: %v", i, basicItem)
			}
			i++
		}
		wg.Done()
	}()
	wg.Wait()
}

func TestBasicItemQueueWrapper50K(t *testing.T) {
	queue := newBasicItemQueue()
	in := queue.In()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 1; i < 50000; i++ {
			basicItem := basicItem{i}
			in <- basicItem
		}
		close(in)
		wg.Done()
	}()
	go func() {
		i := 1
		for item := range queue.Out() {
			if item.Number != i {
				t.Errorf("something went wrong for i: %d and item: %v", i, item)
			}
			i++
		}
		wg.Done()
	}()
	wg.Wait()
}
