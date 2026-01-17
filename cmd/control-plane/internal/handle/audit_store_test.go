package handle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IsaacDSC/auditory/cmd/control-plane/internal/handle/mocks"
	"github.com/IsaacDSC/auditory/internal/audit"
	"github.com/IsaacDSC/auditory/internal/backup"
	"go.uber.org/mock/gomock"
)

func TestAuditStore(t *testing.T) {
	validInput := audit.DataAudit{
		Metadata: audit.MetadataAudit{
			Key:       "user:123",
			EventName: "user.created",
			EventAt:   time.Now(),
		},
		Data: map[string]any{"name": "John Doe"},
	}

	tests := []struct {
		name           string
		body           any
		requestID      string
		correlationID  string
		setupMock      func(m *mocks.MockAuditStoreService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:          "success - returns 201 with idempotency key",
			body:          validInput,
			requestID:     "req-123",
			correlationID: "corr-456",
			setupMock: func(m *mocks.MockAuditStoreService) {
				m.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return("user:123-user.created-req-123-corr-456", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"idempotency_key": "user:123-user.created-req-123-corr-456" , "ttl": "5min"}`,
		},
		{
			name:           "error - invalid JSON body returns 400",
			body:           "invalid-json",
			requestID:      "req-123",
			correlationID:  "corr-456",
			setupMock:      func(m *mocks.MockAuditStoreService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "error - missing request_id returns 400",
			body: validInput,
			// requestID is empty
			correlationID:  "corr-456",
			setupMock:      func(m *mocks.MockAuditStoreService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "error - missing correlation_id returns 400",
			body:      validInput,
			requestID: "req-123",
			// correlationID is empty
			setupMock:      func(m *mocks.MockAuditStoreService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "error - missing key in metadata returns 400",
			body: audit.DataAudit{
				Metadata: audit.MetadataAudit{
					// Key is empty
					EventName: "user.created",
					EventAt:   time.Now(),
				},
				Data: map[string]any{"name": "John Doe"},
			},
			requestID:      "req-123",
			correlationID:  "corr-456",
			setupMock:      func(m *mocks.MockAuditStoreService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "error - idempotency key already exists returns 409",
			body:          validInput,
			requestID:     "req-123",
			correlationID: "corr-456",
			setupMock: func(m *mocks.MockAuditStoreService) {
				m.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return("", backup.ErrIdempotencyKeyAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:          "error - internal server error returns 500",
			body:          validInput,
			requestID:     "req-123",
			correlationID: "corr-456",
			setupMock: func(m *mocks.MockAuditStoreService) {
				m.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return("", errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockAuditStoreService(ctrl)
			tt.setupMock(mockService)

			_, handler := AuditStore(mockService)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/audit", bytes.NewReader(bodyBytes))
			req = req.WithContext(context.Background())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Request-ID", tt.requestID)
			req.Header.Set("X-Correlation-ID", tt.correlationID)

			rr := httptest.NewRecorder()

			handler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" && rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
