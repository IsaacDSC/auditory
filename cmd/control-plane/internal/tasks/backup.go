package tasks

import (
	"context"
	"fmt"
	"log"
	"time"
)

type BackupService interface {
	Backup(ctx context.Context) error
}

func Backup(ctx context.Context, period time.Duration, backupService BackupService) {
	fmt.Println("backup task started", period)
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("backup task stopped")
			return
		case <-ticker.C:
			fmt.Println("backup task running")
			if err := backupService.Backup(ctx); err != nil {
				log.Printf("failed to backup data: %v", err)
			}
		}
	}
}
