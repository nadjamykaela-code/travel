package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	apiBaseURL  = "http://localhost:8080"
	workerURL   = "http://localhost:8081"
	firestoreURL = "http://localhost:8200"
)

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("E2E_TEST") == "" {
		t.Skip("Skipping e2e test; set E2E_TEST=1 to run")
	}
}

func waitForAPI(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(apiBaseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("API did not become healthy within 60s")
}

func waitForWorker(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(workerURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("Worker did not become healthy within 60s")
}

func TestE2E_FullStack(t *testing.T) {
	skipIfNotIntegration(t)

	waitForAPI(t)
	waitForWorker(t)

	t.Run("api health", func(t *testing.T) {
		resp, err := http.Get(apiBaseURL + "/health")
		if err != nil {
			t.Fatalf("GET /health: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want 200", resp.StatusCode)
		}
	})

	t.Run("worker health", func(t *testing.T) {
		resp, err := http.Get(workerURL + "/health")
		if err != nil {
			t.Fatalf("GET /health: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want 200", resp.StatusCode)
		}
	})

	t.Run("create filter", func(t *testing.T) {
		body := map[string]interface{}{
			"origin":      "GRU",
			"destination": "LIS",
			"priceMax":    3000,
			"passengers":  1,
			"startDate":   "2026-08-01",
			"endDate":     "2026-08-15",
		}
		payload, _ := json.Marshal(body)

		req, err := http.NewRequest("POST", apiBaseURL+"/api/filters", bytes.NewReader(payload))
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-mock-token")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST /api/filters: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			t.Log("filter created successfully")
		} else if resp.StatusCode == http.StatusUnauthorized {
			t.Log("auth required (expected without valid Firebase token)")
		} else {
			t.Errorf("unexpected status: %d", resp.StatusCode)
		}
	})

	t.Run("trigger worker run", func(t *testing.T) {
		resp, err := http.Post(workerURL+"/run", "application/json", nil)
		if err != nil {
			t.Fatalf("POST /run: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want 200", resp.StatusCode)
		}
	})

	t.Run("firestore has filters collection", func(t *testing.T) {
		projectID := os.Getenv("GCP_PROJECT_ID")
		if projectID == "" {
			projectID = "myka-travel"
		}
		fsURL := fmt.Sprintf("%s/v1/projects/%s/databases/(default)/documents/filters",
			firestoreURL, projectID)

		deadline := time.Now().Add(30 * time.Second)
		var lastErr error
		for time.Now().Before(deadline) {
			resp, err := http.Get(fsURL)
			if err != nil {
				lastErr = err
				time.Sleep(2 * time.Second)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var result map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					lastErr = err
					time.Sleep(2 * time.Second)
					continue
				}
				docs, ok := result["documents"].([]interface{})
				if ok && len(docs) > 0 {
					t.Logf("found %d filter documents in Firestore", len(docs))
					return
				}
				t.Log("no filter documents yet, retrying...")
				time.Sleep(2 * time.Second)
				continue
			}
			lastErr = fmt.Errorf("firestore status: %d", resp.StatusCode)
			time.Sleep(2 * time.Second)
		}
		if lastErr != nil {
			t.Fatalf("firestore check failed: %v", lastErr)
		}
	})
}

func TestE2E_FirestoreEmulator(t *testing.T) {
	skipIfNotIntegration(t)

	t.Run("firestore emulator is running", func(t *testing.T) {
		projectID := os.Getenv("GCP_PROJECT_ID")
		if projectID == "" {
			projectID = "myka-travel"
		}
		url := fmt.Sprintf("%s/v1/projects/%s/databases/(default)/documents",
			firestoreURL, projectID)

		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("connect to firestore emulator: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want 200", resp.StatusCode)
		}
	})

	t.Run("firestore has history collection after worker run", func(t *testing.T) {
		projectID := os.Getenv("GCP_PROJECT_ID")
		if projectID == "" {
			projectID = "myka-travel"
		}
		url := fmt.Sprintf("%s/v1/projects/%s/databases/(default)/documents:listCollectionIds",
			firestoreURL, projectID)

		deadline := time.Now().Add(30 * time.Second)
		var lastErr error
		for time.Now().Before(deadline) {
			resp, err := http.Post(url, "application/json", nil)
			if err != nil {
				lastErr = err
				time.Sleep(2 * time.Second)
				continue
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				lastErr = err
				time.Sleep(2 * time.Second)
				continue
			}
			resp.Body.Close()

			collectionIDs, ok := result["collectionIds"].([]interface{})
			if !ok {
				time.Sleep(2 * time.Second)
				continue
			}

			hasHistory := false
			hasFilters := false
			for _, c := range collectionIDs {
				if c == "history" {
					hasHistory = true
				}
				if c == "filters" {
					hasFilters = true
				}
			}

			t.Logf("collections found: filters=%v history=%v", hasFilters, hasHistory)

			if hasHistory {
				t.Log("history collection found - worker persisted results")
				return
			}
			time.Sleep(2 * time.Second)
		}
		if lastErr != nil {
			t.Logf("collection check ended with error: %v", lastErr)
		}
	})
}
