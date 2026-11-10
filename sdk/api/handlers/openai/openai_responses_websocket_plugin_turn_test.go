package openai

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MoeForge/moego/ojson"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/api/handlers"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	coreexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdkconfig "github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
)

type websocketNestedModelPlugin struct {
	websocketPluginRouteExecutorHost
	base      *handlers.BaseAPIHandler
	providers []string
	calls     int
}

func (h *websocketNestedModelPlugin) ExecutePluginExecutorStream(ctx context.Context, _ string, req coreexecutor.Request, _ coreexecutor.Options) (*coreexecutor.StreamResult, error) {
	provider := h.providers[h.calls]
	h.calls++
	stream, errMsg := h.base.ExecuteModelStream(ctx, handlers.ModelExecutionRequest{
		EntryProtocol: "openai-response", ExitProtocol: "openai-response",
		Model: req.Model, Body: req.Payload, Stream: true, ForcedProvider: provider,
		SkipRouterPluginID: "test-plugin-executor",
	})
	if errMsg != nil {
		return nil, errMsg.Error
	}
	chunks := make(chan coreexecutor.StreamChunk)
	go func() {
		defer close(chunks)
		for chunk := range stream.Chunks {
			if chunk.Err != nil {
				chunks <- coreexecutor.StreamChunk{Err: chunk.Err}
				return
			}
			chunks <- coreexecutor.StreamChunk{Payload: chunk.Payload}
		}
	}()
	return &coreexecutor.StreamResult{Chunks: chunks}, nil
}

type websocketNestedCaptureExecutor struct {
	websocketProviderCaptureExecutor
	duplexInputs []bool
}

func (e *websocketNestedCaptureExecutor) ExecuteStream(ctx context.Context, auth *coreauth.Auth, req coreexecutor.Request, opts coreexecutor.Options) (*coreexecutor.StreamResult, error) {
	e.duplexInputs = append(e.duplexInputs, coreexecutor.WebsocketInputFromContext(ctx) != nil)
	return e.websocketProviderCaptureExecutor.ExecuteStream(ctx, auth, req, opts)
}

func TestResponsesWebsocketPluginNestedCodexKeepsTurnsOnSameSocket(t *testing.T) {
	for _, steering := range []bool{false, true} {
		t.Run(fmt.Sprintf("steering_%t", steering), func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			const model = "plugin-nested-codex-model"
			codex := &websocketNestedCaptureExecutor{websocketProviderCaptureExecutor: websocketProviderCaptureExecutor{
				provider: "codex", websocketCaptureExecutor: websocketCaptureExecutor{responses: [][]byte{
					[]byte("data: " + `{"type":"response.completed","response":{"id":"r1","output":[{"type":"message","role":"assistant","id":"a1"}]}}` + "\n\n"),
					[]byte("data: " + `{"type":"response.completed","response":{"id":"r2","output":[{"type":"message","role":"assistant","id":"a2"}]}}` + "\n\n"),
				}},
			}}
			fallback := &websocketProviderCaptureExecutor{provider: "test-provider", websocketCaptureExecutor: websocketCaptureExecutor{responses: [][]byte{
				[]byte("data: " + `{"type":"response.completed","response":{"id":"r3","output":[]}}` + "\n\n"),
			}}}
			manager := coreauth.NewManager(nil, nil, nil)
			manager.RegisterExecutor(codex)
			manager.RegisterExecutor(fallback)
			for _, provider := range []string{"codex", "test-provider"} {
				id := fmt.Sprintf("nested-%s-%t", provider, steering)
				if _, err := manager.Register(context.Background(), &coreauth.Auth{ID: id, Provider: provider, Status: coreauth.StatusActive, Attributes: map[string]string{"websockets": "true"}}); err != nil {
					t.Fatal(err)
				}
				registry.GetGlobalRegistry().RegisterClient(id, provider, []*registry.ModelInfo{{ID: model}})
				t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(id) })
			}
			cfg := &sdkconfig.SDKConfig{CodexResponseSteering: steering}
			plugin := &websocketNestedModelPlugin{
				base: handlers.NewBaseAPIHandlers(cfg, manager), providers: []string{"codex", "codex", "test-provider"},
			}
			base := handlers.NewBaseAPIHandlers(cfg, manager)
			base.SetModelRouterHost(plugin)
			h := NewOpenAIResponsesAPIHandler(base)
			router := gin.New()
			router.GET("/v1/responses", h.ResponsesWebsocket)
			server := httptest.NewServer(router)
			defer server.Close()
			conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/v1/responses", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			for i := 1; i <= 3; i++ {
				previous := ""
				if i > 1 {
					previous = fmt.Sprintf(`,"previous_response_id":"r%d"`, i-1)
				}
				request := fmt.Sprintf(`{"type":"response.create","model":"%s","route_to_plugin":true%s,"input":[{"type":"message","role":"user","id":"u%d"}]}`, model, previous, i)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(request)); err != nil {
					t.Fatal(err)
				}
				_, payload, err := conn.ReadMessage()
				if err != nil {
					t.Fatalf("turn %d closed the client socket: %v", i, err)
				}
				if got := ojson.GetBytesNoCopy(payload, "type").String(); got != "response.completed" {
					t.Fatalf("turn %d: %s", i, payload)
				}
			}
			for i, attached := range codex.duplexInputs {
				if attached {
					t.Errorf("plugin turn %d inherited the downstream input reader", i+1)
				}
			}
			payloads := append(codex.payloads, fallback.payloads...)
			for i, payload := range payloads {
				if ojson.GetBytesNoCopy(payload, "previous_response_id").Exists() {
					t.Errorf("turn %d retained connection-scoped upstream state", i+1)
				}
				input := ojson.GetBytesNoCopy(payload, "input").Array()
				if len(input) != 2*i+1 || input[len(input)-1].Get("id").String() != fmt.Sprintf("u%d", i+1) {
					t.Errorf("turn %d lost replay history: %s", i+1, payload)
				}
			}
		})
	}
}
