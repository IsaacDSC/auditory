package backup

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IsaacDSC/auditory/internal/audit"
	"github.com/IsaacDSC/auditory/internal/backup/mocks"
	"github.com/IsaacDSC/auditory/internal/store"
	"github.com/IsaacDSC/auditory/pkg/clock"
	"go.uber.org/mock/gomock"
)

func TestBackup_Backup(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	clock.SetNow(fixedTime)
	defer func() {
		clock.Now = func() time.Time { return time.Now().UTC() }
	}()

	tests := []struct {
		name          string
		setupMocks    func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store)
		expectedError error
	}{
		{
			name: "success - backup data to S3",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				data := map[string]store.Data{
					"user:123": {
						store.Date("2025-1-15"): []audit.DataAudit{
							{
								Metadata: audit.MetadataAudit{
									Key:           "user:123",
									EventName:     "user.created",
									RequestID:     "req-123",
									CorrelationID: "corr-123",
									EventAt:       fixedTime,
								},
								Data: map[string]string{"name": "John"},
							},
						},
					},
				}
				fileStore.EXPECT().GetAll(gomock.Any()).Return(data, nil)
				s3Store.EXPECT().Backup(gomock.Any(), fixedTime, gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "success - backup empty data",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				data := map[string]store.Data{}
				fileStore.EXPECT().GetAll(gomock.Any()).Return(data, nil)
				s3Store.EXPECT().Backup(gomock.Any(), fixedTime, gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "error - fileStore.GetAll fails",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				fileStore.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("failed to get data"))
			},
			expectedError: errors.New("failed to get data"),
		},
		{
			name: "error - s3Store.Backup fails",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				data := map[string]store.Data{
					"user:123": {},
				}
				fileStore.EXPECT().GetAll(gomock.Any()).Return(data, nil)
				s3Store.EXPECT().Backup(gomock.Any(), fixedTime, gomock.Any()).Return(errors.New("s3 backup failed"))
			},
			expectedError: errors.New("s3 backup failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFileStore := mocks.NewMockFileStore(ctrl)
			mockS3Store := mocks.NewMockS3Store(ctrl)

			tt.setupMocks(mockFileStore, mockS3Store)

			backup := NewBackup(mockFileStore, mockS3Store)
			err := backup.Backup(context.Background())

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

func TestBackup_Store(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	last24Hours := fixedTime.Add(-24 * time.Hour)
	clock.SetNow(fixedTime)
	defer func() {
		clock.Now = func() time.Time { return time.Now().UTC() }
	}()

	tests := []struct {
		name          string
		setupMocks    func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store)
		expectedError error
	}{
		{
			name: "success - store single item",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				data := map[string]store.Data{
					"user:123": {
						store.Date("2025-1-14"): []audit.DataAudit{
							{
								Metadata: audit.MetadataAudit{
									Key:           "user:123",
									EventName:     "user.created",
									RequestID:     "req-123",
									CorrelationID: "corr-123",
									EventAt:       last24Hours,
								},
								Data: map[string]string{"name": "John"},
							},
						},
					},
				}
				fileStore.EXPECT().GetAll(gomock.Any()).Return(data, nil)
				s3Store.EXPECT().Save(gomock.Any(), "user:123", last24Hours, gomock.Any()).Return(nil)
				fileStore.EXPECT().DeleteAfterDay(gomock.Any(), last24Hours).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "success - store multiple items",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				data := map[string]store.Data{
					"user:123": {
						store.Date("2025-1-14"): []audit.DataAudit{
							{
								Metadata: audit.MetadataAudit{
									Key:           "user:123",
									EventName:     "user.created",
									RequestID:     "req-123",
									CorrelationID: "corr-123",
									EventAt:       last24Hours,
								},
								Data: map[string]string{"name": "John"},
							},
						},
					},
					"order:456": {
						store.Date("2025-1-14"): []audit.DataAudit{
							{
								Metadata: audit.MetadataAudit{
									Key:           "order:456",
									EventName:     "order.created",
									RequestID:     "req-456",
									CorrelationID: "corr-456",
									EventAt:       last24Hours,
								},
								Data: map[string]string{"total": "100.00"},
							},
						},
					},
				}
				fileStore.EXPECT().GetAll(gomock.Any()).Return(data, nil)
				s3Store.EXPECT().Save(gomock.Any(), gomock.Any(), last24Hours, gomock.Any()).Return(nil).Times(2)
				fileStore.EXPECT().DeleteAfterDay(gomock.Any(), last24Hours).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "success - store empty data",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				data := map[string]store.Data{}
				fileStore.EXPECT().GetAll(gomock.Any()).Return(data, nil)
				fileStore.EXPECT().DeleteAfterDay(gomock.Any(), last24Hours).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "error - fileStore.GetAll fails",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				fileStore.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("failed to get data"))
			},
			expectedError: errors.New("failed to get data"),
		},
		{
			name: "error - fileStore.DeleteAfterDay fails",
			setupMocks: func(fileStore *mocks.MockFileStore, s3Store *mocks.MockS3Store) {
				data := map[string]store.Data{
					"user:123": {
						store.Date("2025-1-14"): []audit.DataAudit{
							{
								Metadata: audit.MetadataAudit{
									Key:           "user:123",
									EventName:     "user.created",
									RequestID:     "req-123",
									CorrelationID: "corr-123",
									EventAt:       last24Hours,
								},
								Data: map[string]string{"name": "John"},
							},
						},
					},
				}
				fileStore.EXPECT().GetAll(gomock.Any()).Return(data, nil)
				s3Store.EXPECT().Save(gomock.Any(), "user:123", last24Hours, gomock.Any()).Return(nil)
				fileStore.EXPECT().DeleteAfterDay(gomock.Any(), last24Hours).Return(errors.New("failed to delete"))
			},
			expectedError: errors.New("failed to delete"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFileStore := mocks.NewMockFileStore(ctrl)
			mockS3Store := mocks.NewMockS3Store(ctrl)

			tt.setupMocks(mockFileStore, mockS3Store)

			backup := NewBackup(mockFileStore, mockS3Store)
			err := backup.Store(context.Background())

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
