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
	"github.com/dmpettyp/artwork/infrastructure/filestorage"
	"github.com/dmpettyp/artwork/infrastructure/imagegen"
	"github.com/dmpettyp/artwork/infrastructure/inmem"
	"github.com/dmpettyp/dorky"
)

func main() {
	// Set log level based on LOG_LEVEL environment variable (default: INFO)
	logLevel := slog.LevelInfo
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		if err := logLevel.UnmarshalText([]byte(levelStr)); err != nil {
			// Invalid level, stick with default
			logLevel = slog.LevelInfo
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	logger.Info("this is artwork")

	uow, err := inmem.NewUnitOfWork()

	if err != nil {
		logger.Error("could not create image graph unit of work", "error", err)
		return
	}

	messageBus := dorky.NewMessageBus(logger)

	// Create image storage
	imageStorage, err := filestorage.NewFilesystemImageStorage("uploads")

	if err != nil {
		logger.Error("could not create image storage", "error", err)
		return
	}

	// Create output setter for ImageGen
	outputSetter := application.NewNodeOutputSetter(messageBus)

	// Create ImageGen with dependencies
	imageGen := imagegen.NewImageGen(imageStorage, outputSetter)

	_, err = application.NewImageGraphCommandHandlers(messageBus, uow)

	if err != nil {
		logger.Error("could not create image graph command handlers", "error", err)
		return
	}

	// Create notifier for real-time graph updates
	notifier := httpgateway.NewImageGraphNotifier(logger)

	_, err = application.NewImageGraphEventHandlers(messageBus, uow, imageGen, notifier)

	if err != nil {
		logger.Error("could not create image graph event handlers", "error", err)
		return
	}

	_, err = application.NewLayoutCommandHandlers(messageBus, uow)

	if err != nil {
		logger.Error("could not create layout command handlers", "error", err)
		return
	}

	_, err = application.NewViewportCommandHandlers(messageBus, uow)

	if err != nil {
		logger.Error("could not create viewport command handlers", "error", err)
		return
	}

	httpServer := httpgateway.NewHTTPServer(
		logger,
		messageBus,
		uow.ImageGraphViews,
		uow.LayoutViews,
		uow.ViewportViews,
		imageStorage,
		notifier,
	)

	httpServer.Start()

	go messageBus.Start(context.Background())

	// Bootstrap the application with default ImageGraph
	if err := bootstrap(context.Background(), logger, messageBus); err != nil {
		logger.Error("bootstrap failed", "error", err)
	}

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
