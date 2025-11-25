package storage

import (
	"bufio"
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"cloud.google.com/go/storage"
)

const mb = 1024 * 1024

type StorageRepository interface {
	Stream(ctx context.Context, bucketName, objectName string) (chan models.StreamResult[string], error)
}

type storageClient struct {
	client *storage.Client
}

func NewStorageRepository(client *storage.Client) StorageRepository {
	return &storageClient{client}
}

func (s *storageClient) Stream(ctx context.Context, bucketName, objectName string) (chan models.StreamResult[string], error) {
	rc, err := s.client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get file: %w", err)
	}
	defer rc.Close()

	result := make(chan models.StreamResult[string])
	go func() {
		defer close(result)

		// Scanner
		scanner := bufio.NewScanner(rc)
		bufferSize := 5 * mb
		buffer := make([]byte, bufferSize)
		scanner.Buffer(buffer, bufferSize)

		// Read
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				result <- models.StreamResult[string]{Data: scanner.Text()}
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				result <- models.StreamResult[string]{Err: err}
			}
		}
	}()

	return result, nil
}
