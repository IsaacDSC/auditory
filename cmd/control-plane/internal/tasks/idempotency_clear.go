package tasks

import (
	"context"
	"fmt"
	"log"
	"time"
)

type MemIdempotency interface {
	Reset()
}

func IdempotencyClear(ctx context.Context, period time.Duration, idempotencyClearService MemIdempotency) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("idempotency clear task stopped")
			return
		case <-ticker.C:
			fmt.Println("idempotency clear task running")
			idempotencyClearService.Reset()
		}
	}
}
