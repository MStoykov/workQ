package workQ

import (
	"sync"
	"testing"
	"time"
)

type BasicItem struct {
	Number int
}

func (b BasicItem) Work() {
	time.Sleep(time.Millisecond * time.Duration(b.Number%50))
}

func _TestBasicUsage(t *testing.T) {
	queue := NewWorkQ()
	in := queue.In()
	basicItem := BasicItem{1}
	in <- basicItem
	item := <-queue.Out()
	if item == nil || item != basicItem {
		t.Fatal("Couldn't get a single element through")
	}
}

func TestBasicTenItems(t *testing.T) {
	queue := NewWorkQ()
	in := queue.In()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 1; i < 50000; i++ {
			basicItem := BasicItem{i}
			in <- basicItem
		}
		close(in)
		wg.Done()
	}()
	go func() {
		i := 1
		for item := range queue.Out() {
			basicItem, ok := item.(BasicItem)
			if !ok || basicItem.Number != i {
				t.Errorf("something went wrong for i: %d and item: %v", i, basicItem)
			}
			i++
		}
		wg.Done()
	}()
	wg.Wait()
}
