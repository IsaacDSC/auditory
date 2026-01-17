package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IsaacDSC/auditory/cmd/control-plane/internal/handle"
	"github.com/IsaacDSC/auditory/cmd/control-plane/internal/tasks"
	"github.com/IsaacDSC/auditory/internal/backup"
	"github.com/IsaacDSC/auditory/internal/cfg"
	"github.com/IsaacDSC/auditory/internal/store"
)

func init() {
	cfg.InitConfig()
}

func main() {
	conf := cfg.GetConfig()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	bucketStore, err := store.NewS3BucketStore(ctx, store.S3Config{
		Bucket:          conf.BucketConfig.Name,
		Endpoint:        conf.BucketConfig.Endpoint,
		AccessKeyID:     conf.BucketConfig.AccessKeyID,
		SecretAccessKey: conf.BucketConfig.SecretAccessKey,
		Region:          conf.BucketConfig.Region,
		UsePathStyle:    conf.BucketConfig.UsePathStyle,
	})
	if err != nil {
		log.Fatalf("failed to create S3 client: %v", err)
	}

	dataStore := store.NewDataFileStore()
	memIdempotency := store.NewMemIdempotency(conf.AppConfig.IdempotencyTTL)
	backupService := backup.NewBackup(dataStore, bucketStore)
	fileAuditService := backup.NewFileAudit(dataStore, memIdempotency)

	mux := http.NewServeMux()
	mux.HandleFunc(handle.Health())
	mux.HandleFunc(handle.ManualBackup(backupService))
	mux.HandleFunc(handle.ManualStore(backupService))
	mux.HandleFunc(handle.AuditStore(fileAuditService))

	//task to reset idempotency keys
	go tasks.IdempotencyClear(ctx, conf.TasksConfig.IdempotencyClearPeriod, memIdempotency)

	//task to backup sent data to storage
	go tasks.Backup(ctx, conf.TasksConfig.BackupPeriod, backupService)

	//task to save sent data to storage
	go tasks.Store(ctx, conf.TasksConfig.StorePeriod, backupService)

	server := &http.Server{
		Addr:              conf.AppConfig.Port,
		Handler:           mux,
		ReadTimeout:       conf.AppConfig.ReadTimeout,
		ReadHeaderTimeout: conf.AppConfig.ReadHeaderTimeout,
		WriteTimeout:      conf.AppConfig.WriteTimeout,
		IdleTimeout:       conf.AppConfig.IdleTimeout,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	go func() {
		log.Println("server is running on port 8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited gracefully")
}
