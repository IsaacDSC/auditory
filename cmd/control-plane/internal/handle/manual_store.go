package handle

//go:generate mockgen -source=manual_store.go -destination=mocks/mock_manual_store.go -package=mocks

import (
	"context"
	"net/http"
)

type ManualStoreService interface {
	Store(ctx context.Context) error
}

func ManualStore(storeService ManualStoreService) (string, func(w http.ResponseWriter, r *http.Request)) {
	return "POST /manual-store", func(w http.ResponseWriter, r *http.Request) {
		if err := storeService.Store(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data saved to storage"))
	}
}
