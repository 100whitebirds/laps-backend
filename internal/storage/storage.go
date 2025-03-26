package storage

import (
	"context"
	"time"
)

type FileStorage interface {
	UploadFile(ctx context.Context, data []byte, filename string) (string, error)

	DeleteFile(ctx context.Context, fileURL string) error

	GetFile(ctx context.Context, fileURL string) ([]byte, error)

	GetPresignedURL(ctx context.Context, fileURL string, expiry time.Duration) (string, error)
}
