package metrics

import (
	"strings"
	"testing"
)

func TestRegistryRendersPrometheusTextFormat(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("esmlc_test_total", "a test counter")
	c.Inc()
	c.Inc()
	g := r.NewGauge("esmlc_test_gauge", "a test gauge")
	g.Set(5)
	vec := r.NewCounterVec("esmlc_test_outcomes_total", "a test outcome counter", "outcome", []string{"a", "b"})
	vec.WithLabelValue("a").Inc()
	vec.WithLabelValue("b").Add(3)
	live := 7
	r.NewGaugeFunc("esmlc_test_live_gauge", "a live-computed gauge", func() int64 { return int64(live) })

	var buf strings.Builder
	if _, err := r.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"# HELP esmlc_test_total a test counter",
		"# TYPE esmlc_test_total counter",
		"esmlc_test_total 2",
		"# TYPE esmlc_test_gauge gauge",
		"esmlc_test_gauge 5",
		`esmlc_test_outcomes_total{outcome="a"} 1`,
		`esmlc_test_outcomes_total{outcome="b"} 3`,
		"esmlc_test_live_gauge 7",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}

	live = 42
	buf.Reset()
	if _, err := r.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "esmlc_test_live_gauge 42") {
		t.Fatalf("GaugeFunc did not reflect updated live value:\n%s", buf.String())
	}
}

func TestCounterVecRejectsUndeclaredLabelValue(t *testing.T) {
	r := NewRegistry()
	vec := r.NewCounterVec("esmlc_test_total", "help", "outcome", []string{"known"})
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for undeclared label value")
		}
	}()
	vec.WithLabelValue("unknown")
}

func TestCounterAndGaugeConcurrentUse(t *testing.T) {
	c := &Counter{}
	g := &Gauge{}
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				c.Inc()
				g.Inc()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if c.Value() != 1000 {
		t.Fatalf("counter: got %d want 1000", c.Value())
	}
	if g.Value() != 1000 {
		t.Fatalf("gauge: got %d want 1000", g.Value())
	}
}
