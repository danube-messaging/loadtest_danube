package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithInterrupt returns a context canceled on SIGINT/SIGTERM.
func WithInterrupt(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer signal.Stop(ch)
		select {
		case <-ch:
			cancel()
		case <-parent.Done():
			cancel()
		}
	}()
	return ctx
}
