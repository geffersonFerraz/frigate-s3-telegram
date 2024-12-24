package s3

import (
	"context"
	"time"

	"os"

	"github.com/minio/minio-go/v7"
)

type (
	file struct {
		client       *s3
		bucket       string
		file         *os.File
		destination  string
		presignedUrl string
	}

	File interface {
		SetFile(ctx context.Context, file *os.File) error
		SetBucket(ctx context.Context, bucket string) error
		SetDestinatoin(ctx context.Context, destination string) error
		GetPresignedURL(ctx context.Context) string
		Upload(ctx context.Context) error
	}
)

func Files(client *s3) (File, error) {
	return &file{client: client}, nil
}

// GetPresignedURL implements File.
func (f *file) GetPresignedURL(ctx context.Context) string {
	return f.presignedUrl
}

// SetBucket implements FileUpload.
func (f *file) SetBucket(ctx context.Context, bucket string) error {
	f.bucket = bucket
	return nil
}

// SetDestinatoin implements FileUpload.
func (f *file) SetDestinatoin(ctx context.Context, destination string) error {
	f.destination = destination
	return nil
}

// SetFile implements FileUpload.
func (f *file) SetFile(ctx context.Context, file *os.File) error {
	f.file = file
	return nil
}

// Upload implements FileUpload.
func (f *file) Upload(ctx context.Context) error {

	if err := f.client.CheckAlive(); err != nil {
		f.client.Reconstructor()
	}

	fileInfo, err := f.file.Stat()
	if err != nil {
		return err
	}

	dur := 7 * 24 * time.Hour
	url, err := f.client.s3.Presign(ctx, "GET", f.bucket, f.destination, dur, nil)
	if err != nil {
		return err
	}

	f.presignedUrl = url.String()
	_, err = f.client.s3.PutObject(ctx, f.bucket, f.destination, f.file, fileInfo.Size(), minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}
