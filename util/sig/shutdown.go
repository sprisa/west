package sig

import (
	"context"
	"os"
	"os/signal"
	l "reship/util/logger/v2"
	"syscall"
)

// ShutdownListener will listen for and block on term and
// interrupt signals once signalled.
// Recommend to use `ShutdownContext` instead`.
func ShutdownListener() <-chan any {
	// Setup Shutdown listeners
	shutdownChan := make(chan any)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)

	// Fatal panic cleanup
	defer func() {
		if r := recover(); r != nil {
			l.Log.Error().Any("from", r).Msg("Recovered from panic, shutting down.")
			close(shutdownChan)
		}
	}()

	// Close channel on shutdown
	go func() {
		rawSig := <-sigChan
		sig := rawSig.String()
		l.Log.Warn().Str("signal", sig).Msg("Caught signal, shutting down.")
		close(shutdownChan)
	}()

	return shutdownChan
}

// ShutdownContext will listen for
// shutdown signals and cancel the context.
func ShutdownContext(ctx context.Context) context.Context {
	shutdownChan := ShutdownListener()
	ctx, cancel := context.WithCancel(ctx)
	// Cancel context if shutdown
	go func() {
		<-shutdownChan
		cancel()
	}()

	return ctx
}
