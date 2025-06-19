package frontier
import "sync"

type Queue struct {
	totalQueued int
	elements    []string
	mu          sync.Mutex
}



func (q *Queue) Enqueue(u string) {
	q.mu.Lock()
	q.elements = append(q.elements, u)
	q.totalQueued++
	q.mu.Unlock()
}

func NewQueue() *Queue {
	return &Queue{
		elements: make([]string, 0),
	}
}

// --- NEW --- pop front (BFS)
func (q *Queue) PopFront() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.elements) == 0 {
		return "", false
	}
	u := q.elements[0]
	q.elements = q.elements[1:]
	return u, true
}

// --- NEW --- pop back (DFS)
func (q *Queue) PopBack() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := len(q.elements)
	if n == 0 {
		return "", false
	}
	u := q.elements[n-1]
	q.elements = q.elements[:n-1]
	return u, true
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.elements)
}

func (q *Queue) TotalQueued() int { return q.totalQueued }

