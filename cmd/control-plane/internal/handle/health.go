package handle

import "net/http"

func Health() (string, func(w http.ResponseWriter, r *http.Request)) {
	return "GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}
}
