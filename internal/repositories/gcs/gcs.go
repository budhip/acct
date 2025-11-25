package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"cloud.google.com/go/storage"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type CloudStorageRepository interface {
	NewWriter(ctx context.Context, payload *models.CloudStoragePayload) io.WriteCloser
	NewReader(ctx context.Context, payload *models.CloudStoragePayload) (reader io.ReadCloser, err error)
	ListFiles(ctx context.Context, payload *models.CloudStoragePayload) (filenames []string, err error)
	GetSignedURL(payload models.CloudStoragePayload, expireDuration time.Duration) (url string, err error)
	FileExists(ctx context.Context, payload *models.CloudStoragePayload) (bool, error)
	Close() error
}

type cloudStorageClient struct {
	config *config.CloudStorageConfig
	client *storage.Client
}

func NewCloudStorageRepository(cfg *config.Configuration, opts ...option.ClientOption) (CloudStorageRepository, error) {
	if cfg.CloudStorageConfig.BucketName == "" {
		return nil, fmt.Errorf("failed to init cloud storage bucket name not set")
	}

	client, err := storage.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return &cloudStorageClient{client: client, config: &cfg.CloudStorageConfig}, nil
}

func (cs *cloudStorageClient) NewWriter(ctx context.Context, payload *models.CloudStoragePayload) io.WriteCloser {
	obj := cs.client.Bucket(cs.config.BucketName).Object(payload.GetFilePath())
	writer := obj.NewWriter(ctx)
	writer.ContentDisposition = fmt.Sprintf("attachment; filename=%s", payload.Filename)
	return writer
}

func (cs *cloudStorageClient) Close() error {
	return cs.client.Close()
}

func (cs *cloudStorageClient) GetSignedURL(payload models.CloudStoragePayload, expireDuration time.Duration) (url string, err error) {
	url, err = cs.client.Bucket(cs.config.BucketName).SignedURL(payload.GetFilePath(), &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expireDuration),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get signed url: %w", err)
	}

	return
}

func (cs *cloudStorageClient) NewReader(ctx context.Context, payload *models.CloudStoragePayload) (reader io.ReadCloser, err error) {
	obj := cs.client.Bucket(cs.config.BucketName).Object(payload.GetFilePath())
	reader, err = obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS reader: %w", err)
	}
	return reader, nil
}

func (cs *cloudStorageClient) ListFiles(ctx context.Context, payload *models.CloudStoragePayload) (filenames []string, err error) {
	it := cs.client.Bucket(cs.config.BucketName).Objects(ctx, &storage.Query{Prefix: payload.Path})

	for {
		obj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing objects: %w", err)
		}
		if strings.HasSuffix(obj.Name, "/") {
			continue
		}

		//to get the filename only without the path
		filename := path.Base(obj.Name)

		filenames = append(filenames, filename)

	}
	return filenames, nil
}

func (cs *cloudStorageClient) FileExists(ctx context.Context, payload *models.CloudStoragePayload) (bool, error) {
	obj := cs.client.Bucket(cs.config.BucketName).Object(payload.GetFilePath())
	_, err := obj.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
