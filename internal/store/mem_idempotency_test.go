package store

import (
	"testing"
	"time"
)

func TestMemIdempotency_Set(t *testing.T) {
	tests := []struct {
		name string
		ttl  time.Duration
		key  string
	}{
		{
			name: "success - set key with 5 minute TTL",
			ttl:  5 * time.Minute,
			key:  "user:123-user.created-req-123-corr-123",
		},
		{
			name: "success - set key with 1 hour TTL",
			ttl:  1 * time.Hour,
			key:  "order:456-order.completed-req-456-corr-456",
		},
		{
			name: "success - set empty key",
			ttl:  5 * time.Minute,
			key:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := NewMemIdempotency(tt.ttl)

			mi.Set(tt.key)

			ttl, ok := mi.Get(tt.key)
			if !ok {
				t.Errorf("expected key %q to exist", tt.key)
				return
			}

			if ttl.IsZero() {
				t.Errorf("expected TTL to be set, got zero time")
			}
		})
	}
}

func TestMemIdempotency_Get(t *testing.T) {
	tests := []struct {
		name          string
		ttl           time.Duration
		setupKey      string
		lookupKey     string
		expectedExist bool
	}{
		{
			name:          "success - get existing key",
			ttl:           5 * time.Minute,
			setupKey:      "user:123-user.created-req-123-corr-123",
			lookupKey:     "user:123-user.created-req-123-corr-123",
			expectedExist: true,
		},
		{
			name:          "success - get non-existing key",
			ttl:           5 * time.Minute,
			setupKey:      "user:123-user.created-req-123-corr-123",
			lookupKey:     "user:456-user.created-req-456-corr-456",
			expectedExist: false,
		},
		{
			name:          "success - get from empty store",
			ttl:           5 * time.Minute,
			setupKey:      "",
			lookupKey:     "user:123-user.created-req-123-corr-123",
			expectedExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := NewMemIdempotency(tt.ttl)

			if tt.setupKey != "" {
				mi.Set(tt.setupKey)
			}

			_, ok := mi.Get(tt.lookupKey)
			if ok != tt.expectedExist {
				t.Errorf("expected exist=%v, got %v", tt.expectedExist, ok)
			}
		})
	}
}

func TestMemIdempotency_Reset(t *testing.T) {
	tests := []struct {
		name        string
		ttl         time.Duration
		keys        []string
		description string
	}{
		{
			name:        "success - reset with no keys",
			ttl:         5 * time.Minute,
			keys:        []string{},
			description: "reset on empty store should not panic",
		},
		{
			name:        "success - reset with multiple keys",
			ttl:         5 * time.Minute,
			keys:        []string{"key1", "key2", "key3"},
			description: "reset with active keys should not panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := NewMemIdempotency(tt.ttl)

			for _, key := range tt.keys {
				mi.Set(key)
			}

			// Should not panic
			mi.Reset()
		})
	}
}

func TestMemIdempotency_Concurrent(t *testing.T) {
	mi := NewMemIdempotency(5 * time.Minute)

	// Test concurrent access
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			mi.Set("concurrent-key")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			mi.Get("concurrent-key")
		}
		done <- true
	}()

	// Reset goroutine
	go func() {
		for i := 0; i < 100; i++ {
			mi.Reset()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}
