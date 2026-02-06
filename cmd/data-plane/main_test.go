package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/IsaacDSC/auditory/internal/cfg"
	"github.com/IsaacDSC/auditory/internal/store"
)

func TestDataPlaneIntegration(t *testing.T) {
	// Limpar dados anteriores
	os.Remove("tmp/test-client.json")

	// Setup config
	cfg.SetConfig(&cfg.GeneralConfig{
		AppConfig: cfg.AppConfig{
			ReplacedAudit: "authorization",
		},
	})

	// Step 1: Criar servidor de destino
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"echo":   string(body),
		})
	}))
	defer targetServer.Close()

	// Step 2: Compilar e startar o data-plane
	proxyPort := "19090"
	binaryPath := filepath.Join(t.TempDir(), "data-plane")

	// Compilar
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build: %v\n%s", err, out)
	}

	// Executar
	cmd := exec.Command(binaryPath)
	cmd.Dir = "." // Rodar no diretório do teste
	cmd.Env = append(os.Environ(),
		"TARGET_URL="+targetServer.URL,
		"PORT="+proxyPort,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start data-plane: %v", err)
	}
	defer cmd.Process.Kill()

	// Aguardar o servidor estar pronto
	proxyURL := "http://localhost:" + proxyPort
	waitForServer(t, proxyURL, 10*time.Second)

	// Step 3: Realizar 10 chamadas
	clientID := "test-client"
	for i := 0; i < 10; i++ {
		payload, _ := json.Marshal(map[string]any{"index": i, "message": fmt.Sprintf("request %d", i)})

		req, _ := http.NewRequest(http.MethodPost, proxyURL+"/api/test", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Client-ID", clientID)
		req.Header.Set("X-Request-ID", fmt.Sprintf("req-%d-%d", i, time.Now().UnixNano()))
		req.Header.Set("X-Correlation-ID", fmt.Sprintf("corr-%d", i))
		req.Header.Set("Authorization", "Bearer secret-token")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}

	// Verificar dados auditados
	time.Sleep(100 * time.Millisecond)
	dataStore := store.NewDataFileStore()
	data, err := dataStore.Get(t.Context(), store.Key(clientID))
	if err != nil {
		t.Fatalf("failed to get audit data: %v", err)
	}

	total := 0
	for _, records := range data {
		total += len(records)
	}

	if total != 10 {
		t.Errorf("expected 10 audit records, got %d", total)
	}

	t.Logf("Total audited records: %d", total)
}

func waitForServer(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("server not ready after %v", timeout)
}
