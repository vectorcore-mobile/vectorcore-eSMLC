package session

import (
	"testing"
	"time"
)

func TestFinishExactlyOnceAndAssociationCleanup(t *testing.T) {
	m := New(2)
	id := ID{Association: "a", Correlation: [4]byte{1}}
	s, e := m.Create(id, time.Second)
	if e != nil {
		t.Fatal(e)
	}
	if !m.Finish(s) || m.Finish(s) || m.Count() != 0 {
		t.Fatal("finish was not exactly once")
	}
	s, e = m.Create(id, time.Second)
	if e != nil {
		t.Fatal(e)
	}
	m.DropAssociation("a")
	if m.Count() != 0 {
		t.Fatal("association sessions leaked")
	}
}
