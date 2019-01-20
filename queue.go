package groundcontrol

import (
	"context"
	"sync"
)

var GlobalQueue = NewQueue(2)

func init() {
	go GlobalQueue.Work(context.Background())
}

// Queue is a simple queue with concurency support.
type Queue struct {
	ch          chan func()
	concurrency int
}

// NewQueue creates a queue with given concurrency.
func NewQueue(concurrency int) *Queue {
	return &Queue{
		ch:          make(chan func()),
		concurrency: concurrency,
	}
}

// Work tells the queue to start executing jobs.
//
// It blocks until the context is canceled.
func (q *Queue) Work(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(q.concurrency)

	for i := 0; i < q.concurrency; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case job := <-q.ch:
					job()
				}
			}
		}()
	}

	wg.Wait()
	return ctx.Err()
}

// Do puts a job at the end of the queue and blocks until executed.
func (q *Queue) Do(job func()) {
	done := make(chan struct{})
	q.ch <- func() {
		job()
		close(done)
	}
	<-done
}

// DoError puts a job that can return an error at the end of the queue and
// blocks until executed.
func (q *Queue) DoError(job func() error) error {
	done := make(chan error, 1)
	q.ch <- func() {
		done <- job()
	}
	return <-done
}
