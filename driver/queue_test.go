package driver_test

import (
	"github.com/autom8ter/dagger/driver"
	"testing"
)

func TestQueue(t *testing.T) {
	q := driver.NewQueue()
	q.Enqueue("hello")
	q.Enqueue("world")
	q.Enqueue("tester")
	val, ok := q.Dequeue()
	if !ok || val != "hello" {
		t.Fatal("fail", val)
	}
	val, ok = q.Dequeue()
	if !ok || val != "world" {
		t.Fatal("fail", val)
	}
	val, ok = q.Dequeue()
	if !ok || val != "tester" {
		t.Fatal("fail", val)
	}
}
