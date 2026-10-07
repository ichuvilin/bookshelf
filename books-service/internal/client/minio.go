package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client         *minio.Client
	bucket         string
	publicEndpoint string
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket, publicEndpoint string, useSSL bool) (*MinIOClient, error) {
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

	return &MinIOClient{
		client:         client,
		bucket:         bucket,
		publicEndpoint: publicEndpoint,
	}, nil
}

func (c *MinIOClient) EnsureBucket(ctx context.Context) error {
	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", c.bucket, err)
	}

	if !exists {
		if err := c.client.MakeBucket(
			ctx,
			c.bucket,
			minio.MakeBucketOptions{},
		); err != nil {
			return fmt.Errorf("create bucket %q: %w", c.bucket, err)
		}
	}

	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    []string{"s3:GetObject"},
				"Resource":  fmt.Sprintf("arn:aws:s3:::%s/*", c.bucket),
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("marshal bucket policy: %w", err)
	}

	if err := c.client.SetBucketPolicy(
		ctx,
		c.bucket,
		string(policyJSON),
	); err != nil {
		return fmt.Errorf("set bucket policy for %q: %w", c.bucket, err)
	}

	return nil
}

func (c *MinIOClient) HealthCheck(ctx context.Context) error {
	if _, err := c.client.BucketExists(ctx, c.bucket); err != nil {
		return fmt.Errorf("MinIO health check: %w", err)
	}

	return nil
}
