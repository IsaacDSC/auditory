package backup

//go:generate mockgen -source=bucket_backup.go -destination=mocks/mock_bucket_backup.go -package=mocks

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/IsaacDSC/auditory/internal/store"
	"github.com/IsaacDSC/auditory/pkg/clock"
)

type FileStore interface {
	GetAll(ctx context.Context) (map[string]store.Data, error)
	DeleteAfterDay(ctx context.Context, timeNow time.Time) error
}

type S3Store interface {
	Backup(ctx context.Context, timeNow time.Time, data []byte) error
	Save(ctx context.Context, dataKey string, timeNow time.Time, data []byte) error
}

type Backup struct {
	fileStore FileStore
	s3Store   S3Store
}

func NewBackup(fileStore FileStore, s3Store S3Store) *Backup {
	return &Backup{
		fileStore: fileStore,
		s3Store:   s3Store,
	}
}

func (b *Backup) Backup(ctx context.Context) error {
	now := clock.Now()
	data, err := b.fileStore.GetAll(ctx)
	if err != nil {
		log.Printf("failed to get all data: %v", err)
		return err
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal data: %v", err)
		return err
	}

	err = b.s3Store.Backup(ctx, now, payload)
	if err != nil {
		log.Printf("failed to save data to storage: %v", err)
		return err
	}

	return nil
}

func (b *Backup) Store(ctx context.Context) error {
	now := clock.Now()
	last24Hours := now.Add(-24 * time.Hour)

	data, err := b.fileStore.GetAll(ctx)
	if err != nil {
		log.Printf("failed to get data: %v", err)
		return err
	}

	counter := 0
	for key, value := range data {

		payload, err := json.Marshal(value)
		if err != nil {
			log.Printf("failed to marshal data: %v", err)
			continue
		}

		// save data to storage
		err = b.s3Store.Save(ctx, key, last24Hours, payload)
		if err != nil {
			log.Printf("failed to save data to storage: %v", err)
			continue
		}

		counter++
	}

	if counter < len(data) {
		// ALERT
		log.Printf("ALERT: failed to save data to storage: %v", err)
		return err
	}

	// delete tmp data last 24 hours
	err = b.fileStore.DeleteAfterDay(ctx, last24Hours)
	if err != nil {
		log.Printf("failed to delete data: %v", err)
		return err
	}

	return nil
}
