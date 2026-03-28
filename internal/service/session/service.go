package session

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	// CookieName is the session cookie name expected by autograder.
	CookieName = "X-Session-Id"
	minSIDLen  = 32
)

// Store defines persistence operations used by session service.
type Store interface {
	Exists(ctx context.Context, sid string) (bool, error)
	CreateIfNotExists(ctx context.Context, sid string, createdAt time.Time, updatedAt time.Time, ttl time.Duration) (bool, error)
	Refresh(ctx context.Context, sid string, updatedAt time.Time, ttl time.Duration) error
}

// IDGenerator defines SID generation behavior.
type IDGenerator interface {
	Generate() (string, error)
}

// Clock provides current time for service logic.
type Clock interface {
	Now() time.Time
}

// SystemClock returns real current time.
type SystemClock struct{}

// Now returns current local time.
func (SystemClock) Now() time.Time {
	return time.Now()
}

// Service contains session use-cases.
type Service struct {
	store      Store
	generator  IDGenerator
	clock      Clock
	ttl        time.Duration
	maxAttempt int
}

// Result represents the outcome of POST /session.
type Result struct {
	SID     string
	Created bool
}

// NewService builds session service.
func NewService(store Store, generator IDGenerator, clock Clock, ttl time.Duration) *Service {
	return &Service{
		store:      store,
		generator:  generator,
		clock:      clock,
		ttl:        ttl,
		maxAttempt: 5,
	}
}

// UpsertSession creates a session when absent or refreshes an existing one.
func (s *Service) UpsertSession(ctx context.Context, sid string) (Result, error) {
	if IsValidSID(sid) {
		exists, err := s.store.Exists(ctx, sid)
		if err != nil {
			return Result{}, err
		}
		if exists {
			now := s.clock.Now().UTC()
			if err := s.store.Refresh(ctx, sid, now, s.ttl); err != nil {
				return Result{}, err
			}
			return Result{SID: sid, Created: false}, nil
		}
	}

	for attempt := 0; attempt < s.maxAttempt; attempt++ {
		newSID, err := s.generator.Generate()
		if err != nil {
			return Result{}, err
		}
		if !IsValidSID(newSID) {
			continue
		}
		now := s.clock.Now().UTC()
		created, err := s.store.CreateIfNotExists(ctx, newSID, now, now, s.ttl)
		if err != nil {
			return Result{}, err
		}
		if created {
			return Result{SID: newSID, Created: true}, nil
		}
	}

	return Result{}, errors.New("failed to create unique session id")
}

// IsValidSID validates a session identifier as hex string with minimum 128 bits.
func IsValidSID(sid string) bool {
	if len(sid) < minSIDLen {
		return false
	}
	for _, chr := range sid {
		if !strings.ContainsRune("0123456789abcdefABCDEF", chr) {
			return false
		}
	}
	return true
}
