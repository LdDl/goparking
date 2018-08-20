package fifo

import (
	"log"
	"sync"
)

type fifo struct {
	items []interface{}
	first int
	last  int
	next  *fifo
}

func newfifo(maxSize int) (f *fifo) {
	return &fifo{
		items: make([]interface{}, maxSize),
	}
}

// FIFOQueue - struct for circular buffer
type FIFOQueue struct {
	head    *fifo
	tail    *fifo
	maxsize int
	count   int
	lock    sync.Mutex
}

// NewQueue - creates new instance of FIFOQueue
func NewQueue(maxSize int) (q *FIFOQueue) {
	init := newfifo(maxSize)
	q = &FIFOQueue{
		head:    init,
		tail:    init,
		maxsize: maxSize,
	}
	return q
}

// Push - pushes element into buffer
func (fq *FIFOQueue) Push(item interface{}) {
	fq.lock.Lock()
	defer fq.lock.Unlock()
	if item == nil {
		log.Panicln("Can not add nil item to fifo queue")
	}
	if fq.tail.last >= fq.maxsize {
		// fq.tail.next = new(fifo)
		fq.tail.next = newfifo(fq.maxsize)
		fq.tail = fq.tail.next
	}
	fq.tail.items[fq.tail.last] = item
	fq.tail.last++
	fq.count++
}

// Pop - ejects element from buffer
func (fq *FIFOQueue) Pop() (item interface{}) {
	fq.lock.Lock()
	defer fq.lock.Unlock()
	if fq.count == 0 {
		return nil
	}
	if fq.head.first >= fq.head.last {
		return nil
	}
	item = fq.head.items[fq.head.first]
	fq.head.first++
	fq.count--
	if fq.head.first >= fq.head.last {
		if fq.count == 0 {
			fq.head.first = 0
			fq.head.last = 0
			fq.head.next = nil
		} else {
			fq.head = fq.head.next
		}
	}
	return item
}

// Len - returns buffer's length
func (fq *FIFOQueue) Len() (length int) {
	fq.lock.Lock()
	defer fq.lock.Unlock()
	length = fq.count
	return length
}
