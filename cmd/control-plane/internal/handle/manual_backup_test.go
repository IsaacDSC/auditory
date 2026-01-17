package handle

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IsaacDSC/auditory/cmd/control-plane/internal/handle/mocks"
	"go.uber.org/mock/gomock"
)

func TestManualBackup(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(m *mocks.MockManualBackupService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success - returns 200 with success message",
			setupMock: func(m *mocks.MockManualBackupService) {
				m.EXPECT().
					Backup(gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "data backed up to storage",
		},
		{
			name: "error - backup fails returns 500",
			setupMock: func(m *mocks.MockManualBackupService) {
				m.EXPECT().
					Backup(gomock.Any()).
					Return(errors.New("failed to backup data"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "error - s3 unavailable returns 500",
			setupMock: func(m *mocks.MockManualBackupService) {
				m.EXPECT().
					Backup(gomock.Any()).
					Return(errors.New("s3 service unavailable"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockManualBackupService(ctrl)
			tt.setupMock(mockService)

			_, handler := ManualBackup(mockService)

			req := httptest.NewRequest(http.MethodPost, "/manual-backup", nil)
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
