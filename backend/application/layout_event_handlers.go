package application

import (
	"context"
	"fmt"

	"github.com/dmpettyp/artwork/domain/ui"
	"github.com/dmpettyp/dorky"
)

type LayoutEventHandlers struct {
	notifier ImageGraphNotifier
}

// NewLayoutEventHandlers initializes the handlers struct that processes
// all Layout Events and registers all handlers with the provided
// message bus
func NewLayoutEventHandlers(
	mb *dorky.MessageBus,
	notifier ImageGraphNotifier,
) (
	*LayoutEventHandlers,
	error,
) {
	handlers := &LayoutEventHandlers{
		notifier: notifier,
	}

	err := dorky.RegisterEventHandler(mb, handlers.HandleLayoutUpdatedEvent)

	if err != nil {
		return nil, fmt.Errorf("could not create layout event handlers: %w", err)
	}

	return handlers, nil
}

func (h *LayoutEventHandlers) HandleLayoutUpdatedEvent(
	ctx context.Context,
	event *ui.LayoutUpdatedEvent,
) (
	[]dorky.Event,
	error,
) {
	// Broadcast that layout was updated
	h.notifier.BroadcastLayoutUpdate(event.GraphID)

	return nil, nil
}
