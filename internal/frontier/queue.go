package frontier

import "sync"

type Queue struct{
	totalQueued int
	elements []string
	mu sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		elements: make([]string, 0),
	}
}

func (q* Queue) Enqueue(item string){
	q.mu.Lock()
	defer q.mu.Unlock()
	q.elements = append(q.elements, item)
	q.totalQueued++
}

func (q *Queue) Dequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if(len(q.elements) == 0) {
		return "", false
	}
	u := q.elements[0]
	q.elements = q.elements[1:]
	return u, true
}

func (q *Queue)Size() int{
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.elements)
}

func (q* Queue) TotalQueued() int{
	return q.totalQueued
}

