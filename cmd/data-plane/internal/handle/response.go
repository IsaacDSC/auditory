package handle

import (
	"context"

	"github.com/IsaacDSC/auditory/cmd/data-plane/internal/proxy"
	"github.com/IsaacDSC/auditory/internal/audit"
)

type RespService interface {
	EnqueueResponse(ctx context.Context, input audit.ResponseAudit) error
}

func Response(respService RespService) func(ctx context.Context, input proxy.InputAudit) error {
	return func(ctx context.Context, input proxy.InputAudit) error {
		return respService.EnqueueResponse(ctx, audit.ResponseAudit{
			StatusCode:     input.StatusCode,
			Headers:        input.Headers,
			Body:           input.Body,
			RequestHeaders: input.RequestHeaders,
		})
	}
}
