package scanner

import (
	"fmt"
	"sync"
	"time"
)

// Spinner provides visual feedback
type Spinner struct {
	mu      sync.Mutex
	active  bool
	delay   time.Duration
	message string
}

// NewSpinner creates a new spinner
func NewSpinner(message string, delay time.Duration) *Spinner {
	return &Spinner{
		active:  false,
		delay:   delay,
		message: message,
	}
}

// Start begins the spinner
func (s *Spinner) Start() {
	s.mu.Lock()
	s.active = true
	s.mu.Unlock()

	go func() {
		for {
			s.mu.Lock()
			if !s.active {
				s.mu.Unlock()
				return
			}
			s.mu.Unlock()

			fmt.Printf("\r%s", s.message)
			time.Sleep(s.delay)
		}
	}()
}

// Stop halts the spinner
func (s *Spinner) Stop() {
	s.mu.Lock()
	s.active = false
	s.mu.Unlock()
	fmt.Print("\r")
}