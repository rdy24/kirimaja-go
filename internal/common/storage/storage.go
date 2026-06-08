// Package storage is the object-storage abstraction (MinIO/S3-compatible).
// Files are private; clients receive short-lived presigned URLs, never a
// public path. The DB stores the object key only.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Store is the behavioral interface consumers depend on (mockable in tests).
type Store interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
	PresignedURL(ctx context.Context, key string) (string, error)
	Remove(ctx context.Context, key string) error
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	// PublicEndpoint is the host clients can actually reach (e.g. when the
	// app talks to "minio:9000" inside Docker but browsers need
	// "localhost:9000"). Empty → same as Endpoint. Only used for presigning,
	// which is computed locally (no connection to PublicEndpoint is made).
	PublicEndpoint string
	PresignExpiry  time.Duration
}

type minioStore struct {
	client  *minio.Client // internal: Put/Get/bucket ops
	presign *minio.Client // public-endpoint signer for presigned URLs
	bucket  string
	expiry  time.Duration
}

// New connects to MinIO and ensures the bucket exists.
func New(cfg Config) (Store, error) {
	creds := credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, "")
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  creds,
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connect: %w", err)
	}

	presignCli := cli
	if cfg.PublicEndpoint != "" && cfg.PublicEndpoint != cfg.Endpoint {
		presignCli, err = minio.New(cfg.PublicEndpoint, &minio.Options{
			Creds:  creds,
			Secure: cfg.UseSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("minio presign client: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, err := cli.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := cli.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio make bucket %q: %w", cfg.Bucket, err)
		}
	}

	expiry := cfg.PresignExpiry
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}
	return &minioStore{client: cli, presign: presignCli, bucket: cfg.Bucket, expiry: expiry}, nil
}

func (s *minioStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("minio put %q: %w", key, err)
	}
	return nil
}

func (s *minioStore) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio get %q: %w", key, err)
	}
	defer obj.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, obj); err != nil {
		return nil, fmt.Errorf("minio read %q: %w", key, err)
	}
	return buf.Bytes(), nil
}

func (s *minioStore) PresignedURL(ctx context.Context, key string) (string, error) {
	u, err := s.presign.PresignedGetObject(ctx, s.bucket, key, s.expiry, nil)
	if err != nil {
		return "", fmt.Errorf("minio presign %q: %w", key, err)
	}
	return u.String(), nil
}

func (s *minioStore) Remove(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
