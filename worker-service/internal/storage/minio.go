package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client         *minio.Client
	bucket         string
	publicEndpoint string
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucket, publicEndpoint string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create MinIO client: %w", err)
	}

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket %q: %w", bucket, err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("create bucket %q: %w", bucket, err)
		}
	}

	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    []string{"s3:GetObject"},
				"Resource":  fmt.Sprintf("arn:aws:s3:::%s/*", bucket),
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("marshal bucket policy: %w", err)
	}

	if err := client.SetBucketPolicy(ctx, bucket, string(policyJSON)); err != nil {
		return nil, fmt.Errorf("set bucket policy for %q: %w", bucket, err)
	}

	return &MinIOStorage{
		client:         client,
		bucket:         bucket,
		publicEndpoint: publicEndpoint,
	}, nil
}

func (s *MinIOStorage) GetFile(ctx context.Context, objectName string) ([]byte, error) {
	object, err := s.client.GetObject(
		ctx,
		s.bucket,
		objectName,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("get file %q: %w", objectName, err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", objectName, err)
	}

	return data, nil
}

func (s *MinIOStorage) UploadFile(ctx context.Context, objectName string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)

	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectName,
		reader,
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("upload file %q: %w", objectName, err)
	}

	return nil
}

func (s *MinIOStorage) GetFileURL(objectName string) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		strings.TrimRight(s.publicEndpoint, "/"),
		s.bucket,
		strings.TrimLeft(objectName, "/"),
	)
}
