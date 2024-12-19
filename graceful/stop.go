package graceful

import (
	"context"
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
// This function will block until server shutdown is complete.
// If the server does not shut down within the timeout, it will return an error. The
// error returned from server.Shutdown() will be returned here.
func WaitForShutdown(s *http.Server, timeout time.Duration, cf ...context.CancelFunc) error {
	waitForKill()
	wchan, err := shutdownServer(s, timeout, cf...)
	<-wchan
	return err
}

func waitForKill() {
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-kill
}

// Shutdown the server and then propagate the shutdown to a cancel function.
// Caller should block on the returned channel.
func shutdownServer(s *http.Server, timeout time.Duration, cf ...context.CancelFunc) (chan bool, error) {
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
	err := s.Shutdown(ctx)
	return wchan, err
}
