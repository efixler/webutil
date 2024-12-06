package graceful

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"
)

func TestWaitForShutdown(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		signal      syscall.Signal
		cancelCount int
	}{
		{
			name:        "SIGINT, no child cancel functions",
			timeout:     500 * time.Millisecond,
			signal:      syscall.SIGINT,
			cancelCount: 0,
		},
		{
			name:        "SIGINT + cancel functions",
			timeout:     500 * time.Millisecond,
			signal:      syscall.SIGINT,
			cancelCount: 2,
		},
		{
			name:        "SIGTERM, no child cancel functions",
			timeout:     500 * time.Millisecond,
			signal:      syscall.SIGTERM,
			cancelCount: 0,
		},
		{
			name:        "SIGTERM + cancel functions",
			timeout:     500 * time.Millisecond,
			signal:      syscall.SIGTERM,
			cancelCount: 2,
		},
	}
	port := getPort()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &http.Server{
				Addr:    fmt.Sprintf("127.0.0.1:%d", port),
				Handler: http.DefaultServeMux,
			}
			var cf []context.CancelFunc
			cancelled := 0
			for i := 0; i < tt.cancelCount; i++ {
				cf = append(cf, func() { cancelled++ })
			}
			var shutdownError error
			var serverDone bool
			var waitDone bool
			go func() {
				if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					shutdownError = err
				}
				serverDone = true
			}()
			go func() {
				WaitForShutdown(s, tt.timeout, cf...)
				waitDone = true
			}()
			time.Sleep(100 * time.Millisecond)
			syscall.Kill(syscall.Getpid(), tt.signal)
			time.Sleep(tt.timeout)
			if !serverDone {
				t.Error("server did not shut down")
			}
			if !waitDone {
				t.Error("wait did not complete")
			}
			if shutdownError != nil {
				t.Error("server shutdown error", shutdownError)
			}
			if cancelled != tt.cancelCount {
				t.Errorf("expected %d cancel functions to be called, got %d", tt.cancelCount, cancelled)
			}
		})
	}
}

func getPort() int {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	listener.Close() // Ensure the listener is closed
	return listener.Addr().(*net.TCPAddr).Port
}
