package storage

type MemStorage struct {
	gauge   map[string]float64
	counter map[string][]int64
}

func (m *MemStorage) SaveGaugeMetric(metricType string, value float64) {
	m.gauge[metricType] = value
}

func (m *MemStorage) SaveCounterMetric(metricType string, value int64) {
	val, ok := m.counter[metricType]

	if (ok) {
		m.counter[metricType] = append(val, value)
	} else {
		m.counter[metricType] = []int64{value}
	}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge: make(map[string]float64),
		counter: make(map[string][]int64),
	}
}