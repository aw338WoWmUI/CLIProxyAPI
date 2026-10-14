package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

type copilotQuotaStorage struct{}

func (copilotQuotaStorage) RawJSON() []byte {
	return []byte(`{"type":"copilot","github_access_token":" github-token "}`)
}

func (copilotQuotaStorage) SaveTokenToFile(string) error { return nil }

func TestAPICallUsesCopilotPluginGitHubToken(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer github-token" {
			t.Errorf("quota request did not receive the selected GitHub token")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()
	manager := coreauth.NewManager(nil, &coreauth.RoundRobinSelector{}, nil)
	auth := &coreauth.Auth{ID: "copilot-test.json", Provider: "copilot", Metadata: map[string]any{"github_login": "test"}, Storage: copilotQuotaStorage{}}
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatal(err)
	}
	h := &Handler{cfg: &config.Config{}, authManager: manager}
	router := gin.New()
	router.POST("/", h.APICall)
	router.POST("/v8", func(c *gin.Context) {
		c.Set(ConfigV8ContextKey, true)
		h.APICall(c)
	})
	body, _ := json.Marshal(map[string]any{
		"method": "GET", "url": upstream.URL, "auth_index": auth.EnsureIndex(),
		"header": map[string]string{"Authorization": "Bearer $TOKEN$"},
	})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("legacy management status = %d, want 400", recorder.Code)
	}
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v8", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("management status = %d, want 200", recorder.Code)
	}
	var response apiCallResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("quota upstream status = %d, want 200", response.StatusCode)
	}
}
