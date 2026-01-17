package handle

//go:generate mockgen -source=manual_backup.go -destination=mocks/mock_manual_backup.go -package=mocks

import (
	"context"
	"net/http"
)

type ManualBackupService interface {
	Backup(ctx context.Context) error
}

func ManualBackup(backupService ManualBackupService) (string, func(w http.ResponseWriter, r *http.Request)) {
	return "POST /manual-backup", func(w http.ResponseWriter, r *http.Request) {
		if err := backupService.Backup(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data backed up to storage"))
	}
}
