package backup

//go:generate mockgen -source=file_audit.go -destination=mocks/mock_file_audit.go -package=mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/IsaacDSC/auditory/internal/audit"
)

type AuditStore interface {
	Upsert(ctx context.Context, input audit.DataAudit) error
}

type IdempotencyStore interface {
	Get(key string) (time.Time, bool)
	Set(key string)
}

type FileAudit struct {
	auditStore       AuditStore
	idempotencyStore IdempotencyStore
}

func NewFileAudit(auditStore AuditStore, idempotencyStore IdempotencyStore) *FileAudit {
	return &FileAudit{
		auditStore:       auditStore,
		idempotencyStore: idempotencyStore,
	}
}

var ErrIdempotencyKeyAlreadyExists = fmt.Errorf("idempotency key already exists")

func (fa *FileAudit) Save(ctx context.Context, input audit.DataAudit) (string, error) {
	idepotency_key := fmt.Sprintf("%s-%s-%s-%s", input.Metadata.Key, input.Metadata.EventName, input.Metadata.RequestID, input.Metadata.CorrelationID)
	if _, ok := fa.idempotencyStore.Get(idepotency_key); ok {
		return "", ErrIdempotencyKeyAlreadyExists
	}

	if err := fa.auditStore.Upsert(ctx, input); err != nil {
		return "", fmt.Errorf("failed to save data: %w", err)
	}

	fa.idempotencyStore.Set(idepotency_key)

	return idepotency_key, nil
}
