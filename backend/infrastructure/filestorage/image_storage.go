package filestorage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

// ImageStorage defines the interface for storing and retrieving images
type ImageStorage interface {
	Save(imageID imagegraph.ImageID, imageData []byte) error
	Get(imageID imagegraph.ImageID) ([]byte, error)
	Remove(imageID imagegraph.ImageID) error
	Exists(imageID imagegraph.ImageID) (bool, error)
}

// FilesystemImageStorage implements ImageStorage using the local filesystem
type FilesystemImageStorage struct {
	baseDir string
}

// NewFilesystemImageStorage creates a new filesystem-based image storage
func NewFilesystemImageStorage(baseDir string) (*FilesystemImageStorage, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FilesystemImageStorage{
		baseDir: baseDir,
	}, nil
}

// Save stores an image to the filesystem
func (s *FilesystemImageStorage) Save(imageID imagegraph.ImageID, imageData []byte) error {
	filePath := s.getFilePath(imageID)

	// Write the file
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		return fmt.Errorf("failed to write image file: %w", err)
	}

	return nil
}

// Get retrieves an image from the filesystem
func (s *FilesystemImageStorage) Get(imageID imagegraph.ImageID) ([]byte, error) {
	filePath := s.getFilePath(imageID)

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("image not found: %w", err)
		}
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	return data, nil
}

// Exists checks if an image exists in storage
func (s *FilesystemImageStorage) Exists(imageID imagegraph.ImageID) (bool, error) {
	filePath := s.getFilePath(imageID)

	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if image exists: %w", err)
	}

	return true, nil
}

func (s *FilesystemImageStorage) Remove(imageID imagegraph.ImageID) error {
	filePath := s.getFilePath(imageID)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove image %q: %w", imageID, err)
	}

	return nil
}

// getFilePath returns the filesystem path for a given image ID
func (s *FilesystemImageStorage) getFilePath(imageID imagegraph.ImageID) string {
	// Store images as {baseDir}/{imageID}.png
	// In the future, we could store the extension in metadata or detect it from content
	return filepath.Join(s.baseDir, imageID.String()+".png")
}
