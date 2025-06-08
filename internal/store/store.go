package store

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (value float64, exists bool)
	GetCounter(name string) (value int64, exists bool)
}

type memStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewStorage() Storage {
	return &memStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *memStorage) UpdateGauge(name string, value float64) {
	delete(s.counters, name)
	s.gauges[name] = value
}

func (s *memStorage) UpdateCounter(name string, value int64) {
	delete(s.gauges, name)
	if existingValue, exists := s.counters[name]; exists {
		s.counters[name] = existingValue + value
	} else {
		s.counters[name] = value
	}
}

func (s *memStorage) GetGauge(name string) (value float64, exists bool) {
	val, exists := s.gauges[name]
	return val, exists
}

func (s *memStorage) GetCounter(name string) (value int64, exists bool) {
	val, exists := s.counters[name]
	return val, exists
}
