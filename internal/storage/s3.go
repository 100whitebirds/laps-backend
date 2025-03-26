package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	"laps/config"
)

type S3Storage struct {
	client *minio.Client
	cfg    config.S3Config
	logger *zap.Logger
}

func NewS3Storage(cfg config.S3Config, logger *zap.Logger) (*S3Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации клиента S3: %w", err)
	}

	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования бакета: %w", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{
			Region: cfg.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("ошибка создания бакета: %w", err)
		}
	}

	return &S3Storage{
		client: client,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (s *S3Storage) UploadFile(ctx context.Context, data []byte, filename string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("пустые данные файла")
	}

	fileType := http.DetectContentType(data)
	if !strings.HasPrefix(fileType, "image/") {
		return "", errors.New("файл не является изображением")
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		switch fileType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".bin"
		}
	}

	objectName := fmt.Sprintf("specialists/%s%s", uuid.New().String(), ext)
	reader := bytes.NewReader(data)
	objectSize := int64(len(data))

	_, err := s.client.PutObject(ctx, s.cfg.Bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: fileType,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки файла в S3: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.Bucket, s.cfg.Region, objectName)

	return url, nil
}

func (s *S3Storage) DeleteFile(ctx context.Context, fileURL string) error {
	if fileURL == "" {
		return nil
	}

	parts := strings.Split(fileURL, "/")
	if len(parts) < 4 || !strings.Contains(parts[2], "amazonaws.com") {
		return fmt.Errorf("некорректный URL файла: %s", fileURL)
	}

	objectName := strings.Join(parts[3:], "/")
	err := s.client.RemoveObject(ctx, s.cfg.Bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка удаления файла из S3: %w", err)
	}

	return nil
}

func (s *S3Storage) GetFile(ctx context.Context, fileURL string) ([]byte, error) {
	if fileURL == "" {
		return nil, errors.New("пустой URL файла")
	}

	parts := strings.Split(fileURL, "/")
	if len(parts) < 4 || !strings.Contains(parts[2], "amazonaws.com") {
		return nil, fmt.Errorf("некорректный URL файла: %s", fileURL)
	}

	objectName := strings.Join(parts[3:], "/")
	object, err := s.client.GetObject(ctx, s.cfg.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения файла из S3: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла из S3: %w", err)
	}

	return data, nil
}

func (s *S3Storage) GetPresignedURL(ctx context.Context, fileURL string, expiry time.Duration) (string, error) {
	if fileURL == "" {
		return "", errors.New("пустой URL файла")
	}

	parts := strings.Split(fileURL, "/")
	if len(parts) < 4 || !strings.Contains(parts[2], "amazonaws.com") {
		return "", fmt.Errorf("некорректный URL файла: %s", fileURL)
	}

	objectName := strings.Join(parts[3:], "/")
	presignedURL, err := s.client.PresignedGetObject(ctx, s.cfg.Bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации пресайн URL: %w", err)
	}

	return presignedURL.String(), nil
}
