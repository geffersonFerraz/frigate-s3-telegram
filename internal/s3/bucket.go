package s3

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type (
	bucket struct {
		client     *s3
		bucketName string
	}
	Bucket interface {
		SetBucket(ctx context.Context, bucket string) error
		Create(ctx context.Context) error
		Delete(ctx context.Context) error
	}
)

func Buckets(client *s3, bucketName string) (Bucket, error) {
	return &bucket{client, bucketName}, nil
}

// SetBucket implements Bucket.
func (b *bucket) SetBucket(ctx context.Context, bucket string) error {
	b.bucketName = bucket
	return nil
}

// Create implements Bucket.
func (b *bucket) Create(ctx context.Context) error {
	err := b.client.s3.MakeBucket(ctx, b.bucketName, minio.MakeBucketOptions{Region: b.client.cfg.BUCKET_REGION})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := b.client.s3.BucketExists(ctx, b.bucketName)
		if errBucketExists == nil && exists {
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

// Delete implements Bucket.
func (b *bucket) Delete(ctx context.Context) error {
	panic("unimplemented")
}
