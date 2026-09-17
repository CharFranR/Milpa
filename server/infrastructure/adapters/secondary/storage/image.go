package storage

import (
	"context"
	"fmt"
	"io"
	"milpa/aplication/dto"
	port "milpa/domain/port/secondary"
	"mime"
	"os"
	"path/filepath"
)

type LocalImageStoreImpl struct {
	baseDir string
}

func NewLocalImageStore(baseDir string) *LocalImageStoreImpl {
	return &LocalImageStoreImpl{baseDir: baseDir}
}

func (LIS *LocalImageStoreImpl) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	path := filepath.Join(LIS.baseDir, filename)

	dstFile, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, file)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (LIS *LocalImageStoreImpl) Load(ctx context.Context, filename string) (*dto.ImageDataDTO, error) {

	path := filepath.Join(LIS.baseDir, filename)

	file, err := os.ReadFile(path)

	if err != nil {

		return nil, fmt.Errorf("Load funcion err: %v", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(path))

	ImageData := dto.ImageDataDTO{
		Content:     file,
		ContentType: contentType,
		Filename:    filename,
		Filepath:    "amazon s3 placeholder",
	}

	return &ImageData, nil
}

func (LIS *LocalImageStoreImpl) Delete(ctx context.Context, filename string) error {
	path := filepath.Join(LIS.baseDir, filename)

	err := os.Remove(path)

	return err
}

var _ port.ImageStore = (*LocalImageStoreImpl)(nil)
