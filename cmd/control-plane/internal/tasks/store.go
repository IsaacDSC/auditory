package tasks

import (
	"context"
	"log"
	"time"

	"github.com/IsaacDSC/auditory/pkg/clock"
)

type StoreService interface {
	Store(ctx context.Context) error
}

func Store(ctx context.Context, period time.Duration, storeService StoreService) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("store task stopped")
			return
		case <-ticker.C:
			now := clock.Now()
			if now.Hour() == 0 && now.Minute() == 0 && now.Second() == 0 {
				if err := storeService.Store(ctx); err != nil {
					log.Printf("failed to store data: %v", err)
				}
			}
		}
	}
}
