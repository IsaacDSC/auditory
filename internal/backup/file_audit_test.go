package backup

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IsaacDSC/auditory/internal/audit"
	"github.com/IsaacDSC/auditory/internal/backup/mocks"
	"go.uber.org/mock/gomock"
)

func TestFileAudit_Save(t *testing.T) {
	fixedTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name                 string
		input                audit.DataAudit
		setupMocks           func(auditStore *mocks.MockAuditStore, idempotencyStore *mocks.MockIdempotencyStore)
		expectedIdempotency  string
		expectedError        error
	}{
		{
			name: "success - save new audit data",
			input: audit.DataAudit{
				Metadata: audit.MetadataAudit{
					Key:           "user:123",
					EventName:     "user.created",
					RequestID:     "req-123",
					CorrelationID: "corr-123",
					EventAt:       fixedTime,
				},
				Data: map[string]string{"name": "John"},
			},
			setupMocks: func(auditStore *mocks.MockAuditStore, idempotencyStore *mocks.MockIdempotencyStore) {
				idempotencyStore.EXPECT().Get("user:123-user.created-req-123-corr-123").Return(time.Time{}, false)
				auditStore.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil)
				idempotencyStore.EXPECT().Set("user:123-user.created-req-123-corr-123")
			},
			expectedIdempotency: "user:123-user.created-req-123-corr-123",
			expectedError:       nil,
		},
		{
			name: "success - save audit data with different key",
			input: audit.DataAudit{
				Metadata: audit.MetadataAudit{
					Key:           "order:456",
					EventName:     "order.completed",
					RequestID:     "req-456",
					CorrelationID: "corr-456",
					EventAt:       fixedTime,
				},
				Data: map[string]string{"total": "100.00"},
			},
			setupMocks: func(auditStore *mocks.MockAuditStore, idempotencyStore *mocks.MockIdempotencyStore) {
				idempotencyStore.EXPECT().Get("order:456-order.completed-req-456-corr-456").Return(time.Time{}, false)
				auditStore.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil)
				idempotencyStore.EXPECT().Set("order:456-order.completed-req-456-corr-456")
			},
			expectedIdempotency: "order:456-order.completed-req-456-corr-456",
			expectedError:       nil,
		},
		{
			name: "error - idempotency key already exists",
			input: audit.DataAudit{
				Metadata: audit.MetadataAudit{
					Key:           "user:123",
					EventName:     "user.created",
					RequestID:     "req-123",
					CorrelationID: "corr-123",
					EventAt:       fixedTime,
				},
				Data: map[string]string{"name": "John"},
			},
			setupMocks: func(auditStore *mocks.MockAuditStore, idempotencyStore *mocks.MockIdempotencyStore) {
				idempotencyStore.EXPECT().Get("user:123-user.created-req-123-corr-123").Return(fixedTime, true)
			},
			expectedIdempotency: "",
			expectedError:       ErrIdempotencyKeyAlreadyExists,
		},
		{
			name: "error - auditStore.Upsert fails",
			input: audit.DataAudit{
				Metadata: audit.MetadataAudit{
					Key:           "user:123",
					EventName:     "user.created",
					RequestID:     "req-123",
					CorrelationID: "corr-123",
					EventAt:       fixedTime,
				},
				Data: map[string]string{"name": "John"},
			},
			setupMocks: func(auditStore *mocks.MockAuditStore, idempotencyStore *mocks.MockIdempotencyStore) {
				idempotencyStore.EXPECT().Get("user:123-user.created-req-123-corr-123").Return(time.Time{}, false)
				auditStore.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			},
			expectedIdempotency: "",
			expectedError:       errors.New("failed to save data: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuditStore := mocks.NewMockAuditStore(ctrl)
			mockIdempotencyStore := mocks.NewMockIdempotencyStore(ctrl)

			tt.setupMocks(mockAuditStore, mockIdempotencyStore)

			fileAudit := NewFileAudit(mockAuditStore, mockIdempotencyStore)
			idempotencyKey, err := fileAudit.Save(context.Background(), tt.input)

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

			if idempotencyKey != tt.expectedIdempotency {
				t.Errorf("expected idempotency key %v, got %v", tt.expectedIdempotency, idempotencyKey)
			}
		})
	}
}
