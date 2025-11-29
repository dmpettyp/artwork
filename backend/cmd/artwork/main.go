package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmpettyp/dorky/messagebus"

	"github.com/dmpettyp/artwork/application"
	httpgateway "github.com/dmpettyp/artwork/gateways/http"
	"github.com/dmpettyp/artwork/infrastructure/filestorage"
	"github.com/dmpettyp/artwork/infrastructure/imagegen"
	"github.com/dmpettyp/artwork/infrastructure/inmem"
	"github.com/dmpettyp/artwork/infrastructure/postgres"
)

func main() {
	storeBackend := flag.String("store", "postgres", "storage backend: postgres or inmem")
	bootstrapFlag := flag.Bool("bootstrap", false, "seed a default graph on startup")
	flag.Parse()

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

	var (
		uow             application.UnitOfWork
		imageGraphViews application.ImageGraphViews
		layoutViews     application.LayoutViews
		viewportViews   application.ViewportViews
	)

	switch *storeBackend {
	case "postgres":
		db, err := postgres.NewDB(postgres.DefaultConfig())
		if err != nil {
			logger.Error("could not create postgres db connection", "error", err)
			return
		}
		uow = postgres.NewUnitOfWork(db)
		imageGraphViews = postgres.NewImageGraphViews(db)
		layoutViews = postgres.NewLayoutViews(db)
		viewportViews = postgres.NewViewportViews(db)
		logger.Info("using postgres backend")
	case "inmem":
		inmemUOW, err := inmem.NewUnitOfWork()
		if err != nil {
			logger.Error("could not create in-memory unit of work", "error", err)
			return
		}
		uow = inmemUOW
		imageGraphViews = inmemUOW.ImageGraphViews
		layoutViews = inmemUOW.LayoutViews
		viewportViews = inmemUOW.ViewportViews
		logger.Info("using in-memory backend")
	default:
		logger.Error("invalid store backend", "value", *storeBackend)
		return
	}

	messageBus := messagebus.New(logger)

	// Create image storage
	imageStorage, err := filestorage.NewFilesystemImageStorage("uploads")

	if err != nil {
		logger.Error("could not create image storage", "error", err)
		return
	}

	// Create node updater for ImageGen
	nodeUpdater := application.NewNodeUpdater(messageBus)

	// Create ImageGen with dependencies
	imageGen := imagegen.NewImageGen(imageStorage, nodeUpdater, logger)

	_, err = application.NewImageGraphCommandHandlers(messageBus, uow)

	if err != nil {
		logger.Error("could not create image graph command handlers", "error", err)
		return
	}

	// Create notifier for real-time graph updates
	notifier := httpgateway.NewImageGraphNotifier(logger)

	_, err = application.NewImageGraphEventHandlers(
		messageBus,
		uow,
		imageGen,
		imageStorage,
		notifier,
	)

	if err != nil {
		logger.Error("could not create image graph event handlers", "error", err)
		return
	}

	_, err = application.NewLayoutCommandHandlers(messageBus, uow)

	if err != nil {
		logger.Error("could not create layout command handlers", "error", err)
		return
	}

	_, err = application.NewLayoutEventHandlers(messageBus, notifier)

	if err != nil {
		logger.Error("could not create layout event handlers", "error", err)
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
		imageGraphViews,
		layoutViews,
		viewportViews,
		imageStorage,
		notifier,
	)

	httpServer.Start()

	go messageBus.Start(context.Background())

	// Bootstrap the application with default ImageGraph if requested
	if *bootstrapFlag {
		if err := bootstrap(context.Background(), logger, messageBus); err != nil {
			logger.Error("bootstrap failed", "error", err)
			return
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down gracefully...")

	messageBus.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Stop(shutdownCtx); err != nil {
		logger.Error("error stopping HTTP server", "error", err)
	}

	logger.Info("shutdown complete")
}
