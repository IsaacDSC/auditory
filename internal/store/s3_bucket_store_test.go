package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IsaacDSC/auditory/internal/cfg"
	"github.com/IsaacDSC/auditory/internal/store/mocks"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/mock/gomock"
)

func setupTestConfig() {
	cfg.SetConfig(&cfg.GeneralConfig{
		BucketConfig: cfg.BucketConfig{
			ExpiresBackupDays: 2,
			ExpiresStoreDays:  365,
		},
	})
}

func TestS3BucketStore_Backup(t *testing.T) {
	setupTestConfig()
	fixedTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		bucket        string
		timeNow       time.Time
		data          []byte
		setupMock     func(client *mocks.MockS3Client)
		expectedError error
	}{
		{
			name:    "success - backup data to S3",
			bucket:  "test-bucket",
			timeNow: fixedTime,
			data:    []byte(`{"key":"value"}`),
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(&s3.PutObjectOutput{}, nil)
			},
			expectedError: nil,
		},
		{
			name:    "success - backup empty data",
			bucket:  "test-bucket",
			timeNow: fixedTime,
			data:    []byte{},
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(&s3.PutObjectOutput{}, nil)
			},
			expectedError: nil,
		},
		{
			name:    "error - S3 client fails",
			bucket:  "test-bucket",
			timeNow: fixedTime,
			data:    []byte(`{"key":"value"}`),
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("s3 connection error"))
			},
			expectedError: errors.New("failed to upload to S3: s3 connection error"),
		},
		{
			name:    "error - bucket not found",
			bucket:  "non-existent-bucket",
			timeNow: fixedTime,
			data:    []byte(`{"key":"value"}`),
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("bucket not found"))
			},
			expectedError: errors.New("failed to upload to S3: bucket not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockS3Client(ctrl)
			tt.setupMock(mockClient)

			s3Store := NewS3BucketStoreWithClient(tt.bucket, mockClient)
			err := s3Store.Backup(context.Background(), tt.timeNow, tt.data)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestS3BucketStore_Save(t *testing.T) {
	setupTestConfig()
	fixedTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		bucket        string
		dataKey       string
		timeNow       time.Time
		data          []byte
		setupMock     func(client *mocks.MockS3Client)
		expectedError error
	}{
		{
			name:    "success - save data to S3",
			bucket:  "test-bucket",
			dataKey: "user:123",
			timeNow: fixedTime,
			data:    []byte(`{"name":"John"}`),
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(&s3.PutObjectOutput{}, nil)
			},
			expectedError: nil,
		},
		{
			name:    "success - save with different data key",
			bucket:  "test-bucket",
			dataKey: "order:456",
			timeNow: fixedTime,
			data:    []byte(`{"total":"100.00"}`),
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(&s3.PutObjectOutput{}, nil)
			},
			expectedError: nil,
		},
		{
			name:    "success - save empty data",
			bucket:  "test-bucket",
			dataKey: "user:789",
			timeNow: fixedTime,
			data:    []byte{},
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(&s3.PutObjectOutput{}, nil)
			},
			expectedError: nil,
		},
		{
			name:    "error - S3 client fails",
			bucket:  "test-bucket",
			dataKey: "user:123",
			timeNow: fixedTime,
			data:    []byte(`{"name":"John"}`),
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("s3 connection timeout"))
			},
			expectedError: errors.New("failed to upload to S3: s3 connection timeout"),
		},
		{
			name:    "error - access denied",
			bucket:  "restricted-bucket",
			dataKey: "user:123",
			timeNow: fixedTime,
			data:    []byte(`{"name":"John"}`),
			setupMock: func(client *mocks.MockS3Client) {
				client.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("access denied"))
			},
			expectedError: errors.New("failed to upload to S3: access denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockS3Client(ctrl)
			tt.setupMock(mockClient)

			s3Store := NewS3BucketStoreWithClient(tt.bucket, mockClient)
			err := s3Store.Save(context.Background(), tt.dataKey, tt.timeNow, tt.data)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestS3Config(t *testing.T) {
	tests := []struct {
		name     string
		config   S3Config
		expected S3Config
	}{
		{
			name: "success - full config",
			config: S3Config{
				Bucket:          "my-bucket",
				Endpoint:        "http://localhost:9000",
				AccessKeyID:     "minioadmin",
				SecretAccessKey: "minioadmin",
				Region:          "us-east-1",
				UsePathStyle:    true,
			},
			expected: S3Config{
				Bucket:          "my-bucket",
				Endpoint:        "http://localhost:9000",
				AccessKeyID:     "minioadmin",
				SecretAccessKey: "minioadmin",
				Region:          "us-east-1",
				UsePathStyle:    true,
			},
		},
		{
			name: "success - AWS config without endpoint",
			config: S3Config{
				Bucket:          "aws-bucket",
				Endpoint:        "",
				AccessKeyID:     "AKIAXXXXXXXX",
				SecretAccessKey: "secret",
				Region:          "us-west-2",
				UsePathStyle:    false,
			},
			expected: S3Config{
				Bucket:          "aws-bucket",
				Endpoint:        "",
				AccessKeyID:     "AKIAXXXXXXXX",
				SecretAccessKey: "secret",
				Region:          "us-west-2",
				UsePathStyle:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Bucket != tt.expected.Bucket {
				t.Errorf("expected bucket %q, got %q", tt.expected.Bucket, tt.config.Bucket)
			}
			if tt.config.Region != tt.expected.Region {
				t.Errorf("expected region %q, got %q", tt.expected.Region, tt.config.Region)
			}
			if tt.config.UsePathStyle != tt.expected.UsePathStyle {
				t.Errorf("expected UsePathStyle %v, got %v", tt.expected.UsePathStyle, tt.config.UsePathStyle)
			}
		})
	}
}
