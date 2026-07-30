package session

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrDuplicate = errors.New("session: duplicate")
	ErrCapacity  = errors.New("session: capacity")
)

type ID struct {
	Association string
	Correlation [4]byte
}
type Session struct {
	ID      ID
	Created time.Time
	Expires time.Time
	done    chan struct{}
	once    sync.Once
}
type Manager struct {
	mu       sync.Mutex
	max      int
	sessions map[ID]*Session
}

func New(max int) *Manager { return &Manager{max: max, sessions: map[ID]*Session{}} }
func (m *Manager) Create(id ID, timeout time.Duration) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sessions) >= m.max {
		return nil, ErrCapacity
	}
	if _, ok := m.sessions[id]; ok {
		return nil, ErrDuplicate
	}
	now := time.Now()
	s := &Session{ID: id, Created: now, Expires: now.Add(timeout), done: make(chan struct{})}
	m.sessions[id] = s
	return s, nil
}
func (m *Manager) Finish(s *Session) bool {
	won := false
	s.once.Do(func() {
		won = true
		close(s.done)
		m.mu.Lock()
		if m.sessions[s.ID] == s {
			delete(m.sessions, s.ID)
		}
		m.mu.Unlock()
	})
	return won
}
func (m *Manager) Wait(ctx context.Context, s *Session) error {
	t := time.NewTimer(time.Until(s.Expires))
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		return nil
	case <-t.C:
		return context.DeadlineExceeded
	}
}
func (m *Manager) DropAssociation(a string) {
	m.mu.Lock()
	all := make([]*Session, 0)
	for id, s := range m.sessions {
		if id.Association == a {
			all = append(all, s)
		}
	}
	m.mu.Unlock()
	for _, s := range all {
		m.Finish(s)
	}
}
func (m *Manager) Count() int { m.mu.Lock(); defer m.mu.Unlock(); return len(m.sessions) }
