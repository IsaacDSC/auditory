package backup

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"slices"
	"strings"

	"github.com/IsaacDSC/auditory/internal/audit"
	"github.com/IsaacDSC/auditory/internal/cfg"
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
		mu:            make(mu.MutexByKey),
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

	input.Headers = sanitizeHeaders(input.Headers)
	input.Query = sanitizeQueryParams(input.Query)
	h.memEventStore[requestID] = input

	return nil
}

func (h *HttpOnCallService) EnqueueResponse(ctx context.Context, input audit.ResponseAudit) error {
	// Extrair valores dos headers do request original
	clientID, err := getValue(input.RequestHeaders, XClientID)
	if err != nil {
		clientID = "unknown"
	}

	mu := h.mu.GetOrCreate(clientID)
	mu.Lock()
	defer mu.Unlock()

	requestID, _ := getValue(input.RequestHeaders, XRequestID)
	request, ok := h.memEventStore[requestID]
	if !ok {
		return fmt.Errorf("request not found for requestID: %s", requestID)
	}

	correlationID, err := getValue(input.RequestHeaders, XCorrelationID)
	if err != nil {
		correlationID = "unknown"
	}

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
	// Busca case-insensitive para headers HTTP
	for key, values := range headers {
		if strings.EqualFold(key, headerKey) {
			if len(values) > 0 {
				return values[0], nil
			}
		}
	}
	return "", fmt.Errorf("%s header is required", headerKey)
}

func sanitizeHeaders(headers map[string][]string) map[string][]string {
	conf := cfg.GetConfig()
	if conf == nil {
		return headers
	}
	listReplacements := strings.Split(strings.ToLower(conf.AppConfig.ReplacedAudit), ",")

	for key, value := range headers {
		if slices.Contains(listReplacements, strings.ToLower(key)) {
			headers[key] = []string{strings.Repeat("*", len(value[0]))}
		}
	}
	return headers
}

func sanitizeQueryParams(queryParams string) string {
	conf := cfg.GetConfig()
	if conf == nil {
		return queryParams
	}
	listReplacements := strings.Split(strings.ToLower(conf.AppConfig.ReplacedAudit), ",")

	queryUrl, err := url.Parse(queryParams)
	if err != nil {
		log.Printf("failed to parse query params: %v", err)
		return ""
	}

	for key, value := range queryUrl.Query() {
		if slices.Contains(listReplacements, strings.ToLower(key)) {
			queryUrl.Query()[key] = []string{strings.Repeat("*", len(value[0]))}
		}
	}
	return queryUrl.String()
}
