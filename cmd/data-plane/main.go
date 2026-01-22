package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/IsaacDSC/auditory/cmd/data-plane/internal/handle"
	"github.com/IsaacDSC/auditory/cmd/data-plane/internal/proxy"
	"github.com/IsaacDSC/auditory/internal/backup"
	"github.com/IsaacDSC/auditory/internal/store"
)

func main() {
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		log.Fatal("TARGET_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("invalid target URL: %v", err)
	}

	dataStore := store.NewDataFileStore()
	onCallService := backup.NewHttpOnCallService(dataStore)
	requestHandler := handle.Request(onCallService)
	responseHandler := handle.Response(onCallService)

	proxy := proxy.NewAuditProxy(target, requestHandler, responseHandler)

	log.Printf("Starting proxy server on :%s -> %s", port, targetURL)
	if err := http.ListenAndServe(":"+port, proxy); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
