package main

import (
	"codealong/store"
	"fmt"
	"time"
)

// MetricsMiddleware wraps a Storer and records call counts and latencies.
type MetricsMiddleware struct {
	inner store.Storer

	getCalls    int
	setCalls    int
	incrCalls   int
	deleteCalls int
	lenCalls    int

	getMisses int
	totalGetLatency time.Duration
	totalSetLatency time.Duration
}

func NewMetricsMiddleware(inner store.Storer) *MetricsMiddleware {
	return &MetricsMiddleware{inner: inner}
}

func (m *MetricsMiddleware) Get(key string) (string, error) {
	val, err := m.inner.Get(key)
	m.getCalls++

	return val, err
}

func (m *MetricsMiddleware) Set(key, value string) error {
	err := m.inner.Set(key, value)
	m.setCalls++
	return err
}

func (m *MetricsMiddleware) Incr(key string) (int, error) {
	n, err := m.inner.Incr(key)
	m.incrCalls++
	return n, err
}

func (m *MetricsMiddleware) Delete(key string) {
	m.inner.Delete(key)

	m.deleteCalls++
}

func (m *MetricsMiddleware) Len() int {
	n := m.inner.Len()

	m.lenCalls++

	return n
}

func (m *MetricsMiddleware) Keys() []string {
	return m.inner.Keys()
}

// Report prints a summary of all recorded metrics.
func (m *MetricsMiddleware) Report() {
	fmt.Println("=== Store Metrics ===")
	fmt.Printf("GET  calls: %d  (misses: %d)\n", m.getCalls, m.getMisses)
	fmt.Printf("SET  calls: %d\n", m.setCalls)
	fmt.Printf("INCR calls: %d\n", m.incrCalls)
	fmt.Printf("DEL  calls: %d\n", m.deleteCalls)
	fmt.Printf("LEN  calls: %d\n", m.lenCalls)

	if m.getCalls > 0 {
		avg := m.totalGetLatency / time.Duration(m.getCalls)
		fmt.Printf("Avg GET latency : %v\n", avg)
	}
	if m.setCalls > 0 {
		avg := m.totalSetLatency / time.Duration(m.setCalls)
		fmt.Printf("Avg SET latency : %v\n", avg)
	}
}
