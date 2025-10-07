package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
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
	defer httpServer.Stop(context.Background())

	go messageBus.Start(context.Background())

	defer messageBus.Stop()

	// id := imagegraph.MustNewImageGraphID()
	// command := application.NewCreateImageGraphCommand(id, "super awesome new image")
	//
	// messageBus.HandleCommand(context.TODO(), command)
}
