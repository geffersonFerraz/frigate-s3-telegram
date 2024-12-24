package s3

import (
	"log"
	"time"

	"github.com/geffersonFerraz/frigate-s3-telegram/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type (
	s3 struct {
		s3  *minio.Client
		cfg *config.Config
	}

	S3 interface {
		CheckAlive() error
		GetClient() *s3
	}
)

func New() (S3, error) {
	cfg := config.New()
	endpoint := cfg.BUCKET_SERVER
	accessKeyID := cfg.KEY_PAIR_ID
	secretAccessKey := cfg.KEY_PAIR_SECRET
	useSSL := true

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &s3{s3: minioClient, cfg: cfg}, nil
}

func (s *s3) CheckAlive() error {
	timeout := time.Duration(1 * time.Second)
	_, err := s.s3.HealthCheck(timeout)
	return err
}

func (s *s3) Reconstructor() {
	S3, err := New()
	if err != nil {
		log.Fatalln(err)
	}
	s.s3 = S3.GetClient().s3
}

func (s *s3) GetClient() *s3 {
	return s
}

// func (s *s3) Create(ctx context.Context, bucket string) error {

// 	err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
// 	if err != nil {
// 		// Check to see if we already own this bucket (which happens if you run this twice)
// 		exists, errBucketExists := s3Client.BucketExists(ctx, bucketName)
// 		if errBucketExists == nil && exists {
// 			log.Printf("We already own %s\n", bucketName)
// 		} else {
// 			return err
// 		}
// 	} else {
// 		log.Printf("Successfully created %s\n", bucketName)
// 	}
// }
