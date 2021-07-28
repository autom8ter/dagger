package ds

type Queue interface {
	Enqueue(val interface{})
	Dequeue() (interface{}, bool)
	IsEmpty() bool
	Len() int
	Range(fn func(element interface{}) bool)
}

type queue struct {
	values []interface{}
}

func NewQueue() Queue {
	vals := &queue{[]interface{}{}}
	return vals
}

func (q *queue) IsEmpty() bool {
	return q == nil || q.Len() == 0
}

func (q *queue) Range(fn func(element interface{}) bool) {
	for {
		r, ok := q.Dequeue()
		if !ok {
			return
		}
		if !fn(r) {
			return
		}
	}
}

func (q *queue) Enqueue(val interface{}) {
	q.values = append(q.values, val)
}

func (q *queue) Dequeue() (interface{}, bool) {
	if q.Len() == 0 {
		return nil, false
	}
	val := q.values[0]
	q.values = q.values[1:]
	return val, true
}

func (q *queue) Len() int {
	return len(q.values)
}
