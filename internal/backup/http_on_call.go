package backup

import (
	"context"
	"fmt"

	"github.com/IsaacDSC/auditory/internal/audit"
	"github.com/IsaacDSC/auditory/pkg/ctxkey"
	"github.com/IsaacDSC/auditory/pkg/mu"
)

const (
	XRequestID     = "X-Request-ID"
	XCorrelationID = "X-Correlation-ID"
	XClientID      = "X-Client-ID"
)

type HttpAuditStore interface {
	Upsert(ctx context.Context, input audit.DataAudit) error
}

type HttpOnCallService struct {
	store         HttpAuditStore
	memEventStore map[string]audit.RequestAudit
	mu            mu.MutexByKey
}

func NewHttpOnCallService(store HttpAuditStore) *HttpOnCallService {
	return &HttpOnCallService{
		store:         store,
		memEventStore: make(map[string]audit.RequestAudit),
	}
}

func (h *HttpOnCallService) EnqueueRequest(ctx context.Context, input audit.RequestAudit) error {
	clientID, err := getValue(input.Headers, XClientID)
	if err != nil {
		clientID = "unknown"
	}

	ctx = ctxkey.SetClientID(ctx, clientID)

	mu := h.mu.GetOrCreate(clientID)
	mu.Lock()
	defer mu.Unlock()

	requestID, _ := getValue(input.Headers, XRequestID)
	ctx = ctxkey.SetRequestID(ctx, requestID)

	correlationID, err := getValue(input.Headers, XCorrelationID)
	if err != nil {
		correlationID = "unknown"
	}

	ctx = ctxkey.SetCorrelationID(ctx, correlationID)

	h.memEventStore[requestID] = input

	return nil
}

func (h *HttpOnCallService) EnqueueResponse(ctx context.Context, input audit.ResponseAudit) error {
	clientID := ctxkey.ClientID(ctx)

	mu := h.mu.GetOrCreate(clientID)
	mu.Lock()
	defer mu.Unlock()

	requestID := ctxkey.RequestID(ctx)
	request, ok := h.memEventStore[requestID]
	if !ok {
		return fmt.Errorf("request not found")
	}

	correlationID := ctxkey.CorrelationID(ctx)

	if err := h.store.Upsert(ctx, audit.DataAudit{
		Metadata: audit.MetadataAudit{
			Key:           clientID,
			EventName:     "http_audit",
			RequestID:     requestID,
			CorrelationID: correlationID,
		},
		Data: audit.HttpAudit{
			Request:  request,
			Response: input,
		},
	}); err != nil {
		return fmt.Errorf("failed to save data: %w", err)
	}

	delete(h.memEventStore, clientID)

	return nil
}

func getValue(headers map[string][]string, headerKey string) (string, error) {
	values, ok := headers[headerKey]
	if !ok {
		return "", fmt.Errorf("%s header is required", headerKey)
	}

	return values[0], nil
}
