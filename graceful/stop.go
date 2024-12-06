package graceful

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// After a SIGINT, SIGTERM, or os.Interrupt, shut down s, wait a bit for
// requests to clear, then call the passed cancel function.
// This will let the requests finish before shutting down any other open
// resources with the passed CancelFunc.
// This function will block until server shutdown is complete
func WaitForShutdown(s *http.Server, timeout time.Duration, cf ...context.CancelFunc) {
	waitForKill()
	<-shutdownServer(s, timeout, cf...)
}

func waitForKill() {
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-kill
}

// Shutdown the server and then propagate the shutdown to a cancel function.
// Caller should block on the returned channel.
func shutdownServer(s *http.Server, timeout time.Duration, cf ...context.CancelFunc) chan bool {
	slog.Info("server shutting down")
	wchan := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	context.AfterFunc(ctx, func() {
		for _, cancel := range cf {
			cancel()
		}
		// without a little bit of sleep here sometimes final I/O doesn't get flushed
		time.Sleep(100 * time.Millisecond)
		close(wchan)
	})
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}
	slog.Info("server stopped")
	return wchan
}
