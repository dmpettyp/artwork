package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/dmpettyp/artwork/backend/application"
	"github.com/dmpettyp/dorky"
)

// UnitOfWork implements application.UnitOfWork using PostgreSQL
type UnitOfWork struct {
	db *sql.DB

	// Views for read-model queries
	ImageGraphViews *ImageGraphViews
	LayoutViews     *LayoutViews
	ViewportViews   *ViewportViews
}

// NewUnitOfWork creates a new PostgreSQL-based unit of work
func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{
		db:              db,
		ImageGraphViews: &ImageGraphViews{db: db},
		LayoutViews:     &LayoutViews{db: db},
		ViewportViews:   &ViewportViews{db: db},
	}
}

// Run executes a function within a transaction boundary
func (uow *UnitOfWork) Run(
	ctx context.Context,
	fn func(repos *application.Repos) error,
) (
	[]dorky.Event,
	error,
) {
	var events []dorky.Event

	err := withTx(ctx, uow.db, func(tx *sql.Tx) error {
		// Create repositories with the transaction
		repos := &application.Repos{
			ImageGraphRepository: &ImageGraphRepository{tx: tx},
			LayoutRepository:     &LayoutRepository{tx: tx},
			ViewportRepository:   &ViewportRepository{tx: tx},
		}

		// Execute the provided function
		if err := fn(repos); err != nil {
			return err
		}

		// Save all modified aggregates back to the database
		if err := saveModifiedAggregates(ctx, repos); err != nil {
			return fmt.Errorf("failed to save modified aggregates: %w", err)
		}

		// Collect events from all repositories
		events = collectEvents(repos)

		// Save events to the events table
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

// saveModifiedAggregates persists all modified aggregates back to the database
func saveModifiedAggregates(ctx context.Context, repos *application.Repos) error {
	// Save modified ImageGraphs
	if igRepo, ok := repos.ImageGraphRepository.(*ImageGraphRepository); ok {
		for _, ig := range igRepo.loadedForUpdate {
			if err := igRepo.save(ig); err != nil {
				return fmt.Errorf("failed to save image graph: %w", err)
			}
		}
	}

	// Layout and Viewport use UPSERT in Add(), so they're already saved
	// No need to save them again

	return nil
}

// collectEvents retrieves and clears events from all modified aggregates
func collectEvents(repos *application.Repos) []dorky.Event {
	var allEvents []dorky.Event

	// Collect from ImageGraphRepository
	if igRepo, ok := repos.ImageGraphRepository.(*ImageGraphRepository); ok {
		for _, ig := range igRepo.modified {
			events := ig.GetEvents()
			allEvents = append(allEvents, events...)
			ig.ClearEvents()
		}
	}

	// Collect from LayoutRepository
	if layoutRepo, ok := repos.LayoutRepository.(*LayoutRepository); ok {
		for _, layout := range layoutRepo.modified {
			events := layout.GetEvents()
			allEvents = append(allEvents, events...)
			layout.ClearEvents()
		}
	}

	// Collect from ViewportRepository
	if vpRepo, ok := repos.ViewportRepository.(*ViewportRepository); ok {
		for _, viewport := range vpRepo.modified {
			events := viewport.GetEvents()
			allEvents = append(allEvents, events...)
			viewport.ClearEvents()
		}
	}

	return allEvents
}

// saveEvents persists events to the events table
func saveEvents(ctx context.Context, tx *sql.Tx, events []dorky.Event) error {
	if len(events) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO events (aggregate_id, aggregate_type, event_type, event_data, aggregate_version, metadata)
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

		// Extract metadata from event if available
		aggregateID := event.GetEntityID()
		aggregateType := event.GetEntityType()
		eventType := event.GetEventType()

		// For now, we'll use nil for aggregate_version and metadata
		// These can be extracted from specific event types if needed
		var aggregateVersion *int64
		var metadata []byte

		_, err = stmt.ExecContext(ctx,
			aggregateID,
			aggregateType,
			eventType,
			eventData,
			aggregateVersion,
			metadata,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return nil
}
