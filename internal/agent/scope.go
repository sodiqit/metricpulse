package agent

import "sync"

type MetricSnapshot struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

type scope struct {
	cm sync.Mutex
	gm sync.Mutex

	counters map[string]*counter
	gauges   map[string]*gauge
}

func NewRootScope() *scope {
	return &scope{
		counters: make(map[string]*counter),
		gauges:   make(map[string]*gauge),
	}
}

func (s *scope) Counter(name string) Counter {
	s.cm.Lock()
	defer s.cm.Unlock()
	val, ok := s.counters[name]

	if !ok {
		val = newCounter()
		s.counters[name] = val
	}

	return val
}

func (s *scope) Gauge(name string) Gauge {
	s.gm.Lock()
	defer s.gm.Unlock()
	val, ok := s.gauges[name]

	if !ok {
		val = newGauge()
		s.gauges[name] = val
	}

	return val
}

func (s *scope) Snapshot() MetricSnapshot {
	s.cm.Lock()
	countersSnapshot := make(map[string]Counter, len(s.counters))
	for k, v := range s.counters {
		countersSnapshot[k] = v
	}
	s.cm.Unlock()

	s.gm.Lock()
	gaugesSnapshot := make(map[string]Gauge, len(s.gauges))
	for k, v := range s.gauges {
		gaugesSnapshot[k] = v
	}
	s.gm.Unlock()

	return MetricSnapshot{
		Counters: countersSnapshot,
		Gauges:   gaugesSnapshot,
	}
}
