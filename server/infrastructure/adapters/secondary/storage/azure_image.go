package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"milpa/aplication/dto"
	port "milpa/domain/port/secondary"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
)

const azureContainerTimeout = 10 * time.Second

type AzureImageStoreImpl struct {
	client        *azblob.Client
	container     string
	publicBaseURL string
}

// NewAzureImageStore construye el cliente con la connection string (Render no
// tiene identidad administrada de Azure) y crea el contenedor si no existe.
func NewAzureImageStore(connectionString, container, publicBaseURL string) (*AzureImageStoreImpl, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("create azure blob client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), azureContainerTimeout)
	defer cancel()

	if _, err := client.CreateContainer(ctx, container, nil); err != nil && !bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
		return nil, fmt.Errorf("create container %q: %w", container, err)
	}

	return &AzureImageStoreImpl{
		client:        client,
		container:     container,
		publicBaseURL: publicBaseURL,
	}, nil
}

func (AIS *AzureImageStoreImpl) Upload(ctx context.Context, file []byte, filename string) (string, error) {
	contentType := resolveContentType(filename, file)

	_, err := AIS.client.UploadBuffer(ctx, AIS.container, filename, file, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{BlobContentType: &contentType},
	})
	if err != nil {
		return "", fmt.Errorf("upload blob %q: %w", filename, err)
	}

	return buildPublicImageURL(AIS.publicBaseURL, filename), nil
}

func (AIS *AzureImageStoreImpl) Load(ctx context.Context, filename string) (*dto.ImageDataDTO, error) {
	response, err := AIS.client.DownloadStream(ctx, AIS.container, filename, nil)
	if err != nil {
		return nil, fmt.Errorf("download blob %q: %w", filename, err)
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read blob %q: %w", filename, err)
	}

	contentType := resolveContentType(filename, content)
	if response.ContentType != nil && *response.ContentType != "" {
		contentType = *response.ContentType
	}

	return &dto.ImageDataDTO{
		Content:     content,
		ContentType: contentType,
		Filename:    filename,
		Filepath:    buildPublicImageURL(AIS.publicBaseURL, filename),
	}, nil
}

func (AIS *AzureImageStoreImpl) Delete(ctx context.Context, filename string) error {
	if _, err := AIS.client.DeleteBlob(ctx, AIS.container, filename, nil); err != nil && !bloberror.HasCode(err, bloberror.BlobNotFound) {
		return fmt.Errorf("delete blob %q: %w", filename, err)
	}

	return nil
}

// buildPublicImageURL arma la URL absoluta que sirve Get: el blob no es
// público, así que el cliente siempre pasa por GET /api/v1/images/{filename}.
func buildPublicImageURL(base, filename string) string {
	return strings.TrimRight(base, "/") + "/api/v1/images/" + url.PathEscape(filename)
}

// resolveContentType prefiere la extensión del archivo y cae al sniffing de
// contenido cuando el nombre no trae una extensión conocida.
func resolveContentType(filename string, content []byte) string {
	if contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); contentType != "" {
		return contentType
	}

	return http.DetectContentType(content)
}

var _ port.ImageStore = (*AzureImageStoreImpl)(nil)
