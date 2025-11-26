package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/dmpettyp/dorky/messages"

	"github.com/dmpettyp/artwork/application"
)

type repository interface {
	SaveAll() error
	CollectEvents() []messages.Event
}

// UnitOfWork implements application.UnitOfWork using PostgreSQL
type UnitOfWork struct {
	db *sql.DB
}

// NewUnitOfWork creates a new PostgreSQL-based unit of work
func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// Run executes a function within a transaction boundary
func (uow *UnitOfWork) Run(
	ctx context.Context,
	fn func(repos *application.Repos) error,
) (
	[]messages.Event,
	error,
) {
	var events []messages.Event

	err := withTx(ctx, uow.db, func(tx *sql.Tx) error {
		igRepo := newImageGraphRepository(tx)
		layoutRepo := newLayoutRepository(tx)
		vpRepo := newViewportRepository(tx)

		repos := &application.Repos{
			ImageGraphRepository: igRepo,
			LayoutRepository:     layoutRepo,
			ViewportRepository:   vpRepo,
		}

		repositories := []repository{igRepo, layoutRepo, vpRepo}

		if err := fn(repos); err != nil {
			return err
		}

		for _, repo := range repositories {
			if err := repo.SaveAll(); err != nil {
				return err
			}
		}

		for _, repo := range repositories {
			events = append(events, repo.CollectEvents()...)
		}

		if err := saveEvents(ctx, tx, events); err != nil {
			return fmt.Errorf("failed to save events: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return events, nil
}

func saveEvents(ctx context.Context, tx *sql.Tx, events []messages.Event) error {
	if len(events) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO events (aggregate_id, aggregate_type, event_type, event_data, aggregate_version, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare event insert statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		aggregateID := event.GetEntityID()
		aggregateType := event.GetEntityType()
		eventType := event.GetType()
		timestamp := event.GetTimestamp()

		var aggregateVersion *int64
		if versioned, ok := event.(interface{ GetAggregateVersion() int64 }); ok {
			version := versioned.GetAggregateVersion()
			aggregateVersion = &version
		}

		_, err = stmt.ExecContext(ctx,
			aggregateID,
			aggregateType,
			eventType,
			eventData,
			aggregateVersion,
			timestamp,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return nil
}
