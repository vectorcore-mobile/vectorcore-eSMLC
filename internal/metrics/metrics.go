// Package metrics is a minimal, dependency-free counter/gauge registry that
// renders the Prometheus text exposition format. It exists so this service
// can expose scrapable metrics without pulling in a metrics client library:
// the surface needed (a handful of counters and gauges, no histograms, no
// arbitrary label cardinality) is small enough to hand-roll and verify
// directly, matching how this codebase already hand-rolls its bit-level
// codecs instead of importing an ASN.1 runtime.
package metrics

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// Counter is a monotonically increasing count. The zero value is usable.
type Counter struct{ v atomic.Uint64 }

func (c *Counter) Inc()             { c.v.Add(1) }
func (c *Counter) Add(delta uint64) { c.v.Add(delta) }
func (c *Counter) Value() uint64    { return c.v.Load() }

// Gauge is a value that can go up or down. The zero value is usable.
type Gauge struct{ v atomic.Int64 }

func (g *Gauge) Set(v int64)  { g.v.Store(v) }
func (g *Gauge) Inc()         { g.v.Add(1) }
func (g *Gauge) Dec()         { g.v.Add(-1) }
func (g *Gauge) Value() int64 { return g.v.Load() }

// CounterVec is a counter split by exactly one label, restricted to a fixed
// set of label values declared at construction time. Unlike a general
// label-value counter, this cannot accumulate unbounded series from
// attacker- or request-controlled strings: WithLabelValue panics on any
// value not in the declared set, the same fail-closed discipline this
// codebase applies to unimplemented protocol fields elsewhere.
type CounterVec struct {
	name, help, label string
	values            map[string]*Counter
	order             []string
}

func newCounterVec(name, help, label string, allowed []string) *CounterVec {
	values := make(map[string]*Counter, len(allowed))
	for _, v := range allowed {
		values[v] = &Counter{}
	}
	return &CounterVec{name: name, help: help, label: label, values: values, order: append([]string(nil), allowed...)}
}

// WithLabelValue returns the counter for value. It panics if value was not
// declared when the CounterVec was registered: this is a programming error
// (a typo'd or newly-added outcome kind that forgot to update the registry),
// not a runtime condition to handle gracefully.
func (c *CounterVec) WithLabelValue(value string) *Counter {
	counter, ok := c.values[value]
	if !ok {
		panic(fmt.Sprintf("metrics: undeclared label value %q for %s", value, c.name))
	}
	return counter
}

type namedGauge struct {
	name, help string
	g          *Gauge
}
type namedCounter struct {
	name, help string
	c          *Counter
}
type namedGaugeFunc struct {
	name, help string
	fn         func() int64
}
type namedCounterFunc struct {
	name, help string
	fn         func() uint64
}

// Registry collects named metrics and renders them in Prometheus text
// exposition format. A Registry is safe for concurrent registration and
// concurrent WriteTo/ServeHTTP.
type Registry struct {
	mu           sync.Mutex
	counters     []namedCounter
	gauges       []namedGauge
	gaugeFuncs   []namedGaugeFunc
	counterFuncs []namedCounterFunc
	counterVecs  []*CounterVec
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) NewCounter(name, help string) *Counter {
	c := &Counter{}
	r.mu.Lock()
	r.counters = append(r.counters, namedCounter{name, help, c})
	r.mu.Unlock()
	return c
}

func (r *Registry) NewGauge(name, help string) *Gauge {
	g := &Gauge{}
	r.mu.Lock()
	r.gauges = append(r.gauges, namedGauge{name, help, g})
	r.mu.Unlock()
	return g
}

// NewGaugeFunc registers a gauge whose value is computed by calling fn at
// scrape time, rather than pushed. Use this for values with an existing
// live accessor (e.g. session.Manager.Count()) so there is no separate copy
// that can drift out of sync with the source of truth.
func (r *Registry) NewGaugeFunc(name, help string, fn func() int64) {
	r.mu.Lock()
	r.gaugeFuncs = append(r.gaugeFuncs, namedGaugeFunc{name, help, fn})
	r.mu.Unlock()
}

// NewCounterFunc registers a monotonic counter whose value is computed by
// calling fn at scrape time, rather than pushed. Use this for values with an
// existing live monotonic accessor (e.g. CatalogStatus.ReloadSuccesses) so
// there is no separate copy that can drift out of sync with the source of
// truth; unlike NewGaugeFunc, it renders as Prometheus TYPE counter.
func (r *Registry) NewCounterFunc(name, help string, fn func() uint64) {
	r.mu.Lock()
	r.counterFuncs = append(r.counterFuncs, namedCounterFunc{name, help, fn})
	r.mu.Unlock()
}

// NewCounterVec registers a counter split by label, restricted to allowed
// values (see CounterVec).
func (r *Registry) NewCounterVec(name, help, label string, allowed []string) *CounterVec {
	cv := newCounterVec(name, help, label, allowed)
	r.mu.Lock()
	r.counterVecs = append(r.counterVecs, cv)
	r.mu.Unlock()
	return cv
}

// WriteTo renders every registered metric in Prometheus text exposition
// format (https://prometheus.io/docs/instrumenting/exposition_formats/).
func (r *Registry) WriteTo(w io.Writer) (int64, error) {
	r.mu.Lock()
	counters := append([]namedCounter(nil), r.counters...)
	gauges := append([]namedGauge(nil), r.gauges...)
	gaugeFuncs := append([]namedGaugeFunc(nil), r.gaugeFuncs...)
	counterFuncs := append([]namedCounterFunc(nil), r.counterFuncs...)
	vecs := append([]*CounterVec(nil), r.counterVecs...)
	r.mu.Unlock()

	var b strings.Builder
	for _, c := range counters {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s counter\n%s %d\n", c.name, c.help, c.name, c.name, c.c.Value())
	}
	for _, g := range gauges {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s gauge\n%s %d\n", g.name, g.help, g.name, g.name, g.g.Value())
	}
	for _, g := range gaugeFuncs {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s gauge\n%s %d\n", g.name, g.help, g.name, g.name, g.fn())
	}
	for _, c := range counterFuncs {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s counter\n%s %d\n", c.name, c.help, c.name, c.name, c.fn())
	}
	for _, v := range vecs {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s counter\n", v.name, v.help, v.name)
		values := append([]string(nil), v.order...)
		sort.Strings(values)
		for _, value := range values {
			fmt.Fprintf(&b, "%s{%s=%q} %d\n", v.name, v.label, value, v.values[value].Value())
		}
	}
	n, err := io.WriteString(w, b.String())
	return int64(n), err
}
