package handle

import (
	"context"

	"github.com/IsaacDSC/auditory/cmd/data-plane/internal/proxy"
	"github.com/IsaacDSC/auditory/internal/audit"
)

type ReqService interface {
	EnqueueRequest(ctx context.Context, input audit.RequestAudit) error
}

func Request(reqService ReqService) func(ctx context.Context, input proxy.InputAudit) error {
	return func(ctx context.Context, input proxy.InputAudit) error {
		return reqService.EnqueueRequest(ctx, audit.RequestAudit{
			Headers: input.Headers,
			Body:    input.Body,
			Method:  input.Method,
			Path:    input.Path,
			Query:   input.Query,
		})
	}
}
