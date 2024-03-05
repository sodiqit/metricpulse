package agent

import (
	"math"
	"sync/atomic"
)

type Counter interface {
	Inc(delta int64)
	Value() int64
}

type Gauge interface {
	Update(value float64)
	Value() float64
}

type counter struct {
	value int64
}

func newCounter() *counter {
	return &counter{}
}

func (c *counter) Inc(delta int64) {
	atomic.AddInt64(&c.value, delta)
}

func (c *counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

type gauge struct {
	floatBits uint64
}

func newGauge() *gauge {
	return &gauge{}
}

func (g *gauge) Update(value float64) {
	atomic.StoreUint64(&g.floatBits, math.Float64bits(value))
}

func (g *gauge) Value() float64 {
	return math.Float64frombits(g.floatBits)
}
