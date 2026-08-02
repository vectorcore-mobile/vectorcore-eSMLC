package sls

import (
	"testing"

	"github.com/vectorcore/esmlc/internal/config"
)

func TestReadyReflectsListenerAndSLsEnablement(t *testing.T) {
	c := config.Default()
	c.SLs.Enabled = false
	s := New(c, nil)
	if !s.Live() {
		t.Fatal("expected live before any shutdown")
	}
	if !s.Ready() {
		t.Fatal("expected ready when SLs is disabled: nothing to wait for")
	}

	c = config.Default() // SLs.Enabled defaults to true
	s = New(c, nil)
	if !s.Live() {
		t.Fatal("expected live before any shutdown")
	}
	if s.Ready() {
		t.Fatal("expected not ready before Listen() binds a listener")
	}
}

func TestLiveAndReadyFalseAfterClose(t *testing.T) {
	c := config.Default()
	c.SLs.Enabled = false
	s := New(c, nil)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if s.Live() {
		t.Fatal("expected not live after Close")
	}
	if s.Ready() {
		t.Fatal("expected not ready after Close")
	}
}
