package service

import (
	"testing"
)

func TestLimiter(t *testing.T) {
	tests := []struct {
		name      string
		max       int
		calls     int
		wantAllow int
		wantRem   int
	}{
		{
			name:      "allows up to max calls",
			max:       5,
			calls:     3,
			wantAllow: 3,
			wantRem:   2,
		},
		{
			name:      "blocks after max calls",
			max:       2,
			calls:     5,
			wantAllow: 2,
			wantRem:   0,
		},
		{
			name:      "zero max blocks everything",
			max:       0,
			calls:     3,
			wantAllow: 0,
			wantRem:   0,
		},
		{
			name:      "single call",
			max:       1,
			calls:     1,
			wantAllow: 1,
			wantRem:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(tt.max)
			allowed := 0
			for i := 0; i < tt.calls; i++ {
				if l.Allow() {
					allowed++
				}
			}
			if allowed != tt.wantAllow {
				t.Errorf("Allow() = %d allowed; want %d", allowed, tt.wantAllow)
			}
			if got := l.Remaining(); got != tt.wantRem {
				t.Errorf("Remaining() = %d; want %d", got, tt.wantRem)
			}
		})
	}
}
