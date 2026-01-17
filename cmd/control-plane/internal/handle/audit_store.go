package handle

//go:generate mockgen -source=audit_store.go -destination=mocks/mock_audit_store.go -package=mocks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/IsaacDSC/auditory/internal/audit"
	"github.com/IsaacDSC/auditory/internal/backup"
)

type AuditStoreService interface {
	Save(ctx context.Context, input audit.DataAudit) (string, error)
}

func AuditStore(auditStoreService AuditStoreService) (string, func(w http.ResponseWriter, r *http.Request)) {
	return "POST /audit", func(w http.ResponseWriter, r *http.Request) {
		var input audit.DataAudit
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		input.Metadata.RequestID = r.Header.Get("X-Request-ID")
		input.Metadata.CorrelationID = r.Header.Get("X-Correlation-ID")

		// validate all fields are not empty
		if err := input.Metadata.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		idepotency_key, err := auditStoreService.Save(r.Context(), input)
		switch err {
		case backup.ErrIdempotencyKeyAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
			return
		case nil:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"idempotency_key": "%s" , "ttl": "5min"}`, idepotency_key)))
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
