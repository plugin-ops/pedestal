package util

import "github.com/gogf/gf/v2/container/gqueue"

type StringQueue struct {
	*gqueue.Queue
}

func NewStringQueue() *StringQueue {
	return &StringQueue{
		Queue: gqueue.New(),
	}
}

func (q *StringQueue) Push(s string) {
	q.Queue.Push(s)
}

func (q *StringQueue) Pull() string {
	return q.Queue.Pop().(string)
}

func (q *StringQueue) Len() int {
	return q.Queue.Len()
}

func (q *StringQueue) Close() {
	q.Queue.Close()
}
