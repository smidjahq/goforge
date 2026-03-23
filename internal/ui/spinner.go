// Package ui provides terminal UI utilities for goforge.
package ui

import (
	"fmt"
	"io"
	"sync"
	"time"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner renders an animated braille spinner with a status message in-place.
type Spinner struct {
	w       io.Writer
	mu      sync.Mutex
	msg     string
	idx     int
	done    chan struct{}
	stopped bool
}

// New creates a new Spinner that writes to w.
func New(w io.Writer) *Spinner {
	return &Spinner{w: w, done: make(chan struct{})}
}

// Start begins the animation with the given initial message.
func (s *Spinner) Start(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				frame := frames[s.idx%len(frames)]
				m := s.msg
				s.idx++
				s.mu.Unlock()
				fmt.Fprintf(s.w, "\r  %s  %s   ", frame, m)
			case <-s.done:
				return
			}
		}
	}()
}

// Update swaps the displayed message while the spinner keeps running.
func (s *Spinner) Update(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

// Stop halts the spinner and prints a green ✓ with the given final message.
func (s *Spinner) Stop(final string) {
	s.halt()
	fmt.Fprintf(s.w, "\r  \033[32m✓\033[0m  %s\n", final)
}

// Fail halts the spinner and prints a red ✗ with the given error message.
func (s *Spinner) Fail(msg string) {
	s.halt()
	fmt.Fprintf(s.w, "\r  \033[31m✗\033[0m  %s\n", msg)
}

func (s *Spinner) halt() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.stopped {
		close(s.done)
		s.stopped = true
	}
	// Clear the spinner line before printing the final status.
	fmt.Fprintf(s.w, "\r%s\r", "                                                  ")
}
