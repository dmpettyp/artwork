package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmpettyp/artwork/application"
	httpgateway "github.com/dmpettyp/artwork/gateways/http"
	"github.com/dmpettyp/artwork/infrastructure/inmem"
	"github.com/dmpettyp/dorky"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("this is artwork")

	uow, err := inmem.NewUnitOfWork()

	if err != nil {
		logger.Error("could not create image graph unit of work", "error", err)
		return
	}

	messageBus := dorky.NewMessageBus(logger)

	_, err = application.NewImageGraphCommandHandlers(messageBus, uow)

	if err != nil {
		logger.Error("could not create image graph command handlers", "error", err)
		return
	}

	_, err = application.NewImageGraphEventHandlers(messageBus, uow)

	if err != nil {
		logger.Error("could not create image graph event handlers", "error", err)
		return
	}

	httpServer := httpgateway.NewHTTPServer(messageBus, logger)
	httpServer.Start()

	go messageBus.Start(context.Background())

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-sigChan

	logger.Info("shutting down gracefully...")

	// Stop the message bus
	messageBus.Stop()

	// Stop the HTTP server with timeout context
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Stop(shutdownCtx); err != nil {
		logger.Error("error stopping HTTP server", "error", err)
	}

	logger.Info("shutdown complete")
}
