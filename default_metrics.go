package metrics

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type simpleCounter struct {
	name        string
	description string
	labels      string
	value       *int64
}

func NewCounter(name string, options ...option) Counter {
	config := configuration{Name: name}
	Options.apply(options...)(&config)
	var value int64
	return simpleCounter{
		name:        config.Name,
		description: config.Description,
		labels:      config.RenderLabels(),
		value:       &value,
	}
}
func (this simpleCounter) Type() string            { return "counter" }
func (this simpleCounter) Name() string            { return this.name }
func (this simpleCounter) Description() string     { return this.description }
func (this simpleCounter) Labels() string          { return this.labels }
func (this simpleCounter) Value() int64            { return atomic.LoadInt64(this.value) }
func (this simpleCounter) Increment()              { atomic.AddInt64(this.value, 1) }
func (this simpleCounter) IncrementN(value uint64) { atomic.AddInt64(this.value, int64(value)) }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type simpleGauge struct {
	name        string
	description string
	labels      string
	value       *int64
}

func NewGauge(name string, options ...option) Gauge {
	config := configuration{Name: name}
	Options.apply(options...)(&config)
	var value int64
	return simpleGauge{
		name:        config.Name,
		description: config.Description,
		labels:      config.RenderLabels(),
		value:       &value,
	}
}

func (this simpleGauge) Type() string           { return "gauge" }
func (this simpleGauge) Name() string           { return this.name }
func (this simpleGauge) Description() string    { return this.description }
func (this simpleGauge) Labels() string         { return this.labels }
func (this simpleGauge) Value() int64           { return atomic.LoadInt64(this.value) }
func (this simpleGauge) Increment()             { atomic.AddInt64(this.value, 1) }
func (this simpleGauge) IncrementN(value int64) { atomic.AddInt64(this.value, value) }
func (this simpleGauge) Measure(value int64)    { atomic.StoreInt64(this.value, value) }

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type simpleHistogram struct {
	name         string
	description  string
	labels       string
	bucketKeys   []uint64
	bucketValues []uint64
	sum          *uint64
	count        *uint64
}

func NewHistogram(name string, options ...option) Histogram {
	config := configuration{Name: name}
	Options.apply(options...)(&config)
	var sum uint64
	var count uint64
	return simpleHistogram{
		name:         config.Name,
		description:  config.Description,
		labels:       config.RenderLabels(),
		bucketKeys:   config.BucketKeys,
		bucketValues: make([]uint64, len(config.BucketKeys)),
		sum:          &sum,
		count:        &count,
	}
}
func (this simpleHistogram) Type() string        { return "histogram" }
func (this simpleHistogram) Name() string        { return this.name }
func (this simpleHistogram) Description() string { return this.description }
func (this simpleHistogram) Labels() string      { return this.labels }
func (this simpleHistogram) Buckets() []uint64   { return this.bucketKeys }
func (this simpleHistogram) Values() []uint64    { return this.bucketValues }
func (this simpleHistogram) Count() uint64       { return atomic.LoadUint64(this.count) }
func (this simpleHistogram) Sum() uint64         { return atomic.LoadUint64(this.sum) }

func (this simpleHistogram) Value() int64 { return 0 }
func (this simpleHistogram) Increment()   {}

func (this simpleHistogram) Measure(value uint64) {
	for index, bucketKey := range this.bucketKeys {
		if value <= bucketKey {
			atomic.AddUint64(&this.bucketValues[index], 1)
		}
	}

	atomic.AddUint64(this.sum, value)
	atomic.AddUint64(this.count, 1)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var Options singleton

type singleton struct{}
type option func(*configuration)
type configuration struct {
	Name        string
	Description string
	Labels      map[string]string
	BucketKeys  []uint64
}

func (singleton) Description(value string) option {
	return func(this *configuration) { this.Description = value }
}
func (singleton) Label(key, value string) option {
	return func(this *configuration) { this.Labels[key] = value }
}
func (singleton) Bucket(value uint64) option {
	return func(this *configuration) {
		this.BucketKeys = append(this.BucketKeys, value)
	}
}
func (singleton) apply(options ...option) option {
	return func(this *configuration) {
		this.Labels = map[string]string{}
		for _, option := range Options.defaults(options...) {
			option(this)
		}
	}
}
func (singleton) defaults(options ...option) []option {
	return append([]option{}, options...)
}

func (this configuration) RenderLabels() (result string) {
	if len(this.Labels) == 0 {
		return ""
	}

	for key, value := range this.Labels {
		result += fmt.Sprintf(`%s="%s", `, key, value)
	}
	result = strings.TrimSuffix(result, ", ")
	return fmt.Sprintf("{ %s }", result)
}
