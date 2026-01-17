package handle

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IsaacDSC/auditory/cmd/control-plane/internal/handle/mocks"
	"go.uber.org/mock/gomock"
)

func TestManualStore(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(m *mocks.MockManualStoreService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success - returns 200 with success message",
			setupMock: func(m *mocks.MockManualStoreService) {
				m.EXPECT().
					Store(gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "data saved to storage",
		},
		{
			name: "error - store fails returns 500",
			setupMock: func(m *mocks.MockManualStoreService) {
				m.EXPECT().
					Store(gomock.Any()).
					Return(errors.New("failed to store data"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "error - connection timeout returns 500",
			setupMock: func(m *mocks.MockManualStoreService) {
				m.EXPECT().
					Store(gomock.Any()).
					Return(errors.New("connection timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockManualStoreService(ctrl)
			tt.setupMock(mockService)

			_, handler := ManualStore(mockService)

			req := httptest.NewRequest(http.MethodPost, "/manual-store", nil)
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
