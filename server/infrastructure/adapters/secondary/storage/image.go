package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"milpa/aplication/dto"
	port "milpa/domain/port/secondary"
)

type LocalImageStoreImpl struct {
	baseDir       string
	publicBaseURL string
}

// NewLocalImageStore es el fallback de desarrollo: guarda en disco local y
// devuelve la misma URL pública que el adaptador de Azure.
func NewLocalImageStore(baseDir, publicBaseURL string) *LocalImageStoreImpl {
	return &LocalImageStoreImpl{baseDir: baseDir, publicBaseURL: publicBaseURL}
}

func (LIS *LocalImageStoreImpl) Upload(ctx context.Context, file []byte, filename string) (string, error) {
	path := filepath.Join(LIS.baseDir, filename)

	if err := os.WriteFile(path, file, 0644); err != nil {
		return "", fmt.Errorf("write image %q: %w", filename, err)
	}

	return buildPublicImageURL(LIS.publicBaseURL, filename), nil
}

func (LIS *LocalImageStoreImpl) Load(ctx context.Context, filename string) (*dto.ImageDataDTO, error) {
	path := filepath.Join(LIS.baseDir, filename)

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read image %q: %w", filename, err)
	}

	return &dto.ImageDataDTO{
		Content:     content,
		ContentType: resolveContentType(filename, content),
		Filename:    filename,
		Filepath:    buildPublicImageURL(LIS.publicBaseURL, filename),
	}, nil
}

func (LIS *LocalImageStoreImpl) Delete(ctx context.Context, filename string) error {
	path := filepath.Join(LIS.baseDir, filename)

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete image %q: %w", filename, err)
	}

	return nil
}

var _ port.ImageStore = (*LocalImageStoreImpl)(nil)
