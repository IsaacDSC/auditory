package store

//go:generate mockgen -source=s3_bucket_store.go -destination=mocks/mock_s3_bucket_store.go -package=mocks

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/IsaacDSC/auditory/internal/cfg"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3BucketStore struct {
	bucket string
	client S3Client
}

type S3Config struct {
	Bucket          string
	Endpoint        string // Optional: for MinIO/LocalStack
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	UsePathStyle    bool // Required for MinIO
}

func NewS3BucketStore(ctx context.Context, s3cfg S3Config) (*S3BucketStore, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s3cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s3cfg.AccessKeyID,
			s3cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	var client *s3.Client
	if s3cfg.Endpoint != "" {
		// Use custom endpoint (MinIO/LocalStack)
		client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s3cfg.Endpoint)
			o.UsePathStyle = s3cfg.UsePathStyle
		})
	} else {
		client = s3.NewFromConfig(awsCfg)
	}

	return &S3BucketStore{
		bucket: s3cfg.Bucket,
		client: client,
	}, nil
}

func NewS3BucketStoreWithClient(bucket string, client S3Client) *S3BucketStore {
	return &S3BucketStore{
		bucket: bucket,
		client: client,
	}
}

func (s3bs *S3BucketStore) Backup(ctx context.Context, timeNow time.Time, data []byte) error {
	key := fmt.Sprintf("audits/%d-%02d-%02d.json", timeNow.Year(), timeNow.Month(), timeNow.Day())
	// expires in 2 days

	cfg := cfg.GetConfig()
	expires := timeNow.Add(time.Hour * 24 * time.Duration(cfg.BucketConfig.ExpiresBackupDays))

	_, err := s3bs.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3bs.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
		Expires:     aws.Time(expires),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

func (s3bs *S3BucketStore) Save(ctx context.Context, dataKey string, timeNow time.Time, data []byte) error {
	cfg := cfg.GetConfig()
	expires := timeNow.Add(time.Hour * 24 * time.Duration(cfg.BucketConfig.ExpiresStoreDays))

	key := fmt.Sprintf("audits/%s/%d-%02d-%02d.json", dataKey, timeNow.Year(), timeNow.Month(), timeNow.Day())

	_, err := s3bs.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3bs.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
		Expires:     aws.Time(expires),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}
