package inmem

import (
	"fmt"
	"sync"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/artwork/domain/ui"
)

type UIMetadataRepository struct {
	mu   sync.RWMutex
	data map[imagegraph.ImageGraphID]*ui.UIMetadata
}

func NewUIMetadataRepository() *UIMetadataRepository {
	return &UIMetadataRepository{
		data: make(map[imagegraph.ImageGraphID]*ui.UIMetadata),
	}
}

func (repo *UIMetadataRepository) Get(
	graphID imagegraph.ImageGraphID,
) (
	*ui.UIMetadata,
	error,
) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	metadata, exists := repo.data[graphID]
	if !exists {
		return nil, application.ErrUIMetadataNotFound
	}

	return metadata, nil
}

func (repo *UIMetadataRepository) Add(metadata *ui.UIMetadata) error {
	if metadata == nil {
		return fmt.Errorf("cannot add nil UIMetadata")
	}

	if metadata.GraphID.IsNil() {
		return fmt.Errorf("cannot add UIMetadata with nil GraphID")
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	// Check if it already exists
	if _, exists := repo.data[metadata.GraphID]; exists {
		return fmt.Errorf("UIMetadata for ImageGraph %q already exists", metadata.GraphID)
	}

	repo.data[metadata.GraphID] = metadata

	return nil
}
