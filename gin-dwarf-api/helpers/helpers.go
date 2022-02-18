package helpers

import (
	"math"
)

type Waiter struct {
	count int64
}

func NewWaiter() *Waiter {
	w := &Waiter{
		count: int64(-10),
	}
	return w
}

func (w *Waiter) GetCount() int64 {
	return w.count
}

func (w *Waiter) GetDelay() float64 {
	w.count++
	return math.Exp(float64(w.count) / 25.0)
}
